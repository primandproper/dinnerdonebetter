package managers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/tracing"
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

func (m *mealPlanningManager) CreateMealPlanOptionVotes(ctx context.Context, creatorID string, input *types.MealPlanOptionVoteCreationRequestInput) ([]*types.MealPlanOptionVote, error) {
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

	// NOTE: rejecting votes for events that are no longer eligible for voting (e.g. after
	// finalization) needs the meal plan and event IDs so it can call the purpose-built
	// MealPlanEventIsEligibleForVoting repository method. Those IDs live on the gRPC request
	// (request.MealPlanId / request.MealPlanEventId) but are not threaded into this manager
	// method's signature, and updating the service handler is outside this change's scope.
	// Deferred: thread the IDs through and gate creation on MealPlanEventIsEligibleForVoting.

	convertedInput := converters.ConvertMealPlanOptionVoteCreationRequestInputToMealPlanOptionVotesDatabaseCreationInput(input)
	logger := m.logger.WithSpan(span).WithValue("vote_count", len(input.Votes))
	tracing.AttachToSpan(span, "vote_count", len(input.Votes))

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
