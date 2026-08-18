package converters

// The conversions in this file are hand-written: each does something the declaration format in
// cmd/tools/codegen/converters cannot express, and the note above each one says what. Everything
// else in this package is generated from those declarations into converters_generated.go.
//
// Adding a conversion here rather than declaring it is a decision, not a default. A conversion
// that is a field copy with a handful of exceptions belongs in the declaration, where the
// generator guarantees no destination field is silently forgotten.

import (
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"

	"github.com/primandproper/platform-go/v11/identifiers"
)

// ConvertAccountInvitationCreationInputToAccountInvitationDatabaseCreationInput creates a AccountInvitationDatabaseCreationInput from a AccountInvitationCreationRequestInput.
//
// Hand-written: an invitation with no expiry given is given one in the year 9999, which is a
// policy about what absence means rather than a copy.
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

// ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput builds the destination
// from its source.
//
// Hand-written: its source is a gRPC request type rather than a domain entity, so there is no
// domain struct for the generator to plan the copy against.
func ConvertGRPCAdminUpdateUserStatusRequestToUserAccountStatusUpdateInput(input *identitysvc.AdminUpdateUserStatusRequest) *identity.UserAccountStatusUpdateInput {
	return &identity.UserAccountStatusUpdateInput{
		NewStatus:    input.NewStatus,
		Reason:       input.Reason,
		TargetUserID: input.TargetUserId,
	}
}
