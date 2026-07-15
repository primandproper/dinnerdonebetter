package notifications

import (
	"database/sql"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	mockdatabase "github.com/primandproper/platform-go/v4/database/mock"
	"github.com/primandproper/platform-go/v4/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v4/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v4/observability/tracing/noop"

	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	exampleQuantity = 3
)

func buildDatabaseClientForTest(t *testing.T) (c *Repository, auditLogEntryRepo audit.Repository, container *pgcontainers.PostgresContainer) {
	t.Helper()

	ctx := t.Context()
	container, db, config := pgtesting.BuildDatabaseContainerForTest(t)
	require.NoError(t, migrations.NewMigrator(loggingnoop.NewLogger()).Migrate(ctx, db))

	pgc, err := postgres.NewDatabaseClient(ctx, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), config, nil)
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo = auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), pgc)

	c = ProvideNotificationsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, config, pgc)

	return c, auditLogEntryRepo, container
}

func buildInertClientForTest(t *testing.T) *Repository {
	t.Helper()

	cfg := &databasecfg.Config{}
	c := ProvideNotificationsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, cfg, &mockdatabase.ClientMock{ReadDBFunc: func() *sql.DB { return nil }, WriteDBFunc: func() *sql.DB { return nil }})

	return c
}
