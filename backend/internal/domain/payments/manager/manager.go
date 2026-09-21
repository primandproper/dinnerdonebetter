package manager

import (
	"context"
	"errors"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/platform-go/v14/billing/standing"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "payments_data_manager"
)

var _ PaymentsDataManager = (*paymentsManager)(nil)

type paymentsManager struct {
	tracer  tracing.Tracer
	logger  logging.Logger
	db      database.Client
	store   billing.Store
	billing platformidentity.BillingWriter
}

// NewPaymentsDataManager returns a new PaymentsDataManager.
//
// Audit entries and data change events are recorded by the repository around
// platform's store; see internal/repositories/postgres/payments.
func NewPaymentsDataManager(
	_ context.Context,
	tracerProvider tracing.Provider,
	logger logging.Logger,
	db database.Client,
	store billing.Store,
	billingWriter platformidentity.BillingWriter,
) (PaymentsDataManager, error) {
	return &paymentsManager{
		tracer:  tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:  logging.NewNamedLogger(logger, o11yName),
		db:      db,
		store:   store,
		billing: billingWriter,
	}, nil
}

// ProcessWebhookEvent applies a provider's event to the subscription it names
// and to the account's billing standing.
//
// Every status write here goes through the store's guarded SetSubscriptionStatus,
// which reports a redelivered event as billing.ErrStatusUnchanged. That is an
// answer rather than a failure — the provider already told us this — so it is
// swallowed and the account's standing is re-derived anyway, which is idempotent.
// Any other error is reported, so that the provider retries the delivery.
func (m *paymentsManager) ProcessWebhookEvent(ctx context.Context, provider string, parsed *payments.ParsedWebhookEvent, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		"provider":                provider,
	}, span, m.logger)

	if parsed == nil {
		return platformerrors.ErrNilInputParameter
	}

	// Use account ID from event payload when not provided in URL (e.g. RevenueCat app_user_id).
	if accountID == "" && parsed.AccountID != "" {
		accountID = parsed.AccountID
	}

	eventType := parsed.EventType
	subscriptionID := parsed.SubscriptionID

	switch eventType {
	case "subscription.updated", "subscription.created", "customer.subscription.updated":
		if subscriptionID == "" {
			return nil
		}

		sub, err := m.store.GetSubscriptionByExternalID(ctx, m.db.Reader(), payments.Scope(), subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		// An event that carries no standing at all is a sync of a subscription the
		// provider still considers live, which is what "updated" has always been read
		// as here. An event carrying a word this module does not recognize is not that,
		// and the two used to be the same branch: !Known() is true for both, so a status
		// Stripe adds next year was rewritten to active and the account came out paid.
		//
		// The comment on the mapping this replaced claimed the opposite — "a word the
		// provider added last week should not entitle an account on its own" — and its
		// unpaid default was unreachable, because the coercion above had already run.
		if parsed.StatusUnrecognized {
			logger.WithValue(paymentskeys.SubscriptionStatusKey, parsed.Status).
				Info("provider reported a standing no adapter could place; leaving the account alone")

			return nil
		}

		status := parsed.Status
		if status == "" {
			status = capitalism.SubscriptionStatusActive
		}

		if err = m.setSubscriptionStatus(ctx, sub.ID, status); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}

		// standing.Strict is platform's reading and this application's rule: active is
		// paid, trialing is a trial, the other six leave the account unpaid. It is passed
		// rather than defaulted so that taking it is this deployment saying "yes, that is
		// our rule" — no dunning window, no grace on past_due.
		//
		// Reporting false means the status is not one it can place, and platform's
		// contract for that is to leave the account alone rather than to pick a standing.
		// That is the fix for the defect above: an unrecognized word now changes nothing
		// about what the account may do, and says so where somebody will see it, instead
		// of silently entitling or silently demoting.
		billingStatus, placed := standing.Strict(status)
		if !placed {
			logger.WithValue(paymentskeys.SubscriptionStatusKey, status).
				Info("subscription status could not be placed; leaving the account's standing alone")
			return nil
		}

		if err = m.recordSubscription(ctx, sub.BelongsToAccount, billingStatus, sub.ProductID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "recording account subscription")
		}
	case "subscription.deleted", "customer.subscription.deleted":
		if subscriptionID == "" {
			return nil
		}

		sub, err := m.store.GetSubscriptionByExternalID(ctx, m.db.Reader(), payments.Scope(), subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusCanceled); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}

		if err = m.recordSubscriptionEnded(ctx, sub.BelongsToAccount); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "recording ended account subscription")
		}

	// RevenueCat events (mobile in-app purchases)
	case "INITIAL_PURCHASE", "RENEWAL", "PRODUCT_CHANGE", "UNCANCELLATION", "SUBSCRIPTION_EXTENDED":
		if accountID == "" || parsed.ProductID == "" {
			return nil
		}

		return m.handleRevenueCatSubscriptionActive(ctx, logger, span, accountID, subscriptionID, parsed.ProductID)
	case "EXPIRATION":
		if accountID == "" {
			return nil
		}

		return m.handleRevenueCatSubscriptionExpired(ctx, logger, span, accountID, subscriptionID)
	case "CANCELLATION":
		// User cancelled; access may persist until EXPIRATION. Optionally mark subscription cancelled.
		if accountID == "" || subscriptionID == "" {
			return nil
		}

		sub, err := m.store.GetSubscriptionByExternalID(ctx, m.db.Reader(), payments.Scope(), subscriptionID)
		if err != nil {
			if errors.Is(err, billing.ErrSubscriptionNotFound) {
				return nil // subscription may not exist yet
			}

			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusCanceled); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}
	case "BILLING_ISSUE":
		// Log; optionally treat as at-risk. No-op for now.
		logger.WithValue("account_id", accountID).Info("RevenueCat billing issue received")
	default:
		// Unknown event type - no-op
	}

	return nil
}

// setSubscriptionStatus moves the subscription's standing, treating the store's
// "already there" as success.
func (m *paymentsManager) setSubscriptionStatus(ctx context.Context, subscriptionID string, status capitalism.SubscriptionStatus) error {
	// One transaction per status write, which is exactly what the store opened
	// for itself before v14. Widening it to span the identity manager's writes
	// as well is now possible and is a behavior change rather than a port; see
	// the note on ProcessWebhookEvent.
	err := m.db.WithTransaction(ctx, func(tx database.Tx) error {
		return m.store.SetSubscriptionStatus(ctx, tx, payments.Scope(), subscriptionID, status)
	})
	if err != nil && !errors.Is(err, billing.ErrStatusUnchanged) {
		return err
	}

	return nil
}

func (m *paymentsManager) handleRevenueCatSubscriptionActive(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	accountID, transactionID, externalProductID string,
) error {
	product, err := m.store.GetProductByExternalID(ctx, m.db.Reader(), payments.Scope(), externalProductID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching product by external ID")
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, product.ID)

	sub, err := m.store.GetSubscriptionByExternalID(ctx, m.db.Reader(), payments.Scope(), transactionID)
	if err != nil {
		if !errors.Is(err, billing.ErrSubscriptionNotFound) {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		// Create new subscription for INITIAL_PURCHASE
		now := time.Now()
		if err = m.db.WithTransaction(ctx, func(tx database.Tx) error {
			_, createErr := m.store.CreateSubscription(ctx, tx, payments.Scope(), &billing.Subscription{
				BelongsToAccount:       accountID,
				ProductID:              product.ID,
				ExternalSubscriptionID: transactionID,
				Status:                 capitalism.SubscriptionStatusActive,
				CurrentPeriodStart:     now,
				CurrentPeriodEnd:       now.AddDate(0, 1, 0), // approximate
			})

			return createErr
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating subscription")
		}
	} else if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusActive); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	billingStatus := platformidentity.BillingPaid
	productID := product.ID

	return observability.PrepareAndLogError(
		m.recordSubscription(ctx, accountID, billingStatus, productID),
		logger, span, "recording account subscription",
	)
}

func (m *paymentsManager) handleRevenueCatSubscriptionExpired(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	accountID, transactionID string,
) error {
	sub, err := m.store.GetSubscriptionByExternalID(ctx, m.db.Reader(), payments.Scope(), transactionID)
	if err != nil {
		// Subscription may not exist; still update account to unpaid
		return observability.PrepareAndLogError(
			m.recordSubscriptionEnded(ctx, accountID),
			logger, span, "recording ended account subscription",
		)
	}

	if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusCanceled); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	return observability.PrepareAndLogError(
		m.recordSubscriptionEnded(ctx, sub.BelongsToAccount),
		logger, span, "recording ended account subscription",
	)
}

// recordSubscription writes what a processor delivery reported about an account: its
// standing, the plan it is on, and that this process just heard from the processor.
//
// The three move together because a delivery reports them together, which is the reason
// platform made this one method rather than three columns a caller sets independently —
// writing them one at a time leaves an account paid on last month's plan between two of
// the writes, and leaves it there for good if the second fails.
//
// The sync stamp is no longer passed in. It is the store's clock, because the fact being
// recorded is that this process heard from the processor, and the moment that happened is
// the moment of the write rather than a time.Now() the caller took earlier.
func (m *paymentsManager) recordSubscription(ctx context.Context, accountID string, status platformidentity.BillingStatus, planID string) error {
	return m.db.WithTransaction(ctx, func(tx database.Tx) error {
		return m.billing.RecordAccountSubscription(ctx, tx, identity.Scope(), accountID, status, planID)
	})
}

// recordSubscriptionEnded writes that an account's subscription is over.
//
// A separate method from the one above rather than that one with an empty plan, and the
// distinction is platform's: the difference between them is a cancellation, so a handler
// passing an unchecked payload through one call would otherwise cancel a subscription
// while believing it had renewed one.
func (m *paymentsManager) recordSubscriptionEnded(ctx context.Context, accountID string) error {
	return m.db.WithTransaction(ctx, func(tx database.Tx) error {
		// Unpaid, which is the only ending this application distinguishes. platform takes
		// the standing because ending is not one status — a customer cancelling, a
		// processor giving up on collection and a trial running out are the same write
		// over different standings — and which of them means what is this application's
		// policy. Here they all mean the account stops being paid for.
		return m.billing.RecordAccountSubscriptionEnded(ctx, tx, identity.Scope(), accountID,
			platformidentity.BillingUnpaid)
	})
}
