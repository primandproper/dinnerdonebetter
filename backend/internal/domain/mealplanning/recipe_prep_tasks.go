package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v4/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// RecipePrepTaskStorageTypeUncovered is a valid storage type for a recipe step task.
	RecipePrepTaskStorageTypeUncovered = "uncovered"
	// RecipePrepTaskStorageTypeCovered is a valid storage type for a recipe step task.
	RecipePrepTaskStorageTypeCovered = "covered"
	// RecipePrepTaskStorageTypeWireRack is a valid storage type for a recipe step task.
	RecipePrepTaskStorageTypeWireRack = "on a wire rack"
	// RecipePrepTaskStorageTypeAirtightContainer is a valid storage type for a recipe step task.
	RecipePrepTaskStorageTypeAirtightContainer = "in an airtight container"

	// RecipePrepTaskCreatedServiceEventType indicates a recipe prep task was created.
	RecipePrepTaskCreatedServiceEventType = "recipe_created"
	// RecipePrepTaskUpdatedServiceEventType indicates a recipe prep task was updated.
	RecipePrepTaskUpdatedServiceEventType = "recipe_updated"
	// RecipePrepTaskArchivedServiceEventType indicates a recipe prep task was archived.
	RecipePrepTaskArchivedServiceEventType = "recipe_archived"
)

func init() {
	gob.Register(new(RecipePrepTask))
	gob.Register(new(RecipePrepTaskCreationRequestInput))
	gob.Register(new(RecipePrepTaskUpdateRequestInput))
}

type (
	// RecipePrepTask represents a recipe prep task.
	RecipePrepTask struct {
		_                                  struct{}              `json:"-"`
		CreatedAt                          time.Time             `json:"createdAt"`
		MinStorageTemperatureInCelsius     *float32              `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius     *float32              `json:"maxStorageTemperatureInCelsius,omitempty"`
		MaxTimeBufferBeforeRecipeInSeconds *uint32               `json:"maxTimeBufferBeforeRecipeInSeconds,omitempty"`
		ArchivedAt                         *time.Time            `json:"archivedAt"`
		LastUpdatedAt                      *time.Time            `json:"lastUpdatedAt"`
		ID                                 string                `json:"id"`
		StorageType                        string                `json:"storageType"`
		BelongsToRecipe                    string                `json:"belongsToRecipe"`
		ExplicitStorageInstructions        string                `json:"explicitStorageInstructions"`
		Notes                              string                `json:"notes"`
		Name                               string                `json:"name"`
		Description                        string                `json:"description"`
		TaskSteps                          []*RecipePrepTaskStep `json:"recipeSteps"`
		MinTimeBufferBeforeRecipeInSeconds uint32                `json:"minTimeBufferBeforeRecipeInSeconds"`
		Optional                           bool                  `json:"optional"`
	}

	// RecipePrepTaskCreationRequestInput represents what a user could set as input for creating recipes.
	RecipePrepTaskCreationRequestInput struct {
		_                                  struct{}                                  `json:"-"`
		MinStorageTemperatureInCelsius     *float32                                  `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius     *float32                                  `json:"maxStorageTemperatureInCelsius,omitempty"`
		MaxTimeBufferBeforeRecipeInSeconds *uint32                                   `json:"maxTimeBufferBeforeRecipeInSeconds,omitempty"`
		ExplicitStorageInstructions        string                                    `json:"explicitStorageInstructions"`
		StorageType                        string                                    `json:"storageType"`
		Notes                              string                                    `json:"notes"`
		Name                               string                                    `json:"name"`
		Description                        string                                    `json:"description"`
		BelongsToRecipe                    string                                    `json:"belongsToRecipe"`
		RecipeSteps                        []*RecipePrepTaskStepCreationRequestInput `json:"recipeSteps"`
		MinTimeBufferBeforeRecipeInSeconds uint32                                    `json:"minTimeBufferBeforeRecipeInSeconds"`
		Optional                           bool                                      `json:"optional"`
	}

	// RecipePrepTaskWithinRecipeCreationRequestInput represents what a user could set as input for creating recipes.
	RecipePrepTaskWithinRecipeCreationRequestInput struct {
		_                                  struct{}                                              `json:"-"`
		MinStorageTemperatureInCelsius     *float32                                              `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius     *float32                                              `json:"maxStorageTemperatureInCelsius,omitempty"`
		MaxTimeBufferBeforeRecipeInSeconds *uint32                                               `json:"maxTimeBufferBeforeRecipeInSeconds,omitempty"`
		Name                               string                                                `json:"name"`
		StorageType                        string                                                `json:"storageType"`
		Description                        string                                                `json:"description"`
		ExplicitStorageInstructions        string                                                `json:"explicitStorageInstructions"`
		Notes                              string                                                `json:"notes"`
		BelongsToRecipe                    string                                                `json:"belongsToRecipe"`
		RecipeSteps                        []*RecipePrepTaskStepWithinRecipeCreationRequestInput `json:"recipeSteps"`
		MinTimeBufferBeforeRecipeInSeconds uint32                                                `json:"minTimeBufferBeforeRecipeInSeconds"`
		Optional                           bool                                                  `json:"optional"`
	}

	// RecipePrepTaskDatabaseCreationInput represents what a user could set as input for creating recipes.
	RecipePrepTaskDatabaseCreationInput struct {
		_                                  struct{}                                   `json:"-"`
		MinStorageTemperatureInCelsius     *float32                                   `json:"-"`
		MaxStorageTemperatureInCelsius     *float32                                   `json:"-"`
		MaxTimeBufferBeforeRecipeInSeconds *uint32                                    `json:"-"`
		Notes                              string                                     `json:"-"`
		ExplicitStorageInstructions        string                                     `json:"-"`
		ID                                 string                                     `json:"-"`
		Name                               string                                     `json:"-"`
		Description                        string                                     `json:"-"`
		StorageType                        string                                     `json:"-"`
		BelongsToRecipe                    string                                     `json:"-"`
		TaskSteps                          []*RecipePrepTaskStepDatabaseCreationInput `json:"-"`
		MinTimeBufferBeforeRecipeInSeconds uint32                                     `json:"-"`
		Optional                           bool                                       `json:"-"`
	}

	// RecipePrepTaskUpdateRequestInput represents what a user could set as input for updating recipes.
	RecipePrepTaskUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes                              *string                                 `json:"notes,omitempty"`
		ExplicitStorageInstructions        *string                                 `json:"explicitStorageInstructions,omitempty"`
		StorageType                        *string                                 `json:"storageType,omitempty"`
		Name                               *string                                 `json:"name"`
		Optional                           *bool                                   `json:"optional"`
		Description                        *string                                 `json:"description"`
		MinStorageTemperatureInCelsius     *float32                                `json:"minStorageTemperatureInCelsius,omitempty"`
		MaxStorageTemperatureInCelsius     *float32                                `json:"maxStorageTemperatureInCelsius,omitempty"`
		MinTimeBufferBeforeRecipeInSeconds *uint32                                 `json:"minTimeBufferBeforeRecipeInSeconds,omitempty"`
		MaxTimeBufferBeforeRecipeInSeconds *uint32                                 `json:"maxTimeBufferBeforeRecipeInSeconds,omitempty"`
		BelongsToRecipe                    *string                                 `json:"belongsToRecipe,omitempty"`
		TaskSteps                          []*RecipePrepTaskStepUpdateRequestInput `json:"recipeSteps,omitempty"`
	}

	// RecipePrepTaskDataManager describes a structure capable of storing recipes permanently.
	RecipePrepTaskDataManager interface {
		RecipePrepTaskExists(ctx context.Context, recipeID, recipePrepTaskID string) (bool, error)
		GetRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (*RecipePrepTask, error)
		GetRecipePrepTasks(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipePrepTask], error)
		GetRecipePrepTasksForRecipe(ctx context.Context, recipeID string) ([]*RecipePrepTask, error)
		CreateRecipePrepTask(ctx context.Context, input *RecipePrepTaskDatabaseCreationInput) (*RecipePrepTask, error)
		UpdateRecipePrepTask(ctx context.Context, updated *RecipePrepTask) error
		ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error
	}
)

// Update merges an RecipePrepTaskUpdateRequestInput with a recipe prep task.
func (x *RecipePrepTask) Update(input *RecipePrepTaskUpdateRequestInput) {
	if input.BelongsToRecipe != nil && *input.BelongsToRecipe != x.BelongsToRecipe {
		x.BelongsToRecipe = *input.BelongsToRecipe
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}

	if input.ExplicitStorageInstructions != nil && *input.ExplicitStorageInstructions != x.ExplicitStorageInstructions {
		x.ExplicitStorageInstructions = *input.ExplicitStorageInstructions
	}

	if input.MinTimeBufferBeforeRecipeInSeconds != nil && *input.MinTimeBufferBeforeRecipeInSeconds != x.MinTimeBufferBeforeRecipeInSeconds {
		x.MinTimeBufferBeforeRecipeInSeconds = *input.MinTimeBufferBeforeRecipeInSeconds
	}

	if input.MaxTimeBufferBeforeRecipeInSeconds != nil && input.MaxTimeBufferBeforeRecipeInSeconds != x.MaxTimeBufferBeforeRecipeInSeconds {
		x.MaxTimeBufferBeforeRecipeInSeconds = input.MaxTimeBufferBeforeRecipeInSeconds
	}

	if input.StorageType != nil && *input.StorageType != x.StorageType {
		x.StorageType = *input.StorageType
	}

	if input.MinStorageTemperatureInCelsius != nil && input.MinStorageTemperatureInCelsius != x.MinStorageTemperatureInCelsius {
		x.MinStorageTemperatureInCelsius = input.MinStorageTemperatureInCelsius
	}

	if input.MaxStorageTemperatureInCelsius != nil && input.MaxStorageTemperatureInCelsius != x.MaxStorageTemperatureInCelsius {
		x.MaxStorageTemperatureInCelsius = input.MaxStorageTemperatureInCelsius
	}

	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}

	if input.Optional != nil && *input.Optional != x.Optional {
		x.Optional = *input.Optional
	}
}

var _ validation.ValidatableWithContext = (*RecipePrepTaskCreationRequestInput)(nil)

// ValidateWithContext validates a RecipePrepTaskCreationRequestInput.
func (x *RecipePrepTaskCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	result := &multierror.Error{}

	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToRecipe, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.StorageType, validation.Required),
		validation.Field(&x.MinTimeBufferBeforeRecipeInSeconds, validation.Required),
		validation.Field(&x.MinStorageTemperatureInCelsius, validation.Required),
	); err != nil {
		result = multierror.Append(err, result)
	}

	// TODO: uncomment me
	// if x.MinStorageTemperatureInCelsius != nil && x.MaxStorageTemperatureInCelsius != nil && *x.MinStorageTemperatureInCelsius > *x.MaxStorageTemperatureInCelsius {
	//	result = multierror.Append(fmt.Errorf("minimum storage temperature (%d) is greater than maximum storage temperature (%d)", x.MinStorageTemperatureInCelsius, x.MaxStorageTemperatureInCelsius))
	// }

	return result.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*RecipePrepTaskDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipePrepTaskDatabaseCreationInput.
func (x *RecipePrepTaskDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.BelongsToRecipe, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipePrepTaskUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipePrepTaskUpdateRequestInput.
func (x *RecipePrepTaskUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToRecipe, validation.Required),
	)
}
