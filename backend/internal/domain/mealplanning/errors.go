package mealplanning

import (
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
)

var (
	// ErrDuplicateMeal is returned when creating a meal that already exists (same name and components for the creator).
	ErrDuplicateMeal = platformerrors.New("meal with same name and components already exists")
	// ErrDuplicateMealInList is returned when adding a meal to a list that already contains it.
	ErrDuplicateMealInList = platformerrors.New("meal already exists in list")
	// ErrDuplicateMealPlanOption is returned when adding a meal as an option to an event that already has it.
	ErrDuplicateMealPlanOption = platformerrors.New("meal already exists as option for this event")

	// ErrNoMatchingMeal is a sentinel returned when FindMealWithSameComponents finds no duplicate.
	// It is not an error; callers should treat it as "no match found" and proceed.
	ErrNoMatchingMeal = platformerrors.New("no meal with matching components found")

	// ErrMealPlanEventNotEligibleForVoting is returned when votes are submitted for an event whose
	// voting deadline has passed or whose meal plan is no longer awaiting votes.
	ErrMealPlanEventNotEligibleForVoting = platformerrors.New("meal plan event is not eligible for voting")
	// ErrMealPlanOptionNotFoundForEvent is returned when a submitted vote names an option that does
	// not belong to the targeted meal plan event.
	ErrMealPlanOptionNotFoundForEvent = platformerrors.New("meal plan option does not belong to event")
)

// ErrInvalidRecipeInput is returned when a recipe a caller sent cannot be stored as
// described: a recipe with one step, a step naming neither an instrument nor a vessel, or
// bridge-table references that do not agree with the steps naming them — a
// ValidPreparationInstrument for one preparation used in a step that runs another, an
// ingredient's preparation bridge naming a different ingredient.
//
// It exists to give those refusals a gRPC code. They are the caller's mistake and read as
// InvalidArgument; without a sentinel they reached a client as Internal, which tells
// somebody the server broke when what broke is the recipe they sent. The detail — which
// step, which instrument, which preparation — travels in the status details, because a
// status message is deliberately generic and a caller fixing this needs to know which of
// thirty references was wrong.
var ErrInvalidRecipeInput = platformerrors.New("recipe cannot be stored as described")
