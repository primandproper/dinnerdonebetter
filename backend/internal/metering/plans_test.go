package metering

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// subscriptionReaderStub records what it was asked and answers with what it was given.
type subscriptionReaderStub struct {
	err           error
	result        *filtering.QueryFilteredResult[payments.Subscription]
	requestedFor  []string
	callCount     int
	returnNilBoth bool
}

func (s *subscriptionReaderStub) GetSubscriptionsForAccount(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
	s.callCount++
	s.requestedFor = append(s.requestedFor, accountID)

	if s.err != nil {
		return nil, s.err
	}

	if s.returnNilBoth {
		return nil, nil
	}

	return s.result, nil
}

// resultOf wraps subscriptions in the shape the repository returns them in.
func resultOf(subscriptions ...*payments.Subscription) *filtering.QueryFilteredResult[payments.Subscription] {
	return &filtering.QueryFilteredResult[payments.Subscription]{Data: subscriptions}
}

// subscriptionWithStatus is a fake subscription forced to one status and product.
func subscriptionWithStatus(status, productID string) *payments.Subscription {
	subscription := paymentsfakes.BuildFakeSubscription(identifiers.New(), productID)
	subscription.Status = status

	return subscription
}

// The ladder these limits feed — unlimited for an unlisted meter, the product's limit for an
// entitled subject, Unsubscribed for everyone else — is platform's and is tested there. What
// is tested here is the only part that is a product decision: which statuses entitle, and which
// row wins when an account holds several.
func TestEntitlementReader(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		accountID := identifiers.New()
		productID := identifiers.New()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusActive, productID)),
		}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, accountID)
		require.NoError(t, err)
		assert.True(t, entitled)
		assert.Equal(t, productID, gotProduct)
		assert.Equal(t, []string{accountID}, subscriptions.requestedFor)
	})

	T.Run("a trialing subscription entitles the same as an active one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		productID := identifiers.New()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusTrialing, productID)),
		}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.NoError(t, err)
		assert.True(t, entitled)
		assert.Equal(t, productID, gotProduct)
	})

	T.Run("a past due subscription does not entitle", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusPastDue, identifiers.New())),
		}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.NoError(t, err)
		assert.False(t, entitled)
		assert.Empty(t, gotProduct)
	})

	T.Run("a live subscription is found past a cancelled one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		liveProductID := identifiers.New()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(
				subscriptionWithStatus(payments.SubscriptionStatusCancelled, identifiers.New()),
				subscriptionWithStatus(payments.SubscriptionStatusActive, liveProductID),
			),
		}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.NoError(t, err)
		assert.True(t, entitled)
		assert.Equal(t, liveProductID, gotProduct)
	})

	T.Run("an account with no subscriptions at all", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		subscriptions := &subscriptionReaderStub{result: resultOf()}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.NoError(t, err)
		assert.False(t, entitled)
		assert.Empty(t, gotProduct)
	})

	T.Run("with a nil result and no error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		subscriptions := &subscriptionReaderStub{returnNilBoth: true}

		gotProduct, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.NoError(t, err)
		assert.False(t, entitled)
		assert.Empty(t, gotProduct)
	})

	T.Run("with error reading subscriptions", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")
		subscriptions := &subscriptionReaderStub{err: expected}

		_, entitled, err := entitlementReader(subscriptions).EntitlingProduct(ctx, identifiers.New())
		require.ErrorIs(t, err, expected)
		assert.False(t, entitled)
	})
}

func TestNewSubscriptionQuotaSource(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		registry, err := NewRegistry()
		require.NoError(t, err)

		source, err := NewSubscriptionQuotaSource(registry, &subscriptionReaderStub{})
		require.NoError(t, err)
		assert.NotNil(t, source)
	})

	T.Run("an ungated meter is answered without reading the subscription", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		registry, err := NewRegistry()
		require.NoError(t, err)

		subscriptions := &subscriptionReaderStub{}

		source, err := NewSubscriptionQuotaSource(registry, subscriptions)
		require.NoError(t, err)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)
		assert.Equal(t, unlimited, quota.Limit)
		assert.Zero(t, subscriptions.callCount)
	})
}

// The shipped table is empty on purpose; see planLimits for why. This is here so that adding
// the first limit is a deliberate act with a test to change, rather than something that lands
// unnoticed.
func TestPlanLimits_ShippedTableIsEmpty(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, planLimits)
	})
}
