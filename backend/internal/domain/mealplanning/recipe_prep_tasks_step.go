package mealplanning

import (
	"encoding/gob"
)

func init() {
	gob.Register(new(RecipePrepTask))
	gob.Register(new(RecipePrepTaskCreationRequestInput))
	gob.Register(new(RecipePrepTaskUpdateRequestInput))
}

type (
	// RecipePrepTaskStep represents a recipe prep task step.
	RecipePrepTaskStep struct {
		_ struct{} `json:"-"`

		ID                      string `json:"id"`
		BelongsToRecipeStep     string `json:"belongsToRecipeStep"`
		BelongsToRecipePrepTask string `json:"belongsToRecipeStepTask"`
		SatisfiesRecipeStep     bool   `json:"satisfiesRecipeStep"`
	}

	// RecipePrepTaskStepWithinRecipeCreationRequestInput represents a recipe prep task step.
	RecipePrepTaskStepWithinRecipeCreationRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToRecipeStepIndex uint32 `json:"belongsToRecipeStepIndex"`
		SatisfiesRecipeStep      bool   `json:"satisfiesRecipeStep"`
	}

	// RecipePrepTaskStepCreationRequestInput represents a recipe prep task step.
	RecipePrepTaskStepCreationRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToRecipeStep string `json:"belongsToRecipeStep"`
		SatisfiesRecipeStep bool   `json:"satisfiesRecipeStep"`
	}

	// RecipePrepTaskStepDatabaseCreationInput represents a recipe prep task step.
	RecipePrepTaskStepDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                      string `json:"-"`
		BelongsToRecipeStep     string `json:"-"`
		BelongsToRecipePrepTask string `json:"-"`
		SatisfiesRecipeStep     bool   `json:"-"`
	}

	// RecipePrepTaskStepUpdateRequestInput represents a recipe prep task step.
	RecipePrepTaskStepUpdateRequestInput struct {
		_ struct{} `json:"-"`

		SatisfiesRecipeStep     *bool   `json:"satisfiesRecipeStep,omitempty"`
		BelongsToRecipeStep     *string `json:"belongsToRecipeStep,omitempty"`
		BelongsToRecipePrepTask *string `json:"belongsToRecipeStepTask,omitempty"`
	}
)
