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

# The transaction the events are not in

Every hand-written repository here emits inside the transaction that wrote the
row, so the event lives or dies with what it describes (see
internal/repositories/postgres/events). This one cannot: platform's writes own
their transactions and take no executor, so the audit entry and the event are a
second transaction after the first has committed.

The gap that opens is the ordinary one — the row lands, the process dies, and
nothing is recorded about it. It is narrow and it is one-directional: a
subscription can exist with no event, but no event can name a subscription that
was not written. Closing it needs platform's write methods to accept a
database.Tx, which is the same gap comments (platform-go #457), waitlists (#458),
settings (#460) and issuereports (#465) have. It is filed for this package as
platform-go #466 rather than worked around here — a gap papered over locally
stops being a gap anyone remembers. See #1419 for what deletes here when it
lands.

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

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"
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
func (r *repository) CreateProduct(ctx context.Context, scope tenancy.Scope, product *billing.Product) (*billing.Product, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateProduct(ctx, scope, product)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, created.ID)

	if err = r.recordProduct(ctx, created, audit.AuditLogEventTypeCreated, ddbpayments.ProductCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateProduct rewrites the product, then records it.
func (r *repository) UpdateProduct(ctx context.Context, scope tenancy.Scope, product *billing.Product) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.UpdateProduct(ctx, scope, product); err != nil {
		return err
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, product.ID)

	return r.recordProduct(ctx, product, audit.AuditLogEventTypeUpdated, ddbpayments.ProductUpdatedServiceEventType)
}

// ArchiveProduct withdraws the product from sale, then records it.
func (r *repository) ArchiveProduct(ctx context.Context, scope tenancy.Scope, productID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, productID)

	product, err := r.GetProduct(ctx, scope, productID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching product to record")
	}

	if err = r.Store.ArchiveProduct(ctx, scope, productID); err != nil {
		return err
	}

	return r.recordProduct(ctx, product, audit.AuditLogEventTypeArchived, ddbpayments.ProductArchivedServiceEventType)
}

// CreateSubscription opens the agreement, then records it.
func (r *repository) CreateSubscription(ctx context.Context, scope tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreateSubscription(ctx, scope, subscription)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, created.ID)

	if err = r.recordSubscription(ctx, created, audit.AuditLogEventTypeCreated, ddbpayments.SubscriptionCreatedServiceEventType); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateSubscription rewrites the subscription, then records it.
func (r *repository) UpdateSubscription(ctx context.Context, scope tenancy.Scope, subscription *billing.Subscription) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.Store.UpdateSubscription(ctx, scope, subscription); err != nil {
		return err
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscription.ID)

	return r.recordSubscription(ctx, subscription, audit.AuditLogEventTypeUpdated, ddbpayments.SubscriptionUpdatedServiceEventType)
}

// SetSubscriptionStatus moves the subscription's standing, then records it.
//
// A redelivered event is reported by the store as billing.ErrStatusUnchanged
// before anything here runs, so a replay records nothing: there is no second
// entry for a change that did not happen.
func (r *repository) SetSubscriptionStatus(ctx context.Context, scope tenancy.Scope, subscriptionID string, status capitalism.SubscriptionStatus) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscriptionID)

	subscription, err := r.GetSubscription(ctx, scope, subscriptionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching subscription to record")
	}

	if err = r.Store.SetSubscriptionStatus(ctx, scope, subscriptionID, status); err != nil {
		return err
	}

	subscription.Status = status

	return r.recordSubscription(ctx, subscription, audit.AuditLogEventTypeUpdated, ddbpayments.SubscriptionUpdatedServiceEventType)
}

// ArchiveSubscription retires the subscription administratively, then records it.
func (r *repository) ArchiveSubscription(ctx context.Context, scope tenancy.Scope, subscriptionID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, subscriptionID)

	subscription, err := r.GetSubscription(ctx, scope, subscriptionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching subscription to record")
	}

	if err = r.Store.ArchiveSubscription(ctx, scope, subscriptionID); err != nil {
		return err
	}

	return r.recordSubscription(ctx, subscription, audit.AuditLogEventTypeArchived, ddbpayments.SubscriptionArchivedServiceEventType)
}

// CreatePurchase writes the purchase, then records it.
func (r *repository) CreatePurchase(ctx context.Context, scope tenancy.Scope, purchase *billing.Purchase) (*billing.Purchase, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	created, err := r.Store.CreatePurchase(ctx, scope, purchase)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, created.ID)

	if err = r.recordPurchase(ctx, created, audit.AuditLogEventTypeCreated); err != nil {
		return nil, err
	}

	return created, nil
}

// CompletePurchase stamps the moment the money arrived, then records it.
func (r *repository) CompletePurchase(ctx context.Context, scope tenancy.Scope, purchaseID string, at time.Time) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, purchaseID)

	purchase, err := r.GetPurchase(ctx, scope, purchaseID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching purchase to record")
	}

	if err = r.Store.CompletePurchase(ctx, scope, purchaseID, at); err != nil {
		return err
	}

	return r.recordPurchase(ctx, purchase, audit.AuditLogEventTypeUpdated)
}

// ArchivePurchase retires the purchase administratively, then records it.
func (r *repository) ArchivePurchase(ctx context.Context, scope tenancy.Scope, purchaseID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PurchaseIDKey, purchaseID)

	purchase, err := r.GetPurchase(ctx, scope, purchaseID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching purchase to record")
	}

	if err = r.Store.ArchivePurchase(ctx, scope, purchaseID); err != nil {
		return err
	}

	return r.recordPurchase(ctx, purchase, audit.AuditLogEventTypeArchived)
}

// RecordTransaction writes the ledger row, then records it.
func (r *repository) RecordTransaction(ctx context.Context, scope tenancy.Scope, transaction *billing.Transaction) (*billing.Transaction, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	recorded, err := r.Store.RecordTransaction(ctx, scope, transaction)
	if err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, recorded.ID)

	if err = r.recordTransaction(ctx, recorded, audit.AuditLogEventTypeCreated); err != nil {
		return nil, err
	}

	return recorded, nil
}

// SetTransactionStatus moves the attempt's outcome, then records it. A replay is
// billing.ErrStatusUnchanged and records nothing — see SetSubscriptionStatus.
func (r *repository) SetTransactionStatus(ctx context.Context, scope tenancy.Scope, transactionID string, status billing.TransactionStatus) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, transactionID)

	transaction, err := r.GetTransaction(ctx, scope, transactionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching transaction to record")
	}

	if err = r.Store.SetTransactionStatus(ctx, scope, transactionID, status); err != nil {
		return err
	}

	return r.recordTransaction(ctx, transaction, audit.AuditLogEventTypeUpdated)
}

// ArchiveTransaction retires the ledger row administratively, then records it.
func (r *repository) ArchiveTransaction(ctx context.Context, scope tenancy.Scope, transactionID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, paymentskeys.PaymentTransactionIDKey, transactionID)

	transaction, err := r.GetTransaction(ctx, scope, transactionID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching transaction to record")
	}

	if err = r.Store.ArchiveTransaction(ctx, scope, transactionID); err != nil {
		return err
	}

	return r.recordTransaction(ctx, transaction, audit.AuditLogEventTypeArchived)
}

// recordProduct writes the audit entry and the data change event for a write to
// the catalog.
//
// The entry names no account, because a product is a catalog row that belongs
// to nobody: who wrote it is the actor on the context, which is what the audit
// recorder resolves. The event names no account either, so it is service-wide —
// it reaches a webhook subscriber under whichever account the requester had
// active, resolved from the context by the emitter.
func (r *repository) recordProduct(ctx context.Context, product *billing.Product, auditEventType, changeEventType string) error {
	return r.recordAndEmit(ctx, "", resourceTypeProducts, product.ID, auditEventType, changeEventType, map[string]any{
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
func (r *repository) recordSubscription(ctx context.Context, subscription *billing.Subscription, auditEventType, changeEventType string) error {
	return r.recordAndEmit(ctx, subscription.BelongsToAccount, resourceTypeSubscriptions, subscription.ID, auditEventType, changeEventType, map[string]any{
		paymentskeys.SubscriptionIDKey: subscription.ID,
		paymentskeys.ProductIDKey:      subscription.ProductID,
		identitykeys.AccountIDKey:      subscription.BelongsToAccount,
	})
}

// recordPurchase writes the audit entry for a write to one account's purchase.
func (r *repository) recordPurchase(ctx context.Context, purchase *billing.Purchase, auditEventType string) error {
	return r.record(ctx, purchase.BelongsToAccount, resourceTypePurchases, purchase.ID, auditEventType)
}

// recordTransaction writes the audit entry for a write to one account's ledger.
func (r *repository) recordTransaction(ctx context.Context, transaction *billing.Transaction, auditEventType string) error {
	return r.record(ctx, transaction.BelongsToAccount, resourceTypePaymentTransactions, transaction.ID, auditEventType)
}

// recordAndEmit writes the audit entry and enqueues the data change event, in
// one transaction of their own.
//
// The two travel together because they answer the same question from opposite
// sides — the audit log for whoever asks later who did this, the outbox for
// whoever needs to know now — and a write that carried one without the other
// would be a write nobody could tell was incomplete.
func (r *repository) recordAndEmit(
	ctx context.Context,
	accountID, resourceType, relevantID, auditEventType, changeEventType string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.WithSpan(span).WithValue(resourceType, relevantID)

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		return r.recorder.RecordAndEmit(ctx, tx, logger, auditEntry(accountID, resourceType, relevantID, auditEventType), changeEventType, accountID, metadata)
	})
}

// record writes the audit entry alone, for the writes that owe an entry and no
// event. See the package documentation for which those are.
func (r *repository) record(ctx context.Context, accountID, resourceType, relevantID, auditEventType string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	return r.client.WithTransaction(ctx, func(tx database.Tx) error {
		if err := r.auditLogEntryRepo.Record(ctx, tx, auditEntry(accountID, resourceType, relevantID, auditEventType)); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		return nil
	})
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
