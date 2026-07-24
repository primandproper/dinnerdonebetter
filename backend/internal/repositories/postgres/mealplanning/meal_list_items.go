package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v6/database"
	platformerrors "github.com/primandproper/platform-go/v6/errors"
	"github.com/primandproper/platform-go/v6/filtering"
	"github.com/primandproper/platform-go/v6/observability"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

var (
	_ types.MealListItemDataManager = (*repository)(nil)
)

// MealExistsInMealList returns true if the meal already exists in the list (non-archived).
func (q *repository) MealExistsInMealList(ctx context.Context, mealListID, mealID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealListID == "" || mealID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealListIDKey, mealListID).WithValue(mealplanningkeys.MealIDKey, mealID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)

	result, err := q.generatedQuerier.CheckMealInMealList(ctx, q.readDB, &generated.CheckMealInMealListParams{
		BelongsToMealList: mealListID,
		MealID:            mealID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking if meal exists in list")
	}
	return result, nil
}

// GetMealListItems fetches meal list items for a given list with filtering.
func (q *repository) GetMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.MealListItem], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealListID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.MealListItem
		filteredCount uint64
		totalCount    uint64
	)
	mealIDs := []string{}

	results, err := q.generatedQuerier.GetMealListItems(ctx, q.readDB, &generated.GetMealListItemsParams{
		MealListID:      mealListID,
		BelongsToUser:   userID,
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal list items list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		data = append(data, &types.MealListItem{
			CreatedAt:         result.CreatedAt,
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			ID:                result.ID,
			Meal:              types.Meal{ID: result.MealID},
			Notes:             result.Notes,
			BelongsToMealList: result.BelongsToMealList,
		})

		if result.MealID != "" {
			mealIDs = append(mealIDs, result.MealID)
		}
	}

	if len(mealIDs) > 0 {
		meals, mealsFetchErr := q.GetMealsWithIDs(ctx, mealIDs)
		if mealsFetchErr != nil {
			return nil, observability.PrepareAndLogError(mealsFetchErr, logger, span, "fetching meals for meal list items")
		}
		mealsByID := map[string]*types.Meal{}
		for _, m := range meals {
			mealsByID[m.ID] = m
		}
		for i, item := range data {
			if m, ok := mealsByID[item.Meal.ID]; ok && m != nil {
				data[i].Meal = *m
			}
		}
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mli *types.MealListItem) string { return mli.ID },
		filter,
	)

	return x, nil
}

// CreateMealListItem creates a meal list item in the database.
func (q *repository) CreateMealListItem(ctx context.Context, input *types.MealListItemDatabaseCreationInput) (*types.MealListItem, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealListItemIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.MealListItemIDKey, input.ID)

	if err := q.generatedQuerier.CreateMealListItem(ctx, q.writeDB, &generated.CreateMealListItemParams{
		ID:                input.ID,
		MealID:            input.MealID,
		Notes:             input.Notes,
		BelongsToMealList: input.BelongsToMealList,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal list item creation query")
	}

	x := &types.MealListItem{
		ID:                input.ID,
		Meal:              types.Meal{ID: input.MealID},
		Notes:             input.Notes,
		BelongsToMealList: input.BelongsToMealList,
		CreatedAt:         q.CurrentTime(),
	}

	logger.Info("meal list item created")

	return x, nil
}

// UpdateMealListItem updates a particular meal list item.
func (q *repository) UpdateMealListItem(ctx context.Context, updated *types.MealListItem) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealListItemIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListItemIDKey, updated.ID)

	rowsAffected, err := q.generatedQuerier.UpdateMealListItem(ctx, q.writeDB, &generated.UpdateMealListItemParams{
		MealID:            updated.Meal.ID,
		Notes:             updated.Notes,
		BelongsToMealList: updated.BelongsToMealList,
		ID:                updated.ID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal list item")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("meal list item updated")

	return nil
}

// ArchiveMealListItem archives a meal list item from the database by its ID.
func (q *repository) ArchiveMealListItem(ctx context.Context, mealListItemID, mealListID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealListID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)

	if mealListItemID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealListItemIDKey, mealListItemID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListItemIDKey, mealListItemID)

	rowsAffected, err := q.generatedQuerier.ArchiveMealListItem(ctx, q.writeDB, &generated.ArchiveMealListItemParams{
		BelongsToMealList: mealListID,
		ID:                mealListItemID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal list item")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("meal list item archived")

	return nil
}
