package identity

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/database/dialect"
	mockdatabase "github.com/primandproper/platform-go/v11/database/mock"
	"github.com/primandproper/platform-go/v11/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"
	"github.com/primandproper/platform-go/v11/outbox"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3

	// testDataChangesTopic is what the emitter writes onto outbox rows in these tests.
	testDataChangesTopic = "data_changes"
)

type sqlmockExpecterWrapper struct {
	sqlmock.Sqlmock
}

func (e *sqlmockExpecterWrapper) AssertExpectations(t assert.TestingT) bool {
	return assert.NoError(t, e.ExpectationsWereMet(), "not all database expectations were met")
}

func buildMockSQLTestClient(t *testing.T) (*repository, *sqlmockExpecterWrapper) {
	t.Helper()

	fakeDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	c := &repository{
		Client:           pgtesting.NewSQLMockDatabaseClient(fakeDB),
		readDB:           fakeDB,
		writeDB:          fakeDB,
		logger:           loggingnoop.NewLogger(),
		generatedQuerier: generated.New(),
		tracer:           tracing.NewTracerForTest("test"),
	}

	return c, &sqlmockExpecterWrapper{Sqlmock: sqlMock}
}

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from, so that a test costs a database
// clone rather than a container start plus a migration replay. See
// pgtesting.RunTestsWithSharedDatabase.
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

	auditLogRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, pgc)
	require.NoError(t, err)

	// A real emitter, so the tests exercise the same path production does: the data change
	// event and the search index event are further statements in the repository's transaction.
	outboxWriter, err := outbox.NewWriter(dialect.Postgres, outbox.WithWriterLogger(loggingnoop.NewLogger()), outbox.WithWriterSideEffect(indexevents.SideEffectName, indexevents.SideEffect))
	require.NoError(t, err)

	c := ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogRepo, pgc, events.NewEmitter(outboxWriter, testDataChangesTopic, nil, indexevents.SideEffect))
	require.NoError(t, err)

	return c.(*repository), auditLogRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil)

	return c.(*repository)
}
