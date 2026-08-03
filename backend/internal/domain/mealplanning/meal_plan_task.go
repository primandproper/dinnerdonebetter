package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v9/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MealPlanTaskCreatedServiceEventType indicates a meal plan task was created.
	MealPlanTaskCreatedServiceEventType = "meal_plan_task_created"
	// MealPlanTaskStatusChangedServiceEventType indicates a meal plan task's status was changed.
	MealPlanTaskStatusChangedServiceEventType = "meal_plan_task_status_changed"

	// MealPlanTaskStatusUnfinished represents the unfinished enum member for meal plan task status in the DB.
	MealPlanTaskStatusUnfinished = "unfinished"
	// MealPlanTaskStatusPostponed represents the postponed enum member for meal plan task status in the DB.
	MealPlanTaskStatusPostponed = "postponed"
	// MealPlanTaskStatusIgnored represents the ignored enum member for meal plan task status in the DB.
	MealPlanTaskStatusIgnored = "ignored"
	// MealPlanTaskStatusCanceled represents the canceled enum member for meal plan task status in the DB.
	MealPlanTaskStatusCanceled = "canceled"
	// MealPlanTaskStatusFinished represents the finished enum member for meal plan task status in the DB.
	MealPlanTaskStatusFinished = "finished"
)

func init() {
	gob.Register(new(MealPlanTask))
}

type (
	// MealPlanTask represents a meal plan task.
	MealPlanTask struct {
		_ struct{} `json:"-"`

		RecipePrepTask      RecipePrepTask `json:"recipePrepTask"`
		CreatedAt           time.Time      `json:"createdAt"`
		LastUpdatedAt       *time.Time     `json:"lastUpdatedAt"`
		CompletedAt         *time.Time     `json:"completedAt"`
		NotificationSentAt  *time.Time     `json:"notificationSentAt"`
		AssignedToUser      *string        `json:"assignedToUser"`
		ID                  string         `json:"id"`
		Status              string         `json:"status"`
		CreationExplanation string         `json:"creationExplanation"`
		StatusExplanation   string         `json:"statusExplanation"`
		MealPlanOption      MealPlanOption `json:"mealPlanOption"`
	}

	// MealPlanTaskCreationRequestInput represents a meal plan task.
	MealPlanTaskCreationRequestInput struct {
		_ struct{} `json:"-"`

		AssignedToUser      *string `json:"assignedToUser"`
		Status              string  `json:"status"`
		CreationExplanation string  `json:"creationExplanation"`
		StatusExplanation   string  `json:"statusExplanation"`
		MealPlanOptionID    string  `json:"mealPlanOptionID"`
		RecipePrepTaskID    string  `json:"recipePrepTaskID"`
	}

	// MealPlanTaskDatabaseCreationInput represents what a user could set as input for creating meal plan tasks.
	MealPlanTaskDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		AssignedToUser      *string `json:"-"`
		CreationExplanation string  `json:"-"`
		StatusExplanation   string  `json:"-"`
		MealPlanOptionID    string  `json:"-"`
		RecipePrepTaskID    string  `json:"-"`
		ID                  string  `json:"-"`
	}

	// MealPlanTaskStatusChangeRequestInput represents what a user could set as input for updating meal plan tasks.
	MealPlanTaskStatusChangeRequestInput struct {
		_ struct{} `json:"-"`

		Status            *string `json:"status"`
		StatusExplanation string  `json:"statusExplanation"`
		AssignedToUser    *string `json:"assignedToUser"`
		MealPlanTaskID    string  `json:"-"`
	}

	// MealPlanTaskDatabaseCreationEstimate represents what a user could set as input for creating meal plan tasks.
	MealPlanTaskDatabaseCreationEstimate struct {
		_ struct{} `json:"-"`

		CreationExplanation string `json:"creationExplanation"`
	}

	// MealPlanTaskDataManager describes a structure capable of storing meal plan tasks permanently.
	MealPlanTaskDataManager interface {
		MealPlanTaskExists(ctx context.Context, mealPlanID, mealPlanTaskID string) (bool, error)
		CreateMealPlanTask(ctx context.Context, input *MealPlanTaskDatabaseCreationInput) (*MealPlanTask, error)
		GetMealPlanTask(ctx context.Context, mealPlanTaskID string) (*MealPlanTask, error)
		GetMealPlanTasksForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealPlanTask], error)
		// CreateMealPlanTasksForMealPlan writes a meal plan's tasks and the flag saying they were
		// written in one transaction. The two are deliberately not separable: a mark that fails
		// after the tasks commit leaves the plan selectable again, and the next run creates every
		// task a second time.
		CreateMealPlanTasksForMealPlan(ctx context.Context, mealPlanID string, inputs []*MealPlanTaskDatabaseCreationInput) ([]*MealPlanTask, error)
		// UndoMealPlanTaskCreation is CreateMealPlanTasksForMealPlan's compensation: it deletes
		// the named tasks and clears the flag, in the one transaction that wrote them. Only the
		// IDs the caller passes are removed, so it cannot reach past the work it is undoing.
		UndoMealPlanTaskCreation(ctx context.Context, mealPlanID string, taskIDs []string) error
		ChangeMealPlanTaskStatus(ctx context.Context, input *MealPlanTaskStatusChangeRequestInput) error
		MealPlanTaskNotificationHasBeenSent(ctx context.Context, mealPlanTaskID string) (bool, error)
		MarkMealPlanTaskNotificationSent(ctx context.Context, mealPlanTaskID string) error
		ClearMealPlanTaskNotificationSentForEvent(ctx context.Context, mealPlanEventID string) error
		GetMealPlanTaskIDsThatNeedNotification(ctx context.Context) ([]string, error)
		GetMealPlanTaskAccountID(ctx context.Context, mealPlanTaskID string) (string, error)
		GetMealPlanTaskNotificationContext(ctx context.Context, mealPlanTaskID string) (*MealPlanTaskNotificationContext, error)
	}

	// MealPlanTaskNotificationContext holds data needed to build a meal plan task notification.
	MealPlanTaskNotificationContext struct {
		StartsAt            time.Time
		PrepTaskName        string
		CreationExplanation string
		MealName            string
	}
)

type MealPlanTaskList []*MealPlanTask

func (m MealPlanTaskList) Len() int {
	return len(m)
}

func (m MealPlanTaskList) Less(i, j int) bool {
	return m[i].ID < m[j].ID
}

func (m MealPlanTaskList) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

// Update merges an MealPlanTaskStatusChangeRequestInput with a meal plan task.
func (x *MealPlanTask) Update(input *MealPlanTaskStatusChangeRequestInput) {
	if input.StatusExplanation != "" && input.StatusExplanation != x.StatusExplanation {
		x.StatusExplanation = input.StatusExplanation
	}

	if input.Status != nil && *input.Status != x.Status {
		x.Status = *input.Status
	}

	if input.AssignedToUser != nil && *input.AssignedToUser != "" {
		if x.AssignedToUser == nil || *x.AssignedToUser != *input.AssignedToUser {
			x.AssignedToUser = input.AssignedToUser
		}
	}
}

var _ validation.ValidatableWithContext = (*MealPlanTaskCreationRequestInput)(nil)

// ValidateWithContext validates a MealPlanTaskDatabaseCreationInput.
func (x *MealPlanTaskCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.MealPlanOptionID, validation.Required),
		validation.Field(&x.RecipePrepTaskID, validation.Required),
		validation.Field(&x.Status, validation.In(
			MealPlanTaskStatusUnfinished,
			MealPlanTaskStatusPostponed,
			MealPlanTaskStatusIgnored,
			MealPlanTaskStatusCanceled,
			MealPlanTaskStatusFinished,
		)),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanTaskDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealPlanTaskDatabaseCreationInput.
// RecipePrepTaskID is deliberately optional: system-generated thaw tasks for frozen
// ingredients have no backing recipe prep task (persisted as NULL).
func (x *MealPlanTaskDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.MealPlanOptionID, validation.Required),
	)
}

// ValidateWithContext validates a MealPlanTaskStatusChangeRequestInput.
func (x *MealPlanTaskStatusChangeRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.MealPlanTaskID, validation.Required),
		validation.Field(&x.Status, validation.In(
			MealPlanTaskStatusUnfinished,
			MealPlanTaskStatusPostponed,
			MealPlanTaskStatusIgnored,
			MealPlanTaskStatusCanceled,
			MealPlanTaskStatusFinished,
		)),
	)
}
