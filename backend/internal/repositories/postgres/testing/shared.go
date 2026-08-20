package testing

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"testing"

	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"

	"github.com/primandproper/platform-go/v12/testutils/containers"
	"github.com/primandproper/platform-go/v12/testutils/containers/pgtest"

	"github.com/stretchr/testify/require"
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
//
// All of that now lives in platform's testutils/containers/pgtest, which is this
// file generalized — pgtest.Start is the container, Instance.NewTemplate is the
// migrated template, and Template.Clone is the per-test database. What remains here
// is the part pgtest cannot know: this repo's image and credentials, its wait
// strategy, its connection budget, and the *dbcfg.Config the repositories are
// configured from.
const (
	// sharedContainerName seeds the container's provisioning credentials and names
	// the template database in a `\l` listing. It is not the name of any database a
	// test connects to: pgtest mints those, each with a random suffix, because
	// postgres truncates identifiers at 63 bytes and two long test names can
	// otherwise collide onto one database with nothing saying so.
	sharedContainerName = "ddb_shared"

	// maxConnections has to cover every parallel test's pools at once now that they
	// share a server. Each test opens a read and a write pool of
	// pgtest.DefaultIsolatedMaxOpenConns, where per-test containers used to give each
	// its own connection budget. Postgres defaults to 100 and pgtest to 200, either
	// of which a wide -parallel run blows through and reports as "too many clients
	// already" from whichever test happens to connect last.
	maxConnections = 500

	// adminMaxOpenConns caps the pool pgtest runs CREATE/DROP DATABASE through. It is
	// one pool for the whole binary, so it is sized like a test's rather than like a
	// server's: the DDL is brief and the budget above belongs to the tests.
	adminMaxOpenConns = pgtest.DefaultIsolatedMaxOpenConns
)

// sharedTemplate is the migrated database every test's is cloned from. It is set by
// RunTestsWithSharedDatabase and read by NewIsolatedDatabaseForTest — written once
// before m.Run and only read afterwards, so parallel tests need no synchronization
// to reach it.
var sharedTemplate *pgtest.Template

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
func RunTestsWithSharedDatabase(m *testing.M, migrate pgtest.MigrateFunc) int {
	// testing.Short() panics ("Short called before Parse") rather than reporting
	// false until flags are parsed, which is why this parses first and why pgtest.Start
	// refuses to consult -short on a caller's behalf. The gate is worth the line: a
	// -short run would otherwise pay for a container and then skip every test that
	// would have queried it.
	flag.Parse()

	if testing.Short() {
		return m.Run()
	}

	if migrate == nil {
		fmt.Fprintln(os.Stderr, "postgres testing: RunTestsWithSharedDatabase requires a non-nil migrate")
		return 1
	}

	cleanup, err := startSharedDatabase(migrate)

	switch {
	case errors.Is(err, pgtest.ErrNoPostgres):
		// The RUN_CONTAINER_TESTS gate is closed, so nothing was started. The tests
		// skip themselves through NewIsolatedDatabaseForTest.
		return m.Run()
	case err != nil:
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
func startSharedDatabase(migrate pgtest.MigrateFunc) (func(), error) {
	ctx := context.Background()

	dbName, username, password := credentialsFor(sharedContainerName)

	instance, stopInstance, err := pgtest.Start(ctx,
		pgtest.WithImage(defaultPostgresImage),
		pgtest.WithCredentials(dbName, username, password),
		pgtest.WithCustomizers(waitStrategy()),
		pgtest.WithMaxConnections(maxConnections),
		pgtest.WithMaxOpenConns(adminMaxOpenConns),
	)
	if err != nil {
		return nil, err
	}

	template, dropTemplate, err := instance.NewTemplate(ctx,
		pgtest.WithMigration(migrate),
		pgtest.WithLabel(sharedContainerName),
	)
	if err != nil {
		return nil, errors.Join(fmt.Errorf("preparing the template database: %w", err), stopInstance())
	}

	sharedTemplate = template

	return func() {
		sharedTemplate = nil

		// The template first: dropping it needs the server still up.
		reportTeardown(dropTemplate)
		reportTeardown(stopInstance)
	}, nil
}

// reportTeardown runs a teardown and reports what it could not do. There is no
// testing.TB to fail by the time these run, and a leftover object on a container
// about to be reaped is not worth an exit code — but a teardown that silently gave
// up is how a leaked container gets discovered by the Docker daemon instead.
func reportTeardown(teardown func() error) {
	if err := teardown(); err != nil {
		fmt.Fprintf(os.Stderr, "postgres testing: %v\n", err)
	}
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

	require.NotNil(t, sharedTemplate, "postgres testing: no shared database; this package's TestMain must call RunTestsWithSharedDatabase")

	isolated := sharedTemplate.Clone(t)

	config, err := databaseConfigForConnectionString(isolated.ConnectionString)
	require.NoError(t, err)

	// The pools a repository opens from this config draw on one server's connection
	// budget rather than on a container of their own, so they are sized to share —
	// matching the pool Clone just handed back, which pgtest sizes the same way.
	config.MaxOpenConns = pgtest.DefaultIsolatedMaxOpenConns
	config.MaxIdleConns = pgtest.DefaultIsolatedMaxIdleConns

	return isolated.DB, config
}
