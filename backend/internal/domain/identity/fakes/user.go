package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUserCreationInput builds a faked UserRegistrationInput.
func BuildFakeUserCreationInput() *identity.UserRegistrationInput {
	exampleUser := BuildFakeUser()

	return &identity.UserRegistrationInput{
		Username:     exampleUser.Username,
		EmailAddress: fake.Email(),
		FirstName:    exampleUser.FirstName,
		LastName:     exampleUser.LastName,
		Password:     buildFakePassword(),
		Birthday:     exampleUser.Birthday,
	}
}

// BuildFakeUserRegistrationInputFromUser builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInputFromUser(user *identity.User) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		EmailAddress: user.EmailAddress,
		Password:     buildFakePassword(),
		Birthday:     user.Birthday,
	}
}

// BuildFakeUserRegistrationInputWithInviteFromUser builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInputWithInviteFromUser(user *identity.User) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Username:        user.Username,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		EmailAddress:    user.EmailAddress,
		Password:        buildFakePassword(),
		Birthday:        user.Birthday,
		InvitationToken: fake.UUID(),
		InvitationID:    BuildFakeID(),
	}
}

// BuildFakeUserPermissionModificationInput builds a faked ModifyUserPermissionsInput.
func BuildFakeUserPermissionModificationInput() *identity.ModifyUserPermissionsInput {
	return &identity.ModifyUserPermissionsInput{
		Reason:  fake.Sentence(10),
		NewRole: authorization.AccountMemberRole.String(),
	}
}
