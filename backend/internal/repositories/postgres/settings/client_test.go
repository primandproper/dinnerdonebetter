package settings

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v8/database"
	mockdatabase "github.com/primandproper/platform-go/v8/database/mock"
	"github.com/primandproper/platform-go/v8/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v8/observability/tracing/noop"

	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity              = 3
	migratedServiceSettingsCount = 1 // user_temperature_unit from migration 00021
)

func buildDatabaseClientForTest(t *testing.T) (c *Repository, auditLogEntryRepo audit.Repository) {
	t.Helper()

	ctx := t.Context()
	db, config := pgtesting.BuildDatabaseContainerForTest(t)
	migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
	require.NoError(t, err)
	require.NoError(t, migrator.Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo = auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	c = ProvideSettingsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)

	return c, auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *Repository {
	t.Helper()

	c := ProvideSettingsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil)

	return c
}
