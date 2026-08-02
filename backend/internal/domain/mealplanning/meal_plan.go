package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v9/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MealPlanElectionMethodSchulze is used to denote the Schulze election method.
	MealPlanElectionMethodSchulze = "schulze"
	// MealPlanElectionMethodInstantRunoff is used to denote the Instant Runoff election method.
	MealPlanElectionMethodInstantRunoff = "instant-runoff"

	// MealPlanCreatedServiceEventType indicates a meal plan was created.
	MealPlanCreatedServiceEventType = "meal_plan_created"
	// MealPlanUpdatedServiceEventType indicates a meal plan was updated.
	MealPlanUpdatedServiceEventType = "meal_plan_updated"
	// MealPlanArchivedServiceEventType indicates a meal plan was archived.
	MealPlanArchivedServiceEventType = "meal_plan_archived"
	// MealPlanFinalizedServiceEventType indicates a meal plan was finalized.
	MealPlanFinalizedServiceEventType = "meal_plan_finalized"

	// MealPlanStatusAwaitingVotes indicates a meal plan is still awaiting votes.
	MealPlanStatusAwaitingVotes MealPlanStatus = "awaiting_votes"
	// MealPlanStatusFinalized indicates a meal plan has been finalized.
	MealPlanStatusFinalized MealPlanStatus = "finalized"
)

var (
	errInvalidVotingDeadline    = errors.New("invalid voting deadline")
	errVotingDeadlineAfterStart = errors.New("voting deadline must be before all event start times")
)

func init() {
	gob.Register(new(MealPlan))
	gob.Register(new(MealPlanCreationRequestInput))
	gob.Register(new(MealPlanUpdateRequestInput))
}

type (
	// MealPlanStatus is the type to use/compare against when checking meal plan status.
	MealPlanStatus string

	// MealPlan represents a meal plan.
	MealPlan struct {
		_                      struct{}                         `json:"-"`
		CreatedAt              time.Time                        `json:"createdAt"`
		VotingDeadline         time.Time                        `json:"votingDeadline"`
		ArchivedAt             *time.Time                       `json:"archivedAt"`
		LastUpdatedAt          *time.Time                       `json:"lastUpdatedAt"`
		Status                 string                           `json:"status"`
		ID                     string                           `json:"id"`
		Notes                  string                           `json:"notes"`
		ElectionMethod         string                           `json:"electionMethod"`
		BelongsToAccount       string                           `json:"belongsToAccount"`
		CreatedByUser          string                           `json:"createdBy"`
		Events                 []*MealPlanEvent                 `json:"events"`
		Selections             []*MealPlanRecipeOptionSelection `json:"selections"`
		GroceryListInitialized bool                             `json:"groceryListInitialized"`
		TasksCreated           bool                             `json:"tasksCreated"`
	}

	// MealPlanCreationRequestInput represents what a user could set as input for creating meal plans.
	MealPlanCreationRequestInput struct {
		_ struct{} `json:"-"`

		VotingDeadline time.Time                                            `json:"votingDeadline"`
		Notes          string                                               `json:"notes"`
		ElectionMethod string                                               `json:"electionMethod"`
		Events         []*MealPlanEventCreationRequestInput                 `json:"events"`
		Selections     []*MealPlanRecipeOptionSelectionCreationRequestInput `json:"selections,omitempty"`
	}

	// MealPlanDatabaseCreationInput represents what a user could set as input for creating meal plans.
	MealPlanDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		VotingDeadline   time.Time                                             `json:"-"`
		BelongsToAccount string                                                `json:"-"`
		Notes            string                                                `json:"-"`
		ID               string                                                `json:"-"`
		ElectionMethod   string                                                `json:"-"`
		CreatedByUser    string                                                `json:"-"`
		Events           []*MealPlanEventDatabaseCreationInput                 `json:"-"`
		Selections       []*MealPlanRecipeOptionSelectionDatabaseCreationInput `json:"-"`
	}

	// MealPlanUpdateRequestInput represents what a user could set as input for updating meal plans.
	MealPlanUpdateRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToAccount *string    `json:"-"`
		Notes            *string    `json:"notes,omitempty"`
		VotingDeadline   *time.Time `json:"votingDeadline,omitempty"`
	}

	// FinalizedMealPlanDatabaseResult represents what is returned by the above query.
	FinalizedMealPlanDatabaseResult struct {
		MealPlanID       string
		MealPlanEventID  string
		MealPlanOptionID string
		MealID           string
		RecipeIDs        []string
	}

	// MealPlanDataManager describes a structure capable of storing meal plans permanently.
	MealPlanDataManager interface {
		MealPlanExists(ctx context.Context, mealPlanID, accountID string) (bool, error)
		GetMealPlan(ctx context.Context, mealPlanID, accountID string) (*MealPlan, error)
		GetMealPlansForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealPlan], error)
		CreateMealPlan(ctx context.Context, input *MealPlanDatabaseCreationInput) (*MealPlan, error)
		UpdateMealPlan(ctx context.Context, updated *MealPlan) error
		ArchiveMealPlan(ctx context.Context, mealPlanID, accountID string) error
		AttemptToFinalizeMealPlan(ctx context.Context, mealPlanID, accountID string) (bool, error)
		MarkMealPlanAsGroceryListInitialized(ctx context.Context, mealPlanID string) error
		GetFinalizedMealPlanIDsForTheNextWeek(ctx context.Context) ([]*FinalizedMealPlanDatabaseResult, error)
		GetUnfinalizedMealPlansWithExpiredVotingPeriods(ctx context.Context) ([]*MealPlan, error)
		GetFinalizedMealPlansWithUninitializedGroceryLists(ctx context.Context) ([]*MealPlan, error)
	}
)

// Update merges an MealPlanUpdateRequestInput with a meal plan.
func (x *MealPlan) Update(input *MealPlanUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.VotingDeadline != nil && !input.VotingDeadline.Equal(x.VotingDeadline) {
		x.VotingDeadline = *input.VotingDeadline
	}
}

var _ validation.ValidatableWithContext = (*MealPlanCreationRequestInput)(nil)

// ValidateWithContext validates a MealPlanCreationRequestInput.
func (x *MealPlanCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	if time.Now().After(x.VotingDeadline) {
		return errInvalidVotingDeadline
	}

	// Validate required fields first
	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.VotingDeadline, validation.Required),
		validation.Field(&x.Events, validation.Required),
		validation.Field(&x.ElectionMethod, validation.In(MealPlanElectionMethodSchulze, MealPlanElectionMethodInstantRunoff)),
	); err != nil {
		return err
	}

	// Validate that voting deadline is before every event's start time
	for _, event := range x.Events {
		if !x.VotingDeadline.Before(event.StartsAt) {
			return errVotingDeadlineAfterStart
		}
	}

	return nil
}

var _ validation.ValidatableWithContext = (*MealPlanDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealPlanDatabaseCreationInput.
func (x *MealPlanDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.VotingDeadline, validation.Required),
		validation.Field(&x.BelongsToAccount, validation.Required),
		validation.Field(&x.CreatedByUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanUpdateRequestInput)(nil)

// ValidateWithContext validates a MealPlanUpdateRequestInput.
func (x *MealPlanUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	// Every field on an update is optional (see MealPlan.Update, which only applies non-nil fields).
	// Now that this validator is actually invoked by the manager, requiring VotingDeadline here would
	// reject legitimate partial updates such as changing only the notes. Mirrors MealPlanOptionUpdateRequestInput.
	return validation.ValidateStructWithContext(
		ctx,
		x,
	)
}
