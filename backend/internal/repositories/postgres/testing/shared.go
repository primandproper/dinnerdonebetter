package testing

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"

	"github.com/primandproper/platform-go/v9/testutils/containers"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
)

// The share-one-container model.
//
// A container per test gives perfect isolation at a price that stops scaling
// around the time a package has a few dozen of them: every test pays a container
// start plus a full migration replay, and a package running its tests in parallel
// asks the Docker daemon for that many postgres instances at once. Past a certain
// width the daemon stops answering and containers fail their readiness wait — not
// because anything is wrong with the test, but because nothing was rationing the
// daemon.
//
// Isolation does not actually require a container, though. Postgres isolates at
// the database level, and cloning a database from a template is a file copy rather
// than a replay of every migration. So a package starts one container in TestMain,
// migrates one template database into it, and hands each test its own clone. Tests
// keep t.Parallel(), keep a private database nothing else writes to, and the
// container count per package drops to one.
const (
	// templateDatabase is migrated once per test binary and cloned per test. Nothing
	// connects to it after prepareTemplate returns: CREATE DATABASE ... TEMPLATE
	// refuses to run while any session is attached to the template.
	templateDatabase = "ddb_template"

	// maxConnections has to cover every parallel test's pools at once now that they
	// share a server. Each test opens a read and a write pool of maxOpenConnsPerTest,
	// where per-test containers used to give each its own connection budget. Postgres
	// defaults to 100, which a wide -parallel run blows through and reports as
	// "too many clients already" from whichever test happens to connect last.
	maxConnections = 500

	// maxOpenConnsPerTest caps each pool a test opens. The platform default is 7 per
	// pool, which is sized for a service that owns its database rather than for a few
	// dozen suites sharing one.
	maxOpenConnsPerTest = 4
	maxIdleConnsPerTest = 2

	// maxSanitizedNameLen bounds the test-name portion of a cloned database's name.
	// Postgres truncates identifiers at 63 bytes, and a silent truncation could
	// collide two long test names onto one database.
	maxSanitizedNameLen = 40
)

// sharedDatabase owns the one container a test binary starts and mints isolated
// databases inside it.
type sharedDatabase struct {
	// admin is connected to the container's provisioning database, never to the
	// template, so it is always free to issue CREATE/DROP DATABASE.
	admin *sql.DB

	// baseDSN addresses the container; dsnForDatabase swaps the database in it.
	baseDSN string

	counter atomic.Uint64
}

// shared is set by RunTestsWithSharedDatabase and read by NewIsolatedDatabaseForTest.
// It is written once before m.Run and only read afterwards, so parallel tests need no
// synchronization to reach it.
var shared *sharedDatabase

// RunTestsWithSharedDatabase starts one postgres container for the calling test
// binary, migrates the template database every test will be cloned from, runs the
// package's tests, and tears the container down. It returns the exit code TestMain
// should pass to os.Exit:
//
//	func TestMain(m *testing.M) {
//		os.Exit(pgtesting.RunTestsWithSharedDatabase(m, migrate))
//	}
//
// migrate is supplied by the caller rather than imported here because the migrations
// package's own tests import this one; taking it as a parameter keeps that from
// becoming an import cycle. It also leaves suites that genuinely need an unmigrated
// server — the migrations tests themselves — free to keep using
// BuildDatabaseContainerForTest.
//
// Without RUN_CONTAINER_TESTS=true this starts nothing and just runs the tests, which
// then skip themselves through NewIsolatedDatabaseForTest.
func RunTestsWithSharedDatabase(m *testing.M, migrate func(ctx context.Context, db *sql.DB) error) int {
	if !containers.RunningTests {
		return m.Run()
	}

	if migrate == nil {
		fmt.Fprintln(os.Stderr, "postgres testing: RunTestsWithSharedDatabase requires a non-nil migrate")
		return 1
	}

	cleanup, err := startSharedDatabase(migrate)
	if err != nil {
		fmt.Fprintf(os.Stderr, "postgres testing: %v\n", err)
		return 1
	}
	defer cleanup()

	return m.Run()
}

// startSharedDatabase brings up the container and template, returning the teardown
// for the caller to defer. Teardown is returned rather than deferred internally
// because it has to outlive this call and still run before TestMain reaches os.Exit,
// which would skip any defer registered above it.
func startSharedDatabase(migrate func(ctx context.Context, db *sql.DB) error) (func(), error) {
	ctx := context.Background()

	container, admin, _, err := BuildDatabaseContainer(ctx, templateDatabase, connectionLimitCustomizer())
	if err != nil {
		return nil, fmt.Errorf("starting shared postgres container: %w", err)
	}

	cleanup := func() {
		if closeErr := admin.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "postgres testing: closing admin pool: %v\n", closeErr)
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), containers.DefaultShutdownTimeout)
		defer cancel()

		if terminateErr := container.Terminate(shutdownCtx); terminateErr != nil {
			fmt.Fprintf(os.Stderr, "postgres testing: terminating shared container: %v\n", terminateErr)
		}
	}

	admin.SetMaxOpenConns(maxOpenConnsPerTest)

	baseDSN, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		return nil, fmt.Errorf("resolving shared container DSN: %w", err)
	}

	if err = prepareTemplate(ctx, admin, baseDSN, migrate); err != nil {
		cleanup()
		return nil, err
	}

	shared = &sharedDatabase{admin: admin, baseDSN: baseDSN}

	return cleanup, nil
}

// prepareTemplate creates the template database and runs the caller's migrations
// against it exactly once. The pool it migrates through is closed before returning,
// which is load-bearing rather than tidiness: a lingering session on the template
// makes every subsequent clone fail with "source database is being accessed by
// other users".
func prepareTemplate(ctx context.Context, admin *sql.DB, baseDSN string, migrate func(ctx context.Context, db *sql.DB) error) error {
	if _, err := admin.ExecContext(ctx, "CREATE DATABASE "+quoteIdentifier(templateDatabase)); err != nil {
		return fmt.Errorf("creating template database: %w", err)
	}

	dsn, err := dsnForDatabase(baseDSN, templateDatabase)
	if err != nil {
		return err
	}

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("opening template database: %w", err)
	}

	defer func() {
		if closeErr := db.Close(); closeErr != nil {
			fmt.Fprintf(os.Stderr, "postgres testing: closing template pool: %v\n", closeErr)
		}
	}()

	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("connecting to template database: %w", err)
	}

	if err = migrate(ctx, db); err != nil {
		return fmt.Errorf("migrating template database: %w", err)
	}

	return nil
}

// NewIsolatedDatabaseForTest gives the calling test its own freshly migrated database
// inside the binary's shared container, along with the config describing it. The
// database is cloned from the already-migrated template, so the test pays a file copy
// instead of a migration replay, and it is dropped when the test ends.
//
// It is the drop-in replacement for BuildDatabaseContainerForTest in suites whose
// TestMain calls RunTestsWithSharedDatabase — with the difference that the database it
// returns is already migrated, so callers no longer run the migrator themselves.
func NewIsolatedDatabaseForTest(t *testing.T) (*sql.DB, *dbcfg.Config) {
	t.Helper()

	containers.SkipIfNotRunning(t)

	require.NotNil(t, shared, "postgres testing: no shared database; this package's TestMain must call RunTestsWithSharedDatabase")

	return shared.clone(t)
}

// clone mints one database for a test and registers its teardown.
func (s *sharedDatabase) clone(t *testing.T) (*sql.DB, *dbcfg.Config) {
	t.Helper()

	ctx := t.Context()
	name := s.nameFor(t)

	_, err := s.admin.ExecContext(ctx, fmt.Sprintf(
		"CREATE DATABASE %s TEMPLATE %s",
		quoteIdentifier(name), quoteIdentifier(templateDatabase),
	))
	require.NoError(t, err, "cloning template database")

	// Registered before the pool's own cleanup so it runs after it: t.Cleanup is LIFO,
	// and the drop wants the test's connections already gone. FORCE covers whatever the
	// test leaked.
	t.Cleanup(func() {
		dropCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), containers.DefaultShutdownTimeout)
		defer cancel()

		if _, dropErr := s.admin.ExecContext(dropCtx, "DROP DATABASE IF EXISTS "+quoteIdentifier(name)+" WITH (FORCE)"); dropErr != nil {
			t.Logf("postgres testing: dropping %s: %v", name, dropErr)
		}
	})

	dsn, err := dsnForDatabase(s.baseDSN, name)
	require.NoError(t, err)

	config, err := databaseConfigForConnectionString(dsn)
	require.NoError(t, err)

	// Every test's pools now draw on one server's connection budget rather than on a
	// container of their own, so they are sized to share.
	config.MaxOpenConns = maxOpenConnsPerTest
	config.MaxIdleConns = maxIdleConnsPerTest

	db, err := sql.Open(driverName, dsn)
	require.NoError(t, err)
	db.SetMaxOpenConns(maxOpenConnsPerTest)
	db.SetMaxIdleConns(maxIdleConnsPerTest)

	t.Cleanup(func() {
		if closeErr := db.Close(); closeErr != nil {
			t.Logf("postgres testing: closing pool for %s: %v", name, closeErr)
		}
	})

	containers.PingUntilReady(t, ctx, db.PingContext)

	return db, config
}

// nameFor builds a unique, valid database name for a test. The counter guarantees
// uniqueness; the sanitized test name is there so that a stray database or a log line
// points back at the test that made it.
func (s *sharedDatabase) nameFor(t *testing.T) string {
	t.Helper()

	sanitized := strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9', r == '_':
			return r
		case r >= 'A' && r <= 'Z':
			return r + ('a' - 'A')
		default:
			return '_'
		}
	}, t.Name())

	if len(sanitized) > maxSanitizedNameLen {
		sanitized = sanitized[:maxSanitizedNameLen]
	}

	return fmt.Sprintf("ddb_%d_%s", s.counter.Add(1), sanitized)
}

// connectionLimitCustomizer raises the server's connection ceiling to cover every
// parallel test's pools. See maxConnections.
func connectionLimitCustomizer() testcontainers.ContainerCustomizer {
	return testcontainers.WithCmdArgs("-c", fmt.Sprintf("max_connections=%d", maxConnections))
}

// dsnForDatabase re-points a container's DSN at a different database on the same server.
func dsnForDatabase(baseDSN, database string) (string, error) {
	parsed, err := url.Parse(baseDSN)
	if err != nil {
		return "", fmt.Errorf("parsing container DSN: %w", err)
	}

	parsed.Path = "/" + database

	return parsed.String(), nil
}

// quoteIdentifier renders name as a quoted SQL identifier. Database names here are
// generated rather than user-supplied, but they are interpolated into DDL that cannot
// take placeholders, so they are quoted rather than trusted.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}
