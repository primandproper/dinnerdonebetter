package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v10/identifiers"
)

// ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput builds an AccountUserMembershipDatabaseCreationInput from a membership.
func ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput(membership *types.AccountUserMembership) *types.AccountUserMembershipDatabaseCreationInput {
	return &types.AccountUserMembershipDatabaseCreationInput{
		ID:        identifiers.New(),
		Reason:    "",
		UserID:    membership.BelongsToUser,
		AccountID: membership.BelongsToAccount,
	}
}
