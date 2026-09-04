package manager

import (
	"context"
	"errors"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/keys"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "payments_data_manager"
)

var _ PaymentsDataManager = (*paymentsManager)(nil)

type paymentsManager struct {
	tracer      tracing.Tracer
	logger      logging.Logger
	store       billing.Store
	identityMgr identitymanager.IdentityDataManager
}

// NewPaymentsDataManager returns a new PaymentsDataManager.
//
// Audit entries and data change events are recorded by the repository around
// platform's store; see internal/repositories/postgres/payments.
func NewPaymentsDataManager(
	_ context.Context,
	tracerProvider tracing.Provider,
	logger logging.Logger,
	store billing.Store,
	identityMgr identitymanager.IdentityDataManager,
) (PaymentsDataManager, error) {
	return &paymentsManager{
		tracer:      tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:      logging.NewNamedLogger(logger, o11yName),
		store:       store,
		identityMgr: identityMgr,
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
	syncNow := time.Now()

	switch eventType {
	case "subscription.updated", "subscription.created", "customer.subscription.updated":
		if subscriptionID == "" {
			return nil
		}

		sub, err := m.store.GetSubscriptionByExternalID(ctx, payments.Scope(), subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		// An event that carries no standing is a sync of a subscription the
		// provider still considers live, which is what "updated" has always been
		// read as here.
		status := parsed.Status
		if !status.Known() {
			status = capitalism.SubscriptionStatusActive
		}

		if err = m.setSubscriptionStatus(ctx, sub.ID, status); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}

		billingStatus := subscriptionStatusToBillingStatus(status)
		if err = m.identityMgr.UpdateAccountBillingFields(ctx, sub.BelongsToAccount, &billingStatus, &sub.ProductID, nil, &syncNow); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating account billing fields")
		}
	case "subscription.deleted", "customer.subscription.deleted":
		if subscriptionID == "" {
			return nil
		}

		sub, err := m.store.GetSubscriptionByExternalID(ctx, payments.Scope(), subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusCanceled); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}

		unpaid := identity.UnpaidAccountBillingStatus
		if err = m.identityMgr.UpdateAccountBillingFields(ctx, sub.BelongsToAccount, &unpaid, nil, nil, &syncNow); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating account billing fields")
		}

	// RevenueCat events (mobile in-app purchases)
	case "INITIAL_PURCHASE", "RENEWAL", "PRODUCT_CHANGE", "UNCANCELLATION", "SUBSCRIPTION_EXTENDED":
		if accountID == "" || parsed.ProductID == "" {
			return nil
		}

		return m.handleRevenueCatSubscriptionActive(ctx, logger, span, accountID, subscriptionID, parsed.ProductID, syncNow)
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

		sub, err := m.store.GetSubscriptionByExternalID(ctx, payments.Scope(), subscriptionID)
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
	err := m.store.SetSubscriptionStatus(ctx, payments.Scope(), subscriptionID, status)
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
	syncNow time.Time,
) error {
	product, err := m.store.GetProductByExternalID(ctx, payments.Scope(), externalProductID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching product by external ID")
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, product.ID)

	sub, err := m.store.GetSubscriptionByExternalID(ctx, payments.Scope(), transactionID)
	if err != nil {
		if !errors.Is(err, billing.ErrSubscriptionNotFound) {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		// Create new subscription for INITIAL_PURCHASE
		now := time.Now()
		if _, err = m.store.CreateSubscription(ctx, payments.Scope(), &billing.Subscription{
			BelongsToAccount:       accountID,
			ProductID:              product.ID,
			ExternalSubscriptionID: transactionID,
			Status:                 capitalism.SubscriptionStatusActive,
			CurrentPeriodStart:     now,
			CurrentPeriodEnd:       now.AddDate(0, 1, 0), // approximate
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating subscription")
		}
	} else if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusActive); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	billingStatus := identity.PaidAccountBillingStatus
	productID := product.ID

	return observability.PrepareAndLogError(
		m.identityMgr.UpdateAccountBillingFields(ctx, accountID, &billingStatus, &productID, nil, &syncNow),
		logger, span, "updating account billing fields",
	)
}

func (m *paymentsManager) handleRevenueCatSubscriptionExpired(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	accountID, transactionID string,
) error {
	unpaid := identity.UnpaidAccountBillingStatus
	syncNow := time.Now()

	sub, err := m.store.GetSubscriptionByExternalID(ctx, payments.Scope(), transactionID)
	if err != nil {
		// Subscription may not exist; still update account to unpaid
		return observability.PrepareAndLogError(
			m.identityMgr.UpdateAccountBillingFields(ctx, accountID, &unpaid, nil, nil, &syncNow),
			logger, span, "updating account billing fields",
		)
	}

	if err = m.setSubscriptionStatus(ctx, sub.ID, capitalism.SubscriptionStatusCanceled); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	return observability.PrepareAndLogError(
		m.identityMgr.UpdateAccountBillingFields(ctx, sub.BelongsToAccount, &unpaid, nil, nil, &syncNow),
		logger, span, "updating account billing fields",
	)
}

// subscriptionStatusToBillingStatus is the mapping onto identity.Account's
// coarse standing — the one platform says a consumer still writes, because it
// includes a suspension no processor reports.
//
// Trialing is the one status that is not paid and not unpaid. Everything else
// that is not active — past due, unpaid, paused, canceled, incomplete in either
// form — is unpaid, because nothing is being collected. A status this module does
// not know is unpaid too, rather than active: a word the provider added last week
// should not entitle an account on its own.
func subscriptionStatusToBillingStatus(status capitalism.SubscriptionStatus) string {
	switch status {
	case capitalism.SubscriptionStatusActive:
		return identity.PaidAccountBillingStatus
	case capitalism.SubscriptionStatusTrialing:
		return identity.TrialAccountBillingStatus
	default:
		return identity.UnpaidAccountBillingStatus
	}
}
