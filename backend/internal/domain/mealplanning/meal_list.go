package mealplanning

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v8/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(MealList))
	gob.Register(new(MealListCreationRequestInput))
	gob.Register(new(MealListUpdateRequestInput))
}

type (
	// MealList represents a collection of meals belonging to a user.
	MealList struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time       `json:"createdAt"`
		LastUpdatedAt *time.Time      `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time      `json:"archivedAt"`
		ID            string          `json:"id"`
		Name          string          `json:"name"`
		Description   string          `json:"description"`
		BelongsToUser string          `json:"belongsToUser"`
		Items         []*MealListItem `json:"items"`
	}

	// MealListCreationRequestInput represents input for creating meal lists.
	MealListCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string                              `json:"name"`
		Description string                              `json:"description"`
		Items       []*MealListItemCreationRequestInput `json:"items"`
	}

	// MealListDatabaseCreationInput represents database input for creating meal lists.
	MealListDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID            string                               `json:"-"`
		Name          string                               `json:"-"`
		Description   string                               `json:"-"`
		BelongsToUser string                               `json:"-"`
		Items         []*MealListItemDatabaseCreationInput `json:"-"`
	}

	// MealListUpdateRequestInput represents input for updating meal lists.
	MealListUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        *string `json:"name,omitempty"`
		Description *string `json:"description,omitempty"`
	}

	// MealListDataManager describes a structure capable of storing meal lists permanently.
	MealListDataManager interface {
		GetMealLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealList], error)
		CreateMealList(ctx context.Context, input *MealListDatabaseCreationInput) (*MealList, error)
		UpdateMealList(ctx context.Context, updated *MealList) error
		ArchiveMealList(ctx context.Context, mealListID, userID string) error
	}
)

// Update merges a MealListUpdateRequestInput with a meal list.
func (x *MealList) Update(input *MealListUpdateRequestInput) {
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

var _ validation.ValidatableWithContext = (*MealListCreationRequestInput)(nil)

// ValidateWithContext validates a MealListCreationRequestInput.
func (x *MealListCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealListDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealListDatabaseCreationInput.
func (x *MealListDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.BelongsToUser, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealListUpdateRequestInput)(nil)

// ValidateWithContext validates a MealListUpdateRequestInput.
func (x *MealListUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
