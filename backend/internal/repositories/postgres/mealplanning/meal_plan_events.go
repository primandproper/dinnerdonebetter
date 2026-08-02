package mealplanning

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypeMealPlanEvents = "meal_plan_events"
)

var (
	_ types.MealPlanEventDataManager = (*repository)(nil)
)

// MealPlanEventExists fetches whether a meal plan event exists from the database.
func (q *repository) MealPlanEventExists(ctx context.Context, mealPlanID, mealPlanEventID string) (exists bool, err error) {
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

	result, err := q.generatedQuerier.CheckMealPlanEventExistence(ctx, q.readDB, &generated.CheckMealPlanEventExistenceParams{
		ID:         mealPlanEventID,
		MealPlanID: mealPlanID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan event existence check")
	}

	return result, nil
}

// GetMealPlanEvent fetches a meal plan event from the database.
func (q *repository) GetMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*types.MealPlanEvent, error) {
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

	result, err := q.generatedQuerier.GetMealPlanEvent(ctx, q.readDB, &generated.GetMealPlanEventParams{
		ID:                mealPlanEventID,
		BelongsToMealPlan: mealPlanID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan event retrieval query")
	}

	mealPlanEvent := &types.MealPlanEvent{
		CreatedAt:         result.CreatedAt,
		StartsAt:          result.StartsAt,
		EndsAt:            result.EndsAt,
		ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
		LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
		MealName:          string(result.MealName),
		Notes:             result.Notes,
		BelongsToMealPlan: result.BelongsToMealPlan,
		ID:                result.ID,
	}

	options, err := q.getMealPlanOptionsForMealPlanEvent(ctx, mealPlanID, mealPlanEventID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching meal plan options for meal plan event")
	}
	mealPlanEvent.Options = options

	return mealPlanEvent, nil
}

// getMealPlanEventsForMealPlan fetches a list of mealPlanEvents from the database that meet a particular filter.
func (q *repository) getMealPlanEventsForMealPlan(ctx context.Context, mealPlanID string) (x []*types.MealPlanEvent, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	x = []*types.MealPlanEvent{}

	results, err := q.generatedQuerier.GetAllMealPlanEventsForMealPlan(ctx, q.readDB, mealPlanID)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan events list retrieval query")
	}

	for _, result := range results {
		event := &types.MealPlanEvent{
			CreatedAt:         result.CreatedAt,
			StartsAt:          result.StartsAt,
			EndsAt:            result.EndsAt,
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			MealName:          string(result.MealName),
			Notes:             result.Notes,
			BelongsToMealPlan: result.BelongsToMealPlan,
			ID:                result.ID,
		}

		mealPlanOptions, mealPlanOptionsErr := q.getMealPlanOptionsForMealPlanEvent(ctx, mealPlanID, event.ID)
		if mealPlanOptionsErr != nil {
			return nil, observability.PrepareAndLogError(mealPlanOptionsErr, logger, span, "fetching options for meal plan events")
		}

		event.Options = mealPlanOptions

		x = append(x, event)
	}

	return x, nil
}

// GetMealPlanEvents fetches a list of meal plan events from the database that meet a particular filter.
func (q *repository) GetMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[types.MealPlanEvent], err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if mealPlanID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	var (
		data          []*types.MealPlanEvent
		filteredCount uint64
		totalCount    uint64
	)

	results, err := q.generatedQuerier.GetMealPlanEvents(ctx, q.readDB, &generated.GetMealPlanEventsParams{
		MealPlanID:      mealPlanID,
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing meal plan events list retrieval query")
	}

	for _, result := range results {
		if totalCount == 0 {
			filteredCount = uint64(result.FilteredCount)
			totalCount = uint64(result.TotalCount)
		}
		data = append(data, &types.MealPlanEvent{
			CreatedAt:         result.CreatedAt,
			StartsAt:          result.StartsAt,
			EndsAt:            result.EndsAt,
			ArchivedAt:        database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:     database.TimePointerFromNullTime(result.LastUpdatedAt),
			MealName:          string(result.MealName),
			Notes:             result.Notes,
			BelongsToMealPlan: result.BelongsToMealPlan,
			ID:                result.ID,
		})
	}

	x = filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(mpe *types.MealPlanEvent) string { return mpe.ID },
		filter,
	)

	return x, nil
}

// MealPlanEventIsEligibleForVoting returns if a meal plan can be voted on.
func (q *repository) MealPlanEventIsEligibleForVoting(ctx context.Context, mealPlanID, mealPlanEventID string) (bool, error) {
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

	result, err := q.generatedQuerier.MealPlanEventIsEligibleForVoting(ctx, q.readDB, &generated.MealPlanEventIsEligibleForVotingParams{
		MealPlanID:      mealPlanID,
		MealPlanEventID: mealPlanEventID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing meal plan event existence check")
	}

	return result, nil
}

// createMealPlanEvent creates a meal plan event in the database.
func (q *repository) createMealPlanEvent(ctx context.Context, querier database.SQLQueryExecutor, input *types.MealPlanEventDatabaseCreationInput) (*types.MealPlanEvent, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanEventIDKey, input.ID)

	// create the meal plan event.
	if err := q.generatedQuerier.CreateMealPlanEvent(ctx, querier, &generated.CreateMealPlanEventParams{
		ID:                input.ID,
		Notes:             input.Notes,
		StartsAt:          input.StartsAt,
		EndsAt:            input.EndsAt,
		MealName:          generated.MealName(input.MealName),
		BelongsToMealPlan: input.BelongsToMealPlan,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing meal plan event creation query")
	}

	x := &types.MealPlanEvent{
		ID:                input.ID,
		Notes:             input.Notes,
		StartsAt:          input.StartsAt,
		EndsAt:            input.EndsAt,
		MealName:          input.MealName,
		BelongsToMealPlan: input.BelongsToMealPlan,
		CreatedAt:         q.CurrentTime(),
	}

	logger.WithValue("quantity", len(input.Options)).Info("creating options for meal plan event")
	for _, option := range input.Options {
		option.BelongsToMealPlanEvent = x.ID
		opt, createErr := q.createMealPlanOption(ctx, querier, option, len(input.Options) == 1)
		if createErr != nil {
			return nil, observability.PrepareError(createErr, span, "creating meal plan option for meal plan event")
		}
		x.Options = append(x.Options, opt)
	}

	if _, err := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, querier, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeMealPlanEvents,
		RelevantID:   input.ID,
		EventType:    audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, x.ID)
	logger.Info("meal plan event created")

	return x, nil
}

// CreateMealPlanEvent creates a meal plan event in the database.
func (q *repository) CreateMealPlanEvent(ctx context.Context, input *types.MealPlanEventDatabaseCreationInput) (*types.MealPlanEvent, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	var x *types.MealPlanEvent
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		created, err := q.createMealPlanEvent(ctx, tx, input)
		if err != nil {
			return observability.PrepareError(err, span, "creating meal plan event")
		}

		x = created

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, q.logger, types.MealPlanEventCreatedServiceEventType, "", map[string]any{
			mealplanningkeys.MealPlanEventIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return x, nil
}

// UpdateMealPlanEvent updates a particular meal plan event.
func (q *repository) UpdateMealPlanEvent(ctx context.Context, updated *types.MealPlanEvent) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(mealplanningkeys.MealPlanEventIDKey, updated.ID)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanEventIDKey, updated.ID)

	if err := q.withEvent(ctx, logger, types.MealPlanEventUpdatedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanIDKey:      updated.BelongsToMealPlan,
		mealplanningkeys.MealPlanEventIDKey: updated.ID,
	}, func(tx database.SQLQueryExecutor) error {
		_, updateErr := q.generatedQuerier.UpdateMealPlanEvent(ctx, tx, &generated.UpdateMealPlanEventParams{
			Notes:             updated.Notes,
			StartsAt:          updated.StartsAt,
			EndsAt:            updated.EndsAt,
			MealName:          generated.MealName(updated.MealName),
			BelongsToMealPlan: updated.BelongsToMealPlan,
			ID:                updated.ID,
		})

		return updateErr
	}); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "updating meal plan event")
	}

	if _, err := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, q.writeDB, &audit.AuditLogEntryDatabaseCreationInput{
		ID:           identifiers.New(),
		ResourceType: resourceTypeMealPlanEvents,
		RelevantID:   updated.ID,
		EventType:    audit.AuditLogEventTypeUpdated,
	}); err != nil {
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	logger.Info("meal plan event updated")

	return nil
}

// SwapMealPlanEvents atomically swaps the startsAt/endsAt of two meal plan events.
func (q *repository) SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if mealPlanID == "" || mealPlanEventIDA == "" || mealPlanEventIDB == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if mealPlanEventIDA == mealPlanEventIDB {
		return platformerrors.ErrInvalidIDProvided
	}

	logger := q.logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlanID).
		WithValue(mealplanningkeys.MealPlanEventIDKey, mealPlanEventIDA+","+mealPlanEventIDB)
	tracing.AttachToSpan(span, mealplanningkeys.MealPlanIDKey, mealPlanID)

	eventA, err := q.GetMealPlanEvent(ctx, mealPlanID, mealPlanEventIDA)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan event A for swap")
	}
	eventB, err := q.GetMealPlanEvent(ctx, mealPlanID, mealPlanEventIDB)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "fetching meal plan event B for swap")
	}

	eventAStartsAt, eventAEndsAt := eventA.StartsAt, eventA.EndsAt
	eventBStartsAt, eventBEndsAt := eventB.StartsAt, eventB.EndsAt

	eventA.StartsAt, eventA.EndsAt = eventBStartsAt, eventBEndsAt
	eventB.StartsAt, eventB.EndsAt = eventAStartsAt, eventAEndsAt

	if err = q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = q.generatedQuerier.UpdateMealPlanEvent(ctx, tx, &generated.UpdateMealPlanEventParams{
			Notes:             eventA.Notes,
			StartsAt:          eventA.StartsAt,
			EndsAt:            eventA.EndsAt,
			MealName:          generated.MealName(eventA.MealName),
			BelongsToMealPlan: eventA.BelongsToMealPlan,
			ID:                eventA.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating meal plan event A during swap")
		}
		if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:           identifiers.New(),
			ResourceType: resourceTypeMealPlanEvents,
			RelevantID:   eventA.ID,
			EventType:    audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry for event A")
		}

		if _, err = q.generatedQuerier.UpdateMealPlanEvent(ctx, tx, &generated.UpdateMealPlanEventParams{
			Notes:             eventB.Notes,
			StartsAt:          eventB.StartsAt,
			EndsAt:            eventB.EndsAt,
			MealName:          generated.MealName(eventB.MealName),
			BelongsToMealPlan: eventB.BelongsToMealPlan,
			ID:                eventB.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating meal plan event B during swap")
		}
		if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:           identifiers.New(),
			ResourceType: resourceTypeMealPlanEvents,
			RelevantID:   eventB.ID,
			EventType:    audit.AuditLogEventTypeUpdated,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry for event B")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.MealPlanEventUpdatedServiceEventType, "", map[string]any{
			mealplanningkeys.MealPlanIDKey: mealPlanID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("meal plan events swapped")
	return nil
}

// ArchiveMealPlanEvent archives a meal plan event from the database by its ID.
func (q *repository) ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error {
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

	return q.withEvent(ctx, logger, types.MealPlanEventArchivedServiceEventType, "", map[string]any{
		mealplanningkeys.MealPlanIDKey:      mealPlanID,
		mealplanningkeys.MealPlanEventIDKey: mealPlanEventID,
	}, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := q.generatedQuerier.ArchiveMealPlanEvent(ctx, tx, &generated.ArchiveMealPlanEventParams{
			ID:                mealPlanEventID,
			BelongsToMealPlan: mealPlanID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving meal plan event")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		// The audit log entry belongs in this transaction too: it describes the same
		// write, and half of a write is not something to record.
		if _, auditErr := q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
			ID:           identifiers.New(),
			ResourceType: resourceTypeMealPlanEvents,
			RelevantID:   mealPlanEventID,
			EventType:    audit.AuditLogEventTypeArchived,
		}); auditErr != nil {
			return observability.PrepareError(auditErr, span, "creating audit log entry")
		}

		return nil
	})
}
