package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v11/fake"
)

// BuildFakeAccountUserMembership builds a faked AccountUserMembership.
func BuildFakeAccountUserMembership() *types.AccountUserMembership {
	return fake.BuildFakeRecord[types.AccountUserMembership]()
}

// BuildFakeAccountUserMembershipWithUser builds a faked AccountUserMembershipWithUser.
func BuildFakeAccountUserMembershipWithUser() *types.AccountUserMembershipWithUser {
	membership := fake.BuildFakeRecord[types.AccountUserMembershipWithUser]()

	u := BuildFakeUser()

	// The membership is read back through the account read path, which does not return
	// anyone's second factor secret, so a fake that carried one would let a test pass
	// against a response that leaked it.
	u.TwoFactorSecret = ""
	membership.BelongsToUser = u

	return membership
}
