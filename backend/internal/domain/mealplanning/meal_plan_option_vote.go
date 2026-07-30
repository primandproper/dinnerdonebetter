package mealplanning

import (
	"context"
	"database/sql"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v8/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// MealPlanOptionVoteCreatedServiceEventType indicates a meal plan option vote was created.
	MealPlanOptionVoteCreatedServiceEventType = "meal_plan_option_vote_created"
	// MealPlanOptionVoteUpdatedServiceEventType indicates a meal plan option vote was updated.
	MealPlanOptionVoteUpdatedServiceEventType = "meal_plan_option_vote_updated"
	// MealPlanOptionVoteArchivedServiceEventType indicates a meal plan option vote was archived.
	MealPlanOptionVoteArchivedServiceEventType = "meal_plan_option_vote_archived"
)

func init() {
	gob.Register(new(MealPlanOptionVote))
	gob.Register(new(MealPlanOptionVoteCreationRequestInput))
	gob.Register(new(MealPlanOptionVoteUpdateRequestInput))
}

type (
	// MealPlanOptionVote represents a meal plan option vote.
	MealPlanOptionVote struct {
		_ struct{} `json:"-"`

		CreatedAt               time.Time  `json:"createdAt"`
		ArchivedAt              *time.Time `json:"archivedAt"`
		LastUpdatedAt           *time.Time `json:"lastUpdatedAt"`
		ID                      string     `json:"id"`
		Notes                   string     `json:"notes"`
		BelongsToMealPlanOption string     `json:"belongsToMealPlanOption"`
		ByUser                  string     `json:"byUser"`
		Rank                    uint8      `json:"rank"`
		Abstain                 bool       `json:"abstain"`
	}

	// NullableMealPlanOptionVote represents a fully nullable meal plan option vote.
	NullableMealPlanOptionVote struct {
		_ struct{} `json:"-"`

		Rank                    *uint8
		ID                      *string
		Notes                   *string
		BelongsToMealPlanOption *string
		ByUser                  *string
		Abstain                 *bool
		LastUpdatedAt           sql.NullTime
		CreatedAt               sql.NullTime
		ArchivedAt              sql.NullTime
	}

	// MealPlanOptionVoteCreationInput represents what a user could set as input for creating meal plan option votes.
	MealPlanOptionVoteCreationInput struct {
		_ struct{} `json:"-"`

		ID                      string `json:"-"`
		Notes                   string `json:"notes"`
		ByUser                  string `json:"-"`
		BelongsToMealPlanOption string `json:"belongsToMealPlanOption"`
		Rank                    uint8  `json:"rank"`
		Abstain                 bool   `json:"abstain"`
	}

	// MealPlanOptionVoteDatabaseCreationInput represents what a user could set as input for creating meal plan option votes.
	MealPlanOptionVoteDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                      string `json:"-"`
		Notes                   string `json:"-"`
		ByUser                  string `json:"-"`
		BelongsToMealPlanOption string `json:"-"`
		Rank                    uint8  `json:"-"`
		Abstain                 bool   `json:"-"`
	}

	// MealPlanOptionVoteCreationRequestInput is a pending container for multiple votes.
	MealPlanOptionVoteCreationRequestInput struct {
		_ struct{} `json:"-"`

		Votes []*MealPlanOptionVoteCreationInput `json:"votes"`
	}

	// MealPlanOptionVotesDatabaseCreationInput represents what a user could set as input for creating meal plan option votes.
	MealPlanOptionVotesDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		Votes []*MealPlanOptionVoteDatabaseCreationInput `json:"-"`
	}

	// MealPlanOptionVoteUpdateRequestInput represents what a user could set as input for updating meal plan option votes.
	MealPlanOptionVoteUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes                   *string `json:"notes,omitempty"`
		Rank                    *uint8  `json:"rank,omitempty"`
		Abstain                 *bool   `json:"abstain,omitempty"`
		BelongsToMealPlanOption string  `json:"belongsToMealPlanOption"`
	}

	// MissingVote represents missing votes in a meal plan.
	MissingVote struct {
		EventID  string `json:"eventID"`
		OptionID string `json:"optionID"`
		UserID   string `json:"userID"`
	}

	// MealPlanOptionVoteDataManager describes a structure capable of storing meal plan option votes permanently.
	MealPlanOptionVoteDataManager interface {
		MealPlanOptionVoteExists(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (bool, error)
		GetMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*MealPlanOptionVote, error)
		GetMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[MealPlanOptionVote], error)
		GetMealPlanOptionVotesForMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (x []*MealPlanOptionVote, err error)
		CreateMealPlanOptionVote(ctx context.Context, input *MealPlanOptionVotesDatabaseCreationInput) ([]*MealPlanOptionVote, error)
		UpdateMealPlanOptionVote(ctx context.Context, updated *MealPlanOptionVote) error
		ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error
	}
)

// Update merges an MealPlanOptionVoteUpdateRequestInput with a meal plan option vote.
func (x *MealPlanOptionVote) Update(input *MealPlanOptionVoteUpdateRequestInput) {
	if input.Rank != nil && *input.Rank != x.Rank {
		x.Rank = *input.Rank
	}

	if input.Abstain != nil && *input.Abstain != x.Abstain {
		x.Abstain = *input.Abstain
	}

	if input.Notes != nil && *input.Notes != x.Notes {
		x.Notes = *input.Notes
	}
}

var _ validation.ValidatableWithContext = (*MealPlanOptionVoteCreationInput)(nil)

// ValidateWithContext validates a MealPlanOptionVoteCreationInput.
func (x *MealPlanOptionVoteCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.BelongsToMealPlanOption, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanOptionVoteCreationRequestInput)(nil)

// ValidateWithContext validates a MealPlanOptionVoteCreationRequestInput.
func (x *MealPlanOptionVoteCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	if err := validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Votes, validation.Required),
	); err != nil {
		return err
	}

	// every vote must name the meal plan option it belongs to.
	for _, vote := range x.Votes {
		if err := vote.ValidateWithContext(ctx); err != nil {
			return err
		}
	}

	return nil
}

var _ validation.ValidatableWithContext = (*MealPlanOptionVotesDatabaseCreationInput)(nil)

// ValidateWithContext validates a MealPlanOptionVotesDatabaseCreationInput.
func (x *MealPlanOptionVotesDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Votes, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*MealPlanOptionVoteUpdateRequestInput)(nil)

// ValidateWithContext validates a MealPlanOptionVoteUpdateRequestInput.
func (x *MealPlanOptionVoteUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		// we cannot validate rank because zero is a valid value.
		validation.Field(&x.BelongsToMealPlanOption, validation.Required),
	)
}
