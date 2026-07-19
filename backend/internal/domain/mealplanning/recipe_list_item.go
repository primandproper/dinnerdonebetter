package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v5/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(RecipeListItem))
	gob.Register(new(RecipeListItemCreationRequestInput))
	gob.Register(new(RecipeListItemUpdateRequestInput))
}

type (
	// RecipeListItem represents a single entry in a recipe list.
	RecipeListItem struct {
		_                   struct{}   `json:"-"`
		CreatedAt           time.Time  `json:"createdAt"`
		LastUpdatedAt       *time.Time `json:"lastUpdatedAt"`
		ArchivedAt          *time.Time `json:"archivedAt"`
		ID                  string     `json:"id"`
		Notes               string     `json:"notes"`
		BelongsToRecipeList string     `json:"belongsToRecipeList"`
		Recipe              Recipe     `json:"recipe"`
	}

	// RecipeListItemCreationRequestInput represents input for creating recipe list items.
	RecipeListItemCreationRequestInput struct {
		_ struct{} `json:"-"`

		RecipeID string `json:"recipeID"`
		Notes    string `json:"notes"`
	}

	// RecipeListItemDatabaseCreationInput represents database input for creating recipe list items.
	RecipeListItemDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                  string `json:"-"`
		RecipeID            string `json:"-"`
		Notes               string `json:"-"`
		BelongsToRecipeList string `json:"-"`
	}

	// RecipeListItemUpdateRequestInput represents input for updating recipe list items.
	RecipeListItemUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes *string `json:"notes,omitempty"`
	}

	// RecipeListItemDataManager describes a structure capable of storing recipe list items permanently.
	RecipeListItemDataManager interface {
		GetRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[RecipeListItem], error)
		CreateRecipeListItem(ctx context.Context, input *RecipeListItemDatabaseCreationInput) (*RecipeListItem, error)
		UpdateRecipeListItem(ctx context.Context, updated *RecipeListItem) error
		ArchiveRecipeListItem(ctx context.Context, recipeListItemID, recipeListID string) error
	}
)

// Update merges a RecipeListItemUpdateRequestInput with a recipe list item.
func (x *RecipeListItem) Update(input *RecipeListItemUpdateRequestInput) {
	if input == nil {
		return
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}
}

var _ validation.ValidatableWithContext = (*RecipeListItemCreationRequestInput)(nil)

// ValidateWithContext validates a RecipeListItemCreationRequestInput.
func (x *RecipeListItemCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.RecipeID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeListItemDatabaseCreationInput)(nil)

// ValidateWithContext validates a RecipeListItemDatabaseCreationInput.
func (x *RecipeListItemDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.RecipeID, validation.Required),
		validation.Field(&x.BelongsToRecipeList, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*RecipeListItemUpdateRequestInput)(nil)

// ValidateWithContext validates a RecipeListItemUpdateRequestInput.
func (x *RecipeListItemUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
	)
}
