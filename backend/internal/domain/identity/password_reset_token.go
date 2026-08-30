package identity

import (
	"context"
	"encoding/gob"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(PasswordResetTokenCreationRequestInput))
}

type (
	// UsernameReminderRequestInput represents what a user could set as input for creating password reset tokens.
	UsernameReminderRequestInput struct {
		_ struct{} `json:"-"`

		EmailAddress string `json:"emailAddress"`
	}

	// PasswordResetTokenCreationRequestInput represents what a user could set as input for creating password reset tokens.
	PasswordResetTokenCreationRequestInput struct {
		_ struct{} `json:"-"`

		EmailAddress string `json:"emailAddress"`
	}

	// PasswordResetTokenRedemptionRequestInput represents what a user could set as input for creating password reset tokens.
	PasswordResetTokenRedemptionRequestInput struct {
		_ struct{} `json:"-"`

		Token       string `json:"token"`
		NewPassword string `json:"newPassword"`
	}
)

var _ validation.ValidatableWithContext = (*UsernameReminderRequestInput)(nil)

// ValidateWithContext validates a UsernameReminderRequestInput.
func (x *UsernameReminderRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.EmailAddress, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PasswordResetTokenCreationRequestInput)(nil)

// ValidateWithContext validates a PasswordResetTokenCreationRequestInput.
func (x *PasswordResetTokenCreationRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.EmailAddress, validation.Required),
	)
}

var _ validation.ValidatableWithContext = (*PasswordResetTokenRedemptionRequestInput)(nil)

// ValidateWithContext validates a PasswordResetTokenRedemptionRequestInput.
func (x *PasswordResetTokenRedemptionRequestInput) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		x,
		validation.Field(&x.Token, validation.Required),
		validation.Field(&x.NewPassword, validation.Required),
	)
}
