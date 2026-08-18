package mealplanning

import (
	"context"
	"database/sql"
	"fmt"
	"math/rand/v2"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v11/database"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"

	"resenje.org/schulze"
)

var (
	_ mealplanning.MealPlanOptionDataManager = (*repository)(nil)
)

// MealPlanOptionExists fetches whether a meal plan option exists from the database.
func (q *repository) MealPlanOptionExists(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if mealPlanOptionID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	result, err := q.generatedQuerier.CheckMealPlanOptionExistence(ctx, q.readDB, &generated.CheckMealPlanOptionExistenceParams{
		MealPlanEventID:  database.NullStringFromString(mealPlanEventID),
		MealPlanOptionID: mealPlanOptionID,
		MealPlanID:       mealPlanID,
	})
	if err != nil {
		logger.Error("performing meal plan option existence check", err)
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan option existence check")
	}

	return result, nil
}

// MealPlanOptionBelongsToAccount checks whether a meal plan option resolves (via its event and meal plan) to the given account.
func (q *repository) MealPlanOptionBelongsToAccount(ctx context.Context, mealPlanOptionID, accountID string) (belongs bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.generatedQuerier.CheckMealPlanOptionBelongsToAccount(ctx, q.readDB, &generated.CheckMealPlanOptionBelongsToAccountParams{
		MealPlanOptionID: mealPlanOptionID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan option account ownership check")
	}

	return result, nil
}

// GetMealPlanOption fetches a meal plan option from the database.
func (q *repository) GetMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if mealPlanOptionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	result, err := q.generatedQuerier.GetMealPlanOption(ctx, q.readDB, &generated.GetMealPlanOptionParams{
		MealPlanID:       mealPlanID,
		MealPlanEventID:  database.NullStringFromString(mealPlanEventID),
		MealPlanOptionID: mealPlanOptionID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan option query")
	}

	mealPlanOption := &mealplanning.MealPlanOption{
		CreatedAt:              result.CreatedAt,
		LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
		AssignedCook:           database.StringPointerFromNullString(result.AssignedCook),
		ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
		AssignedDishwasher:     database.StringPointerFromNullString(result.AssignedDishwasher),
		Notes:                  result.Notes,
		BelongsToMealPlanEvent: database.StringFromNullString(result.BelongsToMealPlanEvent),
		ID:                     result.ID,
		Meal: mealplanning.Meal{
			CreatedAt:            result.MealCreatedAt,
			ArchivedAt:           database.TimePointerFromNullTime(result.MealArchivedAt),
			LastUpdatedAt:        database.TimePointerFromNullTime(result.MealLastUpdatedAt),
			ID:                   result.MealID,
			Description:          result.MealDescription,
			CreatedByUser:        result.MealCreatedByUser,
			Name:                 result.MealName,
			MinEstimatedPortions: database.Float32FromString(result.MealMinEstimatedPortions),
			MaxEstimatedPortions: database.Float32PointerFromNullString(result.MealMaxEstimatedPortions),
			EligibleForMealPlans: result.MealEligibleForMealPlans,
		},
		MealScale: database.Float32FromString(result.MealScale),
		Chosen:    result.Chosen,
		TieBroken: result.Tiebroken,
	}

	return mealPlanOption, nil
}

// getMealPlanOptionByID fetches a meal plan option from the database by its ID.
func (q *repository) getMealPlanOptionByID(ctx context.Context, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanOptionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	result, err := q.generatedQuerier.GetMealPlanOptionByID(ctx, q.readDB, mealPlanOptionID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan option query")
	}

	mealPlanOption := &mealplanning.MealPlanOption{
		CreatedAt:              result.CreatedAt,
		LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
		AssignedCook:           database.StringPointerFromNullString(result.AssignedCook),
		ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
		AssignedDishwasher:     database.StringPointerFromNullString(result.AssignedDishwasher),
		Notes:                  result.Notes,
		BelongsToMealPlanEvent: database.StringFromNullString(result.BelongsToMealPlanEvent),
		ID:                     result.ID,
		Votes:                  nil,
		Meal: mealplanning.Meal{
			CreatedAt:            result.MealCreatedAt,
			ArchivedAt:           database.TimePointerFromNullTime(result.MealArchivedAt),
			LastUpdatedAt:        database.TimePointerFromNullTime(result.MealLastUpdatedAt),
			MinEstimatedPortions: database.Float32FromString(result.MealMinEstimatedPortions),
			MaxEstimatedPortions: database.Float32PointerFromNullString(result.MealMaxEstimatedPortions),
			ID:                   result.MealID,
			Description:          result.MealDescription,
			CreatedByUser:        result.MealCreatedByUser,
			Name:                 result.MealName,
			Components:           []*mealplanning.MealComponent{},
			EligibleForMealPlans: result.MealEligibleForMealPlans,
		},
		MealScale: database.Float32FromString(result.MealScale),
		Chosen:    result.Chosen,
		TieBroken: result.Tiebroken,
	}

	return mealPlanOption, nil
}

// getMealPlanOptionsForMealPlanEvent fetches a list of meal plan options from the database that meet a particular filter.
func (q *repository) getMealPlanOptionsForMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) ([]*mealplanning.MealPlanOption, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanEventID)

	results, err := q.generatedQuerier.GetAllMealPlanOptionsForMealPlanEvent(ctx, q.readDB, &generated.GetAllMealPlanOptionsForMealPlanEventParams{
		MealPlanID:      mealPlanID,
		MealPlanEventID: database.NullStringFromString(mealPlanEventID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan option query")
	}

	x := []*mealplanning.MealPlanOption{}
	for _, result := range results {
		meal, mealErr := q.GetMeal(ctx, result.MealID)
		if mealErr != nil {
			return nil, observability.PrepareAndLogError(mealErr, logger, span, "getting meal for meal plan")
		}

		x = append(x, &mealplanning.MealPlanOption{
			CreatedAt:              result.CreatedAt,
			LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
			AssignedCook:           database.StringPointerFromNullString(result.AssignedCook),
			ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
			AssignedDishwasher:     database.StringPointerFromNullString(result.AssignedDishwasher),
			Notes:                  result.Notes,
			BelongsToMealPlanEvent: database.StringFromNullString(result.BelongsToMealPlanEvent),
			ID:                     result.ID,
			Votes:                  nil,
			Meal:                   *meal,
			MealScale:              database.Float32FromString(result.MealScale),
			Chosen:                 result.Chosen,
			TieBroken:              result.Tiebroken,
		})
	}

	for i, opt := range x {
		votes, voteFetchErr := q.GetMealPlanOptionVotesForMealPlanOption(ctx, mealPlanID, mealPlanEventID, opt.ID)
		if voteFetchErr != nil {
			return nil, observability.PrepareError(voteFetchErr, span, "fetching meal plan option votes for meal plan option")
		}
		x[i].Votes = votes

		m, mealFetchErr := q.GetMeal(ctx, opt.Meal.ID)
		if mealFetchErr != nil {
			return nil, observability.PrepareAndLogError(mealFetchErr, logger, span, "scanning meal plan options for meal plan event")
		}
		x[i].Meal = *m
	}

	logger.WithValue("quantity", len(x)).Info("fetched meal plan options for meal plan event")

	return x, nil
}

// GetMealPlanOptions fetches a list of meal plan options from the database that meet a particular filter.
func (q *repository) GetMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.MealPlanOption], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*mealplanning.MealPlanOption
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetMealPlanOptions(ctx, q.readDB, &generated.GetMealPlanOptionsParams{
		MealPlanID:      mealPlanID,
		MealPlanEventID: database.NullStringFromString(mealPlanEventID),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan options list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &mealplanning.MealPlanOption{
			CreatedAt:              result.CreatedAt,
			LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
			AssignedCook:           database.StringPointerFromNullString(result.AssignedCook),
			ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
			AssignedDishwasher:     database.StringPointerFromNullString(result.AssignedDishwasher),
			Notes:                  result.Notes,
			BelongsToMealPlanEvent: database.StringFromNullString(result.BelongsToMealPlanEvent),
			ID:                     result.ID,
			Votes:                  nil,
			Meal: mealplanning.Meal{
				ID: result.MealID,
			},
			MealScale: database.Float32FromString(result.MealScale),
			Chosen:    result.Chosen,
			TieBroken: result.Tiebroken,
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mpo *mealplanning.MealPlanOption) string { return mpo.ID },
		filter,
	)

	return x, nil
}

// MealExistsAsOptionInEvent returns true if the meal already exists as a non-archived option for the event.
func (q *repository) MealExistsAsOptionInEvent(ctx context.Context, mealPlanEventID, mealID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanEventID == "" || mealID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID).WithValue(mealplanningkeys.MealIDKey, mealID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealIDKey, mealID)

	result, err := q.generatedQuerier.CheckMealInMealPlanEvent(ctx, q.readDB, &generated.CheckMealInMealPlanEventParams{
		BelongsToMealPlanEvent: database.NullStringFromString(mealPlanEventID),
		MealID:                 mealID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "checking if meal exists as option in event")
	}
	return result, nil
}

// createMealPlanOption creates a meal plan option in the database.
func (q *repository) createMealPlanOption(ctx context.Context, db database.SQLQueryExecutor, input *mealplanning.MealPlanOptionDatabaseCreationInput, markAsChosen bool) (*mealplanning.MealPlanOption, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, input.ID)
	logger := q.logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, input.ID)

	// create the meal plan option.
	if err := q.generatedQuerier.CreateMealPlanOption(ctx, db, &generated.CreateMealPlanOptionParams{
		ID:                     input.ID,
		AssignedCook:           database.NullStringFromStringPointer(input.AssignedCook),
		AssignedDishwasher:     database.NullStringFromStringPointer(input.AssignedDishwasher),
		MealID:                 input.MealID,
		Notes:                  input.Notes,
		MealScale:              database.StringFromFloat32(input.MealScale),
		BelongsToMealPlanEvent: database.NullStringFromString(input.BelongsToMealPlanEvent),
		Chosen:                 markAsChosen,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan option")
	}

	x := &mealplanning.MealPlanOption{
		ID:                     input.ID,
		AssignedCook:           input.AssignedCook,
		Meal:                   mealplanning.Meal{ID: input.MealID},
		Notes:                  input.Notes,
		BelongsToMealPlanEvent: input.BelongsToMealPlanEvent,
		CreatedAt:              q.CurrentTime(),
		MealScale:              input.MealScale,
		Votes:                  []*mealplanning.MealPlanOptionVote{},
	}

	logger.Info("meal plan option created")

	return x, nil
}

// CreateMealPlanOption creates a meal plan option in the database.
func (q *repository) CreateMealPlanOption(ctx context.Context, input *mealplanning.MealPlanOptionDatabaseCreationInput) (*mealplanning.MealPlanOption, error) {
	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	var created *mealplanning.MealPlanOption

	// The write and its event share a transaction.
	if err := q.withEvent(ctx, q.logger, mealplanning.MealPlanOptionCreatedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanOptionIDKey: input.ID,
	}, func(tx database.SQLQueryExecutor) error {
		var createErr error
		created, createErr = q.createMealPlanOption(ctx, tx, input, false)

		return createErr
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// UpdateMealPlanOption updates a particular meal plan option.
func (q *repository) UpdateMealPlanOption(ctx context.Context, updated *mealplanning.MealPlanOption) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, mealplanning.MealPlanOptionUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanEventIDKey:  updated.BelongsToMealPlanEvent,
		mealplanningkeys.MealPlanOptionIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateMealPlanOption(ctx, tx, &generated.UpdateMealPlanOptionParams{
			MealID:             updated.Meal.ID,
			Notes:              updated.Notes,
			MealScale:          database.StringFromFloat32(updated.MealScale),
			MealPlanOptionID:   updated.ID,
			AssignedCook:       database.NullStringFromStringPointer(updated.AssignedCook),
			AssignedDishwasher: database.NullStringFromStringPointer(updated.AssignedDishwasher),
			MealPlanEventID:    database.NullStringFromString(updated.BelongsToMealPlanEvent),
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan option")
	}

	logger.Info("meal plan option updated")

	return nil
}

// ArchiveMealPlanOption archives a meal plan option from the database by its ID.
func (q *repository) ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if mealPlanOptionID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if err := q.withEvent(ctx, logger, mealplanning.MealPlanOptionArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveMealPlanOption(ctx, tx, &generated.ArchiveMealPlanOptionParams{
			ID:                     mealPlanOptionID,
			BelongsToMealPlanEvent: sql.NullString{String: mealPlanEventID, Valid: true},
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving meal plan option")
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

func (q *repository) determineWinner(winners []schulze.Result[string]) string {
	var (
		highestScore int
		scoreWinners []string
	)

	for _, winner := range winners {
		if winner.Wins == highestScore {
			scoreWinners = append(scoreWinners, winner.Choice)
		} else if winner.Wins > highestScore {
			highestScore = winner.Wins
			scoreWinners = []string{winner.Choice}
		}
	}

	/* #nosec: G404 */
	return scoreWinners[rand.N(len(scoreWinners))]
}

func (q *repository) decideOptionWinner(ctx context.Context, options []*mealplanning.MealPlanOption) (_ string, _, _ bool) {
	_, span := q.tracer.StartSpan(ctx)
	defer span.End()

	candidateMap := map[string]struct{}{}
	votesByUser := map[string]schulze.Ballot[string]{}

	logger := q.logger.WithValue("options.count", len(options))

	for _, option := range options {
		for _, v := range option.Votes {
			if votesByUser[v.ByUser] == nil {
				votesByUser[v.ByUser] = schulze.Ballot[string]{}
			}

			if !v.Abstain {
				votesByUser[v.ByUser][v.BelongsToMealPlanOption] = int(v.Rank)
			}

			candidateMap[v.BelongsToMealPlanOption] = struct{}{}
		}
	}

	candidates := []string{}
	for c := range candidateMap {
		candidates = append(candidates, c)
	}

	e := schulze.NewVoting(candidates)
	for _, vote := range votesByUser {
		if _, err := e.Vote(vote); err != nil {
			// this actually can never happen because we use uints for ranks, lol
			observability.AcknowledgeError(err, logger, span, "an invalid vote was received")
		}
	}

	winners, _, tie := e.Compute()
	if tie {
		return q.determineWinner(winners), true, true
	}

	if len(winners) > 0 {
		return winners[0].Choice, false, true
	}

	return "", false, false
}

// FinalizeMealPlanOption archives a meal plan option vote from the database by its ID.
func (q *repository) FinalizeMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, accountID string) (changed bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanEventID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)

	if mealPlanOptionID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	mealPlan, err := q.GetMealPlan(ctx, mealPlanID, accountID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "fetching meal plan")
	}

	var (
		mealPlanEvent  *mealplanning.MealPlanEvent
		mealPlanOption *mealplanning.MealPlanOption
	)
	for _, event := range mealPlan.Events {
		if event.ID == mealPlanEventID {
			mealPlanEvent = event
			for _, option := range event.Options {
				if option.ID == mealPlanOptionID {
					mealPlanOption = option
					break
				}
			}
		}
	}

	if mealPlanEvent == nil {
		return false, fmt.Errorf("meal plan event %s for meal plan %s not found", mealPlanEventID, mealPlanID)
	}

	// fetch account data
	account, err := q.identityRepo.GetAccount(ctx, mealPlan.BelongsToAccount)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "fetching account")
	}

	// go through all the votes for this meal plan option and determine if they're all there
	for _, member := range account.Members {
		memberVoteFound := false
		for _, vote := range mealPlanOption.Votes {
			if vote.ByUser == member.BelongsToUser.ID {
				memberVoteFound = true
				break
			}
		}

		if !memberVoteFound {
			return false, nil
		}
	}

	winner, tiebroken, chosen := q.decideOptionWinner(ctx, mealPlanEvent.Options)
	if chosen {
		if err = q.generatedQuerier.FinalizeMealPlanOption(ctx, q.readDB, &generated.FinalizeMealPlanOptionParams{
			MealPlanEventID: database.NullStringFromString(mealPlanEventID),
			ID:              winner,
			Tiebroken:       tiebroken,
		}); err != nil {
			return false, observability.PrepareAndLogError(err, logger, span, "finalizing meal plan option")
		}

		logger.Debug("finalized meal plan option")
	}

	return chosen, nil
}
