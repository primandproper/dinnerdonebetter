package metering

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v10/filtering"
	platformmetering "github.com/primandproper/platform-go/v10/metering"
)

// quotaSourceO11yName names this component's logger.
const quotaSourceO11yName = "metering_quota_source"

// planLimits is empty, and that is the current product decision rather than an oversight.
//
// Everything is unlimited to start, because a limit set before there is usage data to set it
// from is a guess, and the guess that is too low is an outage for a customer who did nothing
// wrong. The counting exists first; the numbers go here once the dashboards over the totals
// table say what real usage looks like.
//
// A meter absent from this map is unlimited for every account, and the platform's QuotaFor
// answers it without reading anything.
var planLimits = map[string]platformmetering.PlanLimits{}

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

// NewSubscriptionQuotaSource builds the QuotaSource an Enforcer consults, resolving each
// account's limits through the subscription it holds.
//
// The three-rung ladder it walks — meter absent from the table is unlimited without reading
// anything, an entitled subject gets its product's limit, an unentitled one gets Unsubscribed —
// is the platform's, and so is deriving each quota's period from the registry rather than from
// a scan over the meter declarations. What stays here is the only half that is a product
// decision: which subscription statuses count as live, and which row wins when an account holds
// several.
func NewSubscriptionQuotaSource(registry *platformmetering.Registry, subscriptions SubscriptionReader, opts ...platformmetering.PlanLimitOption) (platformmetering.QuotaSource, error) {
	return platformmetering.NewPlanLimitSource(registry, planLimits, entitlementReader(subscriptions), opts...)
}

// entitlementReader answers which product entitles an account, if any.
//
// An account can hold several rows — an old cancelled one beside a current one — so this picks
// the first with a live status rather than assuming there is only ever one.
func entitlementReader(subscriptions SubscriptionReader) platformmetering.EntitlementReader {
	return platformmetering.EntitlementReaderFunc(func(ctx context.Context, accountID string) (productID string, entitled bool, err error) {
		result, err := subscriptions.GetSubscriptionsForAccount(ctx, accountID, filtering.DefaultQueryFilter())
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
	})
}
