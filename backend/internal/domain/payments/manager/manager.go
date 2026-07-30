package manager

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	paymentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/keys"

	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"
	"github.com/primandproper/platform-go/v8/observability"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
)

const (
	o11yName = "payments_data_manager"
)

var (
	// ErrUnknownPaymentProvider indicates a webhook arrived for a provider that has no registered processor.
	ErrUnknownPaymentProvider = errors.New("unknown payment provider")
	// ErrInvalidWebhookSignature indicates a webhook payload failed signature verification.
	ErrInvalidWebhookSignature = errors.New("invalid webhook signature")
)

var _ PaymentsDataManager = (*paymentsManager)(nil)

type paymentsManager struct {
	tracer            tracing.Tracer
	logger            logging.Logger
	repo              payments.Repository
	processorRegistry payments.PaymentProcessorRegistry
	identityMgr       identitymanager.IdentityDataManager
}

// NewPaymentsDataManager returns a new PaymentsDataManager.
func NewPaymentsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo payments.Repository,
	processorRegistry payments.PaymentProcessorRegistry,
	identityMgr identitymanager.IdentityDataManager,
) (PaymentsDataManager, error) {
	return &paymentsManager{
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		repo:              repo,
		processorRegistry: processorRegistry,
		identityMgr:       identityMgr,
	}, nil
}

func (m *paymentsManager) CreateProduct(ctx context.Context, input *payments.ProductCreationRequestInput) (*payments.Product, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating product creation input")
	}

	dbInput := &payments.ProductDatabaseCreationInput{
		ID:                    identifiers.New(),
		Name:                  input.Name,
		Description:           input.Description,
		Kind:                  input.Kind,
		AmountCents:           input.AmountCents,
		Currency:              input.Currency,
		BillingIntervalMonths: input.BillingIntervalMonths,
		ExternalProductID:     input.ExternalProductID,
	}
	created, err := m.repo.CreateProduct(ctx, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "create product")
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, created.ID)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *paymentsManager) GetProduct(ctx context.Context, id string) (*payments.Product, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetProduct(ctx, id)
}

func (m *paymentsManager) GetProducts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Product], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetProducts(ctx, filter)
}

func (m *paymentsManager) UpdateProduct(ctx context.Context, id string, input *payments.ProductUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "validating product update input")
	}

	product, err := m.repo.GetProduct(ctx, id)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching product")
	}

	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.Kind != nil {
		product.Kind = *input.Kind
	}
	if input.AmountCents != nil {
		product.AmountCents = *input.AmountCents
	}
	if input.Currency != nil {
		product.Currency = *input.Currency
	}
	if input.BillingIntervalMonths != nil {
		product.BillingIntervalMonths = input.BillingIntervalMonths
	}
	if input.ExternalProductID != nil {
		product.ExternalProductID = *input.ExternalProductID
	}

	if err = m.repo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, id)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *paymentsManager) ArchiveProduct(ctx context.Context, id string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if err := m.repo.ArchiveProduct(ctx, id); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive product")
	}

	tracing.AttachToSpan(span, paymentskeys.ProductIDKey, id)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *paymentsManager) CreateSubscription(ctx context.Context, input *payments.SubscriptionCreationRequestInput) (*payments.Subscription, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating subscription creation input")
	}

	dbInput := &payments.SubscriptionDatabaseCreationInput{
		ID:                     identifiers.New(),
		BelongsToAccount:       input.BelongsToAccount,
		ProductID:              input.ProductID,
		ExternalSubscriptionID: input.ExternalSubscriptionID,
		Status:                 input.Status,
		CurrentPeriodStart:     input.CurrentPeriodStart,
		CurrentPeriodEnd:       input.CurrentPeriodEnd,
	}
	created, err := m.repo.CreateSubscription(ctx, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "create subscription")
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, created.ID)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return created, nil
}

func (m *paymentsManager) GetSubscription(ctx context.Context, id string) (*payments.Subscription, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetSubscription(ctx, id)
}

func (m *paymentsManager) GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetSubscriptionsForAccount(ctx, accountID, filter)
}

func (m *paymentsManager) UpdateSubscription(ctx context.Context, id string, input *payments.SubscriptionUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "validating subscription update input")
	}

	sub, err := m.repo.GetSubscription(ctx, id)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching subscription")
	}

	if input.Status != nil {
		sub.Status = *input.Status
	}
	if input.CurrentPeriodStart != nil {
		sub.CurrentPeriodStart = *input.CurrentPeriodStart
	}
	if input.CurrentPeriodEnd != nil {
		sub.CurrentPeriodEnd = *input.CurrentPeriodEnd
	}

	if err = m.repo.UpdateSubscription(ctx, sub); err != nil {
		return err
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, id)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *paymentsManager) ArchiveSubscription(ctx context.Context, id string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if err := m.repo.ArchiveSubscription(ctx, id); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive subscription")
	}

	tracing.AttachToSpan(span, paymentskeys.SubscriptionIDKey, id)
	// The event is enqueued into the outbox by the repository, inside the same transaction
	// as the write it describes; see internal/repositories/postgres/events.

	return nil
}

func (m *paymentsManager) GetPurchasesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Purchase], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetPurchasesForAccount(ctx, accountID, filter)
}

func (m *paymentsManager) GetPaymentTransactionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.PaymentTransaction], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetPaymentTransactionsForAccount(ctx, accountID, filter)
}

func (m *paymentsManager) ProcessWebhookEvent(ctx context.Context, provider string, payload []byte, signature, accountID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := observability.ObserveValues(map[string]any{
		identitykeys.AccountIDKey: accountID,
		"provider":                provider,
	}, span, m.logger)

	processor, ok := m.processorRegistry.GetProcessor(provider)
	if !ok {
		return observability.PrepareAndLogError(ErrUnknownPaymentProvider, logger, span, "unknown payment provider: %s", provider)
	}

	if !processor.VerifyWebhookSignature(ctx, payload, signature, accountID) {
		return observability.PrepareAndLogError(ErrInvalidWebhookSignature, logger, span, "invalid webhook signature")
	}

	parsed, err := processor.ParseWebhookEvent(ctx, payload)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "parsing webhook event")
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
		var sub *payments.Subscription
		sub, err = m.repo.GetSubscriptionByExternalID(ctx, subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		status := parsed.Status
		if status == "" {
			status = payments.SubscriptionStatusActive
		}

		if err = m.repo.UpdateSubscriptionStatus(ctx, sub.ID, status); err != nil {
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

		var sub *payments.Subscription
		sub, err = m.repo.GetSubscriptionByExternalID(ctx, subscriptionID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		if err = m.repo.UpdateSubscriptionStatus(ctx, sub.ID, payments.SubscriptionStatusCancelled); err != nil {
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
		if err = m.handleRevenueCatSubscriptionActive(ctx, logger, span, accountID, subscriptionID, parsed.ProductID, syncNow); err != nil {
			return err
		}
	case "EXPIRATION":
		if accountID == "" {
			return nil
		}
		if err = m.handleRevenueCatSubscriptionExpired(ctx, logger, span, accountID, subscriptionID); err != nil {
			return err
		}
	case "CANCELLATION":
		// User cancelled; access may persist until EXPIRATION. Optionally mark subscription cancelled.
		if accountID == "" || subscriptionID == "" {
			return nil
		}
		sub, subErr := m.repo.GetSubscriptionByExternalID(ctx, subscriptionID)
		if subErr != nil {
			if errors.Is(subErr, sql.ErrNoRows) {
				return nil // subscription may not exist yet
			}
			return observability.PrepareAndLogError(subErr, logger, span, "fetching subscription by external ID")
		}
		if err = m.repo.UpdateSubscriptionStatus(ctx, sub.ID, payments.SubscriptionStatusCancelled); err != nil {
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

func (m *paymentsManager) handleRevenueCatSubscriptionActive(
	ctx context.Context,
	logger logging.Logger,
	span tracing.Span,
	accountID, transactionID, externalProductID string,
	syncNow time.Time,
) error {
	product, err := m.repo.GetProductByExternalID(ctx, externalProductID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching product by external ID")
	}

	sub, err := m.repo.GetSubscriptionByExternalID(ctx, transactionID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return observability.PrepareAndLogError(err, logger, span, "fetching subscription by external ID")
		}

		// Create new subscription for INITIAL_PURCHASE
		now := time.Now()
		dbInput := &payments.SubscriptionDatabaseCreationInput{
			ID:                     identifiers.New(),
			BelongsToAccount:       accountID,
			ProductID:              product.ID,
			ExternalSubscriptionID: transactionID,
			Status:                 payments.SubscriptionStatusActive,
			CurrentPeriodStart:     now,
			CurrentPeriodEnd:       now.AddDate(0, 1, 0), // approximate
		}
		_, err = m.repo.CreateSubscription(ctx, dbInput)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating subscription")
		}
	} else {
		if err = m.repo.UpdateSubscriptionStatus(ctx, sub.ID, payments.SubscriptionStatusActive); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
		}
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
	sub, err := m.repo.GetSubscriptionByExternalID(ctx, transactionID)
	if err != nil {
		// Subscription may not exist; still update account to unpaid
		unpaid := identity.UnpaidAccountBillingStatus
		syncNow := time.Now()
		return observability.PrepareAndLogError(
			m.identityMgr.UpdateAccountBillingFields(ctx, accountID, &unpaid, nil, nil, &syncNow),
			logger, span, "updating account billing fields",
		)
	}

	if err = m.repo.UpdateSubscriptionStatus(ctx, sub.ID, payments.SubscriptionStatusCancelled); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating subscription status")
	}

	unpaid := identity.UnpaidAccountBillingStatus
	syncNow := time.Now()
	return observability.PrepareAndLogError(
		m.identityMgr.UpdateAccountBillingFields(ctx, sub.BelongsToAccount, &unpaid, nil, nil, &syncNow),
		logger, span, "updating account billing fields",
	)
}

func subscriptionStatusToBillingStatus(status string) string {
	switch status {
	case payments.SubscriptionStatusActive:
		return identity.PaidAccountBillingStatus
	case payments.SubscriptionStatusTrialing:
		return identity.TrialAccountBillingStatus
	case payments.SubscriptionStatusCancelled, payments.SubscriptionStatusPastDue, payments.SubscriptionStatusIncomplete:
		return identity.UnpaidAccountBillingStatus
	default:
		return identity.UnpaidAccountBillingStatus
	}
}
