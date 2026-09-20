package payments

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// What these tests pin is the half this package adds: that every write records
// the entry it owes, under the account it belongs to, and that the store beneath
// is reachable through it. What the store does with a row — paging, uniqueness,
// the guarded status writes — is platform's and is tested there.

func TestRepository_Integration_Products(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	example := fakes.BuildFakeProduct()

	created, err := writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.CreateProduct(ctx, tx, scope, example)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.Equal(t, example.Name, created.Name)

	// A product belongs to nobody, so its entries are recorded under the
	// unattributed actor — the same shape the table this replaced recorded under.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
	})

	fetched, err := dbc.GetProduct(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, example.AmountCents, fetched.AmountCents)
	assert.Equal(t, example.BillingIntervalMonths, fetched.BillingIntervalMonths)

	// The provider-side id is the lookup a catalog sync makes.
	byExternal, err := dbc.GetProductByExternalID(ctx, db.Reader(), scope, created.ExternalProductID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byExternal.ID)

	page, err := dbc.ListProducts(ctx, db.Reader(), scope, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)

	fetched.Name = "renamed"
	fetched.AmountCents++
	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.UpdateProduct(ctx, tx, scope, fetched)
	})
	require.NoError(t, err)

	updated, err := dbc.GetProduct(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.Equal(t, example.AmountCents+1, updated.AmountCents)
	assert.NotNil(t, updated.LastUpdatedAt)

	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.ArchiveProduct(ctx, tx, scope, created.ID)
	})
	require.NoError(t, err)

	afterArchive, err := dbc.GetProduct(ctx, db.Reader(), scope, created.ID)
	require.ErrorIs(t, err, billing.ErrProductNotFound)
	assert.Nil(t, afterArchive)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeProducts, RelevantID: created.ID},
	})

	// Archiving a row that is not there is refused before anything is recorded
	// about it.
	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.ArchiveProduct(ctx, tx, scope, created.ID)
	})
	require.ErrorIs(t, err, billing.ErrProductNotFound)
}

func TestRepository_Integration_Subscriptions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, db)
	product := productForTest(t, ctx, dbc, db)

	example := fakes.BuildFakeSubscription(accountID, product.ID)

	created, err := writeT(ctx, db, func(tx database.Tx) (*billing.Subscription, error) {
		return dbc.CreateSubscription(ctx, tx, scope, example)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, accountID, created.BelongsToAccount)
	assert.Equal(t, capitalism.SubscriptionStatusActive, created.Status)

	// A subscription is an account's, and is recorded under it.
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
	})

	byExternal, err := dbc.GetSubscriptionByExternalID(ctx, db.Reader(), scope, created.ExternalSubscriptionID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byExternal.ID)

	// The fake's period covers now, so the entitlement read finds it.
	current, err := dbc.ListCurrentSubscriptions(ctx, db.Reader(), scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, current.Data, 1)
	assert.Equal(t, created.ID, current.Data[0].ID)

	// The provider's word for where it stands, written and recorded.
	err = execT(ctx, db, func(tx database.Tx) error {
		return dbc.SetSubscriptionStatus(ctx, tx, scope, created.ID, capitalism.SubscriptionStatusPastDue)
	})
	require.NoError(t, err)

	// The same word again is the store's replay answer, and records nothing.
	err = execT(ctx, db, func(tx database.Tx) error {
		return dbc.SetSubscriptionStatus(ctx, tx, scope, created.ID, capitalism.SubscriptionStatusPastDue)
	})
	require.ErrorIs(t, err, billing.ErrStatusUnchanged)

	fetched, err := dbc.GetSubscription(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, capitalism.SubscriptionStatusPastDue, fetched.Status)

	fetched.CurrentPeriodEnd = fetched.CurrentPeriodEnd.AddDate(0, 1, 0)
	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Subscription, error) {
		return dbc.UpdateSubscription(ctx, tx, scope, fetched)
	})
	require.NoError(t, err)

	forAccount, err := dbc.ListSubscriptionsForAccount(ctx, db.Reader(), scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, forAccount.Data, 1)

	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Subscription, error) {
		return dbc.ArchiveSubscription(ctx, tx, scope, created.ID)
	})
	require.NoError(t, err)

	afterArchive, err := dbc.GetSubscription(ctx, db.Reader(), scope, created.ID)
	require.ErrorIs(t, err, billing.ErrSubscriptionNotFound)
	assert.Nil(t, afterArchive)

	// One entry per write that changed something: the create, the status move,
	// the update, and the archive. The refused replay left none.
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
	})
}

func TestRepository_Integration_Purchases(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, db)

	product, err := writeT(ctx, db, func(tx database.Tx) (*billing.Product, error) {
		return dbc.CreateProduct(ctx, tx, scope, fakes.BuildFakeOneTimeProduct())
	})
	require.NoError(t, err)

	created, err := writeT(ctx, db, func(tx database.Tx) (*billing.Purchase, error) {
		return dbc.CreatePurchase(ctx, tx, scope, fakes.BuildFakePurchase(accountID, product.ID))
	})
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Nil(t, created.CompletedAt)

	settledAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Purchase, error) {
		return dbc.CompletePurchase(ctx, tx, scope, created.ID, settledAt)
	})
	require.NoError(t, err)

	// A purchase completes exactly once.
	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Purchase, error) {
		return dbc.CompletePurchase(ctx, tx, scope, created.ID, settledAt)
	})
	require.ErrorIs(t, err, billing.ErrAlreadyCompleted)

	fetched, err := dbc.GetPurchase(ctx, db.Reader(), scope, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.CompletedAt)
	assert.True(t, fetched.CompletedAt.Equal(settledAt))

	forAccount, err := dbc.ListPurchasesForAccount(ctx, db.Reader(), scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, forAccount.Data, 1)

	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Purchase, error) {
		return dbc.ArchivePurchase(ctx, tx, scope, created.ID)
	})
	require.NoError(t, err)

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePurchases, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePurchases, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypePurchases, RelevantID: created.ID},
	})
}

func TestRepository_Integration_Transactions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, db)
	product := productForTest(t, ctx, dbc, db)
	subscription := subscriptionForTest(t, ctx, dbc, db, accountID, product.ID)

	example := fakes.BuildFakeTransaction(accountID)
	example.SubscriptionID = subscription.ID
	example.Status = billing.TransactionPending

	recorded, err := writeT(ctx, db, func(tx database.Tx) (*billing.Transaction, error) {
		return dbc.RecordTransaction(ctx, tx, scope, example)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, recorded.ID)
	assert.Equal(t, subscription.ID, recorded.SubscriptionID)

	// The redelivery the ledger is shaped around: the same provider-side id
	// collides rather than recording a second charge, and records no entry.
	replay := fakes.BuildFakeTransaction(accountID)
	replay.ExternalTransactionID = example.ExternalTransactionID

	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Transaction, error) {
		return dbc.RecordTransaction(ctx, tx, scope, replay)
	})
	require.ErrorIs(t, err, billing.ErrTransactionExists)

	err = execT(ctx, db, func(tx database.Tx) error {
		return dbc.SetTransactionStatus(ctx, tx, scope, recorded.ID, billing.TransactionSucceeded)
	})
	require.NoError(t, err)

	ledger, err := dbc.ListTransactionsForAccount(ctx, db.Reader(), scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, ledger.Data, 1)
	assert.Equal(t, billing.TransactionSucceeded, ledger.Data[0].Status)

	_, err = writeT(ctx, db, func(tx database.Tx) (*billing.Transaction, error) {
		return dbc.ArchiveTransaction(ctx, tx, scope, recorded.ID)
	})
	require.NoError(t, err)

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSubscriptions, RelevantID: subscription.ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
	})
}
