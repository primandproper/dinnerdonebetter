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

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

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
func buildDatabaseClientForTest(t *testing.T) (billing.Store, audit.Repository, database.Client) {
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

	return c, auditLogEntryRepo, pgc
}

// accountForTest creates a user and an account for them, and returns the account.
//
// The account is not incidental: the three account-owned billing tables carry a
// foreign key to accounts, so a subscription naming an account no test created is
// one the database refuses.
func accountForTest(t *testing.T, db database.Client) string {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, db.Writer())

	return account.ID
}

// productForTest adds one recurring product to the catalog.
func productForTest(t *testing.T, ctx context.Context, dbc billing.Store, db database.Client) *billing.Product {
	t.Helper()

	product, err := writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.CreateProduct(ctx, tx, ddbpayments.Scope(), fakes.BuildFakeProduct())
	})
	require.NoError(t, err)

	return product
}

// subscriptionForTest opens one subscription for the account on the product.
func subscriptionForTest(t *testing.T, ctx context.Context, dbc billing.Store, db database.Client, accountID, productID string) *billing.Subscription {
	t.Helper()

	subscription, err := writeT(ctx, db, func(tx database.Tx) (*billing.Subscription, error) {
		return dbc.CreateSubscription(ctx, tx, ddbpayments.Scope(), fakes.BuildFakeSubscription(accountID, productID))
	})
	require.NoError(t, err)

	return subscription
}

// writeT runs one store write on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test that wants one
// row written supplies the transaction the production caller would — and, here, the one that
// carries the audit entry and the outbox event alongside it. It answers with the error rather
// than asserting on it, because a good number of the writes here are supposed to fail.
func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}

// execT is writeT for the two status writes, which answer with nothing.
func execT(ctx context.Context, db database.Client, write func(tx database.Tx) error) error {
	return db.WithTransaction(ctx, write)
}
