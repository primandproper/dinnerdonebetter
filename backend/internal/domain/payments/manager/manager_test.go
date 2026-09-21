package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v14/billing"
	billingmock "github.com/primandproper/platform-go/v14/billing/mock"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// billingUpdate is one write the manager made to the directory's billing surface,
// recorded so a test can say what the account's standing became.
//
// The plan is a pointer where the status is not, because the two answers differ: a
// subscription that ended names no plan, and an empty string would be a plan called "".
type billingUpdate struct {
	planID    *string
	accountID string
	status    platformidentity.BillingStatus
}

// buildPaymentsManagerForTest wires the manager over a billing store mock and an
// identity manager mock that records every billing update it is asked for.
func buildPaymentsManagerForTest(t *testing.T, store *billingmock.StoreMock) (*paymentsManager, *[]billingUpdate) {
	t.Helper()

	updates := &[]billingUpdate{}

	identityMgr := &identitymock.StoreMock{
		RecordAccountSubscriptionFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, accountID string, status platformidentity.BillingStatus, planID string) error {
			*updates = append(*updates, billingUpdate{accountID: accountID, status: status, planID: &planID})

			return nil
		},
		RecordAccountSubscriptionEndedFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, accountID string, status platformidentity.BillingStatus) error {
			*updates = append(*updates, billingUpdate{accountID: accountID, status: status})

			return nil
		},
	}

	m, err := NewPaymentsDataManager(
		t.Context(),
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		testutils.MockDatabaseClient(),
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
		GetSubscriptionByExternalIDFunc: func(_ context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, externalID string) (*billing.Subscription, error) {
			if scope != payments.Scope() || externalID != subscription.ExternalSubscriptionID {
				return nil, billing.ErrSubscriptionNotFound
			}

			return subscription, nil
		},
		SetSubscriptionStatusFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, subscriptionID string, status capitalism.SubscriptionStatus) error {
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
		assert.Equal(t, platformidentity.BillingTrial, (*updates)[0].status)
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
		assert.Equal(t, platformidentity.BillingPaid, (*updates)[0].status)
	})

	// The store reports a replayed event as ErrStatusUnchanged. That is the provider telling
	// us something we already knew, and the delivery has to be acknowledged rather than
	// retried forever — so it is not an error here, and the account's standing is re-derived
	// anyway, which is idempotent.
	T.Run("a redelivered status is acknowledged rather than failed", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, _ := subscriptionLookup(subscription)
		store.SetSubscriptionStatusFunc = func(context.Context, database.Tx, tenancy.Scope, string, capitalism.SubscriptionStatus) error {
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
		assert.Equal(t, platformidentity.BillingUnpaid, (*updates)[0].status)
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
			GetProductByExternalIDFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, externalID string) (*billing.Product, error) {
				if externalID != product.ExternalProductID {
					return nil, billing.ErrProductNotFound
				}

				return product, nil
			},
			GetSubscriptionByExternalIDFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*billing.Subscription, error) {
				return nil, billing.ErrSubscriptionNotFound
			},
			CreateSubscriptionFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
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
		assert.Equal(t, platformidentity.BillingPaid, (*updates)[0].status)
		assert.Equal(t, product.ID, *(*updates)[0].planID)
	})

	T.Run("a renewal of a known subscription reactivates it", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), product.ID)
		store, statuses := subscriptionLookup(subscription)
		store.GetProductByExternalIDFunc = func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*billing.Product, error) {
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
		assert.Equal(t, platformidentity.BillingUnpaid, (*updates)[0].status)
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
		assert.Equal(t, platformidentity.BillingUnpaid, (*updates)[0].status)
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

// An unrecognised standing leaves the account where it was, which is the one judgment this
// application makes about billing/standing's contract.
//
// The mapping itself is platform's now — standing.Strict, tested upstream value by value —
// and this application passes it rather than writing its own, which is a deployment saying
// "yes, that is our rule": no dunning window, no grace on past_due.
//
// What is pinned here is the case Strict reports it cannot place, because getting there at
// all took a change at the adapter boundary. capitalism's SubscriptionStatusUnknown is the
// empty string and its documentation covers both "a status no adapter recognized" and the
// zero value, so a word Stripe adds next year used to arrive looking exactly like an event
// carrying no standing — and that reading is "a sync of a subscription the provider still
// considers live", which made the account paid.
func TestPaymentsManager_UnrecognizedSubscriptionStatus(T *testing.T) {
	T.Parallel()

	T.Run("a standing no adapter could place leaves the account alone", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:          "customer.subscription.updated",
			SubscriptionID:     subscription.ExternalSubscriptionID,
			Status:             capitalism.SubscriptionStatusUnknown,
			StatusUnrecognized: true,
		}, "")
		require.NoError(t, err)

		// Nothing written on either side. The subscription's own status is not moved
		// either, because what the provider reported is not a value this schema holds.
		assert.Empty(t, *statuses)
		assert.Empty(t, *updates)
	})

	// And the case it must not be confused with: an event that genuinely carries no
	// standing is still read as a live subscription, which is what "updated" has always
	// meant here.
	T.Run("an event carrying no standing is still read as active", func(t *testing.T) {
		t.Parallel()

		subscription := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())
		store, statuses := subscriptionLookup(subscription)
		pm, updates := buildPaymentsManagerForTest(t, store)

		err := pm.ProcessWebhookEvent(t.Context(), "stripe", &payments.ParsedWebhookEvent{
			EventType:      "customer.subscription.updated",
			SubscriptionID: subscription.ExternalSubscriptionID,
			Status:         capitalism.SubscriptionStatusUnknown,
		}, "")
		require.NoError(t, err)

		assert.Equal(t, []capitalism.SubscriptionStatus{capitalism.SubscriptionStatusActive}, *statuses)
		require.Len(t, *updates, 1)
		assert.Equal(t, platformidentity.BillingPaid, (*updates)[0].status)
	})
}
