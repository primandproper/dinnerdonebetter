package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"

	"github.com/primandproper/platform-go/v4/identifiers"
)

// ConvertWaitlistToWaitlistUpdateRequestInput creates a WaitlistUpdateRequestInput from a Waitlist.
func ConvertWaitlistToWaitlistUpdateRequestInput(x *types.Waitlist) *types.WaitlistUpdateRequestInput {
	out := &types.WaitlistUpdateRequestInput{
		Name:        &x.Name,
		Description: &x.Description,
		ValidUntil:  &x.ValidUntil,
	}

	return out
}

// ConvertWaitlistCreationRequestInputToWaitlistDatabaseCreationInput creates a WaitlistDatabaseCreationInput from a WaitlistCreationRequestInput.
func ConvertWaitlistCreationRequestInputToWaitlistDatabaseCreationInput(x *types.WaitlistCreationRequestInput) *types.WaitlistDatabaseCreationInput {
	out := &types.WaitlistDatabaseCreationInput{
		ID:          identifiers.New(),
		Name:        x.Name,
		Description: x.Description,
		ValidUntil:  x.ValidUntil,
	}

	return out
}

// ConvertWaitlistToWaitlistCreationRequestInput builds a WaitlistCreationRequestInput from a Waitlist.
func ConvertWaitlistToWaitlistCreationRequestInput(x *types.Waitlist) *types.WaitlistCreationRequestInput {
	return &types.WaitlistCreationRequestInput{
		Name:        x.Name,
		Description: x.Description,
		ValidUntil:  x.ValidUntil,
	}
}

// ConvertWaitlistToWaitlistDatabaseCreationInput builds a WaitlistDatabaseCreationInput from a Waitlist.
func ConvertWaitlistToWaitlistDatabaseCreationInput(x *types.Waitlist) *types.WaitlistDatabaseCreationInput {
	return &types.WaitlistDatabaseCreationInput{
		ID:          x.ID,
		Name:        x.Name,
		Description: x.Description,
		ValidUntil:  x.ValidUntil,
	}
}

// ConvertWaitlistSignupToWaitlistSignupUpdateRequestInput creates a WaitlistSignupUpdateRequestInput from a WaitlistSignup.
func ConvertWaitlistSignupToWaitlistSignupUpdateRequestInput(x *types.WaitlistSignup) *types.WaitlistSignupUpdateRequestInput {
	out := &types.WaitlistSignupUpdateRequestInput{
		Notes: &x.Notes,
	}

	return out
}

// ConvertWaitlistSignupCreationRequestInputToWaitlistSignupDatabaseCreationInput creates a WaitlistSignupDatabaseCreationInput from a WaitlistSignupCreationRequestInput.
func ConvertWaitlistSignupCreationRequestInputToWaitlistSignupDatabaseCreationInput(x *types.WaitlistSignupCreationRequestInput) *types.WaitlistSignupDatabaseCreationInput {
	out := &types.WaitlistSignupDatabaseCreationInput{
		ID:                identifiers.New(),
		Notes:             x.Notes,
		BelongsToWaitlist: x.BelongsToWaitlist,
		BelongsToUser:     x.BelongsToUser,
		BelongsToAccount:  x.BelongsToAccount,
	}

	return out
}

// ConvertWaitlistSignupToWaitlistSignupCreationRequestInput builds a WaitlistSignupCreationRequestInput from a WaitlistSignup.
func ConvertWaitlistSignupToWaitlistSignupCreationRequestInput(x *types.WaitlistSignup) *types.WaitlistSignupCreationRequestInput {
	return &types.WaitlistSignupCreationRequestInput{
		Notes:             x.Notes,
		BelongsToWaitlist: x.BelongsToWaitlist,
		BelongsToUser:     x.BelongsToUser,
		BelongsToAccount:  x.BelongsToAccount,
	}
}

// ConvertWaitlistSignupToWaitlistSignupDatabaseCreationInput builds a WaitlistSignupDatabaseCreationInput from a WaitlistSignup.
func ConvertWaitlistSignupToWaitlistSignupDatabaseCreationInput(x *types.WaitlistSignup) *types.WaitlistSignupDatabaseCreationInput {
	return &types.WaitlistSignupDatabaseCreationInput{
		ID:                x.ID,
		Notes:             x.Notes,
		BelongsToWaitlist: x.BelongsToWaitlist,
		BelongsToUser:     x.BelongsToUser,
		BelongsToAccount:  x.BelongsToAccount,
	}
}
