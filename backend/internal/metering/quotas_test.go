package metering

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	platformmetering "github.com/primandproper/platform-go/v9/metering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"

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

func TestSubscriptionQuotaSource_QuotaFor(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		subscriptions := &subscriptionReaderStub{}
		source := NewSubscriptionQuotaSource(loggingnoop.NewLogger(), subscriptions)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, UploadedMediaBytesMeter, quota.Meter)
		assert.Equal(t, platformmetering.BehaviorAllowOverage, quota.Behavior)
		assert.Equal(t, platformmetering.PeriodMonth, quota.Period)
		assert.Equal(t, unlimited, quota.Limit)
	})

	T.Run("an ungated meter is answered without reading the subscription", func(t *testing.T) {
		t.Parallel()

		// The short circuit is the point: QuotaFor sits on Enforcer.Check's path, and a
		// database round trip per check to reach an answer that is the same for every
		// account would make the cheap path the expensive one.
		ctx := t.Context()
		subscriptions := &subscriptionReaderStub{}
		source := NewSubscriptionQuotaSource(loggingnoop.NewLogger(), subscriptions)

		_, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Zero(t, subscriptions.callCount)
	})

	T.Run("a gated meter takes the limit of the account's active subscription", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		accountID, productID := identifiers.New(), identifiers.New()

		subscription := paymentsfakes.BuildFakeSubscription(accountID, productID)
		subscription.Status = payments.SubscriptionStatusActive

		subscriptions := &subscriptionReaderStub{result: resultOf(subscription)}
		source := newSubscriptionQuotaSource(loggingnoop.NewLogger(), subscriptions, map[string]PlanLimits{
			UploadedMediaBytesMeter: {
				Behavior:     platformmetering.BehaviorBlock,
				ByProduct:    map[string]int64{productID: 5_000},
				Unsubscribed: 100,
			},
		})

		quota, err := source.QuotaFor(ctx, accountID, UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(5_000), quota.Limit)
		assert.Equal(t, platformmetering.BehaviorBlock, quota.Behavior)
		assert.Equal(t, []string{accountID}, subscriptions.requestedFor)
	})

	T.Run("a trialing subscription entitles the same as an active one", func(t *testing.T) {
		t.Parallel()

		// A trial that quietly got the unsubscribed limits would be a trial of a different
		// product than the one somebody signed up for.
		ctx := t.Context()
		accountID, productID := identifiers.New(), identifiers.New()

		subscription := paymentsfakes.BuildFakeSubscription(accountID, productID)
		subscription.Status = payments.SubscriptionStatusTrialing

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{result: resultOf(subscription)},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {
					Behavior:     platformmetering.BehaviorBlock,
					ByProduct:    map[string]int64{productID: 5_000},
					Unsubscribed: 100,
				},
			},
		)

		quota, err := source.QuotaFor(ctx, accountID, UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(5_000), quota.Limit)
	})

	T.Run("a past due subscription does not entitle", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		productID := identifiers.New()

		subscription := paymentsfakes.BuildFakeSubscription(identifiers.New(), productID)
		subscription.Status = payments.SubscriptionStatusPastDue

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{result: resultOf(subscription)},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {
					Behavior:     platformmetering.BehaviorBlock,
					ByProduct:    map[string]int64{productID: 5_000},
					Unsubscribed: 100,
				},
			},
		)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(100), quota.Limit)
	})

	T.Run("a live subscription is found past a cancelled one", func(t *testing.T) {
		t.Parallel()

		// An account keeps its old rows, so "the first subscription" and "the subscription
		// that entitles" are not the same thing.
		ctx := t.Context()
		productID := identifiers.New()

		cancelled := paymentsfakes.BuildFakeSubscription(identifiers.New(), identifiers.New())
		cancelled.Status = payments.SubscriptionStatusCancelled

		active := paymentsfakes.BuildFakeSubscription(identifiers.New(), productID)
		active.Status = payments.SubscriptionStatusActive

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{result: resultOf(cancelled, active)},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {
					Behavior:     platformmetering.BehaviorBlock,
					ByProduct:    map[string]int64{productID: 5_000},
					Unsubscribed: 100,
				},
			},
		)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(5_000), quota.Limit)
	})

	T.Run("a product with no configured limit falls back to unsubscribed", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		subscription := paymentsfakes.BuildFakeSubscription(identifiers.New(), identifiers.New())
		subscription.Status = payments.SubscriptionStatusActive

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{result: resultOf(subscription)},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {
					Behavior:     platformmetering.BehaviorBlock,
					ByProduct:    map[string]int64{identifiers.New(): 5_000},
					Unsubscribed: 100,
				},
			},
		)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(100), quota.Limit)
	})

	T.Run("an account with no subscriptions at all", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{returnNilBoth: true},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {
					Behavior:     platformmetering.BehaviorBlock,
					Unsubscribed: 100,
				},
			},
		)

		quota, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		require.NoError(t, err)

		assert.Equal(t, int64(100), quota.Limit)
	})

	T.Run("with error reading subscriptions", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		source := newSubscriptionQuotaSource(
			loggingnoop.NewLogger(),
			&subscriptionReaderStub{err: expected},
			map[string]PlanLimits{
				UploadedMediaBytesMeter: {Behavior: platformmetering.BehaviorBlock, Unsubscribed: 100},
			},
		)

		_, err := source.QuotaFor(ctx, identifiers.New(), UploadedMediaBytesMeter)
		assert.ErrorIs(t, err, expected)
	})
}

func TestPlanLimits_ShippedTableIsEmpty(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		// Everything is unlimited to start, which is the product decision this whole
		// adoption rests on: count first, set limits from what the totals actually say.
		// A limit that appears here without the dashboards to justify it is a guess, and
		// the guess that is too low is an outage for a customer who did nothing wrong.
		assert.Empty(t, planLimits)
	})
}
