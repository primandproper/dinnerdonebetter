package oauth

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	encryptioncfg "github.com/primandproper/platform-go/v6/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v6/database"
	databasecfg "github.com/primandproper/platform-go/v6/database/config"
	mockdatabase "github.com/primandproper/platform-go/v6/database/mock"
	"github.com/primandproper/platform-go/v6/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"

	"github.com/stretchr/testify/require"
	pgcontainers "github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	exampleQuantity = 3
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

	c := ProvideOAuthRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, config, pgc)
	require.NoError(t, err)

	return c.(*repository), auditLogEntryRepo, container
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	config := &databasecfg.Config{
		Provider:                 databasecfg.ProviderPostgres,
		ReadConnection:           databasecfg.ConnectionDetails{},
		Encryption:               encryptioncfg.Config{Provider: encryptioncfg.ProviderSalsa20},
		Debug:                    false,
		LogQueries:               false,
		RunMigrations:            true,
		MaxPingAttempts:          10,
		PingWaitPeriod:           time.Second,
		OAuth2TokenEncryptionKey: "blahblahblahblahblahblahblahblah",
	}

	c := ProvideOAuthRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, config, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }})

	return c.(*repository)
}
