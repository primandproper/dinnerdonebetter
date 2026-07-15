package mealplanning

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v4/database"
	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/observability"
	platformkeys "github.com/primandproper/platform-go/v4/observability/keys"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

var (
	_ mealplanning.ValidIngredientGroupDataManager = (*repository)(nil)
)

// ValidIngredientGroupExists fetches whether a valid ingredient group exists from the database.
func (q *repository) ValidIngredientGroupExists(ctx context.Context, validIngredientGroupID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if validIngredientGroupID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

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

	logger := q.logger.Clone()

	if validIngredientGroupID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

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
		validIngredientGroup.Members = append(validIngredientGroup.Members, &mealplanning.ValidIngredientGroupMember{
			CreatedAt:      memberResult.CreatedAt,
			ArchivedAt:     database.TimePointerFromNullTime(memberResult.ArchivedAt),
			ID:             memberResult.ID,
			BelongsToGroup: memberResult.BelongsToGroup,
			ValidIngredient: mealplanning.ValidIngredient{
				CreatedAt:                      memberResult.ValidIngredientCreatedAt,
				LastUpdatedAt:                  database.TimePointerFromNullTime(memberResult.ValidIngredientLastUpdatedAt),
				ArchivedAt:                     database.TimePointerFromNullTime(memberResult.ValidIngredientArchivedAt),
				MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
				MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
				IconPath:                       memberResult.ValidIngredientIconPath,
				Warning:                        memberResult.ValidIngredientWarning,
				PluralName:                     memberResult.ValidIngredientPluralName,
				StorageInstructions:            memberResult.ValidIngredientStorageInstructions,
				Name:                           memberResult.ValidIngredientName,
				ID:                             memberResult.ValidIngredientID,
				Description:                    memberResult.ValidIngredientDescription,
				Slug:                           memberResult.ValidIngredientSlug,
				ShoppingSuggestions:            memberResult.ValidIngredientShoppingSuggestions,
				ContainsShellfish:              memberResult.ValidIngredientContainsShellfish,
				IsLiquid:                       database.BoolFromNullBool(memberResult.ValidIngredientIsLiquid),
				ContainsPeanut:                 memberResult.ValidIngredientContainsPeanut,
				ContainsTreeNut:                memberResult.ValidIngredientContainsTreeNut,
				ContainsEgg:                    memberResult.ValidIngredientContainsEgg,
				ContainsWheat:                  memberResult.ValidIngredientContainsWheat,
				ContainsSoy:                    memberResult.ValidIngredientContainsSoy,
				AnimalDerived:                  memberResult.ValidIngredientAnimalDerived,
				RestrictToPreparations:         memberResult.ValidIngredientRestrictToPreparations,
				ContainsSesame:                 memberResult.ValidIngredientContainsSesame,
				ContainsFish:                   memberResult.ValidIngredientContainsFish,
				ContainsGluten:                 memberResult.ValidIngredientContainsGluten,
				ContainsDairy:                  memberResult.ValidIngredientContainsDairy,
				ContainsAlcohol:                memberResult.ValidIngredientContainsAlcohol,
				AnimalFlesh:                    memberResult.ValidIngredientAnimalFlesh,
				IsStarch:                       memberResult.ValidIngredientIsStarch,
				IsProtein:                      memberResult.ValidIngredientIsProtein,
				IsGrain:                        memberResult.ValidIngredientIsGrain,
				IsFruit:                        memberResult.ValidIngredientIsFruit,
				IsSalt:                         memberResult.ValidIngredientIsSalt,
				IsFat:                          memberResult.ValidIngredientIsFat,
				IsAcid:                         memberResult.ValidIngredientIsAcid,
				IsHeat:                         memberResult.ValidIngredientIsHeat,
			},
		})
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

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	tracing.AttachQueryFilterToSpan(span, filter)
	filter.AttachToLogger(logger)

	results, err := q.generatedQuerier.SearchForValidIngredientGroups(ctx, q.readDB, &generated.SearchForValidIngredientGroupsParams{
		Name:            query,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
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

		validIngredientGroup := &mealplanning.ValidIngredientGroup{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Slug:          result.Slug,
			Description:   result.Description,
			Members:       []*mealplanning.ValidIngredientGroupMember{},
		}

		var membersResults []*generated.GetValidIngredientGroupMembersRow
		membersResults, err = q.generatedQuerier.GetValidIngredientGroupMembers(ctx, q.readDB, result.ID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group members from database")
		}

		for _, memberResult := range membersResults {
			validIngredientGroup.Members = append(validIngredientGroup.Members, &mealplanning.ValidIngredientGroupMember{
				CreatedAt:      memberResult.CreatedAt,
				ArchivedAt:     database.TimePointerFromNullTime(memberResult.ArchivedAt),
				ID:             memberResult.ID,
				BelongsToGroup: memberResult.BelongsToGroup,
				ValidIngredient: mealplanning.ValidIngredient{
					CreatedAt:                      memberResult.ValidIngredientCreatedAt,
					LastUpdatedAt:                  database.TimePointerFromNullTime(memberResult.ValidIngredientLastUpdatedAt),
					ArchivedAt:                     database.TimePointerFromNullTime(memberResult.ValidIngredientArchivedAt),
					MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
					MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
					IconPath:                       memberResult.ValidIngredientIconPath,
					Warning:                        memberResult.ValidIngredientWarning,
					PluralName:                     memberResult.ValidIngredientPluralName,
					StorageInstructions:            memberResult.ValidIngredientStorageInstructions,
					Name:                           memberResult.ValidIngredientName,
					ID:                             memberResult.ValidIngredientID,
					Description:                    memberResult.ValidIngredientDescription,
					Slug:                           memberResult.ValidIngredientSlug,
					ShoppingSuggestions:            memberResult.ValidIngredientShoppingSuggestions,
					ContainsShellfish:              memberResult.ValidIngredientContainsShellfish,
					IsLiquid:                       database.BoolFromNullBool(memberResult.ValidIngredientIsLiquid),
					ContainsPeanut:                 memberResult.ValidIngredientContainsPeanut,
					ContainsTreeNut:                memberResult.ValidIngredientContainsTreeNut,
					ContainsEgg:                    memberResult.ValidIngredientContainsEgg,
					ContainsWheat:                  memberResult.ValidIngredientContainsWheat,
					ContainsSoy:                    memberResult.ValidIngredientContainsSoy,
					AnimalDerived:                  memberResult.ValidIngredientAnimalDerived,
					RestrictToPreparations:         memberResult.ValidIngredientRestrictToPreparations,
					ContainsSesame:                 memberResult.ValidIngredientContainsSesame,
					ContainsFish:                   memberResult.ValidIngredientContainsFish,
					ContainsGluten:                 memberResult.ValidIngredientContainsGluten,
					ContainsDairy:                  memberResult.ValidIngredientContainsDairy,
					ContainsAlcohol:                memberResult.ValidIngredientContainsAlcohol,
					AnimalFlesh:                    memberResult.ValidIngredientAnimalFlesh,
					IsStarch:                       memberResult.ValidIngredientIsStarch,
					IsProtein:                      memberResult.ValidIngredientIsProtein,
					IsGrain:                        memberResult.ValidIngredientIsGrain,
					IsFruit:                        memberResult.ValidIngredientIsFruit,
					IsSalt:                         memberResult.ValidIngredientIsSalt,
					IsFat:                          memberResult.ValidIngredientIsFat,
					IsAcid:                         memberResult.ValidIngredientIsAcid,
					IsHeat:                         memberResult.ValidIngredientIsHeat,
				},
			})
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

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetValidIngredientGroups(ctx, q.readDB, &generated.GetValidIngredientGroupsParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
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
		validIngredientGroup := &mealplanning.ValidIngredientGroup{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Name:          result.Name,
			Slug:          result.Slug,
			Description:   result.Description,
			Members:       []*mealplanning.ValidIngredientGroupMember{},
		}

		var membersResults []*generated.GetValidIngredientGroupMembersRow
		membersResults, err = q.generatedQuerier.GetValidIngredientGroupMembers(ctx, q.readDB, result.ID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "fetching valid ingredients group members from database")
		}

		for _, memberResult := range membersResults {
			validIngredientGroup.Members = append(validIngredientGroup.Members, &mealplanning.ValidIngredientGroupMember{
				CreatedAt:      memberResult.CreatedAt,
				ArchivedAt:     database.TimePointerFromNullTime(memberResult.ArchivedAt),
				ID:             memberResult.ID,
				BelongsToGroup: memberResult.BelongsToGroup,
				ValidIngredient: mealplanning.ValidIngredient{
					CreatedAt:                      memberResult.ValidIngredientCreatedAt,
					LastUpdatedAt:                  database.TimePointerFromNullTime(memberResult.ValidIngredientLastUpdatedAt),
					ArchivedAt:                     database.TimePointerFromNullTime(memberResult.ValidIngredientArchivedAt),
					MinStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMinimumIdealStorageTemperatureInCelsius),
					MaxStorageTemperatureInCelsius: database.Float32PointerFromNullString(memberResult.ValidIngredientMaximumIdealStorageTemperatureInCelsius),
					IconPath:                       memberResult.ValidIngredientIconPath,
					Warning:                        memberResult.ValidIngredientWarning,
					PluralName:                     memberResult.ValidIngredientPluralName,
					StorageInstructions:            memberResult.ValidIngredientStorageInstructions,
					Name:                           memberResult.ValidIngredientName,
					ID:                             memberResult.ValidIngredientID,
					Description:                    memberResult.ValidIngredientDescription,
					Slug:                           memberResult.ValidIngredientSlug,
					ShoppingSuggestions:            memberResult.ValidIngredientShoppingSuggestions,
					ContainsShellfish:              memberResult.ValidIngredientContainsShellfish,
					IsLiquid:                       database.BoolFromNullBool(memberResult.ValidIngredientIsLiquid),
					ContainsPeanut:                 memberResult.ValidIngredientContainsPeanut,
					ContainsTreeNut:                memberResult.ValidIngredientContainsTreeNut,
					ContainsEgg:                    memberResult.ValidIngredientContainsEgg,
					ContainsWheat:                  memberResult.ValidIngredientContainsWheat,
					ContainsSoy:                    memberResult.ValidIngredientContainsSoy,
					AnimalDerived:                  memberResult.ValidIngredientAnimalDerived,
					RestrictToPreparations:         memberResult.ValidIngredientRestrictToPreparations,
					ContainsSesame:                 memberResult.ValidIngredientContainsSesame,
					ContainsFish:                   memberResult.ValidIngredientContainsFish,
					ContainsGluten:                 memberResult.ValidIngredientContainsGluten,
					ContainsDairy:                  memberResult.ValidIngredientContainsDairy,
					ContainsAlcohol:                memberResult.ValidIngredientContainsAlcohol,
					AnimalFlesh:                    memberResult.ValidIngredientAnimalFlesh,
					IsStarch:                       memberResult.ValidIngredientIsStarch,
					IsProtein:                      memberResult.ValidIngredientIsProtein,
					IsGrain:                        memberResult.ValidIngredientIsGrain,
					IsFruit:                        memberResult.ValidIngredientIsFruit,
					IsSalt:                         memberResult.ValidIngredientIsSalt,
					IsFat:                          memberResult.ValidIngredientIsFat,
					IsAcid:                         memberResult.ValidIngredientIsAcid,
					IsHeat:                         memberResult.ValidIngredientIsHeat,
				},
			})
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
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, input.ID)

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "starting transaction")
	}

	// create the valid ingredient group.
	if err = q.generatedQuerier.CreateValidIngredientGroup(ctx, tx, &generated.CreateValidIngredientGroupParams{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Slug:        input.Slug,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing valid ingredient group creation query")
	}

	x := &mealplanning.ValidIngredientGroup{
		ID:          input.ID,
		Name:        input.Name,
		Description: input.Description,
		Slug:        input.Slug,
		CreatedAt:   q.CurrentTime(),
	}

	for i := range input.Members {
		m := input.Members[i]
		var member *mealplanning.ValidIngredientGroupMember
		member, err = q.CreateValidIngredientGroupMember(ctx, tx, x.ID, m)
		if err != nil {
			q.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareAndLogError(err, logger, span, "creating valid ingredient group member")
		}

		x.Members = append(x.Members, member)
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.WithValue("member_count", len(input.Members)).Info("valid ingredient group created")

	return x, nil
}

// CreateValidIngredientGroupMember creates a valid ingredient group member in the database.
func (q *repository) CreateValidIngredientGroupMember(ctx context.Context, db database.SQLQueryExecutorAndTransactionManager, groupID string, input *mealplanning.ValidIngredientGroupMemberDatabaseCreationInput) (*mealplanning.ValidIngredientGroupMember, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
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
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, updated.ID)

	if _, err := q.generatedQuerier.UpdateValidIngredientGroup(ctx, q.writeDB, &generated.UpdateValidIngredientGroupParams{
		Name:        updated.Name,
		Description: updated.Description,
		Slug:        updated.Slug,
		ID:          updated.ID,
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

	logger := q.logger.Clone()

	if validIngredientGroupID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)
	tracing.AttachToSpan(span, mealplanningkeys.ValidIngredientGroupIDKey, validIngredientGroupID)

	rowsAffected, err := q.generatedQuerier.ArchiveValidIngredientGroup(ctx, q.writeDB, validIngredientGroupID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving valid ingredient group")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	return nil
}
