package privacy

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
)

// Collector collects meal planning domain data for GDPR/CCPA disclosure.
type Collector struct {
	repo   mealplanning.Repository
	tracer tracing.Tracer
	logger logging.Logger
}

// NewCollector creates a new meal planning data privacy collector.
func NewCollector(repo mealplanning.Repository, logger logging.Logger, tracerProvider tracing.TracerProvider) *Collector {
	return &Collector{
		repo:   repo,
		tracer: tracing.NewNamedTracer(tracerProvider, "mealplanning_privacy_collector"),
		logger: logging.NewNamedLogger(logger, "mealplanning_privacy_collector"),
	}
}

var _ dataprivacy.UserDataCollector = (*Collector)(nil)

// CollectUserData collects user-scoped meal planning data.
func (c *Collector) CollectUserData(ctx context.Context, collection *dataprivacy.UserDataCollection, userID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("user_id", userID)

	recipes, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
		return c.repo.GetRecipesCreatedByUser(ctx, userID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching recipes")
	}
	for _, recipe := range recipes {
		collection.MealPlanning.Recipes = append(collection.MealPlanning.Recipes, *recipe)
	}

	meals, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
		return c.repo.GetMealsCreatedByUser(ctx, userID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meals")
	}
	for _, meal := range meals {
		collection.MealPlanning.Meals = append(collection.MealPlanning.Meals, *meal)
	}

	preferences, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
		return c.repo.GetUserIngredientPreferences(ctx, userID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching ingredient preferences")
	}
	for _, pref := range preferences {
		collection.MealPlanning.UserIngredientPreferences = append(collection.MealPlanning.UserIngredientPreferences, *pref)
	}

	ratings, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
		return c.repo.GetRecipeRatingsForUser(ctx, userID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching recipe ratings")
	}
	for _, rating := range ratings {
		collection.MealPlanning.RecipeRatings = append(collection.MealPlanning.RecipeRatings, *rating)
	}

	return nil
}

// CollectAccountData collects account-scoped meal planning data.
func (c *Collector) CollectAccountData(ctx context.Context, collection *dataprivacy.UserDataCollection, accountID string) error {
	ctx, span := c.tracer.StartSpan(ctx)
	defer span.End()

	logger := c.logger.WithValue("account_id", accountID)

	mealPlans, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
		return c.repo.GetMealPlansForAccount(ctx, accountID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plans")
	}
	for _, mp := range mealPlans {
		collection.MealPlanning.MealPlans = append(collection.MealPlanning.MealPlans, *mp)
	}

	ownerships, err := dataprivacy.CollectAllPages(ctx, func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
		return c.repo.GetAccountInstrumentOwnerships(ctx, accountID, filter)
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching instrument ownerships")
	}
	for _, ownership := range ownerships {
		collection.MealPlanning.AccountInstrumentOwnerships = append(collection.MealPlanning.AccountInstrumentOwnerships, *ownership)
	}

	return nil
}
