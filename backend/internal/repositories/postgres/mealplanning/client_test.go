package mealplanning

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/database/dialect"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"
	"github.com/primandproper/platform-go/v9/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	"github.com/primandproper/platform-go/v9/outbox"

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

func buildDatabaseClientForTest(t *testing.T) (*repository, audit.Repository) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	identitiesRepo := identity.ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)
	require.NoError(t, err)

	// A real emitter, so the tests exercise the same path production does: the event is
	// another statement in the repository's transaction.
	outboxWriter, err := outbox.NewWriter(dialect.Postgres, outbox.WithWriterLogger(loggingnoop.NewLogger()))
	require.NoError(t, err)

	c := ProvideMealPlanningRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		auditLogEntryRepo,
		identitiesRepo,
		pgc,
		events.NewEmitter(outboxWriter, testDataChangesTopic),
	)

	return c.(*repository), auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideMealPlanningRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil)

	return c.(*repository)
}
