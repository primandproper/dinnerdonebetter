package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(RecipeList))
	gob.Register(new(RecipeListCreationRequestInput))
	gob.Register(new(RecipeListUpdateRequestInput))
}

type (
	// RecipeList represents a collection of recipes belonging to a user.
	RecipeList struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time         `json:"createdAt"`
		LastUpdatedAt *time.Time        `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time        `json:"archivedAt"`
		ID            string            `json:"id"`
		Name          string            `json:"name"`
		Description   string            `json:"description"`
		BelongsToUser string            `json:"belongsToUser"`
		Items         []*RecipeListItem `json:"items"`
	}

	// RecipeListCreationRequestInput represents input for creating recipe lists.
	RecipeListCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string                                `json:"name"`
		Description string                                `json:"description"`
		Items       []*RecipeListItemCreationRequestInput `json:"items"`
	}

	// RecipeListDatabaseCreationInput represents database input for creating recipe lists.
	RecipeListDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID            string                                 `json:"-"`
		Name          string                                 `json:"-"`
		Description   string                                 `json:"-"`
		BelongsToUser string                                 `json:"-"`
		Items         []*RecipeListItemDatabaseCreationInput `json:"-"`
	}

	// RecipeListUpdateRequestInput represents input for updating recipe lists.
	RecipeListUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	// RecipeListDataManager describes a structure capable of storing recipe lists permanently.
	RecipeListDataManager interface {
		GetRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeList], error)
		CreateRecipeList(ctx context.Context, input *RecipeListDatabaseCreationInput) (*RecipeList, error)
		UpdateRecipeList(ctx context.Context, updated *RecipeList) error
		ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error
	}
)

// Update merges a RecipeListUpdateRequestInput with a recipe list.
func (x *RecipeList) Update(input *RecipeListUpdateRequestInput) {
	if input == nil {
		return
	}

	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}
}

var _ validation.ValidatableWithContext = (*RecipeListCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeListCreationRequestInput.
func (x *RecipeListCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeListDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeListDatabaseCreationInput.
func (x *RecipeListDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.BelongsToUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeListUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeListUpdateRequestInput.
func (x *RecipeListUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
