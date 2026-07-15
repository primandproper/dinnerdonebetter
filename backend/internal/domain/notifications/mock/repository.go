package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"

	"github.com/primandproper/platform-go/v4/filtering"

	"github.com/stretchr/testify/mock"
)

var _ notifications.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// UserNotificationExists is a mock function.
func (m *Repository) UserNotificationExists(ctx context.Context, userID, userNotificationID string) (bool, error) {
	args := m.Called(ctx, userID, userNotificationID)
	return args.Bool(0), args.Error(1)
}

// GetUserNotification is a mock function.
func (m *Repository) GetUserNotification(ctx context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
	args := m.Called(ctx, userID, userNotificationID)
	return args.Get(0).(*notifications.UserNotification), args.Error(1)
}

// GetUserNotifications is a mock function.
func (m *Repository) GetUserNotifications(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[notifications.UserNotification]), args.Error(1)
}

// CreateUserNotification is a mock function.
func (m *Repository) CreateUserNotification(ctx context.Context, input *notifications.UserNotificationDatabaseCreationInput) (*notifications.UserNotification, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*notifications.UserNotification), args.Error(1)
}

// UpdateUserNotification is a mock function.
func (m *Repository) UpdateUserNotification(ctx context.Context, updated *notifications.UserNotification) error {
	args := m.Called(ctx, updated)
	return args.Error(0)
}

// UserDeviceTokenExists is a mock function.
func (m *Repository) UserDeviceTokenExists(ctx context.Context, userID, tokenID string) (bool, error) {
	args := m.Called(ctx, userID, tokenID)
	return args.Bool(0), args.Error(1)
}

// GetUserDeviceToken is a mock function.
func (m *Repository) GetUserDeviceToken(ctx context.Context, userID, tokenID string) (*notifications.UserDeviceToken, error) {
	args := m.Called(ctx, userID, tokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notifications.UserDeviceToken), args.Error(1)
}

// GetUserDeviceTokens is a mock function.
func (m *Repository) GetUserDeviceTokens(ctx context.Context, userID string, filter *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[notifications.UserDeviceToken], error) {
	args := m.Called(ctx, userID, filter, platformFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[notifications.UserDeviceToken]), args.Error(1)
}

// CreateUserDeviceToken is a mock function.
func (m *Repository) CreateUserDeviceToken(ctx context.Context, input *notifications.UserDeviceTokenDatabaseCreationInput) (*notifications.UserDeviceToken, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*notifications.UserDeviceToken), args.Error(1)
}

// UpdateUserDeviceToken is a mock function.
func (m *Repository) UpdateUserDeviceToken(ctx context.Context, updated *notifications.UserDeviceToken) error {
	args := m.Called(ctx, updated)
	return args.Error(0)
}

// ArchiveUserDeviceToken is a mock function.
func (m *Repository) ArchiveUserDeviceToken(ctx context.Context, userID, tokenID string) error {
	args := m.Called(ctx, userID, tokenID)
	return args.Error(0)
}
