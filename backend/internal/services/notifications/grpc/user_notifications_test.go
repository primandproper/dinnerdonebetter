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
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/filtering"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func buildTestService(t *testing.T) (*serviceImpl, *notificationsmock.Repository) {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())
	notificationsRepo := &notificationsmock.Repository{}

	service := &serviceImpl{
		tracer:               tracer,
		logger:               logger,
		notificationsManager: notificationsRepo,
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: "test-user-id",
				},
			}, nil
		},
	}

	return service, notificationsRepo
}

func buildTestServiceWithSessionError(t *testing.T) *serviceImpl {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracer := tracing.NewTracerForTest(t.Name())

	service := &serviceImpl{
		tracer:               tracer,
		logger:               logger,
		notificationsManager: &notificationsmock.Repository{},
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		},
	}

	return service
}

func TestServiceImpl_GetUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID
		userID := "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return(fakeNotification, nil)

		request := &notificationssvc.GetUserNotificationRequest{
			UserNotificationId: notificationID,
		}

		response, err := service.GetUserNotification(ctx, request)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.NotNil(t, response.ResponseDetails)
		assert.NotNil(t, response.Result)
		assert.Equal(t, fakeNotification.ID, response.Result.Id)

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		notificationID := "nonexistent-notification"
		userID := "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return((*notifications.UserNotification)(nil), errors.New("repository error"))

		request := &notificationssvc.GetUserNotificationRequest{
			UserNotificationId: notificationID,
		}

		response, err := service.GetUserNotification(ctx, request)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Equal(t, codes.Internal, status.Code(err))

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_GetUserNotifications(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeNotifications := notificationsfakes.BuildFakeUserNotificationsList()
		userID := "test-user-id"
		pageSize := uint8(20)
		filter := &filtering.QueryFilter{
			MaxResponseSize: &pageSize,
		}

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotifications), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return(fakeNotifications, nil)

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		userID := "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotifications), testutils.ContextMatcher, userID, testutils.QueryFilterMatcher).Return((*filtering.QueryFilteredResult[notifications.UserNotification])(nil), errors.New("repository error"))

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

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}

func TestServiceImpl_UpdateUserNotification(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID
		userID := "test-user-id"
		newStatus := notifications.UserNotificationStatusTypeRead

		// Mock the first call to get existing notification
		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return(fakeNotification, nil).Once()

		// Mock the update call
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateUserNotification), testutils.ContextMatcher, mock.AnythingOfType("*notifications.UserNotification")).Return(nil).Once()

		// Mock the second call to get updated notification
		updatedNotification := *fakeNotification
		updatedNotification.Status = newStatus
		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return(&updatedNotification, nil).Once()

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

		mock.AssertExpectationsForObjects(t, mockRepo)
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
		service, mockRepo := buildTestService(t)

		notificationID := "nonexistent-notification"
		userID := "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return((*notifications.UserNotification)(nil), errors.New("repository error"))

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

		mock.AssertExpectationsForObjects(t, mockRepo)
	})

	t.Run("repository error on update", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, mockRepo := buildTestService(t)

		fakeNotification := notificationsfakes.BuildFakeUserNotification()
		notificationID := fakeNotification.ID
		userID := "test-user-id"

		mockRepo.On(reflection.GetMethodName(mockRepo.GetUserNotification), testutils.ContextMatcher, userID, notificationID).Return(fakeNotification, nil).Once()
		mockRepo.On(reflection.GetMethodName(mockRepo.UpdateUserNotification), testutils.ContextMatcher, mock.AnythingOfType("*notifications.UserNotification")).Return(errors.New("update error")).Once()

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

		mock.AssertExpectationsForObjects(t, mockRepo)
	})
}
