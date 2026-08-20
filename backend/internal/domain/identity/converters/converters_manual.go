package converters

// The conversions in this file are hand-written: each does something the generator in
// cmd/tools/codegen/converters does not produce — it fails, it fans one value out into many, it
// defaults something, it needs a second entity to make sense of the first. exceptions.go names
// each one and says why.
//
// Everything else in this package is generated. A conversion that is a field copy with a handful
// of exceptions belongs there, where no destination field can be silently forgotten.

import (
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v12/identifiers"
)

// ConvertAccountInvitationCreationRequestInputToAccountInvitationDatabaseCreationInput creates a AccountInvitationDatabaseCreationInput from a AccountInvitationCreationRequestInput.
func ConvertAccountInvitationCreationRequestInputToAccountInvitationDatabaseCreationInput(userID, accountID, token string, input *identity.AccountInvitationCreationRequestInput) *identity.AccountInvitationDatabaseCreationInput {
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

func ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput(input *identitysvc.AdminUpdateUserStatusRequest) *identity.UserAccountStatusUpdateInput {
	return &identity.UserAccountStatusUpdateInput{
		NewStatus:    input.NewStatus,
		Reason:       input.Reason,
		TargetUserID: input.TargetUserId,
	}
}

func ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput(x *identity.UserDetailsUpdateRequestInput) *identity.UserDetailsDatabaseUpdateInput {
	return &identity.UserDetailsDatabaseUpdateInput{
		FirstName: x.FirstName,
		LastName:  x.LastName,
		Birthday:  x.Birthday,
	}
}
