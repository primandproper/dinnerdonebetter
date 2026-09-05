package entitlements

import (
	"context"
	"testing"

	paymentsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v13/billing"
	billingmock "github.com/primandproper/platform-go/v13/billing/mock"
	"github.com/primandproper/platform-go/v13/capitalism"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// subscriptionWithStatus is a fake current subscription forced to one status.
func subscriptionWithStatus(status capitalism.SubscriptionStatus) *billing.Subscription {
	subscription := paymentsfakes.BuildFakeSubscription(identifiers.New(), identifiers.New())
	subscription.Status = status

	return subscription
}

// Which statuses leave an account entitled, and which row wins when it holds several, are the
// product decisions this package exists to hold. The ladder above them — the current-period
// read, the plan cache, the fallback, what a decision is made of — is the platform's and is
// tested there.
func TestChoosePlan(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		plan, ok := ChoosePlan([]*billing.Subscription{subscriptionWithStatus(capitalism.SubscriptionStatusActive)})
		assert.True(t, ok)
		assert.Equal(t, SubscriberPlan, plan)
	})

	T.Run("a trialing subscription entitles the same as an active one", func(t *testing.T) {
		t.Parallel()

		plan, ok := ChoosePlan([]*billing.Subscription{subscriptionWithStatus(capitalism.SubscriptionStatusTrialing)})
		assert.True(t, ok)
		assert.Equal(t, SubscriberPlan, plan)
	})

	T.Run("a past due subscription does not entitle", func(t *testing.T) {
		t.Parallel()

		plan, ok := ChoosePlan([]*billing.Subscription{subscriptionWithStatus(capitalism.SubscriptionStatusPastDue)})
		assert.True(t, ok)
		assert.Equal(t, FreePlan, plan)
	})

	T.Run("a live subscription is found past a canceled one", func(t *testing.T) {
		t.Parallel()

		plan, ok := ChoosePlan([]*billing.Subscription{
			subscriptionWithStatus(capitalism.SubscriptionStatusCanceled),
			subscriptionWithStatus(capitalism.SubscriptionStatusActive),
		})
		assert.True(t, ok)
		assert.Equal(t, SubscriberPlan, plan)
	})

	T.Run("an account with no current subscriptions is on the free plan", func(t *testing.T) {
		t.Parallel()

		// Not a decline: an account that has never paid is a customer of the free tier
		// rather than a customer of nothing, and declining here would reach entitlements
		// as ErrNoPlan and deny it the features that tier includes.
		plan, ok := ChoosePlan(nil)
		assert.True(t, ok)
		assert.Equal(t, FreePlan, plan)
	})

	T.Run("a nil entry is skipped rather than dereferenced", func(t *testing.T) {
		t.Parallel()

		plan, ok := ChoosePlan([]*billing.Subscription{nil, subscriptionWithStatus(capitalism.SubscriptionStatusActive)})
		assert.True(t, ok)
		assert.Equal(t, SubscriberPlan, plan)
	})
}

// The wiring: the platform's source, reading the platform's current-subscription page in this
// application's scope, decided by ChoosePlan.
func TestNewPlanSource(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		accountID := identifiers.New()

		var (
			requestedScope   tenancy.Scope
			requestedAccount string
		)

		store := &billingmock.SubscriptionStoreMock{
			ListCurrentSubscriptionsFunc: func(_ context.Context, scope tenancy.Scope, account string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Subscription], error) {
				requestedScope, requestedAccount = scope, account

				return &filtering.QueryFilteredResult[billing.Subscription]{
					Data: []*billing.Subscription{subscriptionWithStatus(capitalism.SubscriptionStatusActive)},
				}, nil
			},
		}

		source, err := NewPlanSource(store)
		require.NoError(t, err)

		plan, err := source.PlanFor(ctx, accountID)
		require.NoError(t, err)
		assert.Equal(t, SubscriberPlan, plan)
		assert.Equal(t, accountID, requestedAccount)
		assert.Equal(t, tenancy.Global(), requestedScope)
	})

	T.Run("an account with nothing current is on the free plan", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		store := &billingmock.SubscriptionStoreMock{
			ListCurrentSubscriptionsFunc: func(context.Context, tenancy.Scope, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Subscription], error) {
				return &filtering.QueryFilteredResult[billing.Subscription]{}, nil
			},
		}

		source, err := NewPlanSource(store)
		require.NoError(t, err)

		plan, err := source.PlanFor(ctx, identifiers.New())
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
		store := &billingmock.SubscriptionStoreMock{
			ListCurrentSubscriptionsFunc: func(context.Context, tenancy.Scope, string, *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Subscription], error) {
				return nil, expected
			},
		}

		source, err := NewPlanSource(store)
		require.NoError(t, err)

		plan, err := source.PlanFor(ctx, identifiers.New())
		require.ErrorIs(t, err, expected)
		assert.Empty(t, plan)
	})

	T.Run("with nil store", func(t *testing.T) {
		t.Parallel()

		source, err := NewPlanSource(nil)
		require.Error(t, err)
		assert.Nil(t, source)
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

		// Every plan name ChoosePlan can return has to be one the catalog defines: the two are
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
