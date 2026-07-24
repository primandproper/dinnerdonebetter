package uploadedmedia

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/database"
	mockdatabase "github.com/primandproper/platform-go/v6/database/mock"
	"github.com/primandproper/platform-go/v6/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	exampleQuantity = 5
)

func buildDatabaseClientForTest(t *testing.T) (*repository, audit.Repository, *pgcontainers.PostgresContainer) {
	t.Helper()

	ctx := t.Context()
	container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
	require.NoError(t, migrations.NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)
	c := ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc)
	require.NoError(t, err)

	return c.(*repository), auditLogEntryRepo, container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideUploadedMediaRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}

func TestProvideUploadedMediaRepository(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual := ProvideUploadedMediaRepository(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			nil,
			&mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }},
		)

		assert.NotNil(t, actual)
	})
}
