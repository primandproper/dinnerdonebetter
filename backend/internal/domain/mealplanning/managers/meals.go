package managers

import (
	"context"
	"errors"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	platformkeys "github.com/primandproper/platform-go/v9/observability/keys"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
)

func (m *mealPlanningManager) ListMeals(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	results, err := m.db.GetMeals(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, m.logger, span, "fetching meals from database")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMeal(ctx context.Context, creatorID string, input *types.MealCreationRequestInput) (*types.Meal, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	existing, err := m.db.FindMealWithSameComponents(ctx, creatorID, input)
	if err != nil && !errors.Is(err, types.ErrNoMatchingMeal) {
		return nil, observability.PrepareAndLogError(err, m.logger.WithSpan(span), span, "checking for duplicate meal")
	}
	if existing != nil {
		return nil, types.ErrDuplicateMeal
	}

	convertedInput := converters.ConvertMealCreationRequestInputToMealDatabaseCreationInput(input)
	convertedInput.CreatedByUser = creatorID
	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, convertedInput.ID)

	created, err := m.db.CreateMeal(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal")
	}

	return created, nil
}

func (m *mealPlanningManager) ReadMeal(ctx context.Context, mealID string) (*types.Meal, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealIDKey, mealID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)

	meal, err := m.db.GetMeal(ctx, mealID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal")
	}

	return meal, nil
}

func (m *mealPlanningManager) SearchMeals(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(platformkeys.SearchQueryKey, query).WithValue(platformkeys.UseDatabaseKey, !useSearchService)
	tracing.AttachToSpan(span, platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, platformkeys.UseDatabaseKey, !useSearchService)

	var (
		results *filtering.QueryFilteredResult[types.Meal]
		err     error
	)

	if useSearchService {
		results, err = m.searchMealsViaIndex(ctx, query, filter)
	}

	if err != nil || results == nil {
		results, err = m.db.SearchForMeals(ctx, query, filter)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "searching for meals")
		}
	}

	return results, nil
}

// searchMealsViaIndex searches meals via the external search index. Returns (nil, err) on search failure, empty results, or GetMealsWithIDs failure.
func (m *mealPlanningManager) searchMealsViaIndex(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error) {
	indexHits, err := m.mealsSearchIndex.Search(ctx, textsearch.SearchRequest{Query: query})
	if err != nil || len(indexHits.Hits) == 0 {
		return nil, err
	}

	mealSubsets := indexHits.Hits

	ids := make([]string, 0, len(mealSubsets))
	for _, mealSubset := range mealSubsets {
		ids = append(ids, mealSubset.ID)
	}

	idResults, err := m.db.GetMealsWithIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return filtering.NewQueryFilteredResult(idResults, uint64(len(idResults)), uint64(len(idResults)), func(meal *types.Meal) string {
		return meal.ID
	}, filter), nil
}

func (m *mealPlanningManager) ArchiveMeal(ctx context.Context, mealID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealIDKey: mealID,
		identitykeys.UserIDKey:     ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	if err := m.db.ArchiveMeal(ctx, mealID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal")
	}

	return nil
}

func (m *mealPlanningManager) AddMealImage(ctx context.Context, mealID, uploadedMediaID, uploadedByUser string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealIDKey, mealID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)

	if err := m.db.AddMealImage(ctx, mealID, uploadedMediaID, uploadedByUser); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "adding meal image")
	}

	return nil
}
