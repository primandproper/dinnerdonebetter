package mealplanning

import (
	"context"
	"encoding/gob"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v3/database/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/hashicorp/go-multierror"
)

const (
	// RecipeRatingCreatedServiceEventType indicates a recipe rating was created.
	RecipeRatingCreatedServiceEventType = "recipe_rating_created"
	// RecipeRatingUpdatedServiceEventType indicates a recipe rating was updated.
	RecipeRatingUpdatedServiceEventType = "recipe_rating_updated"
	// RecipeRatingArchivedServiceEventType indicates a recipe rating was archived.
	RecipeRatingArchivedServiceEventType = "recipe_rating_archived"
)

var (
	errAtLeastOneRatingRequired = errors.New("recipe rating must have at least one rating")
)

func init() {
	gob.Register(new(RecipeRating))
	gob.Register(new(RecipeRatingCreationRequestInput))
	gob.Register(new(RecipeRatingUpdateRequestInput))
}

type (
	// RecipeRating represents a recipe rating.
	RecipeRating struct {
		_ struct{} `json:"-"`

		CreatedAt       time.Time  `json:"createdAt"`
		LastUpdatedAt   *time.Time `json:"lastUpdatedAt"`
		ArchivedAt      *time.Time `json:"archivedAt"`
		Notes           string     `json:"notes"`
		ID              string     `json:"id"`
		BelongsToRecipe string     `json:"belongsToRecipe"`
		CreatedByUser   string     `json:"createdByUser"`
		Taste           float32    `json:"taste"`
		Instructions    float32    `json:"instructions"`
		Overall         float32    `json:"overall"`
		Cleanup         float32    `json:"cleanup"`
		Difficulty      float32    `json:"difficulty"`
	}

	// RecipeRatingCreationRequestInput represents what a user could set as input for creating recipe ratings.
	RecipeRatingCreationRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToRecipe string  `json:"belongsToRecipe"`
		Notes           string  `json:"notes"`
		CreatedByUser   string  `json:"createdByUser"`
		Taste           float32 `json:"taste"`
		Difficulty      float32 `json:"difficulty"`
		Cleanup         float32 `json:"cleanup"`
		Instructions    float32 `json:"instructions"`
		Overall         float32 `json:"overall"`
	}

	// RecipeRatingDatabaseCreationInput represents what a user could set as input for creating recipe ratings.
	RecipeRatingDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID              string  `json:"-"`
		BelongsToRecipe string  `json:"-"`
		Notes           string  `json:"-"`
		CreatedByUser   string  `json:"-"`
		Taste           float32 `json:"-"`
		Difficulty      float32 `json:"-"`
		Cleanup         float32 `json:"-"`
		Instructions    float32 `json:"-"`
		Overall         float32 `json:"-"`
	}

	// RecipeRatingUpdateRequestInput represents what a user could set as input for updating recipe ratings.
	RecipeRatingUpdateRequestInput struct {
		_ struct{} `json:"-"`

		BelongsToRecipe *string  `json:"belongsToRecipe"`
		Taste           *float32 `json:"taste"`
		Difficulty      *float32 `json:"difficulty"`
		Cleanup         *float32 `json:"cleanup"`
		Instructions    *float32 `json:"instructions"`
		Overall         *float32 `json:"overall"`
		Notes           *string  `json:"notes"`
	}

	// RecipeRatingDataManager describes a structure capable of storing recipe ratings permanently.
	RecipeRatingDataManager interface {
		RecipeRatingExists(ctx context.Context, recipeID, recipeRatingID string) (bool, error)
		GetRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*RecipeRating, error)
		GetRecipeRatingsForRecipe(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeRating], error)
		GetRecipeRatingsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeRating], error)
		CreateRecipeRating(ctx context.Context, input *RecipeRatingDatabaseCreationInput) (*RecipeRating, error)
		UpdateRecipeRating(ctx context.Context, updated *RecipeRating) error
		ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error
	}
)

// Update merges an RecipeRatingUpdateRequestInput with a recipe rating.
func (x *RecipeRating) Update(input *RecipeRatingUpdateRequestInput) {
	if input.BelongsToRecipe != nil && *input.BelongsToRecipe != x.BelongsToRecipe {
		x.BelongsToRecipe = *input.BelongsToRecipe
	}

	if input.Taste != nil && *input.Taste != x.Taste {
		x.Taste = *input.Taste
	}

	if input.Difficulty != nil && *input.Difficulty != x.Difficulty {
		x.Difficulty = *input.Difficulty
	}

	if input.Cleanup != nil && *input.Cleanup != x.Cleanup {
		x.Cleanup = *input.Cleanup
	}

	if input.Instructions != nil && *input.Instructions != x.Instructions {
		x.Instructions = *input.Instructions
	}

	if input.Overall != nil && *input.Overall != x.Overall {
		x.Overall = *input.Overall
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}
}

var _ validation.ValidatableWithContext = (*RecipeRatingCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeRatingCreationRequestInput.
func (x *RecipeRatingCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	errs := &multierror.Error{}

	if x.Cleanup == 0 && x.Difficulty == 0 && x.Instructions == 0 && x.Overall == 0 && x.Taste == 0 {
		errs = multierror.Append(errs, errAtLeastOneRatingRequired)
	}

	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToRecipe, validation.Required),
	); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*RecipeRatingDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeRatingDatabaseCreationInput.
func (x *RecipeRatingDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	errs := &multierror.Error{}

	if x.Cleanup == 0 && x.Difficulty == 0 && x.Instructions == 0 && x.Overall == 0 && x.Taste == 0 {
		errs = multierror.Append(errs, errAtLeastOneRatingRequired)
	}

	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.BelongsToRecipe, validation.Required),
		validation.Field(&x.CreatedByUser, validation.Required),
	); err != nil {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

var _ validation.ValidatableWithContext = (*RecipeRatingUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeRatingUpdateRequestInput.
func (x *RecipeRatingUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToRecipe, validation.Required),
	)
}
