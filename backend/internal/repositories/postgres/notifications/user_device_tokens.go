package notifications

import (
	"context"
	"database/sql"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"

	"github.com/primandproper/platform-go/v11/database"
	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/observability"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

const (
	resourceTypeUserDeviceTokens = "user_device_tokens"
)

var (
	_ types.UserDeviceTokenDataManager = (*Repository)(nil)
)

// UserDeviceTokenExists fetches whether a user device token exists from the database.
func (q *Repository) UserDeviceTokenExists(ctx context.Context, userID, tokenID string) (bool, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if tokenID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(notificationkeys.UserDeviceTokenIDKey, tokenID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, tokenID)

	result, err := q.generatedQuerier.CheckUserDeviceTokenExistence(ctx, q.readDB, &generated.CheckUserDeviceTokenExistenceParams{
		ID:            tokenID,
		BelongsToUser: userID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing user device token existence check")
	}

	return result, nil
}

// GetUserDeviceToken fetches a user device token from the database.
func (q *Repository) GetUserDeviceToken(ctx context.Context, userID, tokenID string) (*types.UserDeviceToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if tokenID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(notificationkeys.UserDeviceTokenIDKey, tokenID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, tokenID)

	result, err := q.generatedQuerier.GetUserDeviceToken(ctx, q.readDB, &generated.GetUserDeviceTokenParams{
		BelongsToUser: userID,
		ID:            tokenID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user device token")
	}

	return &types.UserDeviceToken{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		DeviceToken:   result.DeviceToken,
		Platform:      result.Platform,
		BelongsToUser: result.BelongsToUser,
	}, nil
}

// GetUserDeviceTokens fetches a list of user device tokens from the database that meet a particular filter.
func (q *Repository) GetUserDeviceTokens(ctx context.Context, userID string, filter *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[types.UserDeviceToken], error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if filter == nil {
		filter = filtering.DefaultQueryFilter()
	}
	logger = filter.AttachToLogger(logger)
	tracing.AttachQueryFilterToSpan(span, filter)

	platformFilterNull := sql.NullString{}
	if platformFilter != nil {
		platformFilterNull = database.NullStringFromStringPointer(platformFilter)
	}

	results, err := q.generatedQuerier.GetUserDeviceTokensForUser(ctx, q.readDB, &generated.GetUserDeviceTokensForUserParams{
		UserID:          userID,
		PlatformFilter:  platformFilterNull,
		CreatedAfter:    database.NullTimeFromTimePointer(filter.CreatedAfter),
		CreatedBefore:   database.NullTimeFromTimePointer(filter.CreatedBefore),
		IncludeArchived: database.NullBoolFromBoolPointer(filter.IncludeArchived),
		Cursor:          database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:     database.NullInt32FromUint16Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing user device tokens list retrieval query")
	}

	var (
		data                      = []*types.UserDeviceToken{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		token := &types.UserDeviceToken{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
			ID:            result.ID,
			DeviceToken:   result.DeviceToken,
			Platform:      result.Platform,
			BelongsToUser: result.BelongsToUser,
		}
		data = append(data, token)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
		func(t *types.UserDeviceToken) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateUserDeviceToken creates or updates a user device token in the database (upsert).
// If the same user+token exists and is not archived, updates last_updated_at and returns the existing record.
func (q *Repository) CreateUserDeviceToken(ctx context.Context, input *types.UserDeviceTokenDatabaseCreationInput) (*types.UserDeviceToken, error) {
	return q.UpsertUserDeviceToken(ctx, input)
}

// UpsertUserDeviceToken creates or updates a user device token.
// If the same user+token exists and is not archived, updates last_updated_at and returns the existing record.
func (q *Repository) UpsertUserDeviceToken(ctx context.Context, input *types.UserDeviceTokenDatabaseCreationInput) (*types.UserDeviceToken, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, input.ID)
	logger := q.logger.WithValue(notificationkeys.UserDeviceTokenIDKey, input.ID)

	var result *generated.UserDeviceTokens

	// The upsert and its event share a transaction.
	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		var upsertErr error
		if result, upsertErr = q.generatedQuerier.UpsertUserDeviceToken(ctx, tx, &generated.UpsertUserDeviceTokenParams{
			ID:            input.ID,
			BelongsToUser: input.BelongsToUser,
			DeviceToken:   input.DeviceToken,
			Platform:      input.Platform,
		}); upsertErr != nil {
			return observability.PrepareAndLogError(upsertErr, logger, span, "performing user device token upsert query")
		}

		if emitErr := q.events.Emit(ctx, tx, logger, types.UserDeviceTokenCreatedServiceEventType, "", map[string]any{
			notificationkeys.UserDeviceTokenIDKey: result.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing user device token created event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	x := &types.UserDeviceToken{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ArchivedAt:    database.TimePointerFromNullTime(result.ArchivedAt),
		ID:            result.ID,
		DeviceToken:   result.DeviceToken,
		Platform:      result.Platform,
		BelongsToUser: result.BelongsToUser,
	}
	logger.Info("user device token upserted")
	return x, nil
}

// UpdateUserDeviceToken updates a particular user device token.
func (q *Repository) UpdateUserDeviceToken(ctx context.Context, updated *types.UserDeviceToken) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(notificationkeys.UserDeviceTokenIDKey, updated.ID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, updated.ID)

	var err error
	if err = q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if _, err = q.generatedQuerier.UpdateUserDeviceToken(ctx, tx, &generated.UpdateUserDeviceTokenParams{
			Platform:      updated.Platform,
			ID:            updated.ID,
			BelongsToUser: updated.BelongsToUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user device token")
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUserDeviceTokens,
			RelevantID:    updated.ID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: updated.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.UserDeviceTokenUpdatedServiceEventType, "", map[string]any{
			notificationkeys.UserDeviceTokenIDKey: updated.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user device token updated")
	return nil
}

// ArchiveUserDeviceToken archives a user device token.
func (q *Repository) ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	if tokenID == "" {
		return platformerrors.ErrInvalidIDProvided
	}
	logger := q.logger.WithValue(notificationkeys.UserDeviceTokenIDKey, tokenID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, tokenID)

	if err := q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		rows, err := q.generatedQuerier.ArchiveUserDeviceToken(ctx, tx, &generated.ArchiveUserDeviceTokenParams{
			ID:            tokenID,
			BelongsToUser: userID,
		})
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "archiving user device token")
		}
		if rows == 0 {
			return sql.ErrNoRows
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUserDeviceTokens,
			RelevantID:    tokenID,
			EventType:     audit.AuditLogEventTypeArchived,
			BelongsToUser: userID,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.UserDeviceTokenArchivedServiceEventType, "", map[string]any{
			notificationkeys.UserDeviceTokenIDKey: tokenID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user device token archived")
	return nil
}
