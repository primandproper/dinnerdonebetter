package fakes

import (
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeUser builds a faked User.
func BuildFakeUser() *identity.User {
	user := fake.BuildFakeRecord[identity.User]()

	// The directory this application keeps. A generated scope names a tenancy no read in
	// this deployment is made in, so a user carrying one is a user nothing finds.
	user.Scope = ddbidentity.Scope()

	// Registration validates the address as an email, and a username has to be unique
	// across every user a test suite creates — hence two of them and a number.
	//
	// Both are folded, because the store folds them on write and on every lookup: a fake
	// spelled in mixed case is a fake that does not equal the row it was written as.
	user.EmailAddress = identity.FoldHandle(gofakeit.Email())
	user.Username = identity.FoldHandle(fmt.Sprintf("%s_%d_%s", gofakeit.Username(), gofakeit.Uint8(), gofakeit.Username()))

	// Never empty on a read: a row whose column is blank reads its handle back here, so a
	// fake with a random display name disagrees with a user created from it.
	user.DisplayName = user.Username

	// Good standing, which is what this application's registration writes.
	//
	// It is not platform's default — platform starts a user unverified and refuses
	// sign-in to anything else — and the disagreement is deliberate policy, stated at the
	// registration site. A fake carrying platform's default would be a user no test could
	// log in as, which is a fake of nobody this application creates.
	//
	// The explanation goes with it: nothing on the registration path sets one, so a
	// generated value disagrees with every user read back out of the store.
	user.AccountStatus = identity.StatusGood
	user.AccountStatusExplanation = ""

	// Registration never demands a password change; that flag is raised later by an
	// operator, and cleared by the next password write.
	user.RequiresPasswordChange = false

	// The TOTP secret is decoded as base32 by everything that checks a code against it,
	// so a random string is one every login test fails on.
	user.TwoFactorSecret = base32.StdEncoding.EncodeToString([]byte(gofakeit.Password(false, true, true, false, false, 32)))

	// A user whose second factor was never verified cannot log in.
	user.TwoFactorSecretVerifiedAt = pointer.To(fake.BuildFakeTime())

	// The token travels one way — a read fills in the digest and never the secret — so a
	// fake that carried one would be a fake no read can produce.
	user.EmailAddressVerificationToken = ""
	user.EmailAddressVerificationTokenDigest = ""

	// The role every registered user holds outside any account. Every read that returns a
	// User fills these in, and a random role name is one the policy resolver refuses.
	user.ServiceRoles = []string{authorization.ServiceUserRoleName}

	return user
}

// BuildFakeUsersList builds a faked page of users.
func BuildFakeUsersList() *filtering.QueryFilteredResult[identity.User] {
	return fake.BuildFakePage(BuildFakeUser)
}

// BuildFakeUserWithPassword builds a faked User whose hashed password field carries a
// plaintext password instead.
//
// The asymmetry is deliberate and is what the sign-in tests need: a test that wants to log
// somebody in has to hold the password they registered with, and the User type has nowhere
// else to put it. Nothing writes this user to the directory.
func BuildFakeUserWithPassword() (user *identity.User, password string) {
	user = BuildFakeUser()
	password = fake.BuildFakePassword()
	user.HashedPassword = password

	return user, password
}

// BuildFakeUsername builds a handle in the spelling the directory stores.
func BuildFakeUsername() string {
	return identity.FoldHandle(strings.TrimSpace(fmt.Sprintf("%s_%d", gofakeit.Username(), gofakeit.Uint16())))
}
