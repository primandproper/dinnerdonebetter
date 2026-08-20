package managers

import (
	"context"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/identifiers"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

func (m *mealPlanningManager) AddMealToMealList(ctx context.Context, mealListID, mealID, notes string) (*types.MealListItem, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if mealListID == "" || mealID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	exists, err := m.db.MealExistsInMealList(ctx, mealListID, mealID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "checking if meal exists in list")
	}
	if exists {
		return nil, types.ErrDuplicateMealInList
	}

	input := &types.MealListItemDatabaseCreationInput{
		ID:                identifiers.New(),
		MealID:            mealID,
		Notes:             notes,
		BelongsToMealList: mealListID,
	}

	item, err := m.db.CreateMealListItem(ctx, input)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "adding meal to meal list")
	}

	return item, nil
}

func (m *mealPlanningManager) UpdateMealListItem(ctx context.Context, mealListItemID, mealListID, mealID string, input *types.MealListItemUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealListIDKey:     mealListID,
		mealplanningkeys.MealListItemIDKey: mealListItemID,
		mealplanningkeys.MealIDKey:         mealID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListItemIDKey, mealListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if mealListItemID == "" || mealListID == "" || mealID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	if input.Notes == nil {
		return platformerrors.ErrNilInputParameter
	}

	updated := &types.MealListItem{
		ID:                mealListItemID,
		BelongsToMealList: mealListID,
		Meal:              types.Meal{ID: mealID},
	}
	updated.Update(input)

	if err := m.db.UpdateMealListItem(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal list item")
	}

	return nil
}

func (m *mealPlanningManager) RemoveMealFromMealList(ctx context.Context, mealListID, mealListItemID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if mealListID == "" || mealListItemID == "" {
		return platformerrors.ErrEmptyInputParameter
	}

	if err := m.db.ArchiveMealListItem(ctx, mealListItemID, mealListID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "removing meal from meal list")
	}

	return nil
}

func (m *mealPlanningManager) ListMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealListItem], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if mealListID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	results, err := m.db.GetMealListItems(ctx, mealListID, userID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing meal list items")
	}

	return results, nil
}
