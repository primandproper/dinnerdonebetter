package waitlists

import (
	"context"
	"encoding/gob"
	"time"

	"github.com/primandproper/platform-go/v5/filtering"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	// WaitlistCreatedServiceEventType indicates a waitlist was created.
	WaitlistCreatedServiceEventType = "waitlist_created"
	// WaitlistUpdatedServiceEventType indicates a waitlist was updated.
	WaitlistUpdatedServiceEventType = "waitlist_updated"
	// WaitlistArchivedServiceEventType indicates a waitlist was archived.
	WaitlistArchivedServiceEventType = "waitlist_archived"

	// WaitlistSignupCreatedServiceEventType indicates a waitlist signup was created.
	WaitlistSignupCreatedServiceEventType = "waitlist_signup_created"
	// WaitlistSignupUpdatedServiceEventType indicates a waitlist signup was updated.
	WaitlistSignupUpdatedServiceEventType = "waitlist_signup_updated"
	// WaitlistSignupArchivedServiceEventType indicates a waitlist signup was archived.
	WaitlistSignupArchivedServiceEventType = "waitlist_signup_archived"
)

func init() {
	gob.Register(new(Waitlist))
	gob.Register(new(WaitlistCreationRequestInput))
	gob.Register(new(WaitlistDatabaseCreationInput))
	gob.Register(new(WaitlistUpdateRequestInput))
	gob.Register(new(WaitlistSignup))
	gob.Register(new(WaitlistSignupCreationRequestInput))
	gob.Register(new(WaitlistSignupDatabaseCreationInput))
	gob.Register(new(WaitlistSignupUpdateRequestInput))
}

type (
	// Waitlist represents a waitlist users can join.
	Waitlist struct {
		_             struct{}   `json:"-"`
		CreatedAt     time.Time  `json:"createdAt"`
		ValidUntil    time.Time  `json:"validUntil"`
		LastUpdatedAt *time.Time `json:"lastUpdatedAt"`
		ArchivedAt    *time.Time `json:"archivedAt"`
		ID            string     `json:"id"`
		Name          string     `json:"name"`
		Description   string     `json:"description"`
	}

	// WaitlistSignup represents a signup for a waitlist.
	WaitlistSignup struct {
		_ struct{} `json:"-"`

		CreatedAt         time.Time  `json:"createdAt"`
		LastUpdatedAt     *time.Time `json:"lastUpdatedAt"`
		ArchivedAt        *time.Time `json:"archivedAt"`
		ID                string     `json:"id"`
		Notes             string     `json:"notes"`
		BelongsToWaitlist string     `json:"belongsToWaitlist"`
		BelongsToUser     string     `json:"belongsToUser"`
		BelongsToAccount  string     `json:"belongsToAccount"`
	}

	// WaitlistCreationRequestInput represents input for creating a waitlist.
	WaitlistCreationRequestInput struct {
		_           struct{}  `json:"-"`
		ValidUntil  time.Time `json:"validUntil"`
		Name        string    `json:"name"`
		Description string    `json:"description"`
	}

	// WaitlistDatabaseCreationInput is used for creating a waitlist in persistence.
	WaitlistDatabaseCreationInput struct {
		_           struct{}  `json:"-"`
		ValidUntil  time.Time `json:"-"`
		ID          string    `json:"-"`
		Name        string    `json:"-"`
		Description string    `json:"-"`
	}

	// WaitlistUpdateRequestInput represents input for updating a waitlist.
	WaitlistUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Name        *string    `json:"name,omitempty"`
		Description *string    `json:"description,omitempty"`
		ValidUntil  *time.Time `json:"validUntil,omitempty"`
	}

	// WaitlistSignupCreationRequestInput represents input for creating a waitlist signup.
	WaitlistSignupCreationRequestInput struct {
		_ struct{} `json:"-"`

		Notes             string `json:"notes"`
		BelongsToWaitlist string `json:"belongsToWaitlist"`
		BelongsToUser     string `json:"belongsToUser"`
		BelongsToAccount  string `json:"belongsToAccount"`
	}

	// WaitlistSignupDatabaseCreationInput is used for creating a waitlist signup in persistence.
	WaitlistSignupDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		ID                string `json:"-"`
		Notes             string `json:"-"`
		BelongsToWaitlist string `json:"-"`
		BelongsToUser     string `json:"-"`
		BelongsToAccount  string `json:"-"`
	}

	// WaitlistSignupUpdateRequestInput represents input for updating a waitlist signup.
	WaitlistSignupUpdateRequestInput struct {
		_ struct{} `json:"-"`

		Notes *string `json:"notes,omitempty"`
	}

	// WaitlistDataManager describes a structure capable of storing waitlists.
	WaitlistDataManager interface {
		WaitlistIsNotExpired(ctx context.Context, waitlistID string) (bool, error)
		GetWaitlist(ctx context.Context, waitlistID string) (*Waitlist, error)
		GetWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Waitlist], error)
		GetActiveWaitlists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Waitlist], error)
		CreateWaitlist(ctx context.Context, input *WaitlistDatabaseCreationInput) (*Waitlist, error)
		UpdateWaitlist(ctx context.Context, waitlist *Waitlist) error
		ArchiveWaitlist(ctx context.Context, waitlistID string) error
	}

	// WaitlistSignupDataManager describes a structure capable of storing waitlist signups.
	WaitlistSignupDataManager interface {
		GetWaitlistSignup(ctx context.Context, waitlistSignupID, waitlistID string) (*WaitlistSignup, error)
		GetWaitlistSignupByID(ctx context.Context, waitlistSignupID string) (*WaitlistSignup, error)
		GetWaitlistSignupsForWaitlist(ctx context.Context, waitlistID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[WaitlistSignup], error)
		GetWaitlistSignupsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[WaitlistSignup], error)
		CreateWaitlistSignup(ctx context.Context, input *WaitlistSignupDatabaseCreationInput) (*WaitlistSignup, error)
		UpdateWaitlistSignup(ctx context.Context, waitlistSignup *WaitlistSignup) error
		ArchiveWaitlistSignup(ctx context.Context, waitlistSignupID string) error
	}
)

// Update merges a WaitlistUpdateRequestInput into a Waitlist.
func (w *Waitlist) Update(input *WaitlistUpdateRequestInput) {
	if input.Name != nil && *input.Name != w.Name {
		w.Name = *input.Name
	}
	if input.Description != nil && *input.Description != w.Description {
		w.Description = *input.Description
	}
	if input.ValidUntil != nil && !input.ValidUntil.Equal(w.ValidUntil) {
		w.ValidUntil = *input.ValidUntil
	}
}

// Update merges a WaitlistSignupUpdateRequestInput into a WaitlistSignup.
func (w *WaitlistSignup) Update(input *WaitlistSignupUpdateRequestInput) {
	if input.Notes != nil && *input.Notes != w.Notes {
		w.Notes = *input.Notes
	}
}

var _ validation.ValidatableWithContext = (*WaitlistCreationRequestInput)(nil)

// ValidateWithContext validates a WaitlistCreationRequestInput.
func (w *WaitlistCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.Description, validation.Required),
		validation.Field(&w.ValidUntil, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WaitlistDatabaseCreationInput)(nil)

// ValidateWithContext validates a WaitlistDatabaseCreationInput.
func (w *WaitlistDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.ID, validation.Required),
		validation.Field(&w.Name, validation.Required),
		validation.Field(&w.Description, validation.Required),
		validation.Field(&w.ValidUntil, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WaitlistUpdateRequestInput)(nil)

// ValidateWithContext validates a WaitlistUpdateRequestInput.
func (w *WaitlistUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.Name, validation.When(w.Name != nil, validation.Required)),
		validation.Field(&w.Description, validation.When(w.Description != nil, validation.Required)),
		validation.Field(&w.ValidUntil, validation.When(w.ValidUntil != nil, validation.Required)),
	)
}

var _ validation.ValidatableWithContext = (*WaitlistSignupCreationRequestInput)(nil)

// ValidateWithContext validates a WaitlistSignupCreationRequestInput.
func (w *WaitlistSignupCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.BelongsToWaitlist, validation.Required),
		validation.Field(&w.Notes, validation.Required),
		validation.Field(&w.BelongsToUser, validation.Required),
		validation.Field(&w.BelongsToAccount, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WaitlistSignupDatabaseCreationInput)(nil)

// ValidateWithContext validates a WaitlistSignupDatabaseCreationInput.
func (w *WaitlistSignupDatabaseCreationInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.ID, validation.Required),
		validation.Field(&w.BelongsToWaitlist, validation.Required),
		validation.Field(&w.Notes, validation.Required),
		validation.Field(&w.BelongsToUser, validation.Required),
		validation.Field(&w.BelongsToAccount, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*WaitlistSignupUpdateRequestInput)(nil)

// ValidateWithContext validates a WaitlistSignupUpdateRequestInput.
func (w *WaitlistSignupUpdateRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		w,
		validation.Field(&w.Notes, validation.When(w.Notes != nil, validation.Required)),
	)
}
