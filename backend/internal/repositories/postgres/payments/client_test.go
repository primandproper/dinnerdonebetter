package payments

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/require"
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

// buildDatabaseClientForTest builds the store over a real database.
func buildDatabaseClientForTest(t *testing.T) (billing.Store, audit.Repository, database.SQLQueryExecutor) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvidePaymentsRepository(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		pgc,
		nil,
	)
	require.NoError(t, err)

	return c, auditLogEntryRepo, pgc.Writer()
}

// accountForTest creates a user and an account for them, and returns the account.
//
// The account is not incidental: the three account-owned billing tables carry a
// foreign key to accounts, so a subscription naming an account no test created is
// one the database refuses.
func accountForTest(t *testing.T, writer database.SQLQueryExecutor) string {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, writer)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, writer)

	return account.ID
}

// productForTest adds one recurring product to the catalog.
func productForTest(t *testing.T, ctx context.Context, dbc billing.Store) *billing.Product {
	t.Helper()

	product, err := dbc.CreateProduct(ctx, ddbpayments.Scope(), fakes.BuildFakeProduct())
	require.NoError(t, err)

	return product
}

// subscriptionForTest opens one subscription for the account on the product.
func subscriptionForTest(t *testing.T, ctx context.Context, dbc billing.Store, accountID, productID string) *billing.Subscription {
	t.Helper()

	subscription, err := dbc.CreateSubscription(ctx, ddbpayments.Scope(), fakes.BuildFakeSubscription(accountID, productID))
	require.NoError(t, err)

	return subscription
}
