package mobilenotificationscheduler

import (
	"context"
	"testing"
	"time"

	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	mealplanningnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"

	msgqueuemock "github.com/primandproper/platform-go/v9/messagequeue/mock"
	notifications "github.com/primandproper/platform-go/v9/notifications/mobile"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduler_ScheduleNotifications_publishesMobileNotificationRequest(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	taskID := fakes.BuildFakeID()
	task := fakes.BuildFakeMealPlanTask()
	task.ID = taskID
	task.RecipePrepTask.Name = "Chop onions"
	assignedUser := fakes.BuildFakeID()
	task.AssignedToUser = &assignedUser

	mealPlanRepo := &mealplanningmock.RepositoryMock{}
	identityRepo := &identitymock.RepositoryMock{}
	publisher := &msgqueuemock.PublisherMock{}

	notificationCtx := &mealplanning.MealPlanTaskNotificationContext{
		PrepTaskName:        "Chop onions",
		CreationExplanation: "",
		MealName:            mealplanning.DinnerMealName,
		StartsAt:            time.Date(2025, 3, 3, 18, 0, 0, 0, time.UTC), // Monday
	}

	mealPlanRepo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(_ context.Context) ([]string, error) {
		return []string{taskID}, nil
	}
	mealPlanRepo.GetMealPlanTaskFunc = func(_ context.Context, mealPlanTaskID string) (*mealplanning.MealPlanTask, error) {
		assert.Equal(t, taskID, mealPlanTaskID)

		return task, nil
	}
	mealPlanRepo.GetMealPlanTaskNotificationContextFunc = func(_ context.Context, mealPlanTaskID string) (*mealplanning.MealPlanTaskNotificationContext, error) {
		assert.Equal(t, taskID, mealPlanTaskID)

		return notificationCtx, nil
	}
	// With AssignedToUser set, GetMealPlanTaskAccountID and GetUsersForAccount are not called

	var publishedPayload any
	publisher.PublishFunc = func(_ context.Context, data any) error {
		publishedPayload = data
		return nil
	}

	scheduler := NewScheduler(logger, tracerProvider, mealPlanRepo, identityRepo, publisher)

	err := scheduler.ScheduleNotifications(ctx)

	require.NoError(t, err)
	assert.Len(t, mealPlanRepo.GetMealPlanTaskIDsThatNeedNotificationCalls(), 1)
	assert.Len(t, mealPlanRepo.GetMealPlanTaskCalls(), 1)
	assert.Len(t, mealPlanRepo.GetMealPlanTaskNotificationContextCalls(), 1)

	req, ok := publishedPayload.(*notifications.MobileNotificationRequest)
	require.True(t, ok, "expected MobileNotificationRequest to be published")
	assert.Equal(t, []string{assignedUser}, req.RecipientUserIDs)
	assert.Equal(t, "Meal plan task", req.Title)
	assert.Equal(t, "Chop onions for Dinner on Monday", req.Body)
	assert.NotNil(t, req.Context)
	assert.Equal(t, taskID, req.Context[mealplanningnotifications.MealPlanTaskIDContextKey])
}
