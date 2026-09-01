package managers

import (
	"context"

	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

func (m *mealPlanningManager) ListMealPlans(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.AccountIDKey, ownerID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, ownerID)

	mealPlans, err := m.db.GetMealPlansForAccount(ctx, ownerID, filter)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching list of meal plans for account")
	}

	return mealPlans, nil
}

// AnnotateMealPlanSummaries reads the chosen meal names and the user's votes for a whole
// page of meal plans, in one query each. The list endpoint projects both onto the
// MealPlanSummary it answers with; see the message's comment in
// mealplanning_messages.proto for why the options they come from are not on the wire.
func (m *mealPlanningManager) AnnotateMealPlanSummaries(ctx context.Context, userID string, mealPlans []*types.MealPlan) (*types.MealPlanSummaryAnnotations, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	logger := m.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	annotations := &types.MealPlanSummaryAnnotations{
		ChosenMealNamesByEventID: map[string]string{},
		VotedOnMealPlanIDs:       map[string]bool{},
	}

	mealPlanIDs := make([]string, 0, len(mealPlans))
	for _, mealPlan := range mealPlans {
		mealPlanIDs = append(mealPlanIDs, mealPlan.ID)
	}

	if len(mealPlanIDs) == 0 {
		return annotations, nil
	}

	chosenMealNames, err := m.db.GetChosenMealNamesForMealPlans(ctx, mealPlanIDs)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching chosen meal names for meal plans")
	}
	annotations.ChosenMealNamesByEventID = chosenMealNames

	votedOn, err := m.db.GetMealPlanIDsVotedOnByUser(ctx, userID, mealPlanIDs)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plans voted on by user")
	}
	for _, mealPlanID := range votedOn {
		annotations.VotedOnMealPlanIDs[mealPlanID] = true
	}

	return annotations, nil
}

func (m *mealPlanningManager) CreateMealPlan(ctx context.Context, ownerID, creatorID string, input *types.MealPlanCreationRequestInput) (*types.MealPlan, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareError(err, span, "validating input")
	}

	if creatorID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	if ownerID == "" {
		return nil, platformerrors.ErrEmptyInputParameter
	}

	convertedInput := converters.ConvertMealPlanCreationRequestInputToMealPlanDatabaseCreationInput(input)
	convertedInput.CreatedByUser = creatorID
	convertedInput.BelongsToAccount = ownerID

	logger := m.logger.WithSpan(span).WithValue(mealplanningkeys.MealPlanIDKey, convertedInput.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, convertedInput.ID)

	created, err := m.db.CreateMealPlan(ctx, convertedInput)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan")
	}

	if created.Status == string(types.MealPlanStatusFinalized) {
		m.startFinalizationPipeline(ctx, created.ID, ownerID, logger, span)
	}

	return created, nil
}

func (m *mealPlanningManager) ReadMealPlan(ctx context.Context, mealPlanID, ownerID string) (*types.MealPlan, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: mealPlanID,
		identitykeys.UserIDKey:         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	mealPlan, err := m.db.GetMealPlan(ctx, mealPlanID, ownerID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan")
	}

	return mealPlan, nil
}

func (m *mealPlanningManager) UpdateMealPlan(ctx context.Context, mealPlanID, ownerID string, input *types.MealPlanUpdateRequestInput) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputParameter
	}

	if err := input.ValidateWithContext(ctx); err != nil {
		return observability.PrepareError(err, span, "validating input")
	}

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: mealPlanID,
		identitykeys.UserIDKey:         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	existingMealPlan, err := m.db.GetMealPlan(ctx, mealPlanID, ownerID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan to update")
	}

	existingMealPlan.Update(input)
	if err = m.db.UpdateMealPlan(ctx, existingMealPlan); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan")
	}

	return nil
}

func (m *mealPlanningManager) ArchiveMealPlan(ctx context.Context, mealPlanID, ownerID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: mealPlanID,
		identitykeys.UserIDKey:         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	if err := m.db.ArchiveMealPlan(ctx, mealPlanID, ownerID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan")
	}

	return nil
}

func (m *mealPlanningManager) FinalizeMealPlan(ctx context.Context, mealPlanID, ownerID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValues(map[string]any{
		mealplanningkeys.MealPlanIDKey: mealPlanID,
		identitykeys.UserIDKey:         ownerID,
	})
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, ownerID)

	finalized, err := m.db.AttemptToFinalizeMealPlan(ctx, mealPlanID, ownerID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "finalizing meal plan")
	}

	// only enter the plan into the pipeline when it actually finalized.
	if finalized {
		m.startFinalizationPipeline(ctx, mealPlanID, ownerID, logger, span)
	}

	return finalized, nil
}
