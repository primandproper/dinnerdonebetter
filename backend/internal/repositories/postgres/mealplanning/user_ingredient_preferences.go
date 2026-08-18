package mealplanning

import (
	"context"
	"database/sql"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v11/database"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

var (
	_ mealplanning.UserIngredientPreferenceDataManager = (*repository)(nil)
)

// UserIngredientPreferenceExists fetches whether a user ingredient preference exists from the database.
func (q *repository) UserIngredientPreferenceExists(ctx context.Context, userIngredientPreferenceID, userID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userIngredientPreferenceID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)

	if userID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	exists, err = q.generatedQuerier.CheckUserIngredientPreferenceExistence(ctx, q.readDB, &generated.CheckUserIngredientPreferenceExistenceParams{
		ID:            userIngredientPreferenceID,
		BelongsToUser: userID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing user ingredient preference existence check")
	}

	return exists, nil
}

// GetUserIngredientPreference fetches a user ingredient preference from the database.
func (q *repository) GetUserIngredientPreference(ctx context.Context, userIngredientPreferenceID, userID string) (*mealplanning.UserIngredientPreference, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userIngredientPreferenceID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	result, err := q.generatedQuerier.GetUserIngredientPreference(ctx, q.readDB, &generated.GetUserIngredientPreferenceParams{
		ID:            userIngredientPreferenceID,
		BelongsToUser: userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "scanning userIngredientPreference")
	}

	userIngredientPreference := &mealplanning.UserIngredientPreference{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		Notes:         result.Notes,
		CreatedByUser: result.BelongsToUser,
		Rating:        int8(result.Rating),
		Allergy:       result.Allergy,
		Ingredient: mealplanning.ValidIngredient{
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

	return userIngredientPreference, nil
}

// GetUserIngredientPreferences fetches a list of user ingredient preferences from the database that meet a particular filter.
func (q *repository) GetUserIngredientPreferences(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := q.generatedQuerier.GetUserIngredientPreferencesForUser(ctx, q.readDB, &generated.GetUserIngredientPreferencesForUserParams{
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		BelongsToUser:   userID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing user ingredient preferences list retrieval query")
	}

	var (
		data          []*mealplanning.UserIngredientPreference
		filteredCount uint64
		totalCount    uint64
	)

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.UserIngredientPreference{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			Notes:         result.Notes,
			CreatedByUser: result.BelongsToUser,
			Rating:        int8(result.Rating),
			Allergy:       result.Allergy,
			Ingredient: mealplanning.ValidIngredient{
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
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(uip *mealplanning.UserIngredientPreference) string { return uip.ID },
		filter,
	)

	return x, nil
}

// CreateUserIngredientPreference creates a user ingredient preference in the database.
func (q *repository) CreateUserIngredientPreference(ctx context.Context, input *mealplanning.UserIngredientPreferenceDatabaseCreationInput) ([]*mealplanning.UserIngredientPreference, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.ValidIngredientIDKey, input.ValidIngredientID)

	validIngredientIDs := []string{}
	if input.ValidIngredientGroupID != "" {
		group, err := q.GetValidIngredientGroup(ctx, input.ValidIngredientGroupID)
		if err != nil {
			return nil, observability.PrepareAndLogError(err, logger, span, "getting valid ingredient group")
		}

		for _, member := range group.Members {
			validIngredientIDs = append(validIngredientIDs, member.ValidIngredient.ID)
		}
	} else {
		validIngredientIDs = append(validIngredientIDs, input.ValidIngredientID)
	}

	logger = logger.WithValue("valid_ingredient_ids", validIngredientIDs)
	logger.Debug("creating user ingredient preferences")

	output := []*mealplanning.UserIngredientPreference{}
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		for _, validIngredientID := range validIngredientIDs {
			l := logger.WithValue(mealplanningkeys.ValidIngredientIDKey, validIngredientID)
			if validIngredientID == "" {
				continue
			}

			id := identifiers.New()
			tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, id)

			// create the user ingredient preference.
			if err := q.generatedQuerier.CreateUserIngredientPreference(ctx, tx, &generated.CreateUserIngredientPreferenceParams{
				ID:            id,
				Ingredient:    validIngredientID,
				Notes:         input.Notes,
				BelongsToUser: input.CreatedByUser,
				Rating:        int16(input.Rating),
				Allergy:       input.Allergy,
			}); err != nil {
				return observability.PrepareAndLogError(err, l, span, "performing user ingredient preference creation query")
			}

			x := &mealplanning.UserIngredientPreference{
				ID:            id,
				Rating:        input.Rating,
				Notes:         input.Notes,
				Allergy:       input.Allergy,
				CreatedByUser: input.CreatedByUser,
				Ingredient:    mealplanning.ValidIngredient{ID: input.ValidIngredientID},
				CreatedAt:     q.CurrentTime(),
			}

			l.Info("user ingredient preference created")

			output = append(output, x)
		}

		// The event is another statement in this transaction, so it commits with the
		// preferences it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, mealplanning.UserIngredientPreferenceCreatedServiceEventType, "", map[string]any{
			mealplanningkeys.ValidIngredientGroupIDKey: input.ValidIngredientGroupID,
			mealplanningkeys.ValidIngredientIDKey:      input.ValidIngredientID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing user ingredient preference created event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return output, nil
}

// UpdateUserIngredientPreference updates a particular user ingredient preference.
func (q *repository) UpdateUserIngredientPreference(ctx context.Context, updated *mealplanning.UserIngredientPreference) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.UserIngredientPreferenceIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.UserIngredientPreferenceUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateUserIngredientPreference(ctx, tx, &generated.UpdateUserIngredientPreferenceParams{
			Ingredient:    updated.Ingredient.ID,
			Notes:         updated.Notes,
			ID:            updated.ID,
			BelongsToUser: updated.CreatedByUser,
			Rating:        int16(updated.Rating),
			Allergy:       updated.Allergy,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating user ingredient preference")
	}

	logger.Info("user ingredient preference updated")

	return nil
}

// ArchiveUserIngredientPreference archives a user ingredient preference from the database by its ID.
func (q *repository) ArchiveUserIngredientPreference(ctx context.Context, userIngredientPreferenceID, userID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if userIngredientPreferenceID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)
	tracing.AttachToSpan(span, mealplanningkeys.UserIngredientPreferenceIDKey, userIngredientPreferenceID)

	if err := q.withEvent(ctx, logger, mealplanning.UserIngredientPreferenceArchivedServiceEventType, userID, map[string]any{
		mealplanningkeys.UserIngredientPreferenceIDKey: userIngredientPreferenceID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveUserIngredientPreference(ctx, tx, &generated.ArchiveUserIngredientPreferenceParams{
			ID:            userIngredientPreferenceID,
			BelongsToUser: userID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving user ingredient preference")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
