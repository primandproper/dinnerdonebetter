package converters

import (
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v5/identifiers"
)

// ConvertAccountInvitationCreationInputToAccountInvitationDatabaseCreationInput creates a AccountInvitationDatabaseCreationInput from a AccountInvitationCreationRequestInput.
func ConvertAccountInvitationCreationInputToAccountInvitationDatabaseCreationInput(userID, accountID, token string, input *identity.AccountInvitationCreationRequestInput) *identity.AccountInvitationDatabaseCreationInput {
	// if you don't specify an expiration, then it doesn't expire
	var expiresAt = time.Date(9999, 12, 12, 12, 12, 12, 12, time.UTC)
	if input.ExpiresAt != nil && !input.ExpiresAt.IsZero() {
		expiresAt = *input.ExpiresAt
	}

	x := &identity.AccountInvitationDatabaseCreationInput{
		ID:                   identifiers.New(),
		DestinationAccountID: accountID,
		FromUser:             userID,
		ToUser:               nil,
		Token:                token,
		ExpiresAt:            expiresAt,
		Note:                 input.Note,
		ToEmail:              input.ToEmail,
		ToName:               input.ToName,
	}

	return x
}

// ConvertAccountInvitationToAccountInvitationCreationInput builds a faked AccountInvitationCreationRequestInput.
func ConvertAccountInvitationToAccountInvitationCreationInput(accountInvitation *identity.AccountInvitation) *identity.AccountInvitationCreationRequestInput {
	return &identity.AccountInvitationCreationRequestInput{
		Note:      accountInvitation.Note,
		ToName:    accountInvitation.ToName,
		ToEmail:   accountInvitation.ToEmail,
		ExpiresAt: &accountInvitation.ExpiresAt,
	}
}

// ConvertAccountInvitationToAccountInvitationUpdateInput builds a faked AccountInvitationUpdateRequestInput.
func ConvertAccountInvitationToAccountInvitationUpdateInput(accountInvitation *identity.AccountInvitation) *identity.AccountInvitationUpdateRequestInput {
	return &identity.AccountInvitationUpdateRequestInput{
		Token: accountInvitation.Token,
		Note:  accountInvitation.Note,
	}
}

// ConvertAccountInvitationToAccountInvitationDatabaseCreationInput builds a faked AccountInvitationCreationRequestInput.
func ConvertAccountInvitationToAccountInvitationDatabaseCreationInput(accountInvitation *identity.AccountInvitation) *identity.AccountInvitationDatabaseCreationInput {
	return &identity.AccountInvitationDatabaseCreationInput{
		ID:                   accountInvitation.ID,
		FromUser:             accountInvitation.FromUser.ID,
		ToUser:               accountInvitation.ToUser,
		ToName:               accountInvitation.ToName,
		Note:                 accountInvitation.Note,
		ToEmail:              accountInvitation.ToEmail,
		Token:                accountInvitation.Token,
		DestinationAccountID: accountInvitation.DestinationAccount.ID,
		ExpiresAt:            accountInvitation.ExpiresAt,
	}
}
