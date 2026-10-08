package mealplanning

// AnalyticsEventTypes is this domain's entry in the product analytics allowlist
// (internal/domain/analytics): the events of this domain's a product question is asked of, and
// no others. Each travels as a datachanges.Message, with the actor and the account on the
// message and the context as the properties.
//
// The rest of the domain's events are create/update/archive traffic on catalog tables that no
// product question is ever asked of; see the analytics package for why unclassified means
// unreported.
func AnalyticsEventTypes() []string {
	return []string{
		RecipeCreatedServiceEventType,
		RecipeClonedServiceEventType,
		RecipeRatingCreatedServiceEventType,
		MealCreatedServiceEventType,
		MealPlanCreatedServiceEventType,
		MealPlanOptionVoteCreatedServiceEventType,
		MealPlanFinalizedServiceEventType,
	}
}
