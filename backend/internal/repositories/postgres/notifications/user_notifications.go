package notifications

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"
	generated "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	resourceTypeUserNotifications = "user_notifications"
)

var (
	_ types.UserNotificationDataManager = (*Repository)(nil)
)

// UserNotificationExists fetches whether a user notification exists from the database.
func (q *Repository) UserNotificationExists(ctx context.Context, userID, userNotificationID string) (exists bool, err error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if userNotificationID == "" {
		return false, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(notificationkeys.UserNotificationIDKey, userNotificationID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, userNotificationID)

	result, err := q.generatedQuerier.CheckUserNotificationExistence(ctx, q.readDB, &generated.CheckUserNotificationExistenceParams{
		ID:            userNotificationID,
		BelongsToUser: userID,
	})
	if err != nil {
		return false, observability.PrepareAndLogError(err, logger, span, "performing user notification existence check")
	}

	return result, nil
}

// GetUserNotification fetches a user notification from the database.
func (q *Repository) GetUserNotification(ctx context.Context, userID, userNotificationID string) (*types.UserNotification, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(identitykeys.UserIDKey, userID)
	tracing.AttachToSpan(span, identitykeys.UserIDKey, userID)

	if userNotificationID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}
	logger = logger.WithValue(notificationkeys.UserNotificationIDKey, userNotificationID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, userNotificationID)

	result, err := q.generatedQuerier.GetUserNotification(ctx, q.readDB, &generated.GetUserNotificationParams{
		BelongsToUser: userID,
		ID:            userNotificationID,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "fetching user notification")
	}

	userNotification := &types.UserNotification{
		CreatedAt:     result.CreatedAt,
		LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
		ID:            result.ID,
		Content:       result.Content,
		Status:        string(result.Status),
		BelongsToUser: result.BelongsToUser,
	}

	return userNotification, nil
}

// GetUserNotifications fetches a list of user notifications from the database that meet a particular filter.
func (q *Repository) GetUserNotifications(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UserNotification], error) {
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

	filterArgs := filtering.ToSQLArgs(filter)

	results, err := q.generatedQuerier.GetUserNotificationsForUser(ctx, q.readDB, &generated.GetUserNotificationsForUserParams{
		UserID:        userID,
		CreatedBefore: filterArgs.CreatedBefore,
		CreatedAfter:  filterArgs.CreatedAfter,
		UpdatedBefore: filterArgs.UpdatedBefore,
		UpdatedAfter:  filterArgs.UpdatedAfter,
		PageCursor:    filterArgs.Cursor,
		ResultLimit:   filterArgs.ResultLimit,
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing user notifications list retrieval query")
	}

	x := filtering.Drain(
		results,
		func(result *generated.GetUserNotificationsForUserRow) *types.UserNotification {
			return &types.UserNotification{
				CreatedAt:     result.CreatedAt,
				LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
				ID:            result.ID,
			}
		},
		func(result *generated.GetUserNotificationsForUserRow) (int64, int64) {
			return result.FilteredCount, result.TotalCount
		},
		func(t *types.UserNotification) string {
			return t.ID
		},
		filter,
	)

	return x, nil
}

// CreateUserNotification creates a user notification in the database.
func (q *Repository) CreateUserNotification(ctx context.Context, input *types.UserNotificationDatabaseCreationInput) (*types.UserNotification, error) {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, input.ID)
	logger := q.logger.WithValue(notificationkeys.UserNotificationIDKey, input.ID)

	var err error
	var x *types.UserNotification
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		// create the user notification.
		if err = q.generatedQuerier.CreateUserNotification(ctx, tx, &generated.CreateUserNotificationParams{
			ID:            input.ID,
			Content:       input.Content,
			BelongsToUser: input.BelongsToUser,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "performing user notification creation query")
		}

		x = &types.UserNotification{
			ID:            input.ID,
			CreatedAt:     q.CurrentTime(),
			Content:       input.Content,
			Status:        types.UserNotificationStatusTypeUnread,
			BelongsToUser: input.BelongsToUser,
		}
		tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, x.ID)
		logger.Info("user notification created")

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUserNotifications,
			RelevantID:    x.ID,
			EventType:     audit.AuditLogEventTypeCreated,
			BelongsToUser: x.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.UserNotificationCreatedServiceEventType, "", map[string]any{
			notificationkeys.UserNotificationIDKey: input.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return x, nil
}

// UpdateUserNotification updates a particular user notification.
func (q *Repository) UpdateUserNotification(ctx context.Context, updated *types.UserNotification) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := q.logger.WithValue(notificationkeys.UserNotificationIDKey, updated.ID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, updated.ID)

	var err error
	if err = q.WithTransaction(ctx, func(tx database.Tx) error {
		if _, err = q.generatedQuerier.UpdateUserNotification(ctx, tx, &generated.UpdateUserNotificationParams{
			Status: generated.UserNotificationStatus(updated.Status),
			ID:     updated.ID,
		}); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "updating user notification")
		}

		if err = q.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
			ResourceType:  resourceTypeUserNotifications,
			RelevantID:    updated.ID,
			EventType:     audit.AuditLogEventTypeUpdated,
			BelongsToUser: updated.BelongsToUser,
		}); err != nil {
			return observability.PrepareError(err, span, "creating audit log entry")
		}

		// The event is another statement in this transaction, so it commits with the
		// rows it describes.
		if emitErr := q.events.Emit(ctx, tx, logger, types.UserNotificationUpdatedServiceEventType, "", map[string]any{
			notificationkeys.UserNotificationIDKey: updated.ID,
		}); emitErr != nil {
			return observability.PrepareError(emitErr, span, "enqueuing data change event")
		}

		return nil
	}); err != nil {
		return err
	}

	logger.Info("user notification updated")

	return nil
}
