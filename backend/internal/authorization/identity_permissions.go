package authorization

import (
	identitygrpc "github.com/primandproper/platform-go/v14/identity/grpc"
)

// The identity permissions are platform's, re-exported under the names this
// application's policy already spells.
//
// They are re-exported rather than declared, for the reason the comment
// permissions give: the surface that enforces them is platform's, and a constant
// declared here with a different string would be a policy that grants something
// no method asks for. identitygrpc.Permissions() maps each RPC to one of these
// and is mounted unamended.
//
// The strings changed with the adoption — "update.account" became
// "identity.accounts.update" — because platform namespaces the domain first so
// that two composed domains cannot collide on a bare verb.
//
// Three of this application's old names collapsed into platform's two, and the
// collapse is worth naming rather than leaving to be discovered:
//
//   - RemoveMemberAccountPermission and ModifyMemberPermissionsForAccountPermission
//     are both PermissionManageMembers. platform gates SetMembershipRoles and
//     RemoveMembership on one permission, on the reading that an administrator who
//     can strip somebody's roles can already make their membership worthless.
//   - SearchUserPermission and ReadUserPermission are both PermissionReadUsers,
//     because reading a user and finding one are the same disclosure.
const (
	// PermissionReadUsers gates reading a user who is not you. Reading yourself is
	// GetPrincipal, which needs nothing.
	PermissionReadUsers = identitygrpc.PermissionReadUsers
	// PermissionCreateUsers gates registering somebody on a registrar's behalf. The
	// open sign-up this application serves is on the auth surface and is behind no
	// grant at all, because the caller has no session.
	PermissionCreateUsers = identitygrpc.PermissionCreateUsers
	// PermissionArchiveUsers gates deactivating a user.
	PermissionArchiveUsers = identitygrpc.PermissionArchiveUsers
	// PermissionUpdateUserStatus gates banning and reinstating.
	PermissionUpdateUserStatus = identitygrpc.PermissionUpdateUserStatus
	// PermissionUpdateUserServiceRoles gates granting and revoking operator roles.
	PermissionUpdateUserServiceRoles = identitygrpc.PermissionUpdateUserServiceRoles

	// PermissionReadAccounts gates reading an account, its roster and its memberships.
	PermissionReadAccounts = identitygrpc.PermissionReadAccounts
	// PermissionListAllAccounts gates walking every account in the deployment, which is
	// an operator read rather than an account holder's.
	PermissionListAllAccounts = identitygrpc.PermissionListAllAccounts
	// PermissionUpdateAccounts gates editing an account.
	PermissionUpdateAccounts = identitygrpc.PermissionUpdateAccounts
	// PermissionCreateAccounts gates minting an account beside the one registration made.
	PermissionCreateAccounts = identitygrpc.PermissionCreateAccounts
	// PermissionTransferAccountOwnership gates handing an account to somebody else.
	PermissionTransferAccountOwnership = identitygrpc.PermissionTransferAccountOwnership
	// PermissionArchiveAccounts gates retiring an account.
	PermissionArchiveAccounts = identitygrpc.PermissionArchiveAccounts
	// PermissionManageMembers gates setting a member's roles and removing them.
	PermissionManageMembers = identitygrpc.PermissionManageMembers

	// PermissionInviteMembers gates sending an invitation and withdrawing one.
	PermissionInviteMembers = identitygrpc.PermissionInviteMembers
	// PermissionReadInvitations gates reading one the caller sent.
	PermissionReadInvitations = identitygrpc.PermissionReadInvitations
)

var (
	// IdentityAccountPermissions is what an account administrator holds over their own
	// account: editing it, retiring it, handing it on, and managing who is in it.
	//
	// Creating an account is here too. Every user may mint one — registration already
	// gave them the first — and the authorizer decides which rows an allowed call may
	// touch, which for a creation is none that exist yet.
	IdentityAccountPermissions = []Permission{
		Permission(PermissionReadAccounts),
		Permission(PermissionCreateAccounts),
		Permission(PermissionUpdateAccounts),
		Permission(PermissionArchiveAccounts),
		Permission(PermissionTransferAccountOwnership),
		Permission(PermissionManageMembers),
		Permission(PermissionInviteMembers),
		Permission(PermissionReadInvitations),
	}

	// IdentityOperatorPermissions is what an operator holds over the directory: reading
	// anybody, banning anybody, granting operator roles, and walking every account.
	//
	// None of it is an account holder's. A support engineer reading a user does not
	// become a member of the account they are looking at, which is the distinction
	// platform draws between a service role and a membership role.
	IdentityOperatorPermissions = []Permission{
		Permission(PermissionReadUsers),
		Permission(PermissionCreateUsers),
		Permission(PermissionArchiveUsers),
		Permission(PermissionUpdateUserStatus),
		Permission(PermissionUpdateUserServiceRoles),
		Permission(PermissionListAllAccounts),
	}

	// IdentityPermissions contains all identity-related permissions.
	IdentityPermissions = append(append([]Permission{}, IdentityAccountPermissions...), IdentityOperatorPermissions...)
)
