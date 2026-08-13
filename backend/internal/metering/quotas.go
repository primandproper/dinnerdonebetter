package metering

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/filtering"
	platformmetering "github.com/primandproper/platform-go/v10/metering"
	"github.com/primandproper/platform-go/v10/observability/logging"
)

// quotaSourceO11yName names this component's logger.
const quotaSourceO11yName = "metering_quota_source"

// PlanLimits is what one meter is worth on each product an account can hold.
//
// It is the whole of this application's opinion about plans. The platform models none of them
// on purpose — a plan catalog kept here as well as at the billing provider is two catalogs that
// can disagree — so this is the one place a tier's limits are written down.
type PlanLimits struct {
	_ struct{} `json:"-"`

	// ByProduct is the limit for an account whose active subscription names that product,
	// keyed by payments.Product ID.
	ByProduct map[string]int64

	// Behavior is what happens at the limit for this meter, on every plan. It is per meter
	// rather than per product because it describes the meter's nature — whether going over is
	// a refusal, a warning, or the point at which the price changes — and a meter that blocked
	// on one tier and billed overage on another would be two different meters.
	Behavior platformmetering.QuotaBehavior

	// Unsubscribed is the limit for an account with no subscription the system considers live:
	// none at all, or one that is past due, cancelled, or never completed.
	Unsubscribed int64
}

// planLimits is empty, and that is the current product decision rather than an oversight.
//
// Everything is unlimited to start, because a limit set before there is usage data to set it
// from is a guess, and the guess that is too low is an outage for a customer who did nothing
// wrong. The counting exists first; the numbers go here once the dashboards over the totals
// table say what real usage looks like.
//
// A meter absent from this map is unlimited for every account, and QuotaFor answers it without
// reading anything — see the short circuit there.
var planLimits = map[string]PlanLimits{}

// liveSubscriptionStatuses are the statuses that entitle an account to its product's limits.
//
// Trialing counts: a trial that silently got the unsubscribed limits would be a trial of a
// different product. Past due does not, which is the lever that makes a lapsed payment
// degrade service rather than end it — though with planLimits empty it currently levers
// nothing.
var liveSubscriptionStatuses = map[string]bool{
	payments.SubscriptionStatusActive:   true,
	payments.SubscriptionStatusTrialing: true,
}

// SubscriptionReader is the slice of the payments repository this package needs: one read, by
// account.
type SubscriptionReader interface {
	GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error)
}

var _ platformmetering.QuotaSource = (*subscriptionQuotaSource)(nil)

// subscriptionQuotaSource answers "what is this account allowed" from the subscription it holds.
type subscriptionQuotaSource struct {
	subscriptions SubscriptionReader
	limits        map[string]PlanLimits
	logger        logging.Logger
}

// NewSubscriptionQuotaSource builds the QuotaSource an Enforcer consults, resolving each
// account's limits through the subscription it holds.
func NewSubscriptionQuotaSource(logger logging.Logger, subscriptions SubscriptionReader) platformmetering.QuotaSource {
	return newSubscriptionQuotaSource(logger, subscriptions, planLimits)
}

// newSubscriptionQuotaSource is NewSubscriptionQuotaSource over an explicit limits table, so a
// test can exercise the plan resolution that the shipped table currently leaves inert.
func newSubscriptionQuotaSource(logger logging.Logger, subscriptions SubscriptionReader, limits map[string]PlanLimits) platformmetering.QuotaSource {
	return &subscriptionQuotaSource{
		subscriptions: subscriptions,
		limits:        limits,
		logger:        logging.NewNamedLogger(logger, quotaSourceO11yName),
	}
}

// QuotaFor implements metering.QuotaSource.
func (s *subscriptionQuotaSource) QuotaFor(ctx context.Context, subject, meter string) (platformmetering.Quota, error) {
	limits, gated := s.limits[meter]
	if !gated {
		// No plan varies this meter, so there is nothing a subscription could tell us.
		//
		// Short-circuited before the read rather than after it, because QuotaFor sits on
		// Enforcer.Check's path — the path whose entire reason to exist is being cheaper
		// than a durable round trip. A subscription lookup per check, to reach an answer
		// that is identical for every account, would make the cheap path the expensive one.
		return unlimitedQuota(meter), nil
	}

	limit := limits.Unsubscribed

	productID, entitled, err := s.entitlingProduct(ctx, subject)
	if err != nil {
		return platformmetering.Quota{}, errors.Wrapf(err, "reading subscriptions for account %q", subject)
	}

	if entitled {
		if productLimit, ok := limits.ByProduct[productID]; ok {
			limit = productLimit
		} else {
			// A product nobody wrote a limit for. Treated as the unsubscribed limit and
			// said out loud, because the alternatives are both worse: unlimited would
			// let a new tier ship without limits and nobody notice, and an error would
			// take the endpoint down for a customer whose only mistake was buying the
			// plan we forgot to configure.
			s.logger.WithValue("meter", meter).
				WithValue("account_id", subject).
				WithValue("product_id", productID).
				Info("no plan limit configured for product, applying unsubscribed limit")
		}
	}

	return platformmetering.Quota{
		Meter:    meter,
		Behavior: limits.Behavior,
		Period:   meterPeriod(meter),
		Limit:    limit,
	}, nil
}

// entitlingProduct returns the product the account's live subscription names, and whether it has
// one at all.
//
// An account can hold several rows — an old cancelled one beside a current one — so this picks
// the first with a live status rather than assuming there is only ever one.
func (s *subscriptionQuotaSource) entitlingProduct(ctx context.Context, accountID string) (productID string, entitled bool, err error) {
	result, err := s.subscriptions.GetSubscriptionsForAccount(ctx, accountID, filtering.DefaultQueryFilter())
	if err != nil {
		return "", false, err
	}

	if result == nil {
		return "", false, nil
	}

	for _, subscription := range result.Data {
		if subscription != nil && liveSubscriptionStatuses[subscription.Status] {
			return subscription.ProductID, true, nil
		}
	}

	return "", false, nil
}

// unlimitedQuota is the quota for a meter no plan limits.
func unlimitedQuota(meter string) platformmetering.Quota {
	return platformmetering.Quota{
		Meter:    meter,
		Behavior: platformmetering.BehaviorAllowOverage,
		Period:   meterPeriod(meter),
		Limit:    unlimited,
	}
}

// meterPeriod is the registered period for a meter. A quota whose period differs from its
// meter's is refused at registration, so this reads it off the one declaration rather than
// repeating the constant.
func meterPeriod(name string) platformmetering.Period {
	for idx := range meters {
		if meters[idx].Name == name {
			return meters[idx].Period
		}
	}

	// An unregistered meter never reaches here through an Enforcer, which resolves the meter
	// before asking for its quota. The calendar month is the period every meter in this
	// application uses, so an answer that cannot be right is still not a period mismatch.
	return platformmetering.PeriodMonth
}
