package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	mealplanningnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"
	domainnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"

	"github.com/primandproper/platform-go/v11/filtering"
	notifications "github.com/primandproper/platform-go/v11/notifications/mobile"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMobileNotificationsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), []byte("not json"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding")
	})

	t.Run("missing title", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "title")
	})

	t.Run("missing body", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "body")
	})

	t.Run("missing request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "request type")
	})

	t.Run("unknown request type", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      "unknown_type",
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "unknown request type")
	})

	t.Run("meal plan task requires mealPlanTaskID in context", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		req := notifications.MobileNotificationRequest{
			RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
			RecipientUserIDs: []string{"user-1"},
			Title:            "title",
			Body:             "body",
		}
		raw, _ := json.Marshal(req)

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "mealPlanTaskID")
	})

	t.Run("idempotent skip when meal plan task already sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.RepositoryMock{}
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

		mealPlanRepo.MealPlanTaskNotificationHasBeenSentFunc = func(_ context.Context, mealPlanTaskID string) (bool, error) {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return true, nil
		}

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.NoError(t, err)
		assert.Len(t, mealPlanRepo.MealPlanTaskNotificationHasBeenSentCalls(), 1)
	})

	t.Run("no recipients with meal plan task ID marks sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.RepositoryMock{}
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

		mealPlanRepo.MealPlanTaskNotificationHasBeenSentFunc = func(_ context.Context, mealPlanTaskID string) (bool, error) {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return false, nil
		}
		mealPlanRepo.MarkMealPlanTaskNotificationSentFunc = func(_ context.Context, mealPlanTaskID string) error {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return nil
		}

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.NoError(t, err)
		assert.Len(t, mealPlanRepo.MealPlanTaskNotificationHasBeenSentCalls(), 1)
		assert.Len(t, mealPlanRepo.MarkMealPlanTaskNotificationSentCalls(), 1)
	})

	t.Run("no device tokens with meal plan task ID marks sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.RepositoryMock{}
		notificationsRepo := &notificationsmock.RepositoryMock{}
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

		mealPlanRepo.MealPlanTaskNotificationHasBeenSentFunc = func(_ context.Context, mealPlanTaskID string) (bool, error) {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return false, nil
		}
		notificationsRepo.GetUserDeviceTokensFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, (*string)(nil), platformFilter)

			return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{Data: []*domainnotifications.UserDeviceToken{}}, nil
		}
		mealPlanRepo.MarkMealPlanTaskNotificationSentFunc = func(_ context.Context, mealPlanTaskID string) error {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return nil
		}

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.NoError(t, err)
		assert.Len(t, mealPlanRepo.MealPlanTaskNotificationHasBeenSentCalls(), 1)
		assert.Len(t, mealPlanRepo.MarkMealPlanTaskNotificationSentCalls(), 1)
		assert.Len(t, notificationsRepo.GetUserDeviceTokensCalls(), 1)
	})

	t.Run("success sends push and marks meal plan task sent", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)
		mealPlanRepo := &mealplanningmock.RepositoryMock{}
		notificationsRepo := &notificationsmock.RepositoryMock{}
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

		mealPlanRepo.MealPlanTaskNotificationHasBeenSentFunc = func(_ context.Context, mealPlanTaskID string) (bool, error) {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return false, nil
		}
		notificationsRepo.GetUserDeviceTokensFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter, platformFilter *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
			assert.Equal(t, "user-1", userID)
			assert.Equal(t, (*string)(nil), platformFilter)

			return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{Data: []*domainnotifications.UserDeviceToken{deviceToken}}, nil
		}
		mealPlanRepo.MarkMealPlanTaskNotificationSentFunc = func(_ context.Context, mealPlanTaskID string) error {
			assert.Equal(t, "task-123", mealPlanTaskID)

			return nil
		}

		err := handler.MobileNotificationsEventHandler("mobile_notifications")(t.Context(), raw)

		require.NoError(t, err)
		assert.Len(t, mealPlanRepo.MealPlanTaskNotificationHasBeenSentCalls(), 1)
		assert.Len(t, mealPlanRepo.MarkMealPlanTaskNotificationSentCalls(), 1)
		assert.Len(t, notificationsRepo.GetUserDeviceTokensCalls(), 1)
	})
}
