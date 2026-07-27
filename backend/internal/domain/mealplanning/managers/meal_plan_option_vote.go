package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v7/errors"
	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

func (m *mealPlanningManager) ListMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOptionVote], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:       mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:  mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey: mealPlanOptionID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)

	results, err := m.db.GetMealPlanOptionVotes(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan option votes")
	}

	return results, nil
}

func (m *mealPlanningManager) CreateMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, creatorID string, input *types.MealPlanOptionVoteCreationRequestInput) ([]*types.MealPlanOptionVote, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	// ValidateWithContext rejects an empty vote set and requires every vote to carry a
	// BelongsToMealPlanOption, so a caller can no longer submit votes without naming an option.
	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
		"vote_count":                        len(input.Votes),
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, "vote_count", len(input.Votes))

	eligible, err := m.db.MealPlanEventIsEligibleForVoting(ctx, mealPlanID, mealPlanEventID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "checking meal plan event voting eligibility")
	}
	if !eligible {
		return nil, observability.PrepareError(types.ErrMealPlanEventNotEligibleForVoting, span, "meal plan event %s is not eligible for voting", mealPlanEventID)
	}

	// Every vote must target an option that actually belongs to the targeted event. Resolving the
	// option through the (mealPlanID, mealPlanEventID) pair — which the service layer has already
	// verified belongs to the caller's active account — rejects cross-account option IDs.
	checkedOptions := map[string]struct{}{}
	for _, vote := range input.Votes {
		if _, alreadyChecked := checkedOptions[vote.BelongsToMealPlanOption]; alreadyChecked {
			continue
		}
		if _, optionErr := m.db.GetMealPlanOption(ctx, mealPlanID, mealPlanEventID, vote.BelongsToMealPlanOption); optionErr != nil {
			logger.WithValue(mealplanningkeys.MealPlanOptionIDKey, vote.BelongsToMealPlanOption).Error("fetching meal plan option for vote", optionErr)
			return nil, observability.PrepareError(types.ErrMealPlanOptionNotFoundForEvent, span, "meal plan option %s does not belong to event %s", vote.BelongsToMealPlanOption, mealPlanEventID)
		}
		checkedOptions[vote.BelongsToMealPlanOption] = struct{}{}
	}

	convertedInput := converters.ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput(input)

	for i := range input.Votes {
		convertedInput.Votes[i].ByUser = creatorID
	}

	created, err := m.db.CreateMealPlanOptionVote(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "created meal plan option votes")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionVoteCreatedServiceEventType, map[string]any{
		"vote_count": len(input.Votes),
		"created":    len(created),
	}))

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	mealPlanOptionVote, err := m.db.GetMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan option vote")
	}

	return mealPlanOptionVote, nil
}

func (m *mealPlanningManager) UpdateMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string, input *types.MealPlanOptionVoteUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating input")
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	existingMealPlanOptionVote, err := m.db.GetMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan option vote to update")
	}

	existingMealPlanOptionVote.Update(input)
	if err = m.db.UpdateMealPlanOptionVote(ctx, existingMealPlanOptionVote); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan option vote")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionVoteUpdatedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	}))

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, mealPlanEventID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionIDKey, mealPlanOptionID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanOptionVoteIDKey, mealPlanOptionVoteID)

	if err := m.db.ArchiveMealPlanOptionVote(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan option vote")
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, types.MealPlanOptionVoteArchivedServiceEventType, map[string]any{
		mealplanningkeys.MealPlanIDKey:           mealPlanID,
		mealplanningkeys.MealPlanEventIDKey:      mealPlanEventID,
		mealplanningkeys.MealPlanOptionIDKey:     mealPlanOptionID,
		mealplanningkeys.MealPlanOptionVoteIDKey: mealPlanOptionVoteID,
	}))

	return nil
}
