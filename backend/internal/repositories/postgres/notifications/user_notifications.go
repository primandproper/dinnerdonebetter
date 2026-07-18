package notifications

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/keys"
	generated "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"

	"github.com/primandproper/platform-go/v5/database"
	platformerrors "github.com/primandproper/platform-go/v5/errors"
	"github.com/primandproper/platform-go/v5/filtering"
	"github.com/primandproper/platform-go/v5/identifiers"
	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/tracing"
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

	results, err := q.generatedQuerier.GetUserNotificationsForUser(ctx, q.readDB, &generated.GetUserNotificationsForUserParams{
		UserID:        userID,
		CreatedBefore: database.NullTimeFromTimePointer(filter.CreatedBefore),
		CreatedAfter:  database.NullTimeFromTimePointer(filter.CreatedAfter),
		UpdatedBefore: database.NullTimeFromTimePointer(filter.UpdatedBefore),
		UpdatedAfter:  database.NullTimeFromTimePointer(filter.UpdatedAfter),
		Cursor:        database.NullStringFromStringPointer(filter.Cursor),
		ResultLimit:   database.NullInt32FromUint8Pointer(filter.MaxResponseSize),
	})
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "executing user notifications list retrieval query")
	}

	var (
		data                      = []*types.UserNotification{}
		filteredCount, totalCount uint64
	)
	for _, result := range results {
		userNotification := &types.UserNotification{
			CreatedAt:     result.CreatedAt,
			LastUpdatedAt: database.TimePointerFromNullTime(result.LastUpdatedAt),
			ID:            result.ID,
		}

		data = append(data, userNotification)
		filteredCount = uint64(result.FilteredCount)
		totalCount = uint64(result.TotalCount)
	}

	x := filtering.NewQueryFilteredResult(
		data,
		filteredCount,
		totalCount,
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
		return nil, platformerrors.ErrNilInputProvided
	}
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, input.ID)
	logger := q.logger.WithValue(notificationkeys.UserNotificationIDKey, input.ID)

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	// create the user notification.
	if err = q.generatedQuerier.CreateUserNotification(ctx, tx, &generated.CreateUserNotificationParams{
		ID:            input.ID,
		Content:       input.Content,
		BelongsToUser: input.BelongsToUser,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareAndLogError(err, logger, span, "performing user notification creation query")
	}

	x := &types.UserNotification{
		ID:            input.ID,
		CreatedAt:     q.CurrentTime(),
		Content:       input.Content,
		Status:        types.UserNotificationStatusTypeUnread,
		BelongsToUser: input.BelongsToUser,
	}
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, x.ID)
	logger.Info("user notification created")

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeUserNotifications,
		RelevantID:    x.ID,
		EventType:     audit.AuditLogEventTypeCreated,
		BelongsToUser: x.BelongsToUser,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return nil, observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	return x, nil
}

// UpdateUserNotification updates a particular user notification.
func (q *Repository) UpdateUserNotification(ctx context.Context, updated *types.UserNotification) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputProvided
	}
	logger := q.logger.WithValue(notificationkeys.UserNotificationIDKey, updated.ID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, updated.ID)

	tx, err := q.writeDB.BeginTx(ctx, nil)
	if err != nil {
		return observability.PrepareAndLogError(err, logger, span, "beginning transaction")
	}

	if _, err = q.generatedQuerier.UpdateUserNotification(ctx, tx, &generated.UpdateUserNotificationParams{
		Status: generated.UserNotificationStatus(updated.Status),
		ID:     updated.ID,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return observability.PrepareAndLogError(err, logger, span, "updating user notification")
	}

	if _, err = q.auditLogEntryRepo.CreateAuditLogEntry(ctx, tx, &audit.AuditLogEntryDatabaseCreationInput{
		ID:            identifiers.New(),
		ResourceType:  resourceTypeUserNotifications,
		RelevantID:    updated.ID,
		EventType:     audit.AuditLogEventTypeUpdated,
		BelongsToUser: updated.BelongsToUser,
	}); err != nil {
		q.RollbackTransaction(ctx, tx)
		return observability.PrepareError(err, span, "creating audit log entry")
	}

	if err = tx.Commit(); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "committing transaction")
	}

	logger.Info("user notification updated")

	return nil
}
