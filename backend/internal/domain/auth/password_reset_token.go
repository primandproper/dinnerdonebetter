package auth

import (
	"context"
	"encoding/gob"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

func init() {
	gob.Register(new(PasswordResetTokenCreationRequestInput))
}

// The stored side of a password reset token is not here. The row — its digest, its
// deadline, and its redemption — belongs to platform-go's
// authentication/passwordreset, whose Store is what this domain's manager depends on.
// What remains here is the request surface: what a caller sends to ask for a reset,
// and what they send to spend one.
//
// There is deliberately no type in this package holding a token's plaintext. The
// secret exists once, in the passwordreset.Issuance that produced it, and travels
// from there to exactly one email. A struct with a Token field is how a copy of it
// ends up in a log line, a response body, or a database column.

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
