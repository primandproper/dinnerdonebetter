package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v9/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	maxIngredientsPerStep = 100

	// RecipeStepCreatedServiceEventType indicates a recipe step was created.
	RecipeStepCreatedServiceEventType = "recipe_step_created"
	// RecipeStepUpdatedServiceEventType indicates a recipe step was updated.
	RecipeStepUpdatedServiceEventType = "recipe_step_updated"
	// RecipeStepArchivedServiceEventType indicates a recipe step was archived.
	RecipeStepArchivedServiceEventType = "recipe_step_archived"
)

var (
	errOneInstrumentOrVesselRequired = errors.New("at least one instrument or vessel is required")
)

func init() {
	gob.Register(new(RecipeStep))
	gob.Register(new(RecipeStepCreationRequestInput))
	gob.Register(new(RecipeStepUpdateRequestInput))
}

type (
	// RecipeStep represents a recipe step.
	RecipeStep struct {
		_                         struct{}                         `json:"-"`
		CreatedAt                 time.Time                        `json:"createdAt"`
		ArchivedAt                *time.Time                       `json:"archivedAt"`
		LastUpdatedAt             *time.Time                       `json:"lastUpdatedAt"`
		MinEstimatedTimeInSeconds *uint32                          `json:"minEstimatedTimeInSeconds,omitempty"`
		MaxEstimatedTimeInSeconds *uint32                          `json:"maxEstimatedTimeInSeconds,omitempty"`
		MinTemperatureInCelsius   *float32                         `json:"minTemperatureInCelsius,omitempty"`
		MaxTemperatureInCelsius   *float32                         `json:"maxTemperatureInCelsius,omitempty"`
		BelongsToRecipe           string                           `json:"belongsToRecipe"`
		ConditionExpression       string                           `json:"conditionExpression"`
		ID                        string                           `json:"id"`
		Notes                     string                           `json:"notes"`
		ExplicitInstructions      string                           `json:"explicitInstructions"`
		Products                  []*RecipeStepProduct             `json:"products"`
		Instruments               []*RecipeStepInstrument          `json:"instruments"`
		Vessels                   []*RecipeStepVessel              `json:"vessels"`
		CompletionConditions      []*RecipeStepCompletionCondition `json:"completionConditions"`
		Ingredients               []*RecipeStepIngredient          `json:"ingredients"`
		Media                     []*RecipeMedia                   `json:"media"`
		StepImages                []*uploadedmedia.UploadedMedia   `json:"stepImages"`
		Preparation               ValidPreparation                 `json:"preparation"`
		Index                     uint32                           `json:"index"`
		Optional                  bool                             `json:"optional"`
		StartTimerAutomatically   bool                             `json:"startTimerAutomatically"`
	}

	// RecipeStepList represents a list of recipe steps.
	RecipeStepList []*RecipeStep

	// RecipeStepCreationRequestInput represents what a user could set as input for creating recipe steps.
	RecipeStepCreationRequestInput struct {
		_ struct{} `json:"-"`

		MinEstimatedTimeInSeconds *uint32                                              `json:"minEstimatedTimeInSeconds,omitempty"`
		MaxEstimatedTimeInSeconds *uint32                                              `json:"maxEstimatedTimeInSeconds,omitempty"`
		MinTemperatureInCelsius   *float32                                             `json:"minTemperatureInCelsius,omitempty"`
		MaxTemperatureInCelsius   *float32                                             `json:"maxTemperatureInCelsius,omitempty"`
		PreparationID             string                                               `json:"preparationID"`
		Notes                     string                                               `json:"notes"`
		ConditionExpression       string                                               `json:"conditionExpression"`
		ExplicitInstructions      string                                               `json:"explicitInstructions"`
		Instruments               []*RecipeStepInstrumentCreationRequestInput          `json:"instruments"`
		Vessels                   []*RecipeStepVesselCreationRequestInput              `json:"vessels"`
		Products                  []*RecipeStepProductCreationRequestInput             `json:"products"`
		Ingredients               []*RecipeStepIngredientCreationRequestInput          `json:"ingredients"`
		CompletionConditions      []*RecipeStepCompletionConditionCreationRequestInput `json:"completionConditions"`
		Index                     uint32                                               `json:"index"`
		Optional                  bool                                                 `json:"optional"`
		StartTimerAutomatically   bool                                                 `json:"startTimerAutomatically"`
	}

	// RecipeStepDatabaseCreationInput represents what a user could set as input for creating recipe steps.
	RecipeStepDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		MinEstimatedTimeInSeconds *uint32                                               `json:"-"`
		MaxEstimatedTimeInSeconds *uint32                                               `json:"-"`
		MinTemperatureInCelsius   *float32                                              `json:"-"`
		MaxTemperatureInCelsius   *float32                                              `json:"-"`
		BelongsToRecipe           string                                                `json:"-"`
		PreparationID             string                                                `json:"-"`
		ID                        string                                                `json:"-"`
		Notes                     string                                                `json:"-"`
		ExplicitInstructions      string                                                `json:"-"`
		ConditionExpression       string                                                `json:"-"`
		Ingredients               []*RecipeStepIngredientDatabaseCreationInput          `json:"-"`
		Instruments               []*RecipeStepInstrumentDatabaseCreationInput          `json:"-"`
		Vessels                   []*RecipeStepVesselDatabaseCreationInput              `json:"-"`
		Products                  []*RecipeStepProductDatabaseCreationInput             `json:"-"`
		CompletionConditions      []*RecipeStepCompletionConditionDatabaseCreationInput `json:"-"`
		Index                     uint32                                                `json:"-"`
		Optional                  bool                                                  `json:"-"`
		StartTimerAutomatically   bool                                                  `json:"-"`
	}

	// RecipeStepUpdateRequestInput represents what a user could set as input for updating recipe steps.
	RecipeStepUpdateRequestInput struct {
		_ struct{} `json:"-"`

		MinEstimatedTimeInSeconds *uint32           `json:"minEstimatedTimeInSeconds,omitempty"`
		MaxEstimatedTimeInSeconds *uint32           `json:"maxEstimatedTimeInSeconds,omitempty"`
		MinTemperatureInCelsius   *float32          `json:"minTemperatureInCelsius,omitempty"`
		MaxTemperatureInCelsius   *float32          `json:"maxTemperatureInCelsius,omitempty"`
		Notes                     *string           `json:"notes,omitempty"`
		Preparation               *ValidPreparation `json:"preparation,omitempty"`
		Index                     *uint32           `json:"index,omitempty"`
		Optional                  *bool             `json:"optional,omitempty"`
		ExplicitInstructions      *string           `json:"explicitInstructions,omitempty"`
		ConditionExpression       *string           `json:"conditionExpression,omitempty"`
		StartTimerAutomatically   *bool             `json:"startTimerAutomatically"`
		BelongsToRecipe           *string           `json:"belongsToRecipe"`
	}

	// RecipeStepDataManager describes a structure capable of storing recipe steps permanently.
	RecipeStepDataManager interface {
		RecipeStepExists(ctx context.Context, recipeID, recipeStepID string) (bool, error)
		GetRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*RecipeStep, error)
		GetRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeStep], error)
		CreateRecipeStep(ctx context.Context, input *RecipeStepDatabaseCreationInput) (*RecipeStep, error)
		UpdateRecipeStep(ctx context.Context, updated *RecipeStep) error
		ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error
	}
)

// Update merges an RecipeStepUpdateRequestInput with a recipe step.
func (x *RecipeStep) Update(input *RecipeStepUpdateRequestInput) {
	if input.Index != nil && *input.Index != x.Index {
		x.Index = *input.Index
	}

	if input.Preparation != nil {
		x.Preparation = *input.Preparation
	}

	if input.MinEstimatedTimeInSeconds != nil && input.MinEstimatedTimeInSeconds != x.MinEstimatedTimeInSeconds {
		x.MinEstimatedTimeInSeconds = input.MinEstimatedTimeInSeconds
	}

	if input.MaxEstimatedTimeInSeconds != nil && input.MaxEstimatedTimeInSeconds != x.MaxEstimatedTimeInSeconds {
		x.MaxEstimatedTimeInSeconds = input.MaxEstimatedTimeInSeconds
	}

	if input.MinTemperatureInCelsius != nil && (x.MinTemperatureInCelsius == nil || (*input.MinTemperatureInCelsius != 0 && *input.MinTemperatureInCelsius != *x.MinTemperatureInCelsius)) {
		x.MinTemperatureInCelsius = input.MinTemperatureInCelsius
	}

	if input.MaxTemperatureInCelsius != nil && (x.MaxTemperatureInCelsius == nil || (*input.MaxTemperatureInCelsius != 0 && *input.MaxTemperatureInCelsius != *x.MaxTemperatureInCelsius)) {
		x.MaxTemperatureInCelsius = input.MaxTemperatureInCelsius
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	// TODO: do something about products here

	if input.Optional != nil && *input.Optional != x.Optional {
		x.Optional = *input.Optional
	}

	if input.ExplicitInstructions != nil && *input.ExplicitInstructions != x.ExplicitInstructions {
		x.ExplicitInstructions = *input.ExplicitInstructions
	}

	if input.ConditionExpression != nil && *input.ConditionExpression != x.ConditionExpression {
		x.ConditionExpression = *input.ConditionExpression
	}

	if input.StartTimerAutomatically != nil && *input.StartTimerAutomatically != x.StartTimerAutomatically {
		x.StartTimerAutomatically = *input.StartTimerAutomatically
	}
}

// GetIngredientOptionsByIndex returns all ingredient options for a given index.
// This groups ingredients that represent the same logical position within the step.
func (x *RecipeStep) GetIngredientOptionsByIndex(index uint16) []*RecipeStepIngredient {
	var options []*RecipeStepIngredient
	for _, ingredient := range x.Ingredients {
		if ingredient.Index == index {
			options = append(options, ingredient)
		}
	}
	return options
}

// GetIngredientByIndexAndOptionIndex returns the ingredient with the specified index and optionIndex.
// Returns nil if not found.
func (x *RecipeStep) GetIngredientByIndexAndOptionIndex(index, optionIndex uint16) *RecipeStepIngredient {
	for _, ingredient := range x.Ingredients {
		if ingredient.Index == index && ingredient.OptionIndex == optionIndex {
			return ingredient
		}
	}
	return nil
}

// GetInstrumentOptionsByIndex returns all instrument options for a given index.
// This groups instruments that represent the same logical position within the step.
func (x *RecipeStep) GetInstrumentOptionsByIndex(index uint16) []*RecipeStepInstrument {
	var options []*RecipeStepInstrument
	for _, instrument := range x.Instruments {
		if instrument.Index == index {
			options = append(options, instrument)
		}
	}
	return options
}

// GetInstrumentByIndexAndOptionIndex returns the instrument with the specified index and optionIndex.
// Returns nil if not found.
func (x *RecipeStep) GetInstrumentByIndexAndOptionIndex(index, optionIndex uint16) *RecipeStepInstrument {
	for _, instrument := range x.Instruments {
		if instrument.Index == index && instrument.OptionIndex == optionIndex {
			return instrument
		}
	}
	return nil
}

// GetVesselsByIndex returns all vessels for a given index.
// Note: Vessels don't have OptionIndex, so multiple vessels with the same index
// represent multiple vessels at the same position (e.g., "2 bowls").
func (x *RecipeStep) GetVesselsByIndex(index uint16) []*RecipeStepVessel {
	var vessels []*RecipeStepVessel
	for _, vessel := range x.Vessels {
		if vessel.Index == index {
			vessels = append(vessels, vessel)
		}
	}
	return vessels
}

var _ validation.ValidatableWithContext = (*RecipeStepCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeStepCreationRequestInput.
func (x *RecipeStepCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	err := &multierror.Error{}

	if len(x.Instruments) == 0 && len(x.Vessels) == 0 {
		err = multierror.Append(err, errOneInstrumentOrVesselRequired)
	}

	validationErr := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.PreparationID, validation.Required),
		validation.Field(&x.Products, validation.Required),
	)

	if validationErr != nil {
		err = multierror.Append(err, validationErr)
	}

	return err.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*RecipeStepDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeStepDatabaseCreationInput.
func (x *RecipeStepDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Products, validation.Required),
		validation.Field(&x.PreparationID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeStepUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeStepUpdateRequestInput.
func (x *RecipeStepUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Index, validation.Required),
		validation.Field(&x.Preparation, validation.Required),
		validation.Field(&x.Notes, validation.Required),
	)
}
