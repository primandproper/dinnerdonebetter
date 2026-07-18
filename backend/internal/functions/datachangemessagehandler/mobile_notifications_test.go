package datachangemessagehandler

import (
	"encoding/json"
	"strings"
	"testing"

	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	mealplanningnotifications "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"
	domainnotifications "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/notifications/mock"

	"github.com/primandproper/platform-go/v5/filtering"
	notifications "github.com/primandproper/platform-go/v5/notifications/mobile"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMobileNotificationsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), []byte("not json"))

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding")
	})

	t.Run("missing title", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("missing body", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "body")
	})

	t.Run("missing request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "request type")
	})

	t.Run("unknown request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      "unknown_type",
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown request type")
	})

	t.Run("meal plan task requires mealPlanTaskID in context", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "mealPlanTaskID")
	})

	t.Run("idempotent skip when meal plan task already sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.Repository{}
		handler.mealPlanRepo = mealPlanRepo

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
			Context: map[string]string{
				mealplanningnotifications.MealPlanTaskIDContextKey: "task-123",
			},
		}
		raw, _ := json.Marshal(req)

		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MealPlanTaskNotificationHasBeenSent), mock.Anything, "task-123").Return(true, nil).Once()

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mealPlanRepo)
	})

	t.Run("no recipients with meal plan task ID marks sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.Repository{}
		handler.mealPlanRepo = mealPlanRepo

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{},
			Title:            "title",
			Body:             "body",
			Context: map[string]string{
				mealplanningnotifications.MealPlanTaskIDContextKey: "task-123",
			},
		}
		raw, _ := json.Marshal(req)

		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MealPlanTaskNotificationHasBeenSent), mock.Anything, "task-123").Return(false, nil).Once()
		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MarkMealPlanTaskNotificationSent), mock.Anything, "task-123").Return(nil).Once()

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mealPlanRepo)
	})

	t.Run("no device tokens with meal plan task ID marks sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.Repository{}
		notificationsRepo := &notificationsmock.Repository{}
		handler.mealPlanRepo = mealPlanRepo
		handler.notificationsRepo = notificationsRepo

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
			Context: map[string]string{
				mealplanningnotifications.MealPlanTaskIDContextKey: "task-123",
			},
		}
		raw, _ := json.Marshal(req)

		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MealPlanTaskNotificationHasBeenSent), mock.Anything, "task-123").Return(false, nil).Once()
		notificationsRepo.On(reflection.GetMethodName(notificationsRepo.GetUserDeviceTokens), mock.Anything, "user-1", mock.Anything, (*string)(nil)).Return(&filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{Data: []*domainnotifications.UserDeviceToken{}}, nil).Once()
		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MarkMealPlanTaskNotificationSent), mock.Anything, "task-123").Return(nil).Once()

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mealPlanRepo, notificationsRepo)
	})

	t.Run("success sends push and marks meal plan task sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.Repository{}
		notificationsRepo := &notificationsmock.Repository{}
		handler.mealPlanRepo = mealPlanRepo
		handler.notificationsRepo = notificationsRepo

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "Meal plan task",
			Body:             "Chop onions for Dinner on Monday",
			Context: map[string]string{
				mealplanningnotifications.MealPlanTaskIDContextKey: "task-123",
			},
		}
		raw, _ := json.Marshal(req)

		deviceToken := &domainnotifications.UserDeviceToken{
			ID:            "token-1",
			DeviceToken:   strings.Repeat("a", 64),
			Platform:      domainnotifications.UserDeviceTokenPlatformIOS,
			BelongsToUser: "user-1",
		}

		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MealPlanTaskNotificationHasBeenSent), mock.Anything, "task-123").Return(false, nil).Once()
		notificationsRepo.On(reflection.GetMethodName(notificationsRepo.GetUserDeviceTokens), mock.Anything, "user-1", mock.Anything, (*string)(nil)).Return(&filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{
			Data: []*domainnotifications.UserDeviceToken{deviceToken},
		}, nil).Once()
		mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.MarkMealPlanTaskNotificationSent), mock.Anything, "task-123").Return(nil).Once()

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		assert.NoError(t, err)
		mock.AssertExpectationsForObjects(t, mealPlanRepo, notificationsRepo)
	})
}
