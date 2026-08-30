package webhooks

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/database"
	mockdatabase "github.com/primandproper/platform-go/v13/database/mock"
	"github.com/primandproper/platform-go/v13/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v13/webhooks/config"

	"github.com/stretchr/testify/require"
)

const (
	exampleQuantity = 3
)

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

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, pgc)
	require.NoError(t, err)

	// A real dispatcher over the same database: creating a webhook registers a delivery
	// endpoint, and these tests assert what actually lands in the webhook tables.
	store, err := webhookscfg.NewStore(ctx, &webhookscfg.Config{}, pgc)
	require.NoError(t, err)

	dispatcher, err := webhookscfg.NewDispatcher(
		ctx,
		&webhookscfg.Config{},
		store,
		catalog.Catalog(),
		webhookscfg.WithLogger(loggingnoop.NewLogger()),
		webhookscfg.WithTracerProvider(tracingnoop.NewTracerProvider()),
		webhookscfg.WithMetricsProvider(metricsnoop.NewMetricsProvider()),
		// The fakes' URLs name domains that do not resolve, and resolving them is not what
		// these tests are about. The policy itself is platform-go's, and covered there.
		webhookscfg.WithDispatcherOptions(
			webhooks.WithDispatcherURLChecker(func(context.Context, string) error { return nil }),
		),
	)
	require.NoError(t, err)

	c := ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil, dispatcher, store)

	return c.(*repository), auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil, nil, nil)

	return c.(*repository)
}
