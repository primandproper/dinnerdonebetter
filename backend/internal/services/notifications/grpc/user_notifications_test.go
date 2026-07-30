package grpc

import (
	"context"
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/fakes"
	notificationsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/mock"
	grpcfiltering "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	notificationssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/notifications"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/notifications/grpc/converters"

	"github.com/primandproper/platform-go/v8/filtering"
	loggingnoop "github.com/primandproper/platform-go/v8/observability/logging/noop"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// testSessionUserID is the user reported by the session fetcher that buildTestService installs.
const testSessionUserID = "test-user-id"

// buildTestService builds a service backed by the given repository mock. A nil repo gets an
// unconfigured mock, which panics if any of its methods are called.
func buildTestService(t *testing.T, notificationsRepo *notificationsmock.RepositoryMock) *serviceImpl {
	t.Helper()

	if notificationsRepo == nil {
		notificationsRepo = &notificationsmock.RepositoryMock{}
	}

	return &serviceImpl{
		tracer:               tracing.NewTracerForTest(t.Name()),
		logger:               loggingnoop.NewLogger(),
		notificationsManager: notificationsRepo,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: testSessionUserID,
				},
			}, nil
		},
	}
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer:               tracing.NewTracerForTest(t.Name()),
		logger:               loggingnoop.NewLogger(),
		notificationsManager: &notificationsmock.RepositoryMock{},
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
	}
}

func TestServiceImpl_GetUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationFunc: func(_ context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.Equal(t, notificationID, userNotificationID)

				return fakeNotification, nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &notificationssvc.GetUserNotificationRequest{
			UserNotificationId: notificationID,
		}

		response, err := service.GetUserNotification(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeNotification.ID, response.Result.Id)

		assert.Len(t, mockRepo.GetUserNotificationCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		request := &notificationssvc.GetUserNotificationRequest{
			UserNotificationId: "test-notification-id",
		}

		response, err := service.GetUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		notificationID := "nonexistent-notification"

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationFunc: func(_ context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.Equal(t, notificationID, userNotificationID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		request := &notificationssvc.GetUserNotificationRequest{
			UserNotificationId: notificationID,
		}

		response, err := service.GetUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUserNotificationCalls(), 1)
	})
}

func TestServiceImpl_GetUserNotifications(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeNotifications := notificationsfakes.BuildFakeUserNotificationsList()
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationsFunc: func(_ context.Context, userID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.NotNil(t, actualFilter)

				return fakeNotifications, nil
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(*filter.MaxResponseSize)
		request := &notificationssvc.GetUserNotificationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetUserNotifications(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.Len(t, response.Results, len(fakeNotifications.Data))

		assert.Len(t, mockRepo.GetUserNotificationsCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		grpcPageSize := uint32(20)
		request := &notificationssvc.GetUserNotificationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetUserNotifications(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationsFunc: func(_ context.Context, userID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[notifications.UserNotification], error) {
				assert.Equal(t, testSessionUserID, userID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		grpcPageSize := uint32(20)
		request := &notificationssvc.GetUserNotificationsRequest{
			Filter: &grpcfiltering.QueryFilter{
				MaxResponseSize: &grpcPageSize,
			},
		}

		response, err := service.GetUserNotifications(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUserNotificationsCalls(), 1)
	})
}

func TestServiceImpl_UpdateUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID
		newStatus := notifications.UserNotificationStatusTypeRead

		updatedNotification := *fakeNotification
		updatedNotification.Status = newStatus

		// The handler fetches the existing notification, updates it, then re-reads it.
		getCallCount := 0
		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationFunc: func(_ context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.Equal(t, notificationID, userNotificationID)

				getCallCount++
				if getCallCount == 1 {
					return fakeNotification, nil
				}

				return &updatedNotification, nil
			},
			UpdateUserNotificationFunc: func(_ context.Context, updated *notifications.UserNotification) error {
				assert.NotNil(t, updated)

				return nil
			},
		}
		service := buildTestService(t, mockRepo)

		request := &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: notificationID,
			Input: &notificationssvc.UserNotificationUpdateRequestInput{
				Status: new(converters.ConvertStringToUserNotificationStatus(newStatus)),
			},
		}

		response, err := service.UpdateUserNotification(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Updated)
		assert.Equal(t, newStatus, converters.ConvertUserNotificationStatusToString(response.Updated.Status))

		assert.Len(t, mockRepo.GetUserNotificationCalls(), 2)
		assert.Len(t, mockRepo.UpdateUserNotificationCalls(), 1)
	})

	t.Run("session context error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service := buildTestServiceWithSessionError(t)

		statusValue := notifications.UserNotificationStatusTypeRead
		request := &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: "test-notification-id",
			Input: &notificationssvc.UserNotificationUpdateRequestInput{
				Status: new(converters.ConvertStringToUserNotificationStatus(statusValue)),
			},
		}

		response, err := service.UpdateUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	t.Run("repository error on get", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		notificationID := "nonexistent-notification"

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationFunc: func(_ context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.Equal(t, notificationID, userNotificationID)

				return nil, errors.New("repository error")
			},
		}
		service := buildTestService(t, mockRepo)

		statusValue := notifications.UserNotificationStatusTypeRead
		request := &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: notificationID,
			Input: &notificationssvc.UserNotificationUpdateRequestInput{
				Status: new(converters.ConvertStringToUserNotificationStatus(statusValue)),
			},
		}

		response, err := service.UpdateUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUserNotificationCalls(), 1)
		assert.Empty(t, mockRepo.UpdateUserNotificationCalls())
	})

	t.Run("repository error on update", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID

		mockRepo := &notificationsmock.RepositoryMock{
			GetUserNotificationFunc: func(_ context.Context, userID, userNotificationID string) (*notifications.UserNotification, error) {
				assert.Equal(t, testSessionUserID, userID)
				assert.Equal(t, notificationID, userNotificationID)

				return fakeNotification, nil
			},
			UpdateUserNotificationFunc: func(_ context.Context, updated *notifications.UserNotification) error {
				assert.NotNil(t, updated)

				return errors.New("update error")
			},
		}
		service := buildTestService(t, mockRepo)

		statusValue := notifications.UserNotificationStatusTypeRead
		request := &notificationssvc.UpdateUserNotificationRequest{
			UserNotificationId: notificationID,
			Input: &notificationssvc.UserNotificationUpdateRequestInput{
				Status: new(converters.ConvertStringToUserNotificationStatus(statusValue)),
			},
		}

		response, err := service.UpdateUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		assert.Len(t, mockRepo.GetUserNotificationCalls(), 1)
		assert.Len(t, mockRepo.UpdateUserNotificationCalls(), 1)
	})
}
