package entitlements

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/billing/plans"
	"github.com/primandproper/platform-go/v13/capitalism"
	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
)

const (
	// FreePlan is what an account holding no live subscription is on.
	//
	// It is a plan rather than an absence, and that is a product decision worth stating: the
	// core of this product is free, so an account that has never paid is entitled to it. The
	// platform's other answer — reporting no plan at all — is for a product where not paying
	// means not being a customer, and it would deny an unsubscribed account a feature it is
	// supposed to have.
	//
	// It is also what CheckerConfig.FallbackPlan names, so that a billing database that has
	// stopped answering degrades a paying account to free rather than locking it out.
	FreePlan = "free"

	// SubscriberPlan is what an account holding a live subscription is on.
	//
	// One plan for every product, because there is one tier to sell. The day there are two,
	// this is where the mapping goes: entitlements joins a plan to a catalog entry by string
	// equality, and a product ID cannot be that string — plan names are plain identifiers and
	// a minted ID is not one.
	SubscriberPlan = "subscriber"
)

// liveSubscriptionStatuses are the statuses that put an account on SubscriberPlan.
//
// Trialing counts: a trial that silently got the free plan's grants would be a trial of a
// different product. Past due does not, which is the lever that makes a lapsed payment degrade
// service rather than end it — though with both plans granting the same thing it currently
// levers nothing.
var liveSubscriptionStatuses = map[capitalism.SubscriptionStatus]bool{
	capitalism.SubscriptionStatusActive:   true,
	capitalism.SubscriptionStatusTrialing: true,
}

// ChoosePlan is this application's plans.Choose: which plan an account is on, given the
// subscriptions whose paid period covers now.
//
// It is the seam the platform leaves to the consumer, and the reason is that which of
// capitalism's statuses leaves an account entitled is policy. This one says: any current
// subscription that is active or trialing puts the account on SubscriberPlan, and nothing does
// otherwise. An account can hold several rows — an old one beside a current one — so the first
// live one wins rather than the only one.
//
// It never declines. Every account is on a plan here, because FreePlan is a real tier rather than
// the name for having nothing; see FreePlan. That is what the second return value is for, and it
// is what keeps entitlements' ErrNoPlan out of this deployment.
func ChoosePlan(subscriptions []*billing.Subscription) (string, bool) {
	for _, subscription := range subscriptions {
		if subscription != nil && liveSubscriptionStatuses[subscription.Status] {
			return SubscriberPlan, true
		}
	}

	return FreePlan, true
}

// NewPlanSource builds the PlanSource a Checker and a QuotaSource resolve accounts through: the
// platform's read of one account's current subscriptions, in the scope this application keeps
// its billing under, decided by ChoosePlan.
//
// It reads the billing store rather than asking the payment provider. A provider round trip per
// feature check spends a latency budget on a fact that changes a few times a year, and an outage
// there would take the product down rather than the billing. The provider says when a
// subscription changes; what it changed to is in the store by the time anybody asks.
//
// The read is of *current* subscriptions — those whose paid period covers now — which is the
// platform's reading and a narrower one than the table this replaced was asked for. A
// subscription whose period lapsed without the provider reporting a status is not one anybody
// is paying for, and this is where that stops entitling.
func NewPlanSource(store billing.SubscriptionStore) (*plans.Source, error) {
	return plans.New(store, payments.Scope(), ChoosePlan)
}

// DefaultPlans is the catalog this service ships with: every feature granted without a bound on
// both tiers.
//
// That is the current product decision rather than an oversight, and it is the same one
// internal/metering documents. A limit set before there is usage data to set it from is a
// guess, and the guess that is too low is an outage for a customer who did nothing wrong. The
// numbers go in once the dashboards over the totals table say what real usage looks like.
//
// Plans are configuration — internal/config.DefaultEntitlementsConfig is what puts these into a
// rendered config file, and an operator changes them there. This function is what that config
// defaults to, so that a deployment which configures no plans at all still resolves both of the
// plan names ChoosePlan can return rather than denying every account with ErrUnknownPlan.
func DefaultPlans() []platformentitlements.Plan {
	// Built per plan rather than from one shared slice: a caller that edited the grants of
	// the plan it was handed would otherwise be editing the other plan's too.
	unbounded := func() []platformentitlements.Grant {
		return []platformentitlements.Grant{
			{Feature: UploadedMediaBytesFeature, Unlimited: true},
		}
	}

	return []platformentitlements.Plan{
		{
			Name:        FreePlan,
			Description: "Everything this service does, for an account with no subscription.",
			Includes:    unbounded(),
		},
		{
			Name:        SubscriberPlan,
			Description: "Everything this service does, for an account with a live subscription.",
			Includes:    unbounded(),
		},
	}
}
