package mobilenotificationscheduler

import (
	"context"
	"testing"
	"time"

	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	mealplanningnotifications "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"

	msgqueuemock "github.com/primandproper/platform-go/v5/messagequeue/mock"
	notifications "github.com/primandproper/platform-go/v5/notifications/mobile"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

	mealPlanRepo := &mealplanningmock.Repository{}
	identityRepo := &identitymock.RepositoryMock{}
	publisher := &msgqueuemock.PublisherMock{}

	notificationCtx := &mealplanning.MealPlanTaskNotificationContext{
		PrepTaskName:        "Chop onions",
		CreationExplanation: "",
		MealName:            mealplanning.DinnerMealName,
		StartsAt:            time.Date(2025, 3, 3, 18, 0, 0, 0, time.UTC), // Monday
	}

	mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.GetMealPlanTaskIDsThatNeedNotification), mock.Anything).Return([]string{taskID}, nil).Once()
	mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.GetMealPlanTask), mock.Anything, taskID).Return(task, nil).Once()
	mealPlanRepo.On(reflection.GetMethodName(mealPlanRepo.GetMealPlanTaskNotificationContext), mock.Anything, taskID).Return(notificationCtx, nil).Once()
	// With AssignedToUser set, GetMealPlanTaskAccountID and GetUsersForAccount are not called

	var publishedPayload any
	publisher.PublishFunc = func(_ context.Context, data any) error {
		publishedPayload = data
		return nil
	}

	scheduler := NewScheduler(logger, tracerProvider, mealPlanRepo, identityRepo, publisher)

	err := scheduler.ScheduleNotifications(ctx)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, mealPlanRepo)

	req, ok := publishedPayload.(*notifications.MobileNotificationRequest)
	require.True(t, ok, "expected MobileNotificationRequest to be published")
	assert.Equal(t, []string{assignedUser}, req.RecipientUserIDs)
	assert.Equal(t, "Meal plan task", req.Title)
	assert.Equal(t, "Chop onions for Dinner on Monday", req.Body)
	assert.NotNil(t, req.Context)
	assert.Equal(t, taskID, req.Context[mealplanningnotifications.MealPlanTaskIDContextKey])
}
