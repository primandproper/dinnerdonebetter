package mealplanning

import (
	"context"
	"database/sql"

	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
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
	_ types.MealListDataManager = (*repository)(nil)
)

// GetMealLists fetches a list of meal lists from the database that meet a particular filter.
func (q *repository) GetMealLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.MealList], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.MealList
		filteredCount uint64
		totalCount    uint64
	)
	listsByID := map[string]*types.MealList{}

	results, err := q.generatedQuerier.GetMealLists(ctx, q.readDB, &generated.GetMealListsParams{
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
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal lists list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}

		ml, exists := listsByID[result.ID]
		if !exists {
			ml = &types.MealList{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
				ID:            result.ID,
				Name:          result.Name,
				Description:   result.Description,
				BelongsToUser: result.BelongsToUser,
				Items:         []*types.MealListItem{},
			}
			listsByID[result.ID] = ml
			data = append(data, ml)
		}

		if result.MealListItemID.Valid && result.MealListItemID.String != "" {
			ml.Items = append(ml.Items, &types.MealListItem{
				CreatedAt:         database.TimeFromNullTime(result.MealListItemCreatedAt),
				LastUpdatedAt:     database.TimePointerFromNullTime(result.MealListItemLastUpdatedAt),
				ArchivedAt:        database.TimePointerFromNullTime(result.MealListItemArchivedAt),
				ID:                result.MealListItemID.String,
				Meal:              types.Meal{ID: result.MealListItemMealID.String},
				Notes:             result.MealListItemNotes.String,
				BelongsToMealList: result.MealListItemBelongsToMealList.String,
			})
		}
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(ml *types.MealList) string { return ml.ID },
		filter,
	)

	return x, nil
}

// CreateMealList creates a meal list in the database.
func (q *repository) CreateMealList(ctx context.Context, input *types.MealListDatabaseCreationInput) (*types.MealList, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.MealListIDKey, input.ID)

	if err := q.generatedQuerier.CreateMealList(ctx, q.writeDB, &generated.CreateMealListParams{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: input.BelongsToUser,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal list creation query")
	}

	x := &types.MealList{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: input.BelongsToUser,
		CreatedAt:     q.CurrentTime(),
	}

	logger.Info("meal list created")

	return x, nil
}

// UpdateMealList updates a particular meal list.
func (q *repository) UpdateMealList(ctx context.Context, updated *types.MealList) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealListIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, updated.ID)

	rowsAffected, err := q.generatedQuerier.UpdateMealList(ctx, q.writeDB, &generated.UpdateMealListParams{
		Name:          updated.Name,
		Description:   updated.Description,
		BelongsToUser: updated.BelongsToUser,
		ID:            updated.ID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal list")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("meal list updated")

	return nil
}

// ArchiveMealList archives a meal list from the database by its ID.
func (q *repository) ArchiveMealList(ctx context.Context, mealListID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if mealListID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealListIDKey, mealListID)
	tracing.AttachToSpan(span, mealplanningkeys.MealListIDKey, mealListID)

	rowsAffected, err := q.generatedQuerier.ArchiveMealList(ctx, q.writeDB, &generated.ArchiveMealListParams{
		BelongsToUser: userID,
		ID:            mealListID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal list")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	logger.Info("meal list archived")

	return nil
}
