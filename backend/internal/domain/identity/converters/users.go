package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
)

// ConvertUserToUserDatabaseCreationInput builds a UserDatabaseCreationInput from a User.
func ConvertUserToUserDatabaseCreationInput(user *types.User) *types.UserDatabaseCreationInput {
	return &types.UserDatabaseCreationInput{
		ID:              user.ID,
		EmailAddress:    user.EmailAddress,
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		HashedPassword:  user.HashedPassword,
		TwoFactorSecret: user.TwoFactorSecret,
		Birthday:        user.Birthday,
	}
}

func ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput(x *types.UserDetailsUpdateRequestInput) *types.UserDetailsDatabaseUpdateInput {
	return &types.UserDetailsDatabaseUpdateInput{
		FirstName: x.FirstName,
		LastName:  x.LastName,
		Birthday:  x.Birthday,
	}
}
