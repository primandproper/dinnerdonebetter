package mealplantasknotifications

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/notifications"

	platformerrors "github.com/primandproper/platform-go/v13/errors"
)

// RequestType labels this worker's pushes in the delivery metrics, alongside the kinds that
// still arrive as messages.
const RequestType = mealplanningnotifications.NotificationKindMealPlanTask

// fallbackTaskName is what a task with neither a prep task name nor a creation explanation is
// called. Both are optional in the schema, and "A task for Dinner on Monday" is a worse
// notification than no notification only in theory — the user still has to be told something is
// due.
const fallbackTaskName = "A task"

// notificationTitle is the same for every meal plan task reminder. The body carries what is
// specific to one.
const notificationTitle = "Meal plan task"

var (
	// errTaskHasNoAccount indicates a task whose meal plan belongs to no account. It cannot
	// be resolved to recipients and never will be, so it fails rather than being retried into
	// the attempt ceiling.
	errTaskHasNoAccount = platformerrors.New("meal plan task has no account")

	// errNoDeviceAccepted indicates that every device the recipients have registered refused
	// the push. The task is left unstamped so the queue offers it again — this is the case
	// the release delay exists for.
	errNoDeviceAccepted = platformerrors.New("no device accepted the meal plan task notification")
)

// content renders one task's reminder.
func content(notificationContext *mealplanning.MealPlanTaskNotificationContext) (title, body string) {
	taskName := notificationContext.PrepTaskName
	if taskName == "" {
		taskName = notificationContext.CreationExplanation
	}
	if taskName == "" {
		taskName = fallbackTaskName
	}

	return notificationTitle, fmt.Sprintf(
		"%s for %s on %s",
		taskName,
		mealplanning.FormatMealNameForDisplay(notificationContext.MealName),
		notificationContext.StartsAt.Format("Monday"),
	)
}
