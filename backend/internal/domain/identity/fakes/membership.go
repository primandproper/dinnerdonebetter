package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	identity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
)

// BuildFakeMembership builds a faked Membership.
func BuildFakeMembership() *identity.Membership {
	membership := fake.BuildFakeRecord[identity.Membership]()

	membership.Scope = ddbidentity.Scope()

	// A role the authorization package knows. The names are resolved to permissions by
	// the policy resolver, which refuses the ones it has no policy for, so a generated
	// role name is one every authorization check fails on.
	membership.Roles = []string{authorization.AccountMemberRoleName}

	// Which account a user lands in is decided by SetDefaultAccount, and the store
	// enforces that exactly one live membership carries it — so a fake that claimed to be
	// the default would be claiming something about the user's other memberships.
	membership.DefaultAccount = false

	return membership
}

// BuildFakeMembershipsList builds a faked page of memberships.
func BuildFakeMembershipsList() *filtering.QueryFilteredResult[identity.Membership] {
	return fake.BuildFakePage(BuildFakeMembership)
}

// BuildFakeMembershipWithUser builds a faked MembershipWithUser.
func BuildFakeMembershipWithUser() *identity.MembershipWithUser {
	membership := BuildFakeMembership()

	user := BuildFakeUser()

	// The roster is the read most likely to reach a response body, and the store redacts
	// the user before it answers — so a fake carrying a secret would let a test pass
	// against a response that leaked one.
	user = user.Redacted()

	return &identity.MembershipWithUser{
		User:       user,
		Membership: *membership,
	}
}

// BuildFakeMembershipsWithUserList builds a faked page of memberships joined to their members.
func BuildFakeMembershipsWithUserList() *filtering.QueryFilteredResult[identity.MembershipWithUser] {
	return fake.BuildFakePage(BuildFakeMembershipWithUser)
}
