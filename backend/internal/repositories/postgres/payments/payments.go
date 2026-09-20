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

# The transaction the events are in

Every hand-written repository here emits inside the transaction that wrote
the row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one now does too. It could not
before: platform's writes owned their transactions and took no executor, so
the audit entry and the event were a second transaction after the first had
committed, and a subscription could exist that nothing had recorded.

As of platform-go v14 a store write takes the caller's database.Tx, so the
write, the entry and the event are one transaction and share one fate. The
gap filed for this package as platform-go #466 is closed by that convention
rather than by anything here, which is why this package has no workaround to
delete.

# Which writes emit, and which only record

Products and subscriptions emit, because a subscriber has something to do with
them: a catalog change is a price somebody displays, and a subscription moving is
churn or revenue. A status change reached by SetSubscriptionStatus emits the same
subscription_updated event an administrative edit does — the provider's word for
where an agreement stands is the update a subscriber most wants to hear about,
and the table this replaced recorded it without telling anybody.

Purchases and ledger rows are recorded in the audit log and emit nothing, as they
did before the store was adopted. A purchase is written by a checkout flow this
application does not yet have, and a ledger row is a fact the provider already
holds; the day either needs a subscriber, it is an event constant and a line in
the catalog.

# Reads before writes

The archive and status paths read the row before the store runs, and not only
for the account the audit entry is filed under. An archive hides the row from
every read that does not ask for archived ones, so an entry recorded afterwards
would be an entry about a row the reader has to know to ask for. Reading first
also refuses the write against a row that is not there with the store's own
not-found, before anything is written about it.
*/
package payments

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
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

var _ billing.Store = (*repository)(nil)

// CreateProduct adds the product to the catalog, then records it.
func (r *repository) CreateProduct(ctx context.Context, tx database.Tx, scope tenancy.Scope, product *billing.Product) (*billing.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateProduct(ctx, tx, scope, product)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, created.ID)

	if err = r.recordProduct(ctx, tx, created, audit.AuditLogEventTypeCreated, ddbpayments.ProductCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateProduct rewrites the product, then records it.
func (r *repository) UpdateProduct(ctx context.Context, tx database.Tx, scope tenancy.Scope, product *billing.Product) (*billing.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.Store.UpdateProduct(ctx, tx, scope, product)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, product.ID)

	if err = r.recordProduct(ctx, tx, product, audit.AuditLogEventTypeUpdated, ddbpayments.ProductUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// ArchiveProduct withdraws the product from sale, then records it.
func (r *repository) ArchiveProduct(ctx context.Context, tx database.Tx, scope tenancy.Scope, productID string) (*billing.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, productID)

	product, err := r.GetProduct(ctx, tx, scope, productID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching product to record")
	}

	result, err := r.Store.ArchiveProduct(ctx, tx, scope, productID)
	if err != nil {
		return nil, err
	}

	if err = r.recordProduct(ctx, tx, product, audit.AuditLogEventTypeArchived, ddbpayments.ProductArchivedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// CreateSubscription opens the agreement, then records it.
func (r *repository) CreateSubscription(ctx context.Context, tx database.Tx, scope tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateSubscription(ctx, tx, scope, subscription)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, created.ID)

	if err = r.recordSubscription(ctx, tx, created, audit.AuditLogEventTypeCreated, ddbpayments.SubscriptionCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateSubscription rewrites the subscription, then records it.
func (r *repository) UpdateSubscription(ctx context.Context, tx database.Tx, scope tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.Store.UpdateSubscription(ctx, tx, scope, subscription)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscription.ID)

	if err = r.recordSubscription(ctx, tx, subscription, audit.AuditLogEventTypeUpdated, ddbpayments.SubscriptionUpdatedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// SetSubscriptionStatus moves the subscription's standing, then records it.
//
// A redelivered event is reported by the store as billing.ErrStatusUnchanged
// before anything here runs, so a replay records nothing: there is no second
// entry for a change that did not happen.
func (r *repository) SetSubscriptionStatus(ctx context.Context, tx database.Tx, scope tenancy.Scope, subscriptionID string, status capitalism.SubscriptionStatus) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscriptionID)

	subscription, err := r.GetSubscription(ctx, tx, scope, subscriptionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching subscription to record")
	}

	if err = r.Store.SetSubscriptionStatus(ctx, tx, scope, subscriptionID, status); err != nil {
		return err
	}

	subscription.Status = status

	return r.recordSubscription(ctx, tx, subscription, audit.AuditLogEventTypeUpdated, ddbpayments.SubscriptionUpdatedServiceEventType)
}

// ArchiveSubscription retires the subscription administratively, then records it.
func (r *repository) ArchiveSubscription(ctx context.Context, tx database.Tx, scope tenancy.Scope, subscriptionID string) (*billing.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscriptionID)

	subscription, err := r.GetSubscription(ctx, tx, scope, subscriptionID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching subscription to record")
	}

	result, err := r.Store.ArchiveSubscription(ctx, tx, scope, subscriptionID)
	if err != nil {
		return nil, err
	}

	if err = r.recordSubscription(ctx, tx, subscription, audit.AuditLogEventTypeArchived, ddbpayments.SubscriptionArchivedServiceEventType); err != nil {
		return nil, err
	}

	return result, nil
}

// CreatePurchase writes the purchase, then records it.
func (r *repository) CreatePurchase(ctx context.Context, tx database.Tx, scope tenancy.Scope, purchase *billing.Purchase) (*billing.Purchase, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreatePurchase(ctx, tx, scope, purchase)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, created.ID)

	if err = r.recordPurchase(ctx, tx, created, audit.AuditLogEventTypeCreated); err != nil {
		return nil, err
	}

	return created, nil
}

// CompletePurchase stamps the moment the money arrived, then records it.
func (r *repository) CompletePurchase(ctx context.Context, tx database.Tx, scope tenancy.Scope, purchaseID string, at time.Time) (*billing.Purchase, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, purchaseID)

	purchase, err := r.GetPurchase(ctx, tx, scope, purchaseID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching purchase to record")
	}

	result, err := r.Store.CompletePurchase(ctx, tx, scope, purchaseID, at)
	if err != nil {
		return nil, err
	}

	if err = r.recordPurchase(ctx, tx, purchase, audit.AuditLogEventTypeUpdated); err != nil {
		return nil, err
	}

	return result, nil
}

// ArchivePurchase retires the purchase administratively, then records it.
func (r *repository) ArchivePurchase(ctx context.Context, tx database.Tx, scope tenancy.Scope, purchaseID string) (*billing.Purchase, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, purchaseID)

	purchase, err := r.GetPurchase(ctx, tx, scope, purchaseID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching purchase to record")
	}

	result, err := r.Store.ArchivePurchase(ctx, tx, scope, purchaseID)
	if err != nil {
		return nil, err
	}

	if err = r.recordPurchase(ctx, tx, purchase, audit.AuditLogEventTypeArchived); err != nil {
		return nil, err
	}

	return result, nil
}

// RecordTransaction writes the ledger row, then records it.
func (r *repository) RecordTransaction(ctx context.Context, tx database.Tx, scope tenancy.Scope, transaction *billing.Transaction) (*billing.Transaction, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	recorded, err := r.Store.RecordTransaction(ctx, tx, scope, transaction)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, recorded.ID)

	if err = r.recordTransaction(ctx, tx, recorded, audit.AuditLogEventTypeCreated); err != nil {
		return nil, err
	}

	return recorded, nil
}

// SetTransactionStatus moves the attempt's outcome, then records it. A replay is
// billing.ErrStatusUnchanged and records nothing — see SetSubscriptionStatus.
func (r *repository) SetTransactionStatus(ctx context.Context, tx database.Tx, scope tenancy.Scope, transactionID string, status billing.TransactionStatus) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, transactionID)

	transaction, err := r.GetTransaction(ctx, tx, scope, transactionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching transaction to record")
	}

	if err = r.Store.SetTransactionStatus(ctx, tx, scope, transactionID, status); err != nil {
		return err
	}

	return r.recordTransaction(ctx, tx, transaction, audit.AuditLogEventTypeUpdated)
}

// ArchiveTransaction retires the ledger row administratively, then records it.
func (r *repository) ArchiveTransaction(ctx context.Context, tx database.Tx, scope tenancy.Scope, transactionID string) (*billing.Transaction, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, transactionID)

	transaction, err := r.GetTransaction(ctx, tx, scope, transactionID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching transaction to record")
	}

	result, err := r.Store.ArchiveTransaction(ctx, tx, scope, transactionID)
	if err != nil {
		return nil, err
	}

	if err = r.recordTransaction(ctx, tx, transaction, audit.AuditLogEventTypeArchived); err != nil {
		return nil, err
	}

	return result, nil
}

// recordProduct writes the audit entry and the data change event for a write to
// the catalog.
//
// The entry names no account, because a product is a catalog row that belongs
// to nobody: who wrote it is the actor on the context, which is what the audit
// recorder resolves. The event names no account either, so it is service-wide —
// it reaches a webhook subscriber under whichever account the requester had
// active, resolved from the context by the emitter.
func (r *repository) recordProduct(ctx context.Context, tx database.Tx, product *billing.Product, auditEventType, changeEventType string) error {
	return r.recordAndEmit(ctx, tx, "", resourceTypeProducts, product.ID, auditEventType, changeEventType, map[string]any{
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
func (r *repository) recordSubscription(ctx context.Context, tx database.Tx, subscription *billing.Subscription, auditEventType, changeEventType string) error {
	return r.recordAndEmit(ctx, tx, subscription.BelongsToAccount, resourceTypeSubscriptions, subscription.ID, auditEventType, changeEventType, map[string]any{
		paymentskeys.SubscriptionIDKey: subscription.ID,
		paymentskeys.ProductIDKey:      subscription.ProductID,
		identitykeys.AccountIDKey:      subscription.BelongsToAccount,
	})
}

// recordPurchase writes the audit entry for a write to one account's purchase.
func (r *repository) recordPurchase(ctx context.Context, tx database.Tx, purchase *billing.Purchase, auditEventType string) error {
	return r.record(ctx, tx, purchase.BelongsToAccount, resourceTypePurchases, purchase.ID, auditEventType)
}

// recordTransaction writes the audit entry for a write to one account's ledger.
func (r *repository) recordTransaction(ctx context.Context, tx database.Tx, transaction *billing.Transaction, auditEventType string) error {
	return r.record(ctx, tx, transaction.BelongsToAccount, resourceTypePaymentTransactions, transaction.ID, auditEventType)
}

// recordAndEmit writes the audit entry and enqueues the data change event, in
// one transaction of their own.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (r *repository) recordAndEmit(
	ctx context.Context, tx database.Tx,
	accountID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(resourceType, relevantID)

	return r.recorder.RecordAndEmit(ctx, tx, logger, auditEntry(accountID, resourceType, relevantID, auditEventType), changeEventType, accountID, metadata)
}

// record writes the audit entry alone, for the writes that owe an entry and no
// event. See the package documentation for which those are.
func (r *repository) record(ctx context.Context, tx database.Tx, accountID, resourceType, relevantID, auditEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.auditLogEntryRepo.Record(ctx, tx, auditEntry(accountID, resourceType, relevantID, auditEventType)); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	return nil
}

// auditEntry is the entry every write here records. The account is a pointer
// because the catalog has none, and a product's entry has to be able to say so.
func auditEntry(accountID, resourceType, relevantID, auditEventType string) *audit.AuditLogEntry {
	entry := &audit.AuditLogEntry{
		ID:           identifiers.New(),
		ResourceType: resourceType,
		RelevantID:   relevantID,
		EventType:    auditEventType,
	}

	if accountID != "" {
		entry.BelongsToAccount = &accountID
	}

	return entry
}
