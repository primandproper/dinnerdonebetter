package mealplanning

import (
	"context"
	"database/sql"
	"strings"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/database/filtering"
	platformerrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/identifiers"
	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/tracing"
)

const (
	resourceTypeMealPlans = "meal_plans"
)

var (
	_ types.MealPlanDataManager = (*repository)(nil)

	ErrAlreadyFinalized = platformerrors.New("meal plan already finalized")
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

// GetMealPlansForAccount fetches a list of meal plans from the database that meet a particular filter.
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

	results, err := q.generatedQuerier.GetMealPlansForAccount(ctx, q.readDB, &generated.GetMealPlansForAccountParams{
		BelongsToAccount: accountID,
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
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

	fullMealPlans := []*types.MealPlan{}
	for _, mp := range data {
		fmp, mealPlanFetchErr := q.getMealPlan(ctx, mp.ID, accountID)
		if mealPlanFetchErr != nil {
			return nil, observability.PrepareError(mealPlanFetchErr, span, "scanning meal plans")
		}

		fullMealPlans = append(fullMealPlans, fmp)
	}
	data = fullMealPlans

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mp *types.MealPlan) string { return mp.ID },
		filter,
	)

	return x, nil
}

// CreateMealPlan creates a meal plan in the database.
func (q *repository) CreateMealPlan(ctx context.Context, input *types.MealPlanDatabaseCreationInput) (*types.MealPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.MealPlanIDKey, input.ID)

	status := types.MealPlanStatusFinalized
	for _, event := range input.Events {
		if len(event.Options) > 1 {
			status = types.MealPlanStatusAwaitingVotes
		}
	}

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// create the meal plan.
	if err = q.generatedQuerier.CreateMealPlan(ctx, tx, &generated.CreateMealPlanParams{
		ID:               input.ID,
		Notes:            input.Notes,
		Status:           generated.MealPlanStatus(status),
		VotingDeadline:   input.VotingDeadline,
		BelongsToAccount: input.BelongsToAccount,
		CreatedByUser:    input.CreatedByUser,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "creating meal plan")
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &input.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeMealPlans,
		RelevantID:       input.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	x := &types.MealPlan{
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
			q.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareError(createErr, span, "creating meal plan event for meal plan")
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
			q.RollbackTransaction(ctx, tx)
			return nil, observability.PrepareAndLogError(loadErr, logger, span, "loading meals for selection matching")
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
				q.RollbackTransaction(ctx, tx)
				return nil, observability.PrepareAndLogError(createErr, logger, span, "creating meal plan recipe option selection")
			}
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
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
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, updated.ID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, updated.BelongsToAccount)

	if _, err := q.generatedQuerier.UpdateMealPlan(ctx, q.writeDB, &generated.UpdateMealPlanParams{
		Notes:            updated.Notes,
		Status:           generated.MealPlanStatus(updated.Status),
		VotingDeadline:   updated.VotingDeadline,
		BelongsToAccount: updated.BelongsToAccount,
		ID:               updated.ID,
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan")
	}

	if _, err := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &updated.BelongsToAccount,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeMealPlans,
		RelevantID:       updated.ID,
		EventType:        audit.AuditLogEventTypeUpdated,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
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

	rowsAffected, err := q.generatedQuerier.ArchiveMealPlan(ctx, q.writeDB, &generated.ArchiveMealPlanParams{
		BelongsToAccount: accountID,
		ID:               mealPlanID,
	})
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving meal plan")
	}

	if rowsAffected == 0 {
		return sql.ErrNoRows
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		BelongsToAccount: &accountID,
		ID:               identifiers.New(),
		ResourceType:     resourceTypeMealPlans,
		RelevantID:       mealPlanID,
		EventType:        audit.AuditLogEventTypeArchived,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	return nil
}

// MarkMealPlanAsGroceryListInitialized marks a meal plan as having all its tasks created.
func (q *repository) MarkMealPlanAsGroceryListInitialized(ctx context.Context, mealPlanID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if err := q.generatedQuerier.MarkMealPlanAsGroceryListInitialized(ctx, q.writeDB, mealPlanID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "marking meal plan as having grocery list initialized")
	}

	return nil
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
	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	allVotesAreSubmitted := true
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

			if err = q.generatedQuerier.FinalizeMealPlanOption(ctx, q.readDB, &generated.FinalizeMealPlanOptionParams{
				MealPlanEventID: database.NullStringFromString(event.ID),
				ID:              winner,
				Tiebroken:       tiebroken,
			}); err != nil {
				return false, observability.PrepareAndLogError(err, logger, span, "finalizing meal plan option")
			}

			logger.Info("finalized meal plan option")
		} else {
			logger.Info("no winner chosen")
		}
	}

	if allVotesAreSubmitted || votingDeadlineHasPassed {
		logger.Info("finalizing meal plan")

		if err = q.generatedQuerier.FinalizeMealPlan(ctx, q.readDB, &generated.FinalizeMealPlanParams{
			Status: generated.MealPlanStatus(types.MealPlanStatusFinalized),
			ID:     mealPlanID,
		}); err != nil {
			return false, observability.PrepareAndLogError(err, logger, span, "finalizing meal plan option")
		}

		finalized = true
	}

	if err = tx.Commit(); err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.WithValue("finalized", finalized).
		WithValue("usersWhoHaveNotVoted", usersWhoHaveNotVoted).
		WithValue("allVotesAreSubmitted", allVotesAreSubmitted).
		WithValue("votingDeadlineHasPassed", votingDeadlineHasPassed).
		Info("done attempting to finalize meal plan")

	return finalized, nil
}

// GetUnfinalizedMealPlansWithExpiredVotingPeriods gets unfinalized meal plans with expired voting deadlines.
func (q *repository) GetUnfinalizedMealPlansWithExpiredVotingPeriods(ctx context.Context) ([]*types.MealPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetExpiredAndUnresolvedMealPlans(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing unfinalized meal plans with expired voting periods retrieval query")
	}

	mealPlans := []*types.MealPlan{}
	for _, result := range results {
		mealPlans = append(mealPlans, &types.MealPlan{
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

	return mealPlans, nil
}

// GetFinalizedMealPlanIDsForTheNextWeek gets finalized meal plans for a given duration.
func (q *repository) GetFinalizedMealPlanIDsForTheNextWeek(ctx context.Context) ([]*types.FinalizedMealPlanDatabaseResult, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetFinalizedMealPlansForPlanning(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing finalized meal plan IDs for the week retrieval query")
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

		if r.MealID != databaseResult.MealID &&
			r.MealPlanOptionID != databaseResult.MealPlanOptionID &&
			r.MealPlanEventID != databaseResult.MealPlanEventID &&
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

func (q *repository) GetFinalizedMealPlansWithUninitializedGroceryLists(ctx context.Context) ([]*types.MealPlan, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	results, err := q.generatedQuerier.GetFinalizedMealPlansWithoutGroceryListInit(ctx, q.readDB)
	if err != nil {
		return nil, observability.PrepareError(err, span, "executing finalized meal plans without grocery list initialization query")
	}

	mealPlanDetails := map[string]string{}
	for _, result := range results {
		mealPlanDetails[result.ID] = result.BelongsToAccount
	}

	mealPlans := []*types.MealPlan{}
	for mealPlanID, accountID := range mealPlanDetails {
		mealPlan, getMealPlanErr := q.GetMealPlan(ctx, mealPlanID, accountID)
		if getMealPlanErr != nil {
			return nil, observability.PrepareError(getMealPlanErr, span, "getting meal plan")
		}

		mealPlans = append(mealPlans, mealPlan)
	}

	return mealPlans, nil
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
