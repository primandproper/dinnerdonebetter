package payments

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// What these tests pin is the half this package adds: that every write records
// the entry it owes, under the account it belongs to, and that the store beneath
// is reachable through it. What the store does with a row — paging, uniqueness,
// the guarded status writes — is platform's and is tested there.

func TestRepository_Integration_Products(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, _ := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	example := fakes.BuildFakeProduct()

	created, err := dbc.CreateProduct(ctx, scope, example)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.False(t, created.CreatedAt.IsZero())
	assert.Equal(t, example.Name, created.Name)

	// A product belongs to nobody, so its entries are recorded under the
	// unattributed actor — the same shape the table this replaced recorded under.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
	})

	fetched, err := dbc.GetProduct(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, example.AmountCents, fetched.AmountCents)
	assert.Equal(t, example.BillingIntervalMonths, fetched.BillingIntervalMonths)

	// The provider-side id is the lookup a catalog sync makes.
	byExternal, err := dbc.GetProductByExternalID(ctx, scope, created.ExternalProductID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byExternal.ID)

	page, err := dbc.ListProducts(ctx, scope, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)

	fetched.Name = "renamed"
	fetched.AmountCents++
	require.NoError(t, dbc.UpdateProduct(ctx, scope, fetched))

	updated, err := dbc.GetProduct(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.Equal(t, example.AmountCents+1, updated.AmountCents)
	assert.NotNil(t, updated.LastUpdatedAt)

	require.NoError(t, dbc.ArchiveProduct(ctx, scope, created.ID))

	afterArchive, err := dbc.GetProduct(ctx, scope, created.ID)
	require.ErrorIs(t, err, billing.ErrProductNotFound)
	assert.Nil(t, afterArchive)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, audit.UnattributedActorID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeProducts, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeProducts, RelevantID: created.ID},
	})

	// Archiving a row that is not there is refused before anything is recorded
	// about it.
	require.ErrorIs(t, dbc.ArchiveProduct(ctx, scope, created.ID), billing.ErrProductNotFound)
}

func TestRepository_Integration_Subscriptions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, writer)
	product := productForTest(t, ctx, dbc)

	example := fakes.BuildFakeSubscription(accountID, product.ID)

	created, err := dbc.CreateSubscription(ctx, scope, example)
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Equal(t, accountID, created.BelongsToAccount)
	assert.Equal(t, capitalism.SubscriptionStatusActive, created.Status)

	// A subscription is an account's, and is recorded under it.
	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSubscriptions, RelevantID: created.ID},
	})

	byExternal, err := dbc.GetSubscriptionByExternalID(ctx, scope, created.ExternalSubscriptionID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, byExternal.ID)

	// The fake's period covers now, so the entitlement read finds it.
	current, err := dbc.ListCurrentSubscriptions(ctx, scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, current.Data, 1)
	assert.Equal(t, created.ID, current.Data[0].ID)

	// The provider's word for where it stands, written and recorded.
	require.NoError(t, dbc.SetSubscriptionStatus(ctx, scope, created.ID, capitalism.SubscriptionStatusPastDue))

	// The same word again is the store's replay answer, and records nothing.
	require.ErrorIs(t, dbc.SetSubscriptionStatus(ctx, scope, created.ID, capitalism.SubscriptionStatusPastDue), billing.ErrStatusUnchanged)

	fetched, err := dbc.GetSubscription(ctx, scope, created.ID)
	require.NoError(t, err)
	assert.Equal(t, capitalism.SubscriptionStatusPastDue, fetched.Status)

	fetched.CurrentPeriodEnd = fetched.CurrentPeriodEnd.AddDate(0, 1, 0)
	require.NoError(t, dbc.UpdateSubscription(ctx, scope, fetched))

	forAccount, err := dbc.ListSubscriptionsForAccount(ctx, scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, forAccount.Data, 1)

	require.NoError(t, dbc.ArchiveSubscription(ctx, scope, created.ID))

	afterArchive, err := dbc.GetSubscription(ctx, scope, created.ID)
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
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, writer)

	product, err := dbc.CreateProduct(ctx, scope, fakes.BuildFakeOneTimeProduct())
	require.NoError(t, err)

	created, err := dbc.CreatePurchase(ctx, scope, fakes.BuildFakePurchase(accountID, product.ID))
	require.NoError(t, err)
	assert.NotEmpty(t, created.ID)
	assert.Nil(t, created.CompletedAt)

	settledAt := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	require.NoError(t, dbc.CompletePurchase(ctx, scope, created.ID, settledAt))

	// A purchase completes exactly once.
	require.ErrorIs(t, dbc.CompletePurchase(ctx, scope, created.ID, settledAt), billing.ErrAlreadyCompleted)

	fetched, err := dbc.GetPurchase(ctx, scope, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched.CompletedAt)
	assert.True(t, fetched.CompletedAt.Equal(settledAt))

	forAccount, err := dbc.ListPurchasesForAccount(ctx, scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, forAccount.Data, 1)

	require.NoError(t, dbc.ArchivePurchase(ctx, scope, created.ID))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePurchases, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePurchases, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypePurchases, RelevantID: created.ID},
	})
}

func TestRepository_Integration_Transactions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)
	scope := ddbpayments.Scope()

	accountID := accountForTest(t, writer)
	product := productForTest(t, ctx, dbc)
	subscription := subscriptionForTest(t, ctx, dbc, accountID, product.ID)

	example := fakes.BuildFakeTransaction(accountID)
	example.SubscriptionID = subscription.ID
	example.Status = billing.TransactionPending

	recorded, err := dbc.RecordTransaction(ctx, scope, example)
	require.NoError(t, err)
	assert.NotEmpty(t, recorded.ID)
	assert.Equal(t, subscription.ID, recorded.SubscriptionID)

	// The redelivery the ledger is shaped around: the same provider-side id
	// collides rather than recording a second charge, and records no entry.
	replay := fakes.BuildFakeTransaction(accountID)
	replay.ExternalTransactionID = example.ExternalTransactionID

	_, err = dbc.RecordTransaction(ctx, scope, replay)
	require.ErrorIs(t, err, billing.ErrTransactionExists)

	require.NoError(t, dbc.SetTransactionStatus(ctx, scope, recorded.ID, billing.TransactionSucceeded))

	ledger, err := dbc.ListTransactionsForAccount(ctx, scope, accountID, nil)
	require.NoError(t, err)
	require.Len(t, ledger.Data, 1)
	assert.Equal(t, billing.TransactionSucceeded, ledger.Data[0].Status)

	require.NoError(t, dbc.ArchiveTransaction(ctx, scope, recorded.ID))

	pgtesting.AssertAuditLogContains(t, ctx, auditRepo, accountID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeSubscriptions, RelevantID: subscription.ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypePaymentTransactions, RelevantID: recorded.ID},
	})
}
