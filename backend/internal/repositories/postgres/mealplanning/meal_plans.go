package mealplanning

import (
	"context"
	"database/sql"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/identifiers"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeMealPlans = "meal_plans"
)

var (
	_ types.MealPlanDataManager = (*repository)(nil)

	ErrAlreadyFinalized = platformerrors.New("meal plan already finalized")

	// ErrFinalizationSagaAlreadyAttached indicates a meal plan that already has a finalization
	// saga. It is the losing side of a race between two starters reading the same page of
	// candidates, not a failure: the plan is being finalized, just not by this caller.
	ErrFinalizationSagaAlreadyAttached = platformerrors.New("meal plan already has a finalization saga")
)

// MealPlanExists fetches whether a meal plan exists from the database.
func (q *repository) MealPlanExists(ctx context.Context, mealPlanID, accountID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.generatedQuerier.CheckMealPlanExistence(ctx, q.readDB, &generated.CheckMealPlanExistenceParams{
		MealPlanID:       mealPlanID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan existence check")
	}

	return result, nil
}

// GetMealPlan fetches a meal plan from the database.
func (q *repository) getMealPlan(ctx context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := q.generatedQuerier.GetMealPlan(ctx, q.readDB, &generated.GetMealPlanParams{
		ID:               mealPlanID,
		BelongsToAccount: accountID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan retrieval")
	}

	mealPlan := &types.MealPlan{
		CreatedAt:              result.CreatedAt,
		VotingDeadline:         result.VotingDeadline,
		ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:                     result.ID,
		Status:                 string(result.Status),
		Notes:                  result.Notes,
		ElectionMethod:         string(result.ElectionMethod),
		BelongsToAccount:       result.BelongsToAccount,
		CreatedByUser:          result.CreatedByUser,
		GroceryListInitialized: result.GroceryListInitialized,
		TasksCreated:           result.TasksCreated,
	}

	events, err := q.getMealPlanEventsForMealPlan(ctx, mealPlanID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "populating meal plan events")
	}

	if events != nil {
		mealPlan.Events = events
	}

	// Populate selections for the meal plan
	selections, err := q.GetSelectionsForMealPlan(ctx, mealPlanID, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching selections for meal plan")
	}
	mealPlan.Selections = selections

	return mealPlan, nil
}

// GetMealPlan fetches a meal plan from the database.
func (q *repository) GetMealPlan(ctx context.Context, mealPlanID, accountID string) (*types.MealPlan, error) {
	return q.getMealPlan(ctx, mealPlanID, accountID)
}

// GetMealPlansForAccount fetches a page of an account's meal plans, each carrying its
// events but not the options hanging off them. It answers the list endpoint, which
// returns a MealPlanSummary.
//
// GetHydratedMealPlansForAccount is the same page with everything attached, for the
// one caller that needs whole records.
func (q *repository) GetMealPlansForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.MealPlan], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.MealPlan
		filteredCount uint64
		totalCount    uint64
	)

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetMealPlansForAccount(ctx, q.readDB, &generated.GetMealPlansForAccountParams{
		BelongsToAccount: accountID,
		CreatedBefore:    filterArgs.CreatedBefore,
		CreatedAfter:     filterArgs.CreatedAfter,
		UpdatedBefore:    filterArgs.UpdatedBefore,
		UpdatedAfter:     filterArgs.UpdatedAfter,
		PageCursor:       filterArgs.Cursor,
		ResultLimit:      filterArgs.ResultLimit,
		IncludeArchived:  filterArgs.IncludeArchived,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plans list retrieval query")
	}

	for _, result := range results {
		// Extract counts from the first result (all rows have the same counts)
		if totalCount == 0 {
			totalCount = uint64(result.TotalCount)
			filteredCount = uint64(result.FilteredCount)
		}

		data = append(data, &types.MealPlan{
			CreatedAt:              result.CreatedAt,
			VotingDeadline:         result.VotingDeadline,
			ArchivedAt:             database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:          database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:                     result.ID,
			Status:                 string(result.Status),
			Notes:                  result.Notes,
			ElectionMethod:         string(result.ElectionMethod),
			BelongsToAccount:       result.BelongsToAccount,
			CreatedByUser:          result.CreatedByUser,
			Events:                 nil,
			GroceryListInitialized: result.GroceryListInitialized,
			TasksCreated:           result.TasksCreated,
		})
	}

	// Attach each plan's events, but not the options hanging off them. This used to
	// refetch every plan through getMealPlan, which hydrated options, their meals, and
	// every recipe inside those -- a page of eight plans cleared the 4 MiB gRPC message
	// bound on its own. GetMealPlansForAccount answers with a MealPlanSummary, whose
	// events carry no options, so the hydration was work the converter then discarded.
	// A caller that needs a plan's options fetches the plan by ID.
	if err = q.attachEventsToMealPlans(ctx, data); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "attaching events to meal plans")
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mp *types.MealPlan) string { return mp.ID },
		filter,
	)

	return x, nil
}

// GetChosenMealNamesForMealPlans returns, keyed by meal plan event ID, the name of the
// meal on the option voting settled on. Events still awaiting a decision are absent.
//
// It backs MealPlanSummary.events[].chosen_meal_name, which is the one thing a list of
// plans reads out of the options those summaries drop.
func (q *repository) GetChosenMealNamesForMealPlans(ctx context.Context, mealPlanIDs []string) (map[string]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	names := map[string]string{}
	if len(mealPlanIDs) == 0 {
		return names, nil
	}

	results, err := q.generatedQuerier.GetChosenMealNamesForMealPlans(ctx, q.readDB, mealPlanIDs)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing chosen meal names query")
	}

	for _, result := range results {
		names[result.ID] = result.Name
	}

	return names, nil
}

// GetMealPlanIDsVotedOnByUser returns which of the given meal plans the user has cast a
// non-abstaining vote on. An abstention is deliberately not a vote here, matching what
// the clients counted when they read the votes off a hydrated plan themselves.
//
// It backs MealPlanSummary.current_user_has_voted.
func (q *repository) GetMealPlanIDsVotedOnByUser(ctx context.Context, userID string, mealPlanIDs []string) ([]string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if len(mealPlanIDs) == 0 {
		return []string{}, nil
	}

	results, err := q.generatedQuerier.GetMealPlanIDsVotedOnByUser(ctx, q.readDB, &generated.GetMealPlanIDsVotedOnByUserParams{
		IDs:    mealPlanIDs,
		ByUser: userID,
	})
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing voted-on meal plans query")
	}

	return results, nil
}

// GetHydratedMealPlansForAccount fetches a page of an account's meal plans with every
// event, option, meal and selection attached.
//
// It exists for the data-privacy collector, whose UserDataCollection is serialized to
// blob storage and has to hold the whole record. Nothing a client reads should use it:
// a single hydrated plan runs to hundreds of kilobytes, which is why the list endpoint
// answers with summaries. See GetMealPlansForAccount.
//
// Building on that page means its events are fetched and then replaced, which is one
// wasted query per page. That is cheaper than a second copy of the list query and its
// filter handling, on a path that runs once per export request.
func (q *repository) GetHydratedMealPlansForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	page, err := q.GetMealPlansForAccount(ctx, accountID, filter)
	if err != nil {
		return nil, observability.PrepareError(err, span, "fetching meal plans to hydrate")
	}

	for i, mealPlan := range page.Data {
		hydrated, hydrateErr := q.getMealPlan(ctx, mealPlan.ID, accountID)
		if hydrateErr != nil {
			return nil, observability.PrepareError(hydrateErr, span, "hydrating meal plan")
		}

		page.Data[i] = hydrated
	}

	return page, nil
}

// attachEventsToMealPlans populates the Events of every plan given, in one query for
// the whole page rather than one per plan. The events carry no options: this is the
// list path, and MealPlanSummary drops them.
func (q *repository) attachEventsToMealPlans(ctx context.Context, mealPlans []*types.MealPlan) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if len(mealPlans) == 0 {
		return nil
	}

	byID := make(map[string]*types.MealPlan, len(mealPlans))
	ids := make([]string, 0, len(mealPlans))
	for _, mealPlan := range mealPlans {
		mealPlan.Events = []*types.MealPlanEvent{}
		byID[mealPlan.ID] = mealPlan
		ids = append(ids, mealPlan.ID)
	}

	results, err := q.generatedQuerier.GetAllMealPlanEventsForMealPlans(ctx, q.readDB, ids)
	if err != nil {
		return observability.PrepareError(err, span, "executing meal plan events list retrieval query")
	}

	for _, result := range results {
		mealPlan, ok := byID[result.BelongsToMealPlan]
		if !ok {
			continue
		}

		mealPlan.Events = append(mealPlan.Events, &types.MealPlanEvent{
			CreatedAt:         result.CreatedAt,
			StartsAt:          result.StartsAt,
			EndsAt:            result.EndsAt,
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			MealName:          string(result.MealName),
			Notes:             result.Notes,
			BelongsToMealPlan: result.BelongsToMealPlan,
			ID:                result.ID,
			Options:           []*types.MealPlanOption{},
		})
	}

	return nil
}

// CreateMealPlan creates a meal plan in the database.
func (q *repository) CreateMealPlan(ctx context.Context, input *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	logger := q.logger.WithValue(mealplanningkeys.MealPlanIDKey, input.ID)

	status := types.MealPlanStatusFinalized
	for _, event := range input.Events {
		if len(event.Options) > 1 {
			status = types.MealPlanStatusAwaitingVotes
		}
	}

	var err error
	var x *types.MealPlan
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		// create the meal plan.
		if err = q.generatedQuerier.CreateMealPlan(ctx, tx, &generated.CreateMealPlanParams{
			ID:               input.ID,
			Notes:            input.Notes,
			Status:           generated.MealPlanStatus(status),
			VotingDeadline:   input.VotingDeadline,
			BelongsToAccount: input.BelongsToAccount,
			CreatedByUser:    input.CreatedByUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "creating meal plan")
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &input.BelongsToAccount,
			ResourceType:     resourceTypeMealPlans,
			RelevantID:       input.ID,
			EventType:        audit.AuditLogEventTypeCreated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		x = &types.MealPlan{
			ID:               input.ID,
			Notes:            input.Notes,
			Status:           string(status),
			VotingDeadline:   input.VotingDeadline,
			BelongsToAccount: input.BelongsToAccount,
			ElectionMethod:   input.ElectionMethod,
			CreatedAt:        q.CurrentTime(),
			CreatedByUser:    input.CreatedByUser,
		}

		logger.WithValue("quantity", len(input.Events)).Info("creating events for meal plan")
		// Map to track option ID -> meal ID for matching selections
		optionToMealID := make(map[string]string)
		for _, event := range input.Events {
			event.BelongsToMealPlan = x.ID
			opt, createErr := q.createMealPlanEvent(ctx, tx, event)
			if createErr != nil {
				return observability.PrepareError(createErr, span, "creating meal plan event for meal plan")
			}
			x.Events = append(x.Events, opt)

			// Track option IDs and their meal IDs for selection matching
			for _, option := range opt.Options {
				optionToMealID[option.ID] = option.Meal.ID
			}
		}

		// Create selections if provided
		if len(input.Selections) > 0 {
			logger.WithValue("quantity", len(input.Selections)).Info("creating selections for meal plan")

			// Load all meals to check their components (deduplicate meal IDs)
			mealIDSet := make(map[string]bool)
			for _, mealID := range optionToMealID {
				mealIDSet[mealID] = true
			}
			mealIDs := make([]string, 0, len(mealIDSet))
			for mealID := range mealIDSet {
				mealIDs = append(mealIDs, mealID)
			}

			meals, loadErr := q.GetMealsWithIDs(ctx, mealIDs)
			if loadErr != nil {
				return observability.PrepareAndLogError(loadErr, logger, span, "loading meals for selection matching")
			}

			// Create a map of meal ID -> meal for quick lookup
			mealsByID := make(map[string]*types.Meal)
			for _, meal := range meals {
				mealsByID[meal.ID] = meal
			}

			// Match and create selections
			for _, selection := range input.Selections {
				// Find the option that contains the matching recipe
				var matchedOptionID string
				for optionID, mealID := range optionToMealID {
					meal, exists := mealsByID[mealID]
					if !exists {
						continue
					}

					// Check if this meal has a component with the matching recipe ID
					for _, component := range meal.Components {
						if component.Recipe.ID == selection.RecipeID {
							matchedOptionID = optionID
							break
						}
					}
					if matchedOptionID != "" {
						break
					}
				}

				if matchedOptionID == "" {
					logger.WithValue("recipe_id", selection.RecipeID).
						WithValue("recipe_step_id", selection.RecipeStepID).
						Info("could not find matching option for selection, skipping")
					continue
				}

				// Create the selection
				selectionID := identifiers.New()
				if createErr := q.generatedQuerier.CreateMealPlanRecipeOptionSelection(ctx, tx, &generated.CreateMealPlanRecipeOptionSelectionParams{
					ID:                      selectionID,
					BelongsToMealPlanOption: matchedOptionID,
					RecipeID:                selection.RecipeID,
					RecipeStepID:            selection.RecipeStepID,
					IngredientIndex:         int32(selection.IngredientIndex),
					SelectedOptionIndex:     int32(selection.SelectedOptionIndex),
					SelectionType:           selection.SelectionType,
				}); createErr != nil {
					return observability.PrepareAndLogError(createErr, logger, span, "creating meal plan recipe option selection")
				}
			}
		}

		// The event is another statement in this transaction, so it lives or dies with the
		// meal plan it describes. See internal/repositories/postgres/events.
		if emitErr := q.events.Emit(ctx, tx, logger, types.MealPlanCreatedServiceEventType, input.BelongsToAccount, map[string]any{
			mealplanningkeys.MealPlanIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing meal plan created event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, x.ID)
	logger.Info("meal plan created")

	return x, nil
}

// UpdateMealPlan updates a particular meal plan.
func (q *repository) UpdateMealPlan(ctx context.Context, updated *types.MealPlan) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, updated.ID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, updated.BelongsToAccount)

	// The update, its audit log entry, and its data change event share a transaction so a
	// failure part-way through leaves none of the three rather than some of them.
	if err := q.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, updateErr := q.generatedQuerier.UpdateMealPlan(ctx, tx, &generated.UpdateMealPlanParams{
			Notes:            updated.Notes,
			Status:           generated.MealPlanStatus(updated.Status),
			VotingDeadline:   updated.VotingDeadline,
			BelongsToAccount: updated.BelongsToAccount,
			ID:               updated.ID,
		})
		if updateErr != nil {
			return observability.PrepareAndLogError(updateErr, logger, span, "updating meal plan")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return q.recordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
			BelongsToAccount: &updated.BelongsToAccount,
			ResourceType:     resourceTypeMealPlans,
			RelevantID:       updated.ID,
			EventType:        audit.AuditLogEventTypeUpdated,
		}, types.MealPlanUpdatedServiceEventType, updated.BelongsToAccount, map[string]any{
			mealplanningkeys.MealPlanIDKey: updated.ID,
		})
	}); err != nil {
		return err
	}

	logger.Info("meal plan updated")

	return nil
}

// ArchiveMealPlan archives a meal plan from the database by its ID.
func (q *repository) ArchiveMealPlan(ctx context.Context, mealPlanID, accountID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	// As with UpdateMealPlan: the archive, its audit log entry, and its data change event
	// share a transaction.
	return q.WithTransaction(ctx, func(tx database.Tx) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveMealPlan(ctx, tx, &generated.ArchiveMealPlanParams{
			BelongsToAccount: accountID,
			ID:               mealPlanID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving meal plan")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return q.recordAndEmit(ctx, tx, logger, &audit.AuditLogEntry{
			BelongsToAccount: &accountID,
			ResourceType:     resourceTypeMealPlans,
			RelevantID:       mealPlanID,
			EventType:        audit.AuditLogEventTypeArchived,
		}, types.MealPlanArchivedServiceEventType, accountID, map[string]any{
			mealplanningkeys.MealPlanIDKey: mealPlanID,
		})
	})
}

// AttemptToFinalizeMealPlan finalizes a meal plan if all of its options have a selection.
func (q *repository) AttemptToFinalizeMealPlan(ctx context.Context, mealPlanID, accountID string) (finalized bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	logger.Info("attempting to finalize meal plan")

	account, err := q.identityRepo.GetAccount(ctx, accountID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "fetching account")
	}

	// fetch meal plan
	mealPlan, err := q.getMealPlan(ctx, mealPlanID, accountID)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "fetching meal plan")
	}

	votingDeadlineHasPassed := mealPlan.VotingDeadline.Before(q.CurrentTime())
	if strings.EqualFold(mealPlan.Status, string(types.MealPlanStatusFinalized)) {
		return false, ErrAlreadyFinalized
	}

	usersWhoHaveNotVoted := []string{}
	allVotesAreSubmitted := true
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		for _, event := range mealPlan.Events {
			if len(event.Options) == 0 {
				continue
			}

			// we load this map with false for each member of the account
			// and then iterate through the votes and mark each voter as true
			userHasVoted := map[string]bool{}
			for _, member := range account.Members {
				userHasVoted[member.BelongsToUser.ID] = false
			}

			alreadyChosen := false
			for _, opt := range event.Options {
				if opt.Chosen {
					alreadyChosen = true
					break
				}

				for _, vote := range opt.Votes {
					userHasVoted[vote.ByUser] = true
				}
			}

			// if we've previously marked an event option as chosen, then we don't need to do anything else
			if alreadyChosen {
				continue
			}

			for userID, hasVoted := range userHasVoted {
				if !hasVoted {
					allVotesAreSubmitted = false
					usersWhoHaveNotVoted = append(usersWhoHaveNotVoted, userID)
				}
			}

			// if we're missing votes from account members, and the deadline hasn't passed, then we can't finalize the meal plan.
			if !allVotesAreSubmitted && !votingDeadlineHasPassed {
				logger.WithValue("users_without_votes", usersWhoHaveNotVoted).Info("not all votes are submitted, and the voting deadline hasn't passed yet")
				continue
			}

			// the ballot is ready to be tallied for this event
			winner, tiebroken, chosen := q.decideOptionWinner(ctx, event.Options)
			if chosen {
				logger = logger.WithValue("winner", winner).WithValue("tiebroken", tiebroken)

				if err = q.generatedQuerier.FinalizeMealPlanOption(ctx, tx, &generated.FinalizeMealPlanOptionParams{
					MealPlanEventID: database.NullStringFromString(event.ID),
					ID:              winner,
					Tiebroken:       tiebroken,
				}); err != nil {
					return observability.PrepareAndLogError(err, logger, span, "finalizing meal plan option")
				}

				logger.Info("finalized meal plan option")
			} else {
				logger.Info("no winner chosen")
			}
		}

		if allVotesAreSubmitted || votingDeadlineHasPassed {
			logger.Info("finalizing meal plan")

			if err = q.generatedQuerier.FinalizeMealPlan(ctx, tx, &generated.FinalizeMealPlanParams{
				Status: generated.MealPlanStatus(types.MealPlanStatusFinalized),
				ID:     mealPlanID,
			}); err != nil {
				return observability.PrepareAndLogError(err, logger, span, "finalizing meal plan option")
			}

			// Emitted here rather than by the two callers — the manager on a user request
			// and the finalizer job on a tick — because only this transaction knows the
			// plan actually finalized, and only in here can the event commit with it.
			// The account is passed explicitly: the finalizer job has no session context.
			if emitErr := q.events.Emit(ctx, tx, logger, types.MealPlanFinalizedServiceEventType, accountID, map[string]any{
				mealplanningkeys.MealPlanIDKey: mealPlanID,
				"meal_plan":                    mealPlan,
			}); emitErr != nil {
				return observability.PrepareError(emitErr, span, "enqueuing meal plan finalized event")
			}

			finalized = true
		}

		return nil
	}); err != nil {
		return false, err
	}

	logger.WithValue("finalized", finalized).
		WithValue("usersWhoHaveNotVoted", usersWhoHaveNotVoted).
		WithValue("allVotesAreSubmitted", allVotesAreSubmitted).
		WithValue("votingDeadlineHasPassed", votingDeadlineHasPassed).
		Info("done attempting to finalize meal plan")

	return finalized, nil
}

// GetMealPlansAwaitingFinalizationSaga gets meal plans the finalization pipeline still owes
// something to and that no saga has claimed.
func (q *repository) GetMealPlansAwaitingFinalizationSaga(ctx context.Context, limit uint16) ([]*types.MealPlanFinalizationCandidate, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if limit == 0 {
		return []*types.MealPlanFinalizationCandidate{}, nil
	}

	results, err := q.generatedQuerier.GetMealPlansAwaitingFinalizationSaga(ctx, q.readDB, int32(limit))
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing meal plans awaiting finalization saga retrieval query")
	}

	candidates := make([]*types.MealPlanFinalizationCandidate, 0, len(results))
	for _, result := range results {
		candidates = append(candidates, &types.MealPlanFinalizationCandidate{
			MealPlanID: result.ID,
			AccountID:  result.BelongsToAccount,
		})
	}

	return candidates, nil
}

// AttachMealPlanFinalizationSaga claims a meal plan for a new finalization saga.
func (q *repository) AttachMealPlanFinalizationSaga(ctx context.Context, mealPlanID string, start types.MealPlanFinalizationSagaStarter) (string, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if start == nil {
		return "", platformerrors.ErrNilInputParameter
	}

	var sagaID string
	if err := q.WithTransaction(ctx, func(tx database.Tx) error {
		// The instance row first, so its ID exists to be claimed with. Both writes are in this
		// transaction, so losing the claim below takes the instance with it.
		id, startErr := start(ctx, tx)
		if startErr != nil {
			return observability.PrepareError(startErr, span, "starting meal plan finalization saga")
		}

		rowsAffected, attachErr := q.generatedQuerier.AttachMealPlanFinalizationSaga(ctx, tx, &generated.AttachMealPlanFinalizationSagaParams{
			FinalizationSagaID: database.NullStringFromString(id),
			ID:                 mealPlanID,
		})
		if attachErr != nil {
			return observability.PrepareError(attachErr, span, "attaching finalization saga to meal plan")
		}

		if rowsAffected == 0 {
			// Another replica claimed this plan between our read and this write, or the plan
			// was archived in the gap. Either way there is nothing to do and nothing wrong.
			return ErrFinalizationSagaAlreadyAttached
		}

		sagaID = id

		return nil
	}); err != nil {
		return "", err
	}

	return sagaID, nil
}

// GetFinalizedMealPlanOptionsForMealPlan gets the chosen options of a finalized meal plan, with
// the recipes each one draws on, which is what the prep tasks are derived from.
func (q *repository) GetFinalizedMealPlanOptionsForMealPlan(ctx context.Context, mealPlanID string) ([]*types.FinalizedMealPlanDatabaseResult, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	results, err := q.generatedQuerier.GetFinalizedMealPlanOptionsForMealPlan(ctx, q.readDB, mealPlanID)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing finalized meal plan options retrieval query")
	}

	output := []*types.FinalizedMealPlanDatabaseResult{}
	var databaseResult *types.FinalizedMealPlanDatabaseResult
	for _, result := range results {
		r := &types.FinalizedMealPlanDatabaseResult{
			MealPlanID:       result.MealPlanID,
			MealPlanEventID:  result.MealPlanEventID,
			MealPlanOptionID: result.MealPlanOptionID,
			MealID:           result.MealID,
			RecipeIDs:        nil,
		}

		if databaseResult == nil {
			databaseResult = r
		}

		if r.MealID != databaseResult.MealID ||
			r.MealPlanOptionID != databaseResult.MealPlanOptionID ||
			r.MealPlanEventID != databaseResult.MealPlanEventID ||
			r.MealPlanID != databaseResult.MealPlanID {
			output = append(output, databaseResult)
			databaseResult = r
		}

		databaseResult.RecipeIDs = append(databaseResult.RecipeIDs, result.RecipeID)
	}

	if databaseResult != nil {
		output = append(output, databaseResult)
	}

	return output, nil
}

// FetchMissingVotesForMealPlan determines the missing votes for a given meal plan.
func (q *repository) FetchMissingVotesForMealPlan(ctx context.Context, mealPlanID, accountID string) ([]*types.MissingVote, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	account, err := q.identityRepo.GetAccount(ctx, accountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching account to determine missing votes")
	}

	mealPlan, err := q.GetMealPlan(ctx, mealPlanID, accountID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan to determine missing votes")
	}

	var missingVotes []*types.MissingVote
	for _, event := range mealPlan.Events {
		for _, option := range event.Options {
			for _, membership := range account.Members {
				var voteFoundForMemberForOption bool
				for _, vote := range option.Votes {
					if vote.ByUser == membership.BelongsToUser.ID {
						voteFoundForMemberForOption = true
						break
					}
				}

				if !voteFoundForMemberForOption {
					missingVotes = append(missingVotes, &types.MissingVote{
						EventID:  event.ID,
						OptionID: option.ID,
						UserID:   membership.BelongsToUser.ID,
					})
				}
			}
		}
	}

	return missingVotes, nil
}
