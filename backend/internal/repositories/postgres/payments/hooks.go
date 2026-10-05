/*
Package payments records what a billing write means to the rest of this
application. The catalog, the subscriptions, the purchases and the ledger are
platform-go's: the schema, the paging, the tenancy column, the uniqueness that
turns a redelivered webhook into a collision instead of a second row, and the
guarded status writes all live there, and this package neither reimplements nor
wraps them.

What it adds is the half platform cannot know about — an audit log entry naming
who did what, and, for the catalog and the subscriptions, a data change event on
the outbox that the webhook dispatcher fans out. Every event this emits is in the
webhook event catalog (internal/domain/webhooks/catalog), so a subscriber can
already ask for them; a write that skipped the pair would be a row with no
provenance and a subscriber that never heard.

Both are written from platform's billing.Hooks, which the store calls on the
caller's transaction once each write has landed. A hook's error fails the write,
so the row, the entry and the event commit together or not at all. A redelivery
the store refuses — billing.ErrStatusUnchanged, billing.ErrAlreadyCompleted, an
exists sentinel — calls no hook, so a replay records nothing.

# Which writes emit, and which only record

Products and subscriptions emit, because a subscriber has something to do with
them: a catalog change is a price somebody displays, and a subscription moving is
churn or revenue. A status change reached by SetSubscriptionStatus emits the same
subscription_updated event an administrative edit does — the provider's word for
where an agreement stands is the update a subscriber most wants to hear about.

Purchases and ledger rows are recorded in the audit log and emit nothing, as they
did before the store was adopted. A purchase is written by a checkout flow this
application does not yet have, and a ledger row is a fact the provider already
holds; the day either needs a subscriber, it is an event constant and a line in
the catalog.

# What an entry is written from

Every entry is written from a row platform hands the hook, never from what the
caller passed in. An archive's is the row as the archive left it, which no read
by id reaches afterwards; an update's and a status move's carry the field-level
difference between the row before and the row after, which is what an update
means and is not readable once the write has run.
*/
package payments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/billing"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// What an audit entry about each table names. They are the names of the tables
// this store replaced, so an entry recorded before the adoption and one recorded
// after read the same.
const (
	resourceTypeProducts            = "products"
	resourceTypeSubscriptions       = "subscriptions"
	resourceTypePurchases           = "purchases"
	resourceTypePaymentTransactions = "payment_transactions"
)

// hooks records every billing write.
//
// It implements Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides what it records.
// Embedded, the new write would compile and record nothing, which is the one
// failure an audit log cannot notice.
type hooks struct {
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	recorder          *recording.Recorder
}

var _ billing.Hooks = (*hooks)(nil)

// AfterCreateProduct records a product being added to the catalog.
func (h *hooks) AfterCreateProduct(ctx context.Context, tx database.Tx, _ tenancy.Scope, product *billing.Product) error {
	return h.recordProduct(ctx, tx, product, platformaudit.EventCreated, ddbpayments.ProductCreatedServiceEventType, nil)
}

// AfterUpdateProduct records a product being rewritten, and which of its fields
// moved.
func (h *hooks) AfterUpdateProduct(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *billing.Product) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated product")
	}

	return h.recordProduct(ctx, tx, after, platformaudit.EventUpdated, ddbpayments.ProductUpdatedServiceEventType, changes)
}

// AfterArchiveProduct records a product being withdrawn from sale.
func (h *hooks) AfterArchiveProduct(ctx context.Context, tx database.Tx, _ tenancy.Scope, product *billing.Product) error {
	return h.recordProduct(ctx, tx, product, platformaudit.EventArchived, ddbpayments.ProductArchivedServiceEventType, nil)
}

// AfterCreateSubscription records an agreement being opened.
func (h *hooks) AfterCreateSubscription(ctx context.Context, tx database.Tx, _ tenancy.Scope, subscription *billing.Subscription) error {
	return h.recordSubscription(ctx, tx, subscription, platformaudit.EventCreated, ddbpayments.SubscriptionCreatedServiceEventType, nil)
}

// AfterUpdateSubscription records a sync rewriting an agreement, and which of
// its fields moved. The account it is filed under is the stored one, which a
// sync cannot move.
func (h *hooks) AfterUpdateSubscription(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *billing.Subscription) error {
	return h.recordSubscriptionUpdate(ctx, tx, before, after)
}

// AfterSetSubscriptionStatus records the provider's word for where an agreement
// stands, as the same update an administrative edit is.
func (h *hooks) AfterSetSubscriptionStatus(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *billing.Subscription) error {
	return h.recordSubscriptionUpdate(ctx, tx, before, after)
}

// AfterArchiveSubscription records an agreement being retired administratively.
func (h *hooks) AfterArchiveSubscription(ctx context.Context, tx database.Tx, _ tenancy.Scope, subscription *billing.Subscription) error {
	return h.recordSubscription(ctx, tx, subscription, platformaudit.EventArchived, ddbpayments.SubscriptionArchivedServiceEventType, nil)
}

// AfterCreatePurchase records a sale.
func (h *hooks) AfterCreatePurchase(ctx context.Context, tx database.Tx, _ tenancy.Scope, purchase *billing.Purchase) error {
	return h.record(ctx, tx, purchase.BelongsToAccount, resourceTypePurchases, purchase.ID, platformaudit.EventCreated, nil)
}

// AfterCompletePurchase records the money for a sale arriving. It is handed no
// row from before, because the one before is implied — outstanding — so the
// entry carries no changes.
func (h *hooks) AfterCompletePurchase(ctx context.Context, tx database.Tx, _ tenancy.Scope, purchase *billing.Purchase) error {
	return h.record(ctx, tx, purchase.BelongsToAccount, resourceTypePurchases, purchase.ID, platformaudit.EventUpdated, nil)
}

// AfterArchivePurchase records a sale being retired administratively.
func (h *hooks) AfterArchivePurchase(ctx context.Context, tx database.Tx, _ tenancy.Scope, purchase *billing.Purchase) error {
	return h.record(ctx, tx, purchase.BelongsToAccount, resourceTypePurchases, purchase.ID, platformaudit.EventArchived, nil)
}

// AfterRecordTransaction records a ledger row being written.
func (h *hooks) AfterRecordTransaction(ctx context.Context, tx database.Tx, _ tenancy.Scope, transaction *billing.Transaction) error {
	return h.record(ctx, tx, transaction.BelongsToAccount, resourceTypePaymentTransactions, transaction.ID, platformaudit.EventCreated, nil)
}

// AfterSetTransactionStatus records an attempt's outcome moving, and from what.
func (h *hooks) AfterSetTransactionStatus(ctx context.Context, tx database.Tx, _ tenancy.Scope, before, after *billing.Transaction) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the moved transaction")
	}

	return h.record(ctx, tx, after.BelongsToAccount, resourceTypePaymentTransactions, after.ID, platformaudit.EventUpdated, changes)
}

// AfterArchiveTransaction records a ledger row being retired administratively.
func (h *hooks) AfterArchiveTransaction(ctx context.Context, tx database.Tx, _ tenancy.Scope, transaction *billing.Transaction) error {
	return h.record(ctx, tx, transaction.BelongsToAccount, resourceTypePaymentTransactions, transaction.ID, platformaudit.EventArchived, nil)
}

// recordSubscriptionUpdate is UpdateSubscription and SetSubscriptionStatus: one
// subscription_updated event, and an entry carrying what moved.
func (h *hooks) recordSubscriptionUpdate(ctx context.Context, tx database.Tx, before, after *billing.Subscription) error {
	changes, err := platformaudit.Diff(before, after)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the updated subscription")
	}

	return h.recordSubscription(ctx, tx, after, platformaudit.EventUpdated, ddbpayments.SubscriptionUpdatedServiceEventType, changes)
}

// recordProduct writes the audit entry and the data change event for a write to
// the catalog.
//
// The entry names no account, because a product is a catalog row that belongs
// to nobody: who wrote it is the actor on the context, which is what the audit
// recorder resolves. The event names no account either, so it is service-wide —
// it reaches a webhook subscriber under whichever account the requester had
// active, resolved from the context by the emitter.
func (h *hooks) recordProduct(
	ctx context.Context,
	tx database.Tx,
	product *billing.Product,
	auditEventType platformaudit.EventType,
	changeEventType string,
	changes map[string]platformaudit.Change,
) error {
	return h.recordAndEmit(ctx, tx, "", resourceTypeProducts, product.ID, auditEventType, changeEventType, changes, map[string]any{
		paymentskeys.ProductIDKey: product.ID,
	})
}

// recordSubscription writes the audit entry and the data change event for a
// write to one account's subscription.
//
// Both are filed under the subscription's account rather than whoever made the
// request, and the account is passed to the emitter explicitly rather than read
// off the context. Most of these writes have no session: a provider's webhook
// carries no user, and an event that had to find its account on the context
// would find nobody there.
func (h *hooks) recordSubscription(
	ctx context.Context,
	tx database.Tx,
	subscription *billing.Subscription,
	auditEventType platformaudit.EventType,
	changeEventType string,
	changes map[string]platformaudit.Change,
) error {
	return h.recordAndEmit(ctx, tx, subscription.BelongsToAccount, resourceTypeSubscriptions, subscription.ID, auditEventType, changeEventType, changes, map[string]any{
		paymentskeys.SubscriptionIDKey: subscription.ID,
		paymentskeys.ProductIDKey:      subscription.ProductID,
		identitykeys.AccountIDKey:      subscription.BelongsToAccount,
	})
}

// recordAndEmit writes the audit entry and enqueues the data change event on the
// transaction the write ran in. changes is the field-level diff an update
// carries, and nil for every other write.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (h *hooks) recordAndEmit(
	ctx context.Context, tx database.Tx,
	accountID, resourceType, relevantID string, auditEventType platformaudit.EventType, changeEventType string,
	changes map[string]platformaudit.Change,
	metadata map[string]any,
) error {
	logger := h.logger.WithValue(resourceType, relevantID)

	entry := audit.NewEntry("", accountID, resourceType, relevantID, auditEventType)
	entry.Changes = changes

	return h.recorder.RecordAndEmit(ctx, tx, logger, entry, changeEventType, accountID, metadata)
}

// record writes the audit entry alone, for the writes that owe an entry and no
// event. See the package documentation for which those are.
func (h *hooks) record(
	ctx context.Context, tx database.Tx,
	accountID, resourceType, relevantID string, auditEventType platformaudit.EventType,
	changes map[string]platformaudit.Change,
) error {
	entry := audit.NewEntry("", accountID, resourceType, relevantID, auditEventType)
	entry.Changes = changes

	if err := h.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
		return platformerrors.Wrap(err, "creating audit log entry")
	}

	return nil
}
