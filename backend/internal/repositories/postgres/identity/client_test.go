package identity

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/identity/generated"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v7/database"
	mockdatabase "github.com/primandproper/platform-go/v7/database/mock"
	"github.com/primandproper/platform-go/v7/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
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

func buildDatabaseClientForTest(t *testing.T) (*repository, audit.Repository) {
	t.Helper()

	ctx := t.Context()
	db, config := pgtesting.BuildDatabaseContainerForTest(t)
	migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
	require.NoError(t, err)
	require.NoError(t, migrator.Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)

	c := ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogRepo, pgc)
	require.NoError(t, err)

	return c.(*repository), auditLogRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideIdentityRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}
