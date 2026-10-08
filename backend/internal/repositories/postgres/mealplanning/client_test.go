package mealplanning

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v15/mediaregistry"
	registrymock "github.com/primandproper/platform-go/v15/mediaregistry/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3

	// testDataChangesTopic is what the emitter writes onto outbox rows in these tests.
	testDataChangesTopic = "data_changes"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from. This package has by far the most
// container-backed tests in the repo, and giving each its own container asked the
// Docker daemon for more instances at once than it would serve — see the commentary on
// RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

func buildDatabaseClientForTest(t *testing.T) (*repository, *auditlogentries.Log) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLog(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, pgc)
	require.NoError(t, err)
	// A real registry store over the same database, so the media hydration these
	// tests exercise reads the table a request would.
	uploadsRegistry, err := mediaregistry.NewSQLStore(pgc, mediaregistry.WithTablePrefix(branding.TablePrefix))
	require.NoError(t, err)

	// The real recording spine, so the tests exercise the same path production does: the
	// event and the entry are further statements in the repository's transaction.
	auditRecorder := auditLogEntryRepo.Recorder()

	spine := pgtesting.NewSpineForTest(t, ctx, pgc, auditRecorder)

	c := ProvideMealPlanningRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		pgc,
		spine.Emitter,
		spine.Recorder,
		spine.Writer,
		uploadsRegistry,
	)

	return c.(*repository), auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideMealPlanningRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil, nil, nil, &registrymock.StoreMock{})

	return c.(*repository)
}
