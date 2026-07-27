package internalops

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v7/database"
	mockdatabase "github.com/primandproper/platform-go/v7/database/mock"
	"github.com/primandproper/platform-go/v7/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v7/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v7/observability/tracing/noop"

	"github.com/stretchr/testify/require"
)

func buildDatabaseClientForTest(t *testing.T) *repository {
	t.Helper()

	ctx := t.Context()
	db, config := pgtesting.BuildDatabaseContainerForTest(t)
	migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
	require.NoError(t, err)
	require.NoError(t, migrator.Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	c := ProvideInternalOpsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	require.NoError(t, err)

	return c.(*repository)
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideInternalOpsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}
