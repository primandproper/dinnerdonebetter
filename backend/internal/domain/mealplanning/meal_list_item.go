package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v10/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(MealListItem))
	gob.Register(new(MealListItemCreationRequestInput))
	gob.Register(new(MealListItemUpdateRequestInput))
}

type (
	// MealListItem represents a single entry in a meal list.
	MealListItem struct {
		_                 struct{}   `json:"-"`
		CreatedAt         time.Time  `json:"createdAt"`
		LastUpdatedAt     *time.Time `json:"lastUpdatedAt"`
		ArchivedAt        *time.Time `json:"archivedAt"`
		ID                string     `json:"id"`
		Notes             string     `json:"notes"`
		BelongsToMealList string     `json:"belongsToMealList"`
		Meal              Meal       `json:"meal"`
	}

	// MealListItemCreationRequestInput represents input for creating meal list items.
	MealListItemCreationRequestInput struct {
		_ struct{} `json:"-"`

		MealID string `json:"mealID"`
		Notes  string `json:"notes"`
	}

	// MealListItemDatabaseCreationInput represents database input for creating meal list items.
	MealListItemDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                string `json:"-"`
		MealID            string `json:"-"`
		Notes             string `json:"-"`
		BelongsToMealList string `json:"-"`
	}

	// MealListItemUpdateRequestInput represents input for updating meal list items.
	MealListItemUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes *string `json:"notes,omitempty"`
	}

	// MealListItemDataManager describes a structure capable of storing meal list items permanently.
	MealListItemDataManager interface {
		MealExistsInMealList(ctx context.Context, mealListID, mealID string) (bool, error)
		GetMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealListItem], error)
		CreateMealListItem(ctx context.Context, input *MealListItemDatabaseCreationInput) (*MealListItem, error)
		UpdateMealListItem(ctx context.Context, updated *MealListItem) error
		ArchiveMealListItem(ctx context.Context, mealListItemID, mealListID string) error
	}
)

// Update merges a MealListItemUpdateRequestInput with a meal list item.
func (x *MealListItem) Update(input *MealListItemUpdateRequestInput) {
	if input == nil {
		return
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}
}

var _ validation.ValidatableWithContext = (*MealListItemCreationRequestInput)(nil)

// ValidateWithContext validates a MealListItemCreationRequestInput.
func (x *MealListItemCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.MealID, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealListItemDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealListItemDatabaseCreationInput.
func (x *MealListItemDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.MealID, validation.Required),
		validation.Field(&x.BelongsToMealList, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealListItemUpdateRequestInput)(nil)

// ValidateWithContext validates a MealListItemUpdateRequestInput.
func (x *MealListItemUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
	)
}
