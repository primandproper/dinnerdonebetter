/*
Package analytics decides which of the application's events are worth reporting to the product
analytics platform.

Every data change event carrying a user ID used to be reported, because the data changes topic
carried every event the application emitted and the consumer had no reason to be selective. That
was never a decision anyone made; it was the absence of one. It sent roughly a hundred and sixty
event types to a third-party vendor, of which all but a dozen were create/update/archive traffic
on catalog tables that no product question is ever asked of, and it did so unfiltered — a
password change and a two-factor secret rotation went the same way a recipe creation did.

# Why an allowlist rather than a denylist

The webhook catalog is a denylist, deliberately: a new domain event nobody remembered to
classify should be deliverable, because that is what webhooks are for, and the events that must
not leak are a known, enumerable set that lives beside the constants declaring them.

The default has to run the other way here. A new event nobody remembered to classify should not
silently become a metric: the cost of forgetting to add one is a question you cannot answer
until you add it, which is recoverable, while the cost of forgetting to leave one out is vendor
spend and a dashboard full of noise that nobody notices is noise. Unclassified means unreported.

# Why this is not the webhook exclusion list

The two answer different questions against different threats. A webhook endpoint is a URL an
account member supplied, so shipping an account's authentication activity there hands a live
security feed to whoever has a foothold in that account. The analytics platform is a single
vendor the operator chose and configured. So the sets are related but neither contains the
other: user_signed_up and user_archived are excluded from webhooks and are the two most
important numbers the product has.

What is true of both is that the events are named by their constants rather than their strings,
so deleting an event type from a domain fails this package's build rather than leaving a metric
that silently stops arriving.
*/
package analytics

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
)

// reportable is the set of event types reported to the analytics platform.
//
// It is short on purpose, and it is a product decision rather than a technical one: these are
// the events someone has an actual question about. Adding one is a line here; the cost of
// adding one is a line here and a number on a dashboard, which is the point.
var reportable = map[string]struct{}{
	// Account lifecycle — activation, churn, and how an account grows.
	identity.UserSignedUpServiceEventType:              {},
	identity.UserArchivedServiceEventType:              {},
	identity.AccountCreatedServiceEventType:            {},
	identity.AccountInvitationCreatedServiceEventType:  {},
	identity.AccountInvitationAcceptedServiceEventType: {},

	// The product loop: a user writes recipes, composes them into meals, plans a week of
	// them, votes, and finalizes. These are the events a question about engagement is asked
	// of. The rest of meal planning is create/update/archive traffic underneath them.
	mealplanning.RecipeCreatedServiceEventType:             {},
	mealplanning.RecipeClonedServiceEventType:              {},
	mealplanning.RecipeRatingCreatedServiceEventType:       {},
	mealplanning.MealCreatedServiceEventType:               {},
	mealplanning.MealPlanCreatedServiceEventType:           {},
	mealplanning.MealPlanOptionVoteCreatedServiceEventType: {},
	mealplanning.MealPlanFinalizedServiceEventType:         {},

	// Revenue.
	payments.SubscriptionCreatedServiceEventType: {},
}

// Reportable reports whether eventType is one the analytics platform should receive.
//
// An unknown event type is not reportable, which is what makes the allowlist the whole of the
// decision: an event added to a domain reaches the analytics platform when someone puts it in
// the map above and not before.
func Reportable(eventType string) bool {
	_, found := reportable[eventType]

	return found
}
