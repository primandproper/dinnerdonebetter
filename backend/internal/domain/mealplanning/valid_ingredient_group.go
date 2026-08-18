package mealplanning

import (
	"context"
	"encoding/gob"
	"math"
	"time"

	"github.com/primandproper/platform-go/v11/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// ValidIngredientGroupCreatedServiceEventType indicates a valid ingredient group was created.
	ValidIngredientGroupCreatedServiceEventType = "valid_ingredient_group_created"
	// ValidIngredientGroupUpdatedServiceEventType indicates a valid ingredient group was updated.
	ValidIngredientGroupUpdatedServiceEventType = "valid_ingredient_group_updated"
	// ValidIngredientGroupArchivedServiceEventType indicates a valid ingredient group was archived.
	ValidIngredientGroupArchivedServiceEventType = "valid_ingredient_group_archived"
)

func init() {
	gob.Register(new(ValidIngredientGroup))
	gob.Register(new(ValidIngredientGroupCreationRequestInput))
	gob.Register(new(ValidIngredientGroupUpdateRequestInput))
}

type (
	// ValidIngredientGroup represents a valid ingredient group.
	ValidIngredientGroup struct {
		_ struct{} `json:"-"`

		CreatedAt     time.Time                     `json:"createdAt"`
		LastUpdatedAt *time.Time                    `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time                    `json:"archivedAt"`
		ID            string                        `json:"id"`
		Name          string                        `json:"name"`
		Slug          string                        `json:"slug"`
		Description   string                        `json:"description"`
		Members       []*ValidIngredientGroupMember `json:"members"`
	}

	// ValidIngredientGroupMember represents a valid ingredient group member.
	ValidIngredientGroupMember struct {
		_ struct{} `json:"-"`

		CreatedAt       time.Time       `json:"createdAt"`
		ArchivedAt      *time.Time      `json:"archivedAt"`
		ID              string          `json:"id"`
		BelongsToGroup  string          `json:"belongsToGroup"`
		ValidIngredient ValidIngredient `json:"validIngredient"`
	}

	// ValidIngredientGroupCreationRequestInput represents what a user could set as input for creating valid ingredient groups.
	ValidIngredientGroupCreationRequestInput struct {
		_ struct{} `json:"-"`

		Name        string                                            `json:"name"`
		Slug        string                                            `json:"slug"`
		Description string                                            `json:"description"`
		Members     []*ValidIngredientGroupMemberCreationRequestInput `json:"members"`
	}

	// ValidIngredientGroupMemberCreationRequestInput represents what a user could set as input for creating valid ingredient group members.
	ValidIngredientGroupMemberCreationRequestInput struct {
		_ struct{} `json:"-"`

		ValidIngredientID string `json:"validIngredientID"`
	}

	// ValidIngredientGroupDatabaseCreationInput represents what a user could set as input for creating valid ingredient groups.
	ValidIngredientGroupDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID          string                                             `json:"-"`
		Name        string                                             `json:"-"`
		Slug        string                                             `json:"-"`
		Description string                                             `json:"-"`
		Members     []*ValidIngredientGroupMemberDatabaseCreationInput `json:"-"`
	}

	// ValidIngredientGroupMemberDatabaseCreationInput represents what a user could set as input for creating valid ingredient groups.
	ValidIngredientGroupMemberDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                string `json:"-"`
		ValidIngredientID string `json:"-"`
	}

	// ValidIngredientGroupUpdateRequestInput represents what a user could set as input for updating valid ingredient groups.
	ValidIngredientGroupUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        *string `json:"name,omitempty"`
		Slug        *string `json:"slug"`
		Description *string `json:"description,omitempty"`
	}

	// ValidIngredientGroupDataManager describes a structure capable of storing valid ingredient groups permanently.
	ValidIngredientGroupDataManager interface {
		ValidIngredientGroupExists(ctx context.Context, validIngredientID string) (bool, error)
		GetValidIngredientGroup(ctx context.Context, validIngredientID string) (*ValidIngredientGroup, error)
		GetValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientGroup], error)
		SearchForValidIngredientGroups(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[ValidIngredientGroup], error)
		CreateValidIngredientGroup(ctx context.Context, input *ValidIngredientGroupDatabaseCreationInput) (*ValidIngredientGroup, error)
		UpdateValidIngredientGroup(ctx context.Context, updated *ValidIngredientGroup) error
		ArchiveValidIngredientGroup(ctx context.Context, validIngredientID string) error
	}
)

// Update merges a ValidIngredientGroupUpdateRequestInput with a valid ingredient group.
func (x *ValidIngredientGroup) Update(input *ValidIngredientGroupUpdateRequestInput) {
	if input.Name != nil && *input.Name != x.Name {
		x.Name = *input.Name
	}

	if input.Slug != nil && *input.Slug != x.Slug {
		x.Slug = *input.Slug
	}

	if input.Description != nil && *input.Description != x.Description {
		x.Description = *input.Description
	}
}

var _ validation.ValidatableWithContext = (*ValidIngredientGroupCreationRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientGroupCreationRequestInput.
func (x *ValidIngredientGroupCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
		validation.Field(&x.Members, validation.Length(1, math.MaxUint8)),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientGroupDatabaseCreationInput)(nil)

// ValidateWithContext validates a ValidIngredientGroupDatabaseCreationInput.
func (x *ValidIngredientGroupDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.ID, validation.Required),
		validation.Field(&x.Name, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*ValidIngredientGroupUpdateRequestInput)(nil)

// ValidateWithContext validates a ValidIngredientGroupUpdateRequestInput.
func (x *ValidIngredientGroupUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Name, validation.Required),
	)
}
