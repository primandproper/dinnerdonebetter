package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/keys"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "notifications_data_manager"
)

// notificationsRepo avoids wire cycles: manager takes this interface and produces notifications.Repository.
type notificationsRepo interface {
	notifications.Repository
}

var (
	_ notifications.Repository = (*notificationsManager)(nil)
	_ NotificationsDataManager = (*notificationsManager)(nil)
)

type notificationsManager struct {
	tracer tracing.Tracer
	logger logging.Logger
	repo   notificationsRepo
}

// NewNotificationsDataManager returns a new NotificationsDataManager implementing notifications.Repository.
//
// Data change events are enqueued into the outbox by the repository, inside the same
// transaction as the write they describe; see internal/repositories/postgres/events.
func NewNotificationsDataManager(
	ctx context.Context,
	tracerProvider tracing.Provider,
	logger logging.Logger,
	repo notificationsRepo,
) (NotificationsDataManager, error) {
	return &notificationsManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
	}, nil
}

func (m *notificationsManager) UserNotificationExists(ctx context.Context, userID, userNotificationID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.UserNotificationExists(ctx, userID, userNotificationID)
}

func (m *notificationsManager) GetUserNotification(ctx context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserNotification(ctx, userID, userNotificationID)
}

func (m *notificationsManager) GetUserNotifications(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserNotifications(ctx, userID, filter)
}

func (m *notificationsManager) CreateUserNotification(ctx context.Context, input *notifications.UserNotificationDatabaseCreationInput) (*notifications.UserNotification, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(notificationkeys.UserNotificationIDKey, input.ID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, input.ID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating user notification creation input")
	}

	created, err := m.repo.CreateUserNotification(ctx, input)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (m *notificationsManager) UpdateUserNotification(ctx context.Context, updated *notifications.UserNotification) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(notificationkeys.UserNotificationIDKey, updated.ID)
	tracing.AttachToSpan(span, notificationkeys.UserNotificationIDKey, updated.ID)

	if err := m.repo.UpdateUserNotification(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "update user notification")
	}

	return nil
}

func (m *notificationsManager) UserDeviceTokenExists(ctx context.Context, userID, tokenID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.UserDeviceTokenExists(ctx, userID, tokenID)
}

func (m *notificationsManager) GetUserDeviceToken(ctx context.Context, userID, tokenID string) (*notifications.UserDeviceToken, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserDeviceToken(ctx, userID, tokenID)
}

func (m *notificationsManager) GetUserDeviceTokens(ctx context.Context, userID string, filter *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[notifications.UserDeviceToken], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	return m.repo.GetUserDeviceTokens(ctx, userID, filter, platformFilter)
}

func (m *notificationsManager) CreateUserDeviceToken(ctx context.Context, input *notifications.UserDeviceTokenDatabaseCreationInput) (*notifications.UserDeviceToken, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(notificationkeys.UserDeviceTokenIDKey, input.ID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, input.ID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating user device token creation input")
	}

	created, err := m.repo.CreateUserDeviceToken(ctx, input)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (m *notificationsManager) UpdateUserDeviceToken(ctx context.Context, updated *notifications.UserDeviceToken) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(notificationkeys.UserDeviceTokenIDKey, updated.ID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, updated.ID)

	if err := m.repo.UpdateUserDeviceToken(ctx, updated); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "update user device token")
	}

	return nil
}

func (m *notificationsManager) ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" || tokenID == "" {
		return platformerrors.ErrEmptyInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(notificationkeys.UserDeviceTokenIDKey, tokenID)
	tracing.AttachToSpan(span, notificationkeys.UserDeviceTokenIDKey, tokenID)

	if err := m.repo.ArchiveUserDeviceToken(ctx, userID, tokenID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive user device token")
	}

	return nil
}
