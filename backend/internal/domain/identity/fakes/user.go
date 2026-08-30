package fakes

import (
	"encoding/base32"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *identity.User {
	user := fake.BuildFakeRecord[identity.User]()

	// Registration validates the address as an email, and a username has to be unique
	// across every user a test suite creates — hence two of them and a number.
	user.EmailAddress = gofakeit.Email()
	user.Username = fmt.Sprintf("%s_%d_%s", gofakeit.Username(), gofakeit.Uint8(), gofakeit.Username())

	// A user who has not verified their email yet, which is what a user who just
	// registered is. The explanation goes with it: UserDatabaseCreationInput has no
	// field for one, so registration always writes the empty string and only an admin
	// setting a status later fills it in. A generated value here disagrees with every
	// user read back out of the store.
	user.AccountStatus = string(identity.UnverifiedAccountStatus)
	user.AccountStatusExplanation = ""

	// Registration never demands a password change; that flag is raised later, and the
	// creation input has no field for it either.
	user.RequiresPasswordChange = false

	// The TOTP secret is decoded as base32 by everything that checks a code against it,
	// so a random string is one every login test fails on.
	user.TwoFactorSecret = base32.StdEncoding.EncodeToString([]byte(gofakeit.Password(false, true, true, false, false, 32)))

	// Both are optional on the type and set here anyway: a user whose second factor was
	// never verified cannot log in, and the birthday is what the age checks read.
	user.TwoFactorSecretVerifiedAt = pointer.To(fake.BuildFakeTime())
	user.Birthday = pointer.To(fake.BuildFakeTime())

	return user
}

// BuildFakeUsersList builds a faked UserList.
func BuildFakeUsersList() *filtering.QueryFilteredResult[identity.User] {
	return fake.BuildFakePage(BuildFakeUser)
}

// BuildFakeUserCreationInput builds a faked UserRegistrationInput.
func BuildFakeUserCreationInput() *identity.UserRegistrationInput {
	exampleUser := BuildFakeUser()

	input := BuildFakeUserRegistrationInputFromUser(exampleUser)
	input.EmailAddress = gofakeit.Email()

	return input
}

// BuildFakeUserRegistrationInput builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInput() *identity.UserRegistrationInput {
	return BuildFakeUserRegistrationInputFromUser(BuildFakeUser())
}

// BuildFakeUserRegistrationInputFromUser builds a faked UserRegistrationInput.
//
// Hand-written because it takes the user: registering the same person twice, or
// registering someone a test already has a row for, is a fake of a user rather than a
// fake of an input, and only the caller knows which user that is.
func BuildFakeUserRegistrationInputFromUser(user *identity.User) *identity.UserRegistrationInput {
	return &identity.UserRegistrationInput{
		Username:     user.Username,
		FirstName:    user.FirstName,
		LastName:     user.LastName,
		EmailAddress: user.EmailAddress,
		Password:     fake.BuildFakePassword(),
		Birthday:     user.Birthday,
	}
}

// BuildFakeUserRegistrationInputWithInviteFromUser builds a faked UserRegistrationInput.
func BuildFakeUserRegistrationInputWithInviteFromUser(user *identity.User) *identity.UserRegistrationInput {
	input := BuildFakeUserRegistrationInputFromUser(user)
	input.InvitationToken = fake.BuildFakeString()
	input.InvitationID = fake.BuildFakeID()

	return input
}

// BuildFakeUserCreationResponse builds a faked UserCreationResponse.
func BuildFakeUserCreationResponse() *identity.UserCreationResponse {
	user := BuildFakeUser()

	return &identity.UserCreationResponse{
		CreatedAt:       user.CreatedAt,
		Birthday:        user.Birthday,
		Username:        user.Username,
		EmailAddress:    user.EmailAddress,
		TwoFactorQRCode: fake.BuildFakeString(),
		CreatedUserID:   user.ID,
		AccountStatus:   user.AccountStatus,
		TwoFactorSecret: user.TwoFactorSecret,
		FirstName:       user.FirstName,
		LastName:        user.LastName,
	}
}

// BuildFakeAvatarUpdateInput builds a faked AvatarUpdateInput.
func BuildFakeAvatarUpdateInput() *identity.AvatarUpdateInput {
	return fake.BuildFakeRecord[identity.AvatarUpdateInput]()
}

// BuildFakeUserDetailsUpdateRequestInput builds a faked UserDetailsUpdateRequestInput.
func BuildFakeUserDetailsUpdateRequestInput() *identity.UserDetailsUpdateRequestInput {
	input := fake.BuildFakeRecord[identity.UserDetailsUpdateRequestInput]()

	input.CurrentPassword = gofakeit.Password(true, true, true, false, false, 32)

	// Six digits, which is the length the type validates and the length of the codes
	// the authenticator app produces.
	input.TOTPToken = "123456"

	return input
}

// BuildFakeUserDetailsDatabaseUpdateInput builds a faked UserDetailsDatabaseUpdateInput.
func BuildFakeUserDetailsDatabaseUpdateInput() *identity.UserDetailsDatabaseUpdateInput {
	return fake.BuildFakeRecord[identity.UserDetailsDatabaseUpdateInput]()
}

// BuildFakeUserPermissionModificationInput builds a faked ModifyUserPermissionsInput.
func BuildFakeUserPermissionModificationInput() *identity.ModifyUserPermissionsInput {
	input := fake.BuildFakeRecord[identity.ModifyUserPermissionsInput]()

	// A role the authorization package knows, since the handler resolves it to a set of
	// permissions and refuses the ones it cannot.
	input.NewRole = authorization.AccountMemberRole.String()

	return input
}
