package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v13/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MealPlanRecipeOptionSelectionTypeIngredient represents the ingredient selection type.
	MealPlanRecipeOptionSelectionTypeIngredient = "ingredient"
	// MealPlanRecipeOptionSelectionTypeInstrument represents the instrument selection type.
	MealPlanRecipeOptionSelectionTypeInstrument = "instrument"
	// MealPlanRecipeOptionSelectionTypeVessel represents the vessel selection type.
	MealPlanRecipeOptionSelectionTypeVessel = "vessel"

	// MealPlanRecipeOptionSelectionCreatedServiceEventType indicates a meal plan recipe option selection was created.
	MealPlanRecipeOptionSelectionCreatedServiceEventType = "meal_plan_recipe_option_selection_created"
	// MealPlanRecipeOptionSelectionUpdatedServiceEventType indicates a meal plan recipe option selection was updated.
	MealPlanRecipeOptionSelectionUpdatedServiceEventType = "meal_plan_recipe_option_selection_updated"
	// MealPlanRecipeOptionSelectionArchivedServiceEventType indicates a meal plan recipe option selection was archived.
	MealPlanRecipeOptionSelectionArchivedServiceEventType = "meal_plan_recipe_option_selection_archived"
)

func init() {
	gob.Register(new(MealPlanRecipeOptionSelection))
	gob.Register(new(MealPlanRecipeOptionSelectionDatabaseCreationInput))
	gob.Register(new(MealPlanRecipeOptionSelectionCreationRequestInput))
	gob.Register(new(MealPlanRecipeOptionSelectionUpdateRequestInput))
}

type (
	// MealPlanRecipeOptionSelection represents a user's selection for a recipe option group.
	MealPlanRecipeOptionSelection struct {
		_ struct{} `json:"-"`

		CreatedAt               time.Time  `json:"createdAt"`
		LastUpdatedAt           *time.Time `json:"lastUpdatedAt"`
		ArchivedAt              *time.Time `json:"archivedAt"`
		ID                      string     `json:"id"`
		BelongsToMealPlanOption string     `json:"belongsToMealPlanOption"`
		RecipeID                string     `json:"recipeID"`
		RecipeStepID            string     `json:"recipeStepID"`
		SelectionType           string     `json:"selectionType"`
		IngredientIndex         uint16     `json:"ingredientIndex"`
		SelectedOptionIndex     uint16     `json:"selectedOptionIndex"`
	}

	// MealPlanRecipeOptionSelectionDatabaseCreationInput represents what a user could set as input for creating meal plan recipe option selections.
	MealPlanRecipeOptionSelectionDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                      string `json:"-"`
		BelongsToMealPlanOption string `json:"-"`
		RecipeID                string `json:"-"`
		RecipeStepID            string `json:"-"`
		SelectionType           string `json:"-"`
		IngredientIndex         uint16 `json:"-"`
		SelectedOptionIndex     uint16 `json:"-"`
	}

	// MealPlanRecipeOptionSelectionCreationRequestInput represents what a user could set as input for creating meal plan recipe option selections.
	MealPlanRecipeOptionSelectionCreationRequestInput struct {
		_ struct{} `json:"-"`

		RecipeID            string `json:"recipeID"`
		RecipeStepID        string `json:"recipeStepID"`
		SelectionType       string `json:"selectionType"`
		IngredientIndex     uint16 `json:"ingredientIndex"`
		SelectedOptionIndex uint16 `json:"selectedOptionIndex"`
	}

	// MealPlanRecipeOptionSelectionUpdateRequestInput represents what a user could set as input for updating meal plan recipe option selections.
	MealPlanRecipeOptionSelectionUpdateRequestInput struct {
		_ struct{} `json:"-"`

		SelectedOptionIndex *uint16 `json:"selectedOptionIndex,omitempty"`
	}

	// MealPlanRecipeOptionSelectionDataManager describes a structure capable of storing meal plan recipe option selections permanently.
	MealPlanRecipeOptionSelectionDataManager interface {
		GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*MealPlanRecipeOptionSelection, error)
		GetSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealPlanRecipeOptionSelection], error)
		CreateMealPlanRecipeOptionSelection(ctx context.Context, input *MealPlanRecipeOptionSelectionDatabaseCreationInput) (*MealPlanRecipeOptionSelection, error)
		UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *MealPlanRecipeOptionSelectionUpdateRequestInput) error
		ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error
	}
)

// Update updates a MealPlanRecipeOptionSelection with the provided input.
func (x *MealPlanRecipeOptionSelection) Update(input *MealPlanRecipeOptionSelectionUpdateRequestInput) {
	if input.SelectedOptionIndex != nil {
		x.SelectedOptionIndex = *input.SelectedOptionIndex
	}
}

var _ validation.ValidatableWithContext = (*MealPlanRecipeOptionSelectionCreationRequestInput)(nil)

// ValidateWithContext validates a MealPlanRecipeOptionSelectionCreationRequestInput.
func (x *MealPlanRecipeOptionSelectionCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.RecipeID, validation.Required),
		validation.Field(&x.RecipeStepID, validation.Required),
		validation.Field(&x.SelectionType, validation.Required, validation.In(
			MealPlanRecipeOptionSelectionTypeIngredient,
			MealPlanRecipeOptionSelectionTypeInstrument,
			MealPlanRecipeOptionSelectionTypeVessel,
		)),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanRecipeOptionSelectionDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealPlanRecipeOptionSelectionDatabaseCreationInput.
func (x *MealPlanRecipeOptionSelectionDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.BelongsToMealPlanOption, validation.Required),
		validation.Field(&x.RecipeID, validation.Required),
		validation.Field(&x.RecipeStepID, validation.Required),
		validation.Field(&x.SelectionType, validation.Required, validation.In(
			MealPlanRecipeOptionSelectionTypeIngredient,
			MealPlanRecipeOptionSelectionTypeInstrument,
			MealPlanRecipeOptionSelectionTypeVessel,
		)),
	)
}
