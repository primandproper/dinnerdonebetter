package mealplanning

import (
	platformerrors "github.com/primandproper/platform-go/v7/errors"
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
