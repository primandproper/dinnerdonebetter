package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/observability"
	platformkeys "github.com/primandproper/primitives-go/v2/observability/keys"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

var (
	_ mealplanning.ValidIngredientGroupDataManager = (*repository)(nil)
)

// ValidIngredientGroupExists fetches whether a valid ingredient group exists from the database.
func (q *repository) ValidIngredientGroupExists(ctx context.Context, validIngredientGroupID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID})
	if err != nil {
		return false, err
	}

	result, err := q.generatedQuerier.CheckValidIngredientGroupExistence(ctx, q.readDB, validIngredientGroupID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient group existence check")
	}

	return result, nil
}

// GetValidIngredientGroup fetches a valid ingredient group from the database.
func (q *repository) GetValidIngredientGroup(ctx context.Context, validIngredientGroupID string) (*mealplanning.ValidIngredientGroup, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID})
	if err != nil {
		return nil, err
	}

	result, err := q.generatedQuerier.GetValidIngredientGroup(ctx, q.readDB, validIngredientGroupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group from database")
	}

	validIngredientGroup := &mealplanning.ValidIngredientGroup{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Slug:          result.Slug,
		Description:   result.Description,
		Members:       nil,
	}

	membersResults, err := q.generatedQuerier.GetValidIngredientGroupMembers(ctx, q.readDB, validIngredientGroupID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group members from database")
	}

	for _, memberResult := range membersResults {
		validIngredientGroup.Members = append(validIngredientGroup.Members, validIngredientGroupMemberFromRow(memberResult))
	}

	return validIngredientGroup, nil
}

// SearchForValidIngredientGroups fetches a valid ingredient group from the database.
func (q *repository) SearchForValidIngredientGroups(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if query == "" {
		return nil, platformerrors.ErrEmptyInputProvided
	}
	logger = logger.WithValue(platformkeys.SearchQueryKey, query)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, query)

	filter, logger = filtering.Observe(ctx, logger, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.SearchForValidIngredientGroups(ctx, q.readDB, &generated.SearchForValidIngredientGroupsParams{
		Name:            query,
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhook from database")
	}

	var (
		validIngredientGroups     = []*mealplanning.ValidIngredientGroup{}
		filteredCount, totalCount uint64
	)

	for _, result := range results {
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)

		validIngredientGroup := validIngredientGroupFromListRow((*generated.GetValidIngredientGroupsRow)(result))

		var membersResults []*generated.GetValidIngredientGroupMembersRow
		membersResults, err = q.generatedQuerier.GetValidIngredientGroupMembers(ctx, q.readDB, result.ID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group members from database")
		}

		for _, memberResult := range membersResults {
			validIngredientGroup.Members = append(validIngredientGroup.Members, validIngredientGroupMemberFromRow(memberResult))
		}

		validIngredientGroups = append(validIngredientGroups, validIngredientGroup)
	}

	x := filtering.NewQueryFilteredResult(validIngredientGroups, filteredCount, totalCount, func(vig *mealplanning.ValidIngredientGroup) string { return vig.ID }, filter)

	return x, nil
}

// GetValidIngredientGroups fetches a list of valid ingredients group from the database that meet a particular filter.
func (q *repository) GetValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	filter, logger = filtering.Observe(ctx, logger, filter)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetValidIngredientGroups(ctx, q.readDB, &generated.GetValidIngredientGroupsParams{
		CreatedBefore:   filterArgs.CreatedBefore,
		CreatedAfter:    filterArgs.CreatedAfter,
		UpdatedBefore:   filterArgs.UpdatedBefore,
		UpdatedAfter:    filterArgs.UpdatedAfter,
		PageCursor:      filterArgs.Cursor,
		ResultLimit:     filterArgs.ResultLimit,
		IncludeArchived: filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhook from database")
	}

	var (
		data          []*mealplanning.ValidIngredientGroup
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		validIngredientGroup := validIngredientGroupFromListRow(result)

		var membersResults []*generated.GetValidIngredientGroupMembersRow
		membersResults, err = q.generatedQuerier.GetValidIngredientGroupMembers(ctx, q.readDB, result.ID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group members from database")
		}

		for _, memberResult := range membersResults {
			validIngredientGroup.Members = append(validIngredientGroup.Members, validIngredientGroupMemberFromRow(memberResult))
		}

		data = append(data, validIngredientGroup)
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(vig *mealplanning.ValidIngredientGroup) string { return vig.ID },
		filter,
	)

	return x, nil
}

// CreateValidIngredientGroup creates a valid ingredient group in the database.
func (q *repository) CreateValidIngredientGroup(ctx context.Context, input *mealplanning.ValidIngredientGroupDatabaseCreationInput) (*mealplanning.ValidIngredientGroup, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, input.ID)

	x := &mealplanning.ValidIngredientGroup{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Slug:        input.Slug,
		CreatedAt:   q.CurrentTime(),
	}

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientGroupCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientGroupIDKey: input.ID,
	}, func(tx database.Tx) error {
		// create the valid ingredient group.
		if err := q.generatedQuerier.CreateValidIngredientGroup(ctx, tx, &generated.CreateValidIngredientGroupParams{
			ID:          input.ID,
			Name:        input.Name,
			Description: input.Description,
			Slug:        input.Slug,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing valid ingredient group creation query")
		}

		for i := range input.Members {
			member, err := q.createValidIngredientGroupMember(ctx, tx, x.ID, input.Members[i])
			if err != nil {
				return observability.PrepareAndLogError(err, logger, span, "creating valid ingredient group member")
			}

			x.Members = append(x.Members, member)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	logger.WithValue("member_count", len(input.Members)).Info("valid ingredient group created")

	return x, nil
}

// createValidIngredientGroupMember creates a valid ingredient group member in the database.
func (q *repository) createValidIngredientGroupMember(ctx context.Context, db database.SQLQueryExecutor, groupID string, input *mealplanning.ValidIngredientGroupMemberDatabaseCreationInput) (*mealplanning.ValidIngredientGroupMember, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, input.ID).WithValue(mealplanningkeys.ValidIngredientIDKey, input.ValidIngredientID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, input.ID)

	// create the valid ingredient group.
	if err := q.generatedQuerier.CreateValidIngredientGroupMember(ctx, db, &generated.CreateValidIngredientGroupMemberParams{
		ID:              input.ID,
		BelongsToGroup:  groupID,
		ValidIngredient: input.ValidIngredientID,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient group member creation query")
	}

	x := &mealplanning.ValidIngredientGroupMember{
		ID:              input.ID,
		BelongsToGroup:  groupID,
		ValidIngredient: mealplanning.ValidIngredient{ID: input.ValidIngredientID},
		CreatedAt:       q.CurrentTime(),
	}

	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, x.ID)
	logger.Info("valid ingredient group member created")

	return x, nil
}

// UpdateValidIngredientGroup updates a particular valid ingredient group.
func (q *repository) UpdateValidIngredientGroup(ctx context.Context, updated *mealplanning.ValidIngredientGroup) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.ValidIngredientGroupUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientGroupIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateValidIngredientGroup(ctx, tx, &generated.UpdateValidIngredientGroupParams{
			Name:        updated.Name,
			Description: updated.Description,
			Slug:        updated.Slug,
			ID:          updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating valid ingredient group")
	}

	logger.Info("valid ingredient group updated")

	return nil
}

// ArchiveValidIngredientGroup archives a valid ingredient group from the database by its ID.
func (q *repository) ArchiveValidIngredientGroup(ctx context.Context, validIngredientGroupID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger, err := guardIDs(q.logger.Clone(), span, idArg{mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID})
	if err != nil {
		return err
	}

	return q.withEvent(ctx, logger, mealplanning.ValidIngredientGroupArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.ValidIngredientGroupIDKey: validIngredientGroupID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveValidIngredientGroup(ctx, tx, validIngredientGroupID)
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving valid ingredient group")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	})
}

// validIngredientGroupFromListRow maps a row from the paginated valid ingredient group reads to its
// domain type, with an empty (not nil) member list for the caller to fill. The search and list rows
// carry the same columns, so both convert to generated.GetValidIngredientGroupsRow. The single-group
// read leaves Members nil instead and keeps its own literal.
func validIngredientGroupFromListRow(result *generated.GetValidIngredientGroupsRow) *mealplanning.ValidIngredientGroup {
	return &mealplanning.ValidIngredientGroup{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Name:          result.Name,
		Slug:          result.Slug,
		Description:   result.Description,
		Members:       []*mealplanning.ValidIngredientGroupMember{},
	}
}

// validIngredientGroupMemberFromRow maps a valid ingredient group member row, with its joined valid
// ingredient, to its domain type. Every group read loads its members through
// GetValidIngredientGroupMembers, so this takes that query's row type.
func validIngredientGroupMemberFromRow(result *generated.GetValidIngredientGroupMembersRow) *mealplanning.ValidIngredientGroupMember {
	return &mealplanning.ValidIngredientGroupMember{
		CreatedAt:      result.CreatedAt,
		ArchivedAt:     database.TimePointerFromNullTime(result.ArchivedAt),
		ID:             result.ID,
		BelongsToGroup: result.BelongsToGroup,
		ValidIngredient: mealplanning.ValidIngredient{
			CreatedAt:                      result.ValidIngredientCreatedAt,
			LastUpdatedAt:                  database.TimePointerFromNullTime(result.ValidIngredientLastUpdatedAt),
			ArchivedAt:                     database.TimePointerFromNullTime(result.ValidIngredientArchivedAt),
			MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
			MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(result.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
			IconPath:                       result.ValidIngredientIconPath,
			Warning:                        result.ValidIngredientWarning,
			PluralName:                     result.ValidIngredientPluralName,
			StorageInstructions:            result.ValidIngredientStorageInstructions,
			Name:                           result.ValidIngredientName,
			ID:                             result.ValidIngredientID,
			Description:                    result.ValidIngredientDescription,
			Slug:                           result.ValidIngredientSlug,
			ShoppingSuggestions:            result.ValidIngredientShoppingSuggestions,
			ContainsShellfish:              result.ValidIngredientContainsShellfish,
			IsLiquid:                       database.BoolFromNullBool(result.ValidIngredientIsLiquid),
			ContainsPeanut:                 result.ValidIngredientContainsPeanut,
			ContainsTreeNut:                result.ValidIngredientContainsTreeNut,
			ContainsEgg:                    result.ValidIngredientContainsEgg,
			ContainsWheat:                  result.ValidIngredientContainsWheat,
			ContainsSoy:                    result.ValidIngredientContainsSoy,
			AnimalDerived:                  result.ValidIngredientAnimalDerived,
			RestrictToPreparations:         result.ValidIngredientRestrictToPreparations,
			ContaminatesEquipment:          result.ValidIngredientContaminatesEquipment,
			ContainsSesame:                 result.ValidIngredientContainsSesame,
			ContainsFish:                   result.ValidIngredientContainsFish,
			ContainsGluten:                 result.ValidIngredientContainsGluten,
			ContainsDairy:                  result.ValidIngredientContainsDairy,
			ContainsAlcohol:                result.ValidIngredientContainsAlcohol,
			AnimalFlesh:                    result.ValidIngredientAnimalFlesh,
			IsStarch:                       result.ValidIngredientIsStarch,
			IsProtein:                      result.ValidIngredientIsProtein,
			IsGrain:                        result.ValidIngredientIsGrain,
			IsFruit:                        result.ValidIngredientIsFruit,
			IsSalt:                         result.ValidIngredientIsSalt,
			IsFat:                          result.ValidIngredientIsFat,
			IsAcid:                         result.ValidIngredientIsAcid,
			IsHeat:                         result.ValidIngredientIsHeat,
		},
	}
}
