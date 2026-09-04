package manager

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v13/billing"
	billingmock "github.com/primandproper/platform-go/v13/billing/mock"
	"github.com/primandproper/platform-go/v13/capitalism"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/fake"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// billingUpdate is one call the manager made to the identity manager, recorded
// so a test can say what the account's standing became.
type billingUpdate struct {
	status    *string
	planID    *string
	accountID string
}

// buildPaymentsManagerForTest wires the manager over a billing store mock and an
// identity manager mock that records every billing update it is asked for.
func buildPaymentsManagerForTest(t *testing.T, store *billingmock.StoreMock) (*paymentsManager, *[]billingUpdate) {
	t.Helper()

	updates := &[]billingUpdate{}

	identityMgr := &identitymock.IdentityDataManagerMock{
		UpdateAccountBillingFieldsFunc: func(_ context.Context, accountID string, billingStatus, subscriptionPlanID, _ *string, _ *time.Time) error {
			*updates = append(*updates, billingUpdate{accountID: accountID, status: billingStatus, planID: subscriptionPlanID})

			return nil
		},
	}

	m, err := NewPaymentsDataManager(
		t.Context(),
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		store,
		identityMgr,
	)
	require.NoError(t, err)

	return m.(*paymentsManager), updates
}

// subscriptionLookup is a store that knows one subscription by its provider-side
// id and records the status writes made against it.
func subscriptionLookup(subscription *billing.Subscription) (*billingmock.StoreMock, *[]capitalism.SubscriptionStatus) {
	statuses := &[]capitalism.SubscriptionStatus{}

	return &billingmock.StoreMock{
		GetSubscriptionByExternalIDFunc: func(_ context.Context, scope tenancy.Scope, externalID string) (*billing.Subscription, error) {
			if scope != payments.Scope() || externalID != subscription.ExternalSubscriptionID {
				return nil, billing.ErrSubscriptionNotFound
			}

			return subscription, nil
		},
		SetSubscriptionStatusFunc: func(_ context.Context, _ tenancy.Scope, subscriptionID string, status capitalism.SubscriptionStatus) error {
			if subscriptionID != subscription.ID {
				return billing.ErrSubscriptionNotFound
			}

			*statuses = append(*statuses, status)

			return nil
		},
	}, statuses
}

func TestPaymentsManager_ProcessWebhookEvent(T *testing.T) {
	T.Parallel()

	T.Run("a subscription update writes the reported status and the account's standing", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "customer.subscription.updated",
			SubscriptionID: subscription.ExternalSubscriptionID,
			Status:         capitalism.SubscriptionStatusTrialing,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusTrialing}, *statuses)
		require.Len(t, *updates, 1)
		assert.Equal(t, subscription.BelongsToAccount, (*updates)[0].accountID)
		assert.Equal(t, identity.TrialAccountBillingStatus, *(*updates)[0].status)
		assert.Equal(t, subscription.ProductID, *(*updates)[0].planID)
	})

	T.Run("an update carrying no status is read as active", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "subscription.updated",
			SubscriptionID: subscription.ExternalSubscriptionID,
			Status:         capitalism.SubscriptionStatusUnknown,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusActive}, *statuses)
		require.Len(t, *updates, 1)
		assert.Equal(t, identity.PaidAccountBillingStatus, *(*updates)[0].status)
	})

	// The store reports a replayed event as ErrStatusUnchanged. That is the provider telling
	// us something we already knew, and the delivery has to be acknowledged rather than
	// retried forever — so it is not an error here, and the account's standing is re-derived
	// anyway, which is idempotent.
	T.Run("a redelivered status is acknowledged rather than failed", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, _ := subscriptionLookup(subscription)
		store.SetSubscriptionStatusFunc = func(context.Context, tenancy.Scope, string, capitalism.SubscriptionStatus) error {
			return billing.ErrStatusUnchanged
		}
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "customer.subscription.updated",
			SubscriptionID: subscription.ExternalSubscriptionID,
			Status:         capitalism.SubscriptionStatusActive,
		}, "")
		require.NoError(t, err)
		assert.Len(t, *updates, 1)
	})

	T.Run("a deletion cancels the subscription and marks the account unpaid", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "customer.subscription.deleted",
			SubscriptionID: subscription.ExternalSubscriptionID,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusCanceled}, *statuses)
		require.Len(t, *updates, 1)
		assert.Equal(t, identity.UnpaidAccountBillingStatus, *(*updates)[0].status)
		assert.Nil(t, (*updates)[0].planID)
	})

	T.Run("an update for a subscription nobody has is reported", func(t *testing.T) {
		t.Parallel()

		store, _ := subscriptionLookup(fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID()))
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "customer.subscription.updated",
			SubscriptionID: fake.BuildFakeID(),
			Status:         capitalism.SubscriptionStatusActive,
		}, "")
		require.ErrorIs(t, err, billing.ErrSubscriptionNotFound)
		assert.Empty(t, *updates)
	})

	T.Run("an initial purchase opens a subscription on the matching product", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		product := fakes.BuildFakeProduct()
		transactionID := fake.BuildFakeID()

		var created *billing.Subscription

		store := &billingmock.StoreMock{
			GetProductByExternalIDFunc: func(_ context.Context, _ tenancy.Scope, externalID string) (*billing.Product, error) {
				if externalID != product.ExternalProductID {
					return nil, billing.ErrProductNotFound
				}

				return product, nil
			},
			GetSubscriptionByExternalIDFunc: func(context.Context, tenancy.Scope, string) (*billing.Subscription, error) {
				return nil, billing.ErrSubscriptionNotFound
			},
			CreateSubscriptionFunc: func(_ context.Context, _ tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
				created = subscription

				return subscription, nil
			},
		}
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "INITIAL_PURCHASE",
			AccountID:      accountID,
			SubscriptionID: transactionID,
			ProductID:      product.ExternalProductID,
			Status:         capitalism.SubscriptionStatusActive,
		}, "")
		require.NoError(t, err)

		require.NotNil(t, created)
		assert.Equal(t, accountID, created.BelongsToAccount)
		assert.Equal(t, product.ID, created.ProductID)
		assert.Equal(t, transactionID, created.ExternalSubscriptionID)
		assert.Equal(t, capitalism.SubscriptionStatusActive, created.Status)
		assert.True(t, created.CurrentPeriodEnd.After(created.CurrentPeriodStart))

		require.Len(t, *updates, 1)
		assert.Equal(t, accountID, (*updates)[0].accountID)
		assert.Equal(t, identity.PaidAccountBillingStatus, *(*updates)[0].status)
		assert.Equal(t, product.ID, *(*updates)[0].planID)
	})

	T.Run("a renewal of a known subscription reactivates it", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), product.ID)
		store, statuses := subscriptionLookup(subscription)
		store.GetProductByExternalIDFunc = func(context.Context, tenancy.Scope, string) (*billing.Product, error) {
			return product, nil
		}
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "RENEWAL",
			AccountID:      subscription.BelongsToAccount,
			SubscriptionID: subscription.ExternalSubscriptionID,
			ProductID:      product.ExternalProductID,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusActive}, *statuses)
		assert.Len(t, *updates, 1)
	})

	T.Run("an expiration cancels the subscription and marks the account unpaid", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "EXPIRATION",
			AccountID:      subscription.BelongsToAccount,
			SubscriptionID: subscription.ExternalSubscriptionID,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusCanceled}, *statuses)
		require.Len(t, *updates, 1)
		assert.Equal(t, identity.UnpaidAccountBillingStatus, *(*updates)[0].status)
	})

	T.Run("an expiration of a subscription nobody has still marks the account unpaid", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		store, _ := subscriptionLookup(fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID()))
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "EXPIRATION",
			AccountID:      accountID,
			SubscriptionID: fake.BuildFakeID(),
		}, "")
		require.NoError(t, err)

		require.Len(t, *updates, 1)
		assert.Equal(t, accountID, (*updates)[0].accountID)
		assert.Equal(t, identity.UnpaidAccountBillingStatus, *(*updates)[0].status)
	})

	T.Run("a cancellation of a subscription nobody has yet is a no-op", func(t *testing.T) {
		t.Parallel()

		store, statuses := subscriptionLookup(fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID()))
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "CANCELLATION",
			AccountID:      fake.BuildFakeID(),
			SubscriptionID: fake.BuildFakeID(),
		}, "")
		require.NoError(t, err)
		assert.Empty(t, *statuses)
		assert.Empty(t, *updates)
	})

	T.Run("a cancellation marks the subscription canceled and leaves the account alone", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "revenuecat", &payments.ParsedWebhookEvent{
			EventType:      "CANCELLATION",
			AccountID:      subscription.BelongsToAccount,
			SubscriptionID: subscription.ExternalSubscriptionID,
		}, "")
		require.NoError(t, err)

		// Access persists until EXPIRATION, so the standing is not touched here.
		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusCanceled}, *statuses)
		assert.Empty(t, *updates)
	})

	T.Run("an unrecognized event is a no-op", func(t *testing.T) {
		t.Parallel()

		pm, updates := buildPaymentsManagerForTest(t, &billingmock.StoreMock{})

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{EventType: "something.new"}, "")
		require.NoError(t, err)
		assert.Empty(t, *updates)
	})

	T.Run("with nil event", func(t *testing.T) {
		t.Parallel()

		pm, _ := buildPaymentsManagerForTest(t, &billingmock.StoreMock{})

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", nil, "")
		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

// The mapping onto the account's coarse standing is the one judgment platform
// says a consumer still writes, so it is pinned value by value.
func TestSubscriptionStatusToBillingStatus(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := map[capitalism.SubscriptionStatus]string{
			capitalism.SubscriptionStatusActive:            identity.PaidAccountBillingStatus,
			capitalism.SubscriptionStatusTrialing:          identity.TrialAccountBillingStatus,
			capitalism.SubscriptionStatusPastDue:           identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusCanceled:          identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusIncomplete:        identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusIncompleteExpired: identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusUnpaid:            identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusPaused:            identity.UnpaidAccountBillingStatus,
			capitalism.SubscriptionStatusUnknown:           identity.UnpaidAccountBillingStatus,
		}

		for status, want := range expected {
			assert.Equal(t, want, subscriptionStatusToBillingStatus(status), "mapping %s", status.String())
		}
	})
}
