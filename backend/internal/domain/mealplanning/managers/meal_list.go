package managers

import (
	"context"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/identifiers"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

func (m *mealPlanningManager) ListMealLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealList], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	results, err := m.db.GetMealLists(ctx, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "listing meal lists")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMealList(ctx context.Context, userID string, input *types.MealListCreationRequestInput) (*types.MealList, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	if userID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating meal list input")
	}

	seenMealIDs := map[string]bool{}
	dbInput := &types.MealListDatabaseCreationInput{
		ID:            identifiers.New(),
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: userID,
	}

	for _, item := range input.Items {
		if item == nil {
			continue
		}
		if seenMealIDs[item.MealID] {
			return nil, types.ErrDuplicateMealInList
		}
		seenMealIDs[item.MealID] = true

		dbInput.Items = append(dbInput.Items, &types.MealListItemDatabaseCreationInput{
			ID:                identifiers.New(),
			MealID:            item.MealID,
			Notes:             item.Notes,
			BelongsToMealList: dbInput.ID,
		})
	}

	created, err := m.db.CreateMealList(ctx, dbInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal list")
	}

	return created, nil
}

func (m *mealPlanningManager) UpdateMealList(ctx context.Context, mealListID, userID string, input *types.MealListUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealListIDKey: mealListID,
		identitykeys.UserIDKey:         userID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}
	if mealListID == "" || userID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	if input.Name == nil || input.Description == nil {
		return platformerrors.ErrNilInputParameter
	}

	updated := &types.MealList{
		ID:            mealListID,
		BelongsToUser: userID,
	}
	updated.Update(input)

	if err := m.db.UpdateMealList(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal list")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveMealList(ctx context.Context, mealListID, userID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span)

	if mealListID == "" || userID == "" {
		return platformerrors.ErrEmptyInputParameter
	}

	if err := m.db.ArchiveMealList(ctx, mealListID, userID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal list")
	}

	return nil
}
