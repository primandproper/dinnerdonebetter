package auth

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auth/generated"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v8/database"
	mockdatabase "github.com/primandproper/platform-go/v8/database/mock"
	"github.com/primandproper/platform-go/v8/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	"github.com/primandproper/platform-go/v8/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"

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

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)

	c := ProvideAuthRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogRepo, pgc)

	return c.(*repository), auditLogRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideAuthRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}
