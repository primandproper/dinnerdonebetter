// Package privacy is the meal planning domain's contribution to a subject access request.
package privacy

import (
	"context"
	"encoding/json"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	platformdataprivacy "github.com/primandproper/platform-go/v13/dataprivacy"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const o11yName = "mealplanning_privacy_collector"

// Collector collects meal planning data about a subject.
type Collector struct {
	repo            mealplanning.Repository
	resolveAccounts dataprivacy.AccountIDResolver
	tracer          tracing.Tracer
	logger          logging.Logger
}

var _ platformdataprivacy.Collector = (*Collector)(nil)

// NewCollector builds the meal planning collector.
func NewCollector(
	repo mealplanning.Repository,
	resolveAccounts dataprivacy.AccountIDResolver,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) *Collector {
	return &Collector{
		repo:            repo,
		resolveAccounts: resolveAccounts,
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:          logging.NewNamedLogger(logger, o11yName),
	}
}

// Collect implements platformdataprivacy.Collector.
//
// Meal planning straddles the user/account line: recipes, meals, preferences,
// and ratings are authored by a person, while meal plans and instrument
// ownerships belong to an account. Both halves are one fragment because both are
// this domain's answer, and splitting them into two sections would make a
// subject reading their export reassemble what one query returned.
func (c *Collector) Collect(ctx context.Context, subject platformdataprivacy.Subject) (json.RawMessage, error) {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithSpan(span)

	recipes, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
		return c.repo.GetRecipesCreatedByUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipes")
	}

	meals, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
		return c.repo.GetMealsCreatedByUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meals")
	}

	preferences, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
		return c.repo.GetUserIngredientPreferences(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching ingredient preferences")
	}

	ratings, err := platformdataprivacy.CollectAll(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
		return c.repo.GetRecipeRatingsForUser(ctx, subject.ID, filter)
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching recipe ratings")
	}

	accountIDs, err := c.resolveAccounts(ctx, subject.ID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "resolving accounts")
	}

	// The hydrated variant: this collection is the user's own copy of their data, so a
	// meal plan in it has to carry its options, votes, meals and selections. The list
	// endpoint's GetMealPlansForAccount drops all of those.
	mealPlans, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetHydratedMealPlansForAccount)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plans")
	}

	ownerships, err := dataprivacy.CollectAcrossAccounts(ctx, accountIDs, c.repo.GetAccountInstrumentOwnerships)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching instrument ownerships")
	}

	held := len(recipes) > 0 || len(meals) > 0 || len(preferences) > 0 ||
		len(ratings) > 0 || len(mealPlans) > 0 || len(ownerships) > 0

	return platformdataprivacy.Fragment(held, &mealplanning.UserDataCollection{
		Recipes:                     recipes,
		Meals:                       meals,
		UserIngredientPreferences:   preferences,
		RecipeRatings:               ratings,
		MealPlans:                   mealPlans,
		AccountInstrumentOwnerships: ownerships,
	})
}
