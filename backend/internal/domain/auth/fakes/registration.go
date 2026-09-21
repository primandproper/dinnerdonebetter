package fakes

import (
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/fake"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUserRegistrationInput builds a faked sign-up.
//
// Both agreements are accepted, because the input validates that they are: a registration
// that declined them is a refusal rather than a fake of a registration, and a test that
// wanted one would be asserting on the refusal.
func BuildFakeUserRegistrationInput() *auth.UserRegistrationInput {
	return BuildFakeUserRegistrationInputFromUser(identityfakes.BuildFakeUser())
}

// BuildFakeUserRegistrationInputFromUser builds a sign-up for a particular person.
//
// Hand-written because it takes the user: registering the same person twice, or registering
// somebody a test already has a row for, is a fake of a user rather than a fake of an
// input, and only the caller knows which user that is.
func BuildFakeUserRegistrationInputFromUser(user *identity.User) *auth.UserRegistrationInput {
	return &auth.UserRegistrationInput{
		Username:              user.Username,
		FirstName:             user.FirstName,
		LastName:              user.LastName,
		EmailAddress:          user.EmailAddress,
		AccountName:           fmt.Sprintf("%s's account", user.Username),
		Password:              fake.BuildFakePassword(),
		AcceptedTOS:           true,
		AcceptedPrivacyPolicy: true,
	}
}

// BuildFakeUserRegistrationInputWithInviteFromUser builds a sign-up answering an invitation.
func BuildFakeUserRegistrationInputWithInviteFromUser(user *identity.User) *auth.UserRegistrationInput {
	input := BuildFakeUserRegistrationInputFromUser(user)
	input.InvitationToken = fake.BuildFakeString()
	input.InvitationID = fake.BuildFakeID()

	return input
}

// BuildFakeUserCreationResponse builds a faked UserCreationResponse.
func BuildFakeUserCreationResponse() *auth.UserCreationResponse {
	user := identityfakes.BuildFakeUser()

	return &auth.UserCreationResponse{
		CreatedAt:        user.CreatedAt,
		Username:         user.Username,
		EmailAddress:     user.EmailAddress,
		TwoFactorQRCode:  gofakeit.URL(),
		CreatedUserID:    user.ID,
		CreatedAccountID: fake.BuildFakeID(),
		AccountStatus:    string(user.AccountStatus),
		TwoFactorSecret:  user.TwoFactorSecret,
		FirstName:        user.FirstName,
		LastName:         user.LastName,
	}
}
