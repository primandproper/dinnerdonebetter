package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v13/fake"
)

// BuildFakeAccountUserMembership builds a faked AccountUserMembership.
func BuildFakeAccountUserMembership() *types.AccountUserMembership {
	membership := fake.BuildFakeRecord[types.AccountUserMembership]()

	// Which account a user defaults to is decided by MarkAccountUserMembershipAsUserDefault,
	// not at creation — the creation input has no field for it, so a new membership is
	// never the default one.
	membership.DefaultAccount = false

	return membership
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
