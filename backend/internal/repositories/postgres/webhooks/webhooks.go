package webhooks

import (
	"context"
	"database/sql"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v9/database"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	resourceTypeWebhooks              = "webhooks"
	resourceTypeWebhookTriggerConfigs = "webhook_trigger_configs"
	resourceTypeWebhookTriggerEvents  = "webhook_trigger_events"
)

var (
	_ types.WebhookDataManager = (*repository)(nil)
)

// WebhookExists fetches whether a webhook exists from the database.
func (r *repository) WebhookExists(ctx context.Context, webhookID, accountID string) (exists bool, err error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if webhookID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	if accountID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	result, err := r.generatedQuerier.CheckWebhookExistence(ctx, r.readDB, &generated.CheckWebhookExistenceParams{
		BelongsToAccount: accountID,
		ID:               webhookID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing webhook existence check")
	}

	return result, nil
}

// GetWebhook fetches a webhook from the database.
func (r *repository) GetWebhook(ctx context.Context, webhookID, accountID string) (*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if webhookID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	results, err := r.generatedQuerier.GetWebhook(ctx, r.readDB, &generated.GetWebhookParams{
		BelongsToAccount: accountID,
		ID:               webhookID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhook")
	}

	if len(results) == 0 {
		return nil, sql.ErrNoRows
	}

	var webhook *types.Webhook
	for _, result := range results {
		if webhook == nil {
			webhook = &types.Webhook{
				CreatedAt:        result.WebhookCreatedAt,
				ArchivedAt:       database.TimePointerFromNullTime(result.WebhookArchivedAt),
				LastUpdatedAt:    database.TimePointerFromNullTime(result.WebhookLastUpdatedAt),
				Name:             result.WebhookName,
				URL:              result.WebhookUrl,
				Method:           string(result.WebhookMethod),
				ID:               result.WebhookID,
				BelongsToAccount: result.WebhookBelongsToAccount,
				CreatedByUser:    result.WebhookCreatedByUser,
				ContentType:      string(result.WebhookContentType),
				TriggerConfigs:   []*types.WebhookTriggerConfig{},
			}
		}

		if result.WebhookTriggerConfigID.Valid {
			webhook.TriggerConfigs = append(webhook.TriggerConfigs, &types.WebhookTriggerConfig{
				CreatedAt:        database.TimeFromNullTime(result.WebhookTriggerConfigCreatedAt),
				ArchivedAt:       database.TimePointerFromNullTime(result.WebhookTriggerConfigArchivedAt),
				ID:               database.StringFromNullString(result.WebhookTriggerConfigID),
				BelongsToWebhook: database.StringFromNullString(result.WebhookTriggerConfigBelongsToWebhook),
				TriggerEventID:   database.StringFromNullString(result.WebhookTriggerConfigTriggerEvent),
			})
		}
	}

	return webhook, nil
}

// GetWebhooks fetches a list of webhooks from the database that meet a particular filter.
func (r *repository) GetWebhooks(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Webhook], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	results, err := r.generatedQuerier.GetWebhooksForAccount(ctx, r.readDB, &generated.GetWebhooksForAccountParams{
		BelongsToAccount: accountID,
		CreatedBefore:    database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:     database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore:    database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:     database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:           database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:      database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
		IncludeArchived:  database.NullBoolFromBoolPointer(filter.IncludeArchived),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhooks from database")
	}

	var (
		data                      []*types.Webhook
		filteredCount, totalCount uint64
		seen                      = make(map[string]struct{})
	)
	for _, result := range results {
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
		if _, ok := seen[result.ID]; ok {
			continue
		}
		seen[result.ID] = struct{}{}
		data = append(data, &types.Webhook{
			CreatedAt:        result.CreatedAt_2,
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt_2),
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			Name:             result.Name,
			URL:              result.URL,
			Method:           string(result.Method),
			ID:               result.ID,
			BelongsToAccount: result.BelongsToAccount,
			CreatedByUser:    result.CreatedByUser,
			ContentType:      string(result.ContentType),
		})
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.Webhook) string {
			return t.ID
		},
		filter,
	), nil
}

// GetWebhooksForAccountAndEvent fetches webhooks for an account that are subscribed to the given trigger event (catalog) ID.
func (r *repository) GetWebhooksForAccountAndEvent(ctx context.Context, accountID, triggerEventID string) ([]*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	databaseResults, err := r.generatedQuerier.GetWebhooksForAccountAndEvent(ctx, r.readDB, &generated.GetWebhooksForAccountAndEventParams{
		BelongsToAccount: accountID,
		TriggerEvent:     triggerEventID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching webhooks from database")
	}

	results := make([]*types.Webhook, 0, len(databaseResults))
	for _, result := range databaseResults {
		results = append(results, &types.Webhook{
			CreatedAt:        result.CreatedAt,
			ArchivedAt:       database.TimePointerFromNullTime(result.ArchivedAt),
			LastUpdatedAt:    database.TimePointerFromNullTime(result.LastUpdatedAt),
			Name:             result.Name,
			URL:              result.URL,
			Method:           string(result.Method),
			ID:               result.ID,
			BelongsToAccount: result.BelongsToAccount,
			CreatedByUser:    result.CreatedByUser,
			ContentType:      string(result.ContentType),
		})
	}

	return results, nil
}

// CreateWebhook creates a webhook in the database.
func (r *repository) CreateWebhook(ctx context.Context, input *types.WebhookDatabaseCreationInput) (*types.Webhook, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(identitykeys.AccountIDKey, input.BelongsToAccount)

	logger.Debug("CreateWebhook invoked")

	var err error
	var x *types.Webhook
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err = r.generatedQuerier.CreateWebhook(ctx, tx, &generated.CreateWebhookParams{
			ID:               input.ID,
			Name:             input.Name,
			ContentType:      generated.WebhookContentType(input.ContentType),
			URL:              input.URL,
			Method:           generated.WebhookMethod(input.Method),
			CreatedByUser:    input.CreatedByUser,
			BelongsToAccount: input.BelongsToAccount,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing webhook creation query")
		}

		x = &types.Webhook{
			ID:               input.ID,
			Name:             input.Name,
			ContentType:      input.ContentType,
			URL:              input.URL,
			Method:           input.Method,
			BelongsToAccount: input.BelongsToAccount,
			CreatedByUser:    input.CreatedByUser,
			CreatedAt:        r.CurrentTime(),
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			Scope:        x.BelongsToAccount,
			ResourceType: resourceTypeWebhooks,
			ResourceID:   x.ID,
			EventType:    audit.EventCreated,
			Actor:        audit.SystemActor(),
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		for i := range input.TriggerConfigs {
			cfg := input.TriggerConfigs[i]
			cfg.BelongsToWebhook = input.ID

			created, createErr := r.createWebhookTriggerConfig(ctx, tx, x.BelongsToAccount, cfg)
			if createErr != nil {
				return observability.PrepareAndLogError(createErr, logger, span, "performing webhook trigger config creation")
			}

			x.TriggerConfigs = append(x.TriggerConfigs, created)
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookCreatedServiceEventType, input.BelongsToAccount, map[string]any{
			webhookkeys.WebhookIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, x.ID)

	return x, nil
}

// createWebhookTriggerConfig creates a webhook trigger config (join table row) in the database.
func (r *repository) createWebhookTriggerConfig(ctx context.Context, querier database.SQLQueryExecutor, accountID string, input *types.WebhookTriggerConfigDatabaseCreationInput) (*types.WebhookTriggerConfig, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, input.BelongsToWebhook)

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)
	logger = logger.WithValue(identitykeys.AccountIDKey, accountID)

	if err := r.generatedQuerier.CreateWebhookTriggerConfig(ctx, querier, &generated.CreateWebhookTriggerConfigParams{
		ID:               input.ID,
		TriggerEvent:     input.TriggerEventID,
		BelongsToWebhook: input.BelongsToWebhook,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing webhook trigger config creation query")
	}

	if err := r.auditLogEntryRepo.Record(ctx, querier, &audit.Entry{
		Scope:        accountID,
		ResourceType: resourceTypeWebhookTriggerConfigs,
		ResourceID:   input.ID,
		EventType:    audit.EventCreated,
		Actor:        audit.SystemActor(),
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	return &types.WebhookTriggerConfig{
		ID:               input.ID,
		TriggerEventID:   input.TriggerEventID,
		BelongsToWebhook: input.BelongsToWebhook,
		CreatedAt:        r.CurrentTime(),
		ArchivedAt:       nil,
	}, nil
}

// ArchiveWebhook archives a webhook in the database.
func (r *repository) ArchiveWebhook(ctx context.Context, webhookID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	if accountID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		webhookkeys.WebhookIDKey:  webhookID,
		identitykeys.AccountIDKey: accountID,
	})

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, err := r.generatedQuerier.ArchiveWebhook(ctx, tx, &generated.ArchiveWebhookParams{
			BelongsToAccount: accountID,
			ID:               webhookID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving webhook")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			Scope:        accountID,
			ResourceType: resourceTypeWebhooks,
			ResourceID:   webhookID,
			EventType:    audit.EventArchived,
			Actor:        audit.SystemActor(),
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookArchivedServiceEventType, accountID, map[string]any{
			webhookkeys.WebhookIDKey: webhookID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// AddWebhookTriggerConfig adds a webhook trigger config (subscription to a trigger event) for a webhook.
func (r *repository) AddWebhookTriggerConfig(ctx context.Context, accountID string, input *types.WebhookTriggerConfigDatabaseCreationInput) (*types.WebhookTriggerConfig, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if accountID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, input.BelongsToWebhook)

	logger := r.logger.WithValues(map[string]any{
		webhookkeys.WebhookIDKey:             input.BelongsToWebhook,
		webhookkeys.WebhookTriggerEventIDKey: input.ID,
		identitykeys.AccountIDKey:            accountID,
	})

	var created *types.WebhookTriggerConfig
	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		result, err := r.createWebhookTriggerConfig(ctx, tx, accountID, input)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing webhook trigger config creation")
		}

		created = result

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookTriggerConfigCreatedServiceEventType, "", map[string]any{
			webhookkeys.WebhookIDKey:              input.BelongsToWebhook,
			webhookkeys.WebhookTriggerConfigIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return created, nil
}

// ArchiveWebhookTriggerConfig archives a webhook trigger config.
func (r *repository) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, configID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)

	if configID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerEventIDKey, configID)

	logger := r.logger.WithValues(map[string]any{
		webhookkeys.WebhookIDKey:             webhookID,
		webhookkeys.WebhookTriggerEventIDKey: configID,
	})

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = r.generatedQuerier.ArchiveWebhookTriggerConfig(ctx, tx, &generated.ArchiveWebhookTriggerConfigParams{
			BelongsToWebhook: webhookID,
			ID:               configID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving webhook trigger config")
		}

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeWebhookTriggerConfigs,
			ResourceID:   configID,
			EventType:    audit.EventArchived,
			Actor:        audit.SystemActor(),
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the row
		// it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookTriggerConfigArchivedServiceEventType, "", map[string]any{
			webhookkeys.WebhookIDKey:              webhookID,
			webhookkeys.WebhookTriggerConfigIDKey: configID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing webhook trigger config archived event")
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}

// CreateWebhookTriggerEvent creates a catalog trigger event in the database.
func (r *repository) CreateWebhookTriggerEvent(ctx context.Context, input *types.WebhookTriggerEventDatabaseCreationInput) (*types.WebhookTriggerEvent, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}

	// The row and the entry describing it commit together, which also gives
	// Record's lock on the chain-head row a transaction to be held for. Against the
	// pool that lock is released before the INSERT it exists to protect, and two
	// concurrent writers claim the same chain position.
	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err := r.generatedQuerier.CreateWebhookTriggerEvent(ctx, tx, &generated.CreateWebhookTriggerEventParams{
			ID:          input.ID,
			Name:        input.Name,
			Description: input.Description,
		}); err != nil {
			return err
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeWebhookTriggerEvents,
			ResourceID:   input.ID,
			EventType:    audit.EventCreated,
			Actor:        audit.SystemActor(),
		})
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "creating webhook trigger event")
	}

	return &types.WebhookTriggerEvent{
		ID:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		CreatedAt:     r.CurrentTime(),
		LastUpdatedAt: nil,
		ArchivedAt:    nil,
	}, nil
}

// GetWebhookTriggerEvent fetches a catalog trigger event by ID.
func (r *repository) GetWebhookTriggerEvent(ctx context.Context, id string) (*types.WebhookTriggerEvent, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	result, err := r.generatedQuerier.GetWebhookTriggerEvent(ctx, r.readDB, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching webhook trigger event")
	}

	return &types.WebhookTriggerEvent{
		ID:            result.ID,
		Name:          result.Name,
		Description:   result.Description,
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
	}, nil
}

// GetWebhookTriggerEvents fetches catalog trigger events with optional filter.
func (r *repository) GetWebhookTriggerEvents(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.WebhookTriggerEvent], error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}

	results, err := r.generatedQuerier.GetWebhookTriggerEvents(ctx, r.readDB, &generated.GetWebhookTriggerEventsParams{
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		UpdatedAfter:    database.NullTimeFromTimePointer(filter.UpdatedAfter),
		UpdatedBefore:   database.NullTimeFromTimePointer(filter.UpdatedBefore),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, r.logger.Clone(), span, "fetching webhook trigger events")
	}

	data := make([]*types.WebhookTriggerEvent, 0, len(results))
	var filteredCount, totalCount uint64
	for _, row := range results {
		data = append(data, &types.WebhookTriggerEvent{
			ID:            row.ID,
			Name:          row.Name,
			Description:   row.Description,
			CreatedAt:     row.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(row.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(row.ArchivedAt),
		})
		filteredCount = uint64(row.FilteredCount)
		totalCount = uint64(row.TotalCount)
	}

	return filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.WebhookTriggerEvent) string { return t.ID },
		filter,
	), nil
}

// UpdateWebhookTriggerEvent updates a catalog trigger event.
func (r *repository) UpdateWebhookTriggerEvent(ctx context.Context, id string, input *types.WebhookTriggerEventUpdateRequestInput) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return platformerrors.ErrNilInputProvided
	}

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, updateErr := r.generatedQuerier.UpdateWebhookTriggerEvent(ctx, tx, &generated.UpdateWebhookTriggerEventParams{
			ID:          id,
			Name:        input.Name,
			Description: input.Description,
		})
		if updateErr != nil {
			return updateErr
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeWebhookTriggerEvents,
			ResourceID:   id,
			EventType:    audit.EventUpdated,
			Actor:        audit.SystemActor(),
		})
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		return observability.PrepareAndLogError(err, r.logger.Clone(), span, "updating webhook trigger event")
	}

	return nil
}

// ArchiveWebhookTriggerEvent archives a catalog trigger event.
func (r *repository) ArchiveWebhookTriggerEvent(ctx context.Context, id string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if id == "" {
		return platformerrors.ErrInvalidIDProvided
	}

	if err := r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := r.generatedQuerier.ArchiveWebhookTriggerEvent(ctx, tx, id)
		if archiveErr != nil {
			return archiveErr
		}
		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		return r.auditLogEntryRepo.Record(ctx, tx, &audit.Entry{
			ResourceType: resourceTypeWebhookTriggerEvents,
			ResourceID:   id,
			EventType:    audit.EventArchived,
			Actor:        audit.SystemActor(),
		})
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return err
		}

		return observability.PrepareAndLogError(err, r.logger.Clone(), span, "archiving webhook trigger event")
	}

	return nil
}
