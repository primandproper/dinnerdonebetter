package mobilenotificationscheduler

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningnotifications "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"

	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/messagequeue"
	notifications "github.com/primandproper/platform-go/v3/notifications/mobile"
	"github.com/primandproper/platform-go/v3/observability"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const schedulerTracerName = "mobile_notification_scheduler"

// Scheduler publishes meal plan task notification requests to the mobile_notifications queue.
type Scheduler struct {
	logger                       logging.Logger
	tracer                       tracing.Tracer
	mealPlanningRepo             mealplanning.Repository
	identityRepo                 identity.Repository
	mobileNotificationsPublisher messagequeue.Publisher
}

// NewScheduler creates a new mobile notification scheduler.
func NewScheduler(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	mealPlanningRepo mealplanning.Repository,
	identityRepo identity.Repository,
	mobileNotificationsPublisher messagequeue.Publisher,
) *Scheduler {
	return &Scheduler{
		logger:                       logging.NewNamedLogger(logger, schedulerTracerName),
		tracer:                       tracing.NewNamedTracer(tracerProvider, schedulerTracerName),
		mealPlanningRepo:             mealPlanningRepo,
		identityRepo:                 identityRepo,
		mobileNotificationsPublisher: mobileNotificationsPublisher,
	}
}

// ScheduleNotifications runs all notification schedulers. Each scheduler queries for items
// that need notifications and publishes them to the queue. Meal plan tasks are the first
// use case; additional notification types can be added by registering more handlers.
func (s *Scheduler) ScheduleNotifications(ctx context.Context) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	errs := &multierror.Error{}
	if err := s.scheduleMealPlanTaskNotifications(ctx); err != nil {
		errs = multierror.Append(errs, err)
	}
	// Future notification types: add more handler calls here.

	return errs.ErrorOrNil()
}

// scheduleMealPlanTaskNotifications queries for meal plan tasks that need notifications,
// formats the message, resolves recipients, and publishes MobileNotificationRequest.
func (s *Scheduler) scheduleMealPlanTaskNotifications(ctx context.Context) error {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	taskIDs, err := s.mealPlanningRepo.GetMealPlanTaskIDsThatNeedNotification(ctx)
	if err != nil {
		return observability.PrepareAndLogError(err, s.logger, span, "getting meal plan task IDs that need notification")
	}

	if len(taskIDs) == 0 {
		return nil
	}

	s.logger.WithValue("count", len(taskIDs)).WithValue("type", "meal_plan_task").Info("publishing mobile notification requests")

	errs := &multierror.Error{}
	for _, taskID := range taskIDs {
		var req *notifications.MobileNotificationRequest
		req, err = s.buildMealPlanTaskNotificationRequest(ctx, taskID)
		if err != nil {
			errs = multierror.Append(errs, fmt.Errorf("building meal plan task notification for %s: %w", taskID, err))
			continue
		}
		if err = s.mobileNotificationsPublisher.Publish(ctx, req); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("publishing meal plan task notification for %s: %w", taskID, err))
		}
	}

	return errs.ErrorOrNil()
}

func (s *Scheduler) buildMealPlanTaskNotificationRequest(ctx context.Context, taskID string) (*notifications.MobileNotificationRequest, error) {
	task, err := s.mealPlanningRepo.GetMealPlanTask(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("fetching meal plan task: %w", err)
	}
	if task == nil {
		return nil, fmt.Errorf("meal plan task %s not found", taskID)
	}

	notificationCtx, err := s.mealPlanningRepo.GetMealPlanTaskNotificationContext(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("fetching meal plan task notification context: %w", err)
	}
	if notificationCtx == nil {
		return nil, fmt.Errorf("meal plan task notification context %s not found", taskID)
	}

	recipientUserIDs, err := s.resolveNotificationRecipients(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("resolving recipients: %w", err)
	}

	title, body := buildMealPlanTaskNotificationContent(notificationCtx)

	return &notifications.MobileNotificationRequest{
		RequestType:      mealplanningnotifications.MobileNotificationRequestTypeMealPlanTask,
		RecipientUserIDs: recipientUserIDs,
		Title:            title,
		Body:             body,
		Context: map[string]string{
			mealplanningnotifications.MealPlanTaskIDContextKey: taskID,
		},
	}, nil
}

func (s *Scheduler) resolveNotificationRecipients(ctx context.Context, task *mealplanning.MealPlanTask) ([]string, error) {
	if task.AssignedToUser != nil && *task.AssignedToUser != "" {
		return []string{*task.AssignedToUser}, nil
	}

	accountID, err := s.mealPlanningRepo.GetMealPlanTaskAccountID(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("getting account ID for task: %w", err)
	}
	if accountID == "" {
		return nil, fmt.Errorf("task has no account")
	}

	usersResult, err := s.identityRepo.GetUsersForAccount(ctx, accountID, filtering.DefaultQueryFilter())
	if err != nil {
		return nil, fmt.Errorf("getting users for account: %w", err)
	}
	userIDs := make([]string, 0, len(usersResult.Data))
	for _, u := range usersResult.Data {
		if u != nil && u.ID != "" {
			userIDs = append(userIDs, u.ID)
		}
	}
	return userIDs, nil
}

func buildMealPlanTaskNotificationContent(ctx *mealplanning.MealPlanTaskNotificationContext) (title, body string) {
	taskName := ctx.PrepTaskName
	if taskName == "" {
		taskName = ctx.CreationExplanation
	}
	if taskName == "" {
		taskName = "A task"
	}
	mealName := mealplanning.FormatMealNameForDisplay(ctx.MealName)
	dayName := ctx.StartsAt.Format("Monday")
	return "Meal plan task", fmt.Sprintf("%s for %s on %s", taskName, mealName, dayName)
}
