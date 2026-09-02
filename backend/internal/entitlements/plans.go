package entitlements

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	platformentitlements "github.com/primandproper/platform-go/v13/entitlements"
	"github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
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
	// It is also what CheckerConfig.FallbackPlan names, so that a payments database that has
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

// ErrNilSubscriptionReader indicates a plan source built without anything to read
// subscriptions from.
var ErrNilSubscriptionReader = errors.Wrap(errors.ErrNilInputParameter, "nil subscription reader")

// liveSubscriptionStatuses are the statuses that put an account on SubscriberPlan.
//
// Trialing counts: a trial that silently got the free plan's grants would be a trial of a
// different product. Past due does not, which is the lever that makes a lapsed payment degrade
// service rather than end it — though with both plans granting the same thing it currently
// levers nothing.
var liveSubscriptionStatuses = map[string]bool{
	payments.SubscriptionStatusActive:   true,
	payments.SubscriptionStatusTrialing: true,
}

// SubscriptionReader is the slice of the payments repository this package needs: one read, by
// account.
type SubscriptionReader interface {
	GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error)
}

var _ platformentitlements.PlanSource = (*SubscriptionPlanSource)(nil)

// SubscriptionPlanSource answers which plan an account is on from the subscriptions it holds.
type SubscriptionPlanSource struct {
	subscriptions SubscriptionReader
}

// NewSubscriptionPlanSource builds the PlanSource a Checker and a QuotaSource resolve accounts
// through.
//
// This is the seam the platform cannot fill, and the reason is that the join between an account
// and a purchased plan is application data: it is written by the handler that processes the
// billing provider's subscription webhook, and it lives in a table next to the account. It is
// deliberately not read from capitalism — a provider round trip per feature check spends a
// latency budget on a fact that changes a few times a year, and an outage there would take the
// product down rather than the billing.
func NewSubscriptionPlanSource(subscriptions SubscriptionReader) (*SubscriptionPlanSource, error) {
	if subscriptions == nil {
		return nil, ErrNilSubscriptionReader
	}

	return &SubscriptionPlanSource{subscriptions: subscriptions}, nil
}

// PlanFor implements entitlements.PlanSource.
//
// An account can hold several rows — an old cancelled one beside a current one — so this takes
// the first with a live status rather than assuming there is only ever one.
//
// It never reports ErrNoPlan. Every account is on a plan here, because FreePlan is a real tier
// rather than the name for having nothing; see FreePlan. A read that fails is a different
// matter and is reported as the failure it is, which is what CheckerConfig.FallbackPlan then
// answers for boolean features.
func (s *SubscriptionPlanSource) PlanFor(ctx context.Context, account string) (string, error) {
	result, err := s.subscriptions.GetSubscriptionsForAccount(ctx, account, filtering.DefaultQueryFilter())
	if err != nil {
		return "", errors.Wrapf(err, "reading subscriptions for account %q", account)
	}

	if result != nil {
		for _, subscription := range result.Data {
			if subscription != nil && liveSubscriptionStatuses[subscription.Status] {
				return SubscriberPlan, nil
			}
		}
	}

	return FreePlan, nil
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
// plan names SubscriptionPlanSource can return rather than denying every account with
// ErrUnknownPlan.
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
