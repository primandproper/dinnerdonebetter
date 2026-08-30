package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v13/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MealPlanEventCreatedServiceEventType indicates a meal plan event was created.
	MealPlanEventCreatedServiceEventType = "meal_plan_event_created"
	// MealPlanEventUpdatedServiceEventType indicates a meal plan event was updated.
	MealPlanEventUpdatedServiceEventType = "meal_plan_event_updated"
	// MealPlanEventArchivedServiceEventType indicates a meal plan event was archived.
	MealPlanEventArchivedServiceEventType = "meal_plan_event_archived"

	// BreakfastMealName represents breakfast.
	BreakfastMealName = "breakfast"
	// SecondBreakfastMealName represents second breakfast.
	SecondBreakfastMealName = "second_breakfast"
	// BrunchMealName represents brunch.
	BrunchMealName = "brunch"
	// LunchMealName represents lunch.
	LunchMealName = "lunch"
	// SupperMealName represents supper.
	SupperMealName = "supper"
	// DinnerMealName represents dinner.
	DinnerMealName = "dinner"
)

var (
	errStartsAfterItEnds = errors.New("invalid start and end dates")
)

func init() {
	gob.Register(new(MealPlanEvent))
	gob.Register(new(MealPlanEventCreationRequestInput))
	gob.Register(new(MealPlanEventUpdateRequestInput))
}

type (
	// MealPlanEvent represents a meal plan event.
	MealPlanEvent struct {
		_ struct{} `json:"-"`

		CreatedAt         time.Time         `json:"createdAt"`
		StartsAt          time.Time         `json:"startsAt"`
		EndsAt            time.Time         `json:"endsAt"`
		ArchivedAt        *time.Time        `json:"archivedAt"`
		LastUpdatedAt     *time.Time        `json:"lastUpdatedAt"`
		MealName          string            `json:"mealName"`
		Notes             string            `json:"notes"`
		BelongsToMealPlan string            `json:"belongsToMealPlan"`
		ID                string            `json:"id"`
		Options           []*MealPlanOption `json:"options"`
	}

	// MealPlanEventCreationRequestInput represents what a user could set as input for creating meal plan events.
	MealPlanEventCreationRequestInput struct {
		_ struct{} `json:"-"`

		EndsAt   time.Time                             `json:"endsAt"`
		StartsAt time.Time                             `json:"startsAt"`
		Notes    string                                `json:"notes"`
		MealName string                                `json:"mealName"`
		Options  []*MealPlanOptionCreationRequestInput `json:"options"`
	}

	// MealPlanEventDatabaseCreationInput represents what a user could set as input for creating meal plan events.
	MealPlanEventDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		StartsAt          time.Time                              `json:"-"`
		EndsAt            time.Time                              `json:"-"`
		BelongsToMealPlan string                                 `json:"-"`
		Notes             string                                 `json:"-"`
		MealName          string                                 `json:"-"`
		ID                string                                 `json:"-"`
		Options           []*MealPlanOptionDatabaseCreationInput `json:"-"`
	}

	// MealPlanEventUpdateRequestInput represents what a user could set as input for updating meal plan events.
	MealPlanEventUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes             *string    `json:"notes,omitempty"`
		StartsAt          *time.Time `json:"startsAt,omitempty"`
		MealName          *string    `json:"mealName,omitempty"`
		EndsAt            *time.Time `json:"endsAt,omitempty"`
		BelongsToMealPlan string     `json:"-"`
	}

	// MealPlanEventDataManager describes a structure capable of storing meal plan events permanently.
	MealPlanEventDataManager interface {
		MealPlanEventExists(ctx context.Context, mealPlanID, mealPlanEventID string) (bool, error)
		GetMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*MealPlanEvent, error)
		GetMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealPlanEvent], error)
		MealPlanEventIsEligibleForVoting(ctx context.Context, mealPlanID, mealPlanEventID string) (bool, error)
		CreateMealPlanEvent(ctx context.Context, input *MealPlanEventDatabaseCreationInput) (*MealPlanEvent, error)
		UpdateMealPlanEvent(ctx context.Context, updated *MealPlanEvent) error
		SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error
		ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error
	}
)

// Update merges an MealPlanEventUpdateRequestInput with a meal plan event.
func (x *MealPlanEvent) Update(input *MealPlanEventUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.StartsAt != nil && *input.StartsAt != x.StartsAt {
		x.StartsAt = *input.StartsAt
	}

	if input.EndsAt != nil && *input.EndsAt != x.EndsAt {
		x.EndsAt = *input.EndsAt
	}

	if input.MealName != nil && *input.MealName != x.MealName {
		x.MealName = *input.MealName
	}
}

var _ validation.ValidatableWithContext = (*MealPlanEventCreationRequestInput)(nil)

// ValidateWithContext validates a MealPlanEventCreationRequestInput.
func (x *MealPlanEventCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	if x.StartsAt.Equal(x.EndsAt) || x.StartsAt.After(x.EndsAt) {
		return errStartsAfterItEnds
	}

	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.StartsAt, validation.Required),
		validation.Field(&x.EndsAt, validation.Required),
		validation.Field(&x.MealName, validation.In(
			BreakfastMealName,
			SecondBreakfastMealName,
			BrunchMealName,
			LunchMealName,
			SupperMealName,
			DinnerMealName,
		)),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanEventDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealPlanEventDatabaseCreationInput.
func (x *MealPlanEventDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.StartsAt, validation.Required),
		validation.Field(&x.EndsAt, validation.Required),
		validation.Field(&x.MealName, validation.In(
			BreakfastMealName,
			SecondBreakfastMealName,
			BrunchMealName,
			LunchMealName,
			SupperMealName,
			DinnerMealName,
		)),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanEventUpdateRequestInput)(nil)

// ValidateWithContext validates a MealPlanEventUpdateRequestInput.
func (x *MealPlanEventUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.StartsAt, validation.Required),
		validation.Field(&x.EndsAt, validation.Required),
		validation.Field(&x.MealName, validation.In(
			BreakfastMealName,
			SecondBreakfastMealName,
			BrunchMealName,
			LunchMealName,
			SupperMealName,
			DinnerMealName,
		)),
	)
}

// FormatMealNameForDisplay converts a meal name enum value to a user-friendly display string.
func FormatMealNameForDisplay(mealName string) string {
	switch mealName {
	case BreakfastMealName:
		return "Breakfast"
	case SecondBreakfastMealName:
		return "Second Breakfast"
	case BrunchMealName:
		return "Brunch"
	case LunchMealName:
		return "Lunch"
	case SupperMealName:
		return "Supper"
	case DinnerMealName:
		return "Dinner"
	default:
		return "Meal"
	}
}
