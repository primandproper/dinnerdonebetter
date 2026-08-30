package mealplanning

import (
	"context"
	"database/sql"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

var (
	_ types.MealPlanOptionVoteDataManager = (*repository)(nil)
)

// MealPlanOptionVoteExists fetches whether a meal plan option vote exists from the database.
func (q *repository) MealPlanOptionVoteExists(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (exists bool, err error) {
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

	if mealPlanOptionVoteID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	result, err := q.generatedQuerier.CheckMealPlanOptionVoteExistence(ctx, q.readDB, &generated.CheckMealPlanOptionVoteExistenceParams{
		MealPlanOptionID:     mealPlanOptionID,
		MealPlanOptionVoteID: mealPlanOptionVoteID,
		MealPlanEventID:      database.NullStringFromString(mealPlanEventID),
		MealPlanID:           mealPlanID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan option vote existence check")
	}

	return result, nil
}

// GetMealPlanOptionVote fetches a meal plan option vote from the database.
func (q *repository) GetMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if mealPlanOptionID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	if mealPlanOptionVoteID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	result, err := q.generatedQuerier.GetMealPlanOptionVote(ctx, q.readDB, &generated.GetMealPlanOptionVoteParams{
		MealPlanOptionID:     mealPlanOptionID,
		MealPlanOptionVoteID: mealPlanOptionVoteID,
		MealPlanID:           mealPlanID,
		MealPlanEventID:      database.NullStringFromString(mealPlanEventID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "getting meal plan option vote")
	}

	mealPlanOptionVote := &types.MealPlanOptionVote{
		CreatedAt:               result.CreatedAt,
		ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:                      result.ID,
		Notes:                   result.Notes,
		BelongsToMealPlanOption: result.BelongsToMealPlanOption,
		ByUser:                  result.ByUser,
		Rank:                    uint8(result.Rank),
		Abstain:                 result.Abstain,
	}

	return mealPlanOptionVote, nil
}

// GetMealPlanOptionVotesForMealPlanOption fetches a list of meal plan option votes from the database that meet a particular filter.
func (q *repository) GetMealPlanOptionVotesForMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (x []*types.MealPlanOptionVote, err error) {
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

	results, err := q.generatedQuerier.GetMealPlanOptionVotesForMealPlanOption(ctx, q.readDB, &generated.GetMealPlanOptionVotesForMealPlanOptionParams{
		MealPlanID:       mealPlanID,
		MealPlanOptionID: mealPlanOptionID,
		MealPlanEventID:  database.NullStringFromString(mealPlanEventID),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan option votes for meal plan option")
	}

	x = make([]*types.MealPlanOptionVote, len(results))
	for i, result := range results {
		x[i] = &types.MealPlanOptionVote{
			CreatedAt:               result.CreatedAt,
			ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:                      result.ID,
			Notes:                   result.Notes,
			BelongsToMealPlanOption: result.BelongsToMealPlanOption,
			ByUser:                  result.ByUser,
			Rank:                    uint8(result.Rank),
			Abstain:                 result.Abstain,
		}
	}

	return x, nil
}

// GetMealPlanOptionVotes fetches a list of meal plan option votes from the database that meet a particular filter.
func (q *repository) GetMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.MealPlanOptionVote], err error) {
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

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.MealPlanOptionVote
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetMealPlanOptionVotes(ctx, q.readDB, &generated.GetMealPlanOptionVotesParams{
		MealPlanID:       mealPlanID,
		MealPlanOptionID: mealPlanOptionID,
		MealPlanEventID:  database.NullStringFromString(mealPlanEventID),
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		PageCursor:       database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan option votes list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.MealPlanOptionVote{
			CreatedAt:               result.CreatedAt,
			ArchivedAt:              database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:           database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:                      result.ID,
			Notes:                   result.Notes,
			BelongsToMealPlanOption: result.BelongsToMealPlanOption,
			ByUser:                  result.ByUser,
			Rank:                    uint8(result.Rank),
			Abstain:                 result.Abstain,
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mpov *types.MealPlanOptionVote) string { return mpov.ID },
		filter,
	)

	return x, nil
}

// CreateMealPlanOptionVote creates a meal plan option vote in the database.
func (q *repository) CreateMealPlanOptionVote(ctx context.Context, input *types.MealPlanOptionVotesDatabaseCreationInput) ([]*types.MealPlanOptionVote, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue("vote_count", len(input.Votes))

	// begin transaction
	var (
		err   error
		votes []*types.MealPlanOptionVote
	)
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		votes = []*types.MealPlanOptionVote{}
		for _, vote := range input.Votes {
			l := logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, vote.BelongsToMealPlanOption).
				WithValue(mealplanningkeys.MealPlanOptionVoteIDKey, vote.ID)

			// create the meal plan option vote.
			if err = q.generatedQuerier.CreateMealPlanOptionVote(ctx, tx, &generated.CreateMealPlanOptionVoteParams{
				ID:                      vote.ID,
				Notes:                   vote.Notes,
				ByUser:                  vote.ByUser,
				BelongsToMealPlanOption: vote.BelongsToMealPlanOption,
				Rank:                    int32(vote.Rank),
				Abstain:                 vote.Abstain,
			}); err != nil {
				return observability.PrepareAndLogError(err, l, span, "creating meal plan option vote")
			}

			x := &types.MealPlanOptionVote{
				ID:                      vote.ID,
				Rank:                    vote.Rank,
				Abstain:                 vote.Abstain,
				Notes:                   vote.Notes,
				ByUser:                  vote.ByUser,
				BelongsToMealPlanOption: vote.BelongsToMealPlanOption,
				CreatedAt:               q.CurrentTime(),
			}

			tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, x.ID)
			l.Info("meal plan option vote created")

			votes = append(votes, x)
		}

		// The summary is a statement about this whole batch, and this transaction is the
		// batch — so it commits with the votes it counts.
		if emitErr := q.events.Emit(ctx, tx, logger, types.MealPlanOptionVoteCreatedServiceEventType, "", map[string]any{
			"vote_count": len(input.Votes),
			"created":    len(votes),
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing meal plan option votes created event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return votes, nil
}

// UpdateMealPlanOptionVote updates a particular meal plan option vote.
func (q *repository) UpdateMealPlanOptionVote(ctx context.Context, updated *types.MealPlanOptionVote) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanOptionVoteIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.MealPlanOptionVoteUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanOptionIDKey:     updated.BelongsToMealPlanOption,
		mealplanningkeys.MealPlanOptionVoteIDKey: updated.ID,
	}, func(tx database.Tx) error {
		_, updateErr := q.generatedQuerier.UpdateMealPlanOptionVote(ctx, tx, &generated.UpdateMealPlanOptionVoteParams{
			Notes:                   updated.Notes,
			ByUser:                  updated.ByUser,
			BelongsToMealPlanOption: updated.BelongsToMealPlanOption,
			ID:                      updated.ID,
			Rank:                    int32(updated.Rank),
			Abstain:                 updated.Abstain,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan option vote")
	}

	logger.Info("meal plan option vote updated")

	return nil
}

// ArchiveMealPlanOptionVote archives a meal plan option vote from the database by its ID.
func (q *repository) ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error {
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

	if mealPlanOptionVoteID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	if err := q.withEvent(ctx, logger, types.MealPlanOptionVoteArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	}, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveMealPlanOptionVote(ctx, tx, &generated.ArchiveMealPlanOptionVoteParams{
			BelongsToMealPlanOption: mealPlanOptionID,
			ID:                      mealPlanOptionVoteID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "updating meal plan option vote")
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
