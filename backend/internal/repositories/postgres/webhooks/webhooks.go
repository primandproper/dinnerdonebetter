package webhooks

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	webhookkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"
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
				EventType:        database.StringFromNullString(result.WebhookTriggerConfigTriggerEvent),
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

// CreateWebhook creates a webhook in the database and registers it as a delivery endpoint.
//
// The two writes cannot share a transaction — the endpoint store owns its own statements — so
// they are ordered so that a failure between them leaves the system delivering less than
// intended rather than more. The webhook row commits first; if registration then fails, the
// caller gets an error and the account is left with a webhook that does not yet deliver.
//
// The reverse order fails in the direction that matters: an endpoint whose subscriptions were
// live but whose webhook row had rolled back would receive the account's events with nothing in
// the API to show it exists, and no way for the account to find or remove it.
func (r *repository) CreateWebhook(ctx context.Context, input *types.WebhookDatabaseCreationInput) (*types.WebhookCreationResponse, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	logger := r.logger.Clone()

	if input == nil {
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, input.BelongsToAccount)
	logger = logger.WithValue(identitykeys.AccountIDKey, input.BelongsToAccount)

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

		if err = r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &x.BelongsToAccount,
			BelongsToUser:    x.CreatedByUser,
			ResourceType:     resourceTypeWebhooks,
			RelevantID:       x.ID,
			EventType:        audit.AuditLogEventTypeCreated,
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

	secret, err := r.dispatcher.Register(ctx, &webhookdispatch.Registration{
		ID:          x.ID,
		AccountID:   x.BelongsToAccount,
		URL:         x.URL,
		ContentType: x.ContentType,
		EventTypes:  x.EventTypes(),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "registering webhook delivery endpoint")
	}

	return &types.WebhookCreationResponse{Webhook: x, Secret: secret}, nil
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
		TriggerEvent:     input.EventType,
		BelongsToWebhook: input.BelongsToWebhook,
	}); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "performing webhook trigger config creation query")
	}

	if err := r.auditLogEntryRepo.Record(ctx, querier, &audit.AuditLogEntry{
		BelongsToAccount: &accountID,
		ResourceType:     resourceTypeWebhookTriggerConfigs,
		RelevantID:       input.ID,
		EventType:        audit.AuditLogEventTypeCreated,
	}); err != nil {
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	return &types.WebhookTriggerConfig{
		ID:               input.ID,
		EventType:        input.EventType,
		BelongsToWebhook: input.BelongsToWebhook,
		CreatedAt:        r.CurrentTime(),
		ArchivedAt:       nil,
	}, nil
}

// ArchiveWebhook archives a webhook in the database and stops delivering to it.
//
// The endpoint is retired first, which is the opposite of CreateWebhook's ordering and for the
// same reason: each direction is ordered so that a failure between the two writes leaves less
// delivery than intended. A webhook row archived first, with the endpoint retirement then
// failing, is a subscriber that keeps receiving an account's events after the account removed it.
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

	// Ownership is established before the endpoint is retired, because the endpoint store is
	// not account-aware: it is keyed by the webhook's ID alone, and retiring one without
	// checking would let anyone who knows an ID silence another account's webhook.
	exists, err := r.WebhookExists(ctx, webhookID, accountID)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "checking webhook existence")
	}

	if !exists {
		return sql.ErrNoRows
	}

	if err = r.dispatcher.Archive(ctx, webhookID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archiving webhook delivery endpoint")
	}

	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := r.generatedQuerier.ArchiveWebhook(ctx, tx, &generated.ArchiveWebhookParams{
			BelongsToAccount: accountID,
			ID:               webhookID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving webhook")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if auditErr := r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &accountID,
			ResourceType:     resourceTypeWebhooks,
			RelevantID:       webhookID,
			EventType:        audit.AuditLogEventTypeArchived,
		}); auditErr != nil {
			return observability.PrepareError(auditErr, span, "creating audit log entry")
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

// AddWebhookTriggerConfig subscribes a webhook to one more event type.
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
		webhookkeys.WebhookIDKey:              input.BelongsToWebhook,
		webhookkeys.WebhookTriggerConfigIDKey: input.ID,
		identitykeys.AccountIDKey:             accountID,
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
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookTriggerConfigCreatedServiceEventType, accountID, map[string]any{
			webhookkeys.WebhookIDKey:              input.BelongsToWebhook,
			webhookkeys.WebhookTriggerConfigIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	// Subscriptions are added after the row commits: a subscription live against a config row
	// that rolled back would deliver an event the account has no record of asking for.
	if err := r.syncSubscriptions(ctx, input.BelongsToWebhook, accountID); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "syncing webhook endpoint subscriptions")
	}

	return created, nil
}

// ArchiveWebhookTriggerConfig unsubscribes a webhook from one event type.
func (r *repository) ArchiveWebhookTriggerConfig(ctx context.Context, webhookID, accountID, configID string) error {
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

	if configID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookTriggerConfigIDKey, configID)

	logger := r.logger.WithValues(map[string]any{
		webhookkeys.WebhookIDKey:              webhookID,
		identitykeys.AccountIDKey:             accountID,
		webhookkeys.WebhookTriggerConfigIDKey: configID,
	})

	var err error
	if err = r.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rowsAffected, archiveErr := r.generatedQuerier.ArchiveWebhookTriggerConfig(ctx, tx, &generated.ArchiveWebhookTriggerConfigParams{
			BelongsToWebhook: webhookID,
			BelongsToAccount: accountID,
			ID:               configID,
		})
		if archiveErr != nil {
			return observability.PrepareAndLogError(archiveErr, logger, span, "archiving webhook trigger config")
		}

		if rowsAffected == 0 {
			return sql.ErrNoRows
		}

		if auditErr := r.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			BelongsToAccount: &accountID,
			ResourceType:     resourceTypeWebhookTriggerConfigs,
			RelevantID:       configID,
			EventType:        audit.AuditLogEventTypeArchived,
		}); auditErr != nil {
			return observability.PrepareError(auditErr, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the row
		// it describes.
		if emitErr := r.events.Emit(ctx, tx, logger, types.WebhookTriggerConfigArchivedServiceEventType, accountID, map[string]any{
			webhookkeys.WebhookIDKey:              webhookID,
			webhookkeys.WebhookTriggerConfigIDKey: configID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing webhook trigger config archived event")
		}

		return nil
	}); err != nil {
		return err
	}

	// Removing a subscription after the row is archived means the window between them
	// delivers an event the account has just unsubscribed from, which is a duplicate a
	// subscriber already has to tolerate. Removing it first and then failing to archive the
	// row would leave a subscription the API says exists and that never fires, which is the
	// harder failure to notice.
	if err = r.syncSubscriptions(ctx, webhookID, accountID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "syncing webhook endpoint subscriptions")
	}

	return nil
}

// RotateWebhookSecret mints a new signing secret for a webhook and returns it.
func (r *repository) RotateWebhookSecret(ctx context.Context, webhookID, accountID string) (string, error) {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if webhookID == "" || accountID == "" {
		return "", platformerrors.ErrInvalidIDProvided
	}
	tracing.AttachToSpan(span, webhookkeys.WebhookIDKey, webhookID)
	tracing.AttachToSpan(span, identitykeys.AccountIDKey, accountID)

	logger := r.logger.WithValues(map[string]any{
		webhookkeys.WebhookIDKey:  webhookID,
		identitykeys.AccountIDKey: accountID,
	})

	// The endpoint store is keyed by webhook ID alone and knows nothing about accounts, so
	// ownership is established here. Without it, knowing an ID would be enough to roll another
	// account's signing key and break every delivery to it.
	//
	// The full webhook is read rather than just its existence, because the endpoint may not
	// exist yet: a webhook created before delivery worked has no endpoint, and rotating is how
	// its owner adopts it. Registering one needs the URL, content type, and subscriptions.
	webhook, err := r.GetWebhook(ctx, webhookID, accountID)
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "reading webhook")
	}

	secret, err := r.dispatcher.RotateSecret(ctx, webhookID, &webhookdispatch.Registration{
		ID:          webhook.ID,
		AccountID:   webhook.BelongsToAccount,
		URL:         webhook.URL,
		ContentType: webhook.ContentType,
		EventTypes:  webhook.EventTypes(),
	})
	if err != nil {
		return "", observability.PrepareAndLogError(err, logger, span, "rotating webhook signing secret")
	}

	return secret, nil
}

// syncSubscriptions rewrites an endpoint's subscription set from the webhook's trigger configs.
//
// The trigger configs are the account-facing record and the platform's subscriptions are what
// fan-out reads, so one of them has to be derived from the other. Deriving in this direction
// means a subscription can never exist for an event the API does not show — the reverse would be
// invisible to the account it delivers on behalf of.
func (r *repository) syncSubscriptions(ctx context.Context, webhookID, accountID string) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	webhook, err := r.GetWebhook(ctx, webhookID, accountID)
	if err != nil {
		return observability.PrepareError(err, span, "reading webhook for subscription sync")
	}

	return r.dispatcher.SetEventTypes(ctx, webhookID, accountID, webhook.EventTypes())
}
