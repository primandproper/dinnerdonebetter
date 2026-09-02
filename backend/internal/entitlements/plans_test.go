package entitlements

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
	returnNilBoth bool
}

func (s *subscriptionReaderStub) GetSubscriptionsForAccount(_ context.Context, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
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

// subscriptionWithStatus is a fake subscription forced to one status.
func subscriptionWithStatus(status string) *payments.Subscription {
	subscription := paymentsfakes.BuildFakeSubscription(identifiers.New(), identifiers.New())
	subscription.Status = status

	return subscription
}

// sourceOver builds a plan source over a stubbed reader.
func sourceOver(t *testing.T, subscriptions SubscriptionReader) *SubscriptionPlanSource {
	t.Helper()

	source, err := NewSubscriptionPlanSource(subscriptions)
	require.NoError(t, err)

	return source
}

func TestNewSubscriptionPlanSource(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		source, err := NewSubscriptionPlanSource(&subscriptionReaderStub{})
		require.NoError(t, err)
		assert.NotNil(t, source)
	})

	T.Run("with nil subscription reader", func(t *testing.T) {
		t.Parallel()

		source, err := NewSubscriptionPlanSource(nil)
		require.ErrorIs(t, err, ErrNilSubscriptionReader)
		assert.Nil(t, source)
	})
}

// Which statuses leave an account entitled, and which row wins when it holds several, are the
// product decisions this package exists to hold. The ladder above them — the plan cache, the
// fallback, what a decision is made of — is the platform's and is tested there.
func TestSubscriptionPlanSource_PlanFor(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		accountID := identifiers.New()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusActive)),
		}

		plan, err := sourceOver(t, subscriptions).PlanFor(ctx, accountID)
		require.NoError(t, err)
		assert.Equal(t, SubscriberPlan, plan)
		assert.Equal(t, []string{accountID}, subscriptions.requestedFor)
	})

	T.Run("a trialing subscription entitles the same as an active one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusTrialing)),
		}

		plan, err := sourceOver(t, subscriptions).PlanFor(ctx, identifiers.New())
		require.NoError(t, err)
		assert.Equal(t, SubscriberPlan, plan)
	})

	T.Run("a past due subscription does not entitle", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(subscriptionWithStatus(payments.SubscriptionStatusPastDue)),
		}

		plan, err := sourceOver(t, subscriptions).PlanFor(ctx, identifiers.New())
		require.NoError(t, err)
		assert.Equal(t, FreePlan, plan)
	})

	T.Run("a live subscription is found past a cancelled one", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		subscriptions := &subscriptionReaderStub{
			result: resultOf(
				subscriptionWithStatus(payments.SubscriptionStatusCancelled),
				subscriptionWithStatus(payments.SubscriptionStatusActive),
			),
		}

		plan, err := sourceOver(t, subscriptions).PlanFor(ctx, identifiers.New())
		require.NoError(t, err)
		assert.Equal(t, SubscriberPlan, plan)
	})

	T.Run("an account with no subscriptions at all is on the free plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		// Not ErrNoPlan: an account that has never paid is a customer of the free tier
		// rather than a customer of nothing, and reporting an absence would deny it the
		// features that tier includes.
		plan, err := sourceOver(t, &subscriptionReaderStub{result: resultOf()}).PlanFor(ctx, identifiers.New())
		require.NoError(t, err)
		assert.Equal(t, FreePlan, plan)
	})

	T.Run("with a nil result and no error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		plan, err := sourceOver(t, &subscriptionReaderStub{returnNilBoth: true}).PlanFor(ctx, identifiers.New())
		require.NoError(t, err)
		assert.Equal(t, FreePlan, plan)
	})

	T.Run("with error reading subscriptions", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		// Reported rather than answered with the free plan. A failed read is an outage, and
		// what to do about one is CheckerConfig.FallbackPlan's decision to make with the
		// whole picture — degrading to free here would take that decision away and would do
		// it for the quota path too, which has no fallback on purpose.
		plan, err := sourceOver(t, &subscriptionReaderStub{err: expected}).PlanFor(ctx, identifiers.New())
		require.ErrorIs(t, err, expected)
		assert.Empty(t, plan)
	})
}

func TestDefaultPlans(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		names := make([]string, 0, len(DefaultPlans()))
		for _, plan := range DefaultPlans() {
			names = append(names, plan.Name)
		}

		// Every plan name PlanFor can return has to be one the catalog defines: the two are
		// joined by string equality and by nothing else, and a plan the catalog does not
		// know denies with ErrUnknownPlan.
		assert.ElementsMatch(t, []string{FreePlan, SubscriberPlan}, names)
	})

	T.Run("each plan carries its own grants", func(t *testing.T) {
		t.Parallel()

		// The two plans grant the same thing today, and they must not grant it through the
		// same slice — an operator editing one tier's limits would otherwise edit both.
		plans := DefaultPlans()
		require.Len(t, plans, 2)
		require.NotEmpty(t, plans[0].Includes)

		plans[0].Includes[0].Unlimited = false

		assert.True(t, plans[1].Includes[0].Unlimited)
	})
}
