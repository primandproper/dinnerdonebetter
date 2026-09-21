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
	// PermissionRequirePasswordChange gates forcing a password change at a user's next
	// sign-in, and withdrawing the requirement again.
	PermissionRequirePasswordChange = identitygrpc.PermissionRequirePasswordChange

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
	//
	// Reading is not here. It is a member's, below: the grant says a caller may make this
	// kind of call and the authorizer says whose rows, so holding the read hostage to
	// administration would mean a member could not see the household they are in.
	IdentityAccountPermissions = []Permission{
		Permission(PermissionCreateAccounts),
		Permission(PermissionUpdateAccounts),
		Permission(PermissionArchiveAccounts),
		Permission(PermissionTransferAccountOwnership),
		Permission(PermissionManageMembers),
		Permission(PermissionInviteMembers),
		Permission(PermissionReadInvitations),
	}

	// IdentityOperatorPermissions is what an operator holds over the directory: reading
	// anybody, banning anybody, granting operator roles, forcing a password change, and
	// walking every account.
	//
	// None of it is an account holder's. A support engineer reading a user does not
	// become a member of the account they are looking at, which is the distinction
	// platform draws between a service role and a membership role.
	//
	// All four operator writes sit in one set here because this application has one
	// operator role. platform gives the password-change write its own grant on the
	// argument that it is the lightest of the four — it destroys nothing, discloses
	// nothing, and the subject clears it by choosing a password — so a deployment with a
	// support desk can hand out that one without also handing out the ban. If this
	// application ever grows that role, this is the permission to move first.
	IdentityOperatorPermissions = []Permission{
		Permission(PermissionReadUsers),
		Permission(PermissionCreateUsers),
		Permission(PermissionArchiveUsers),
		Permission(PermissionUpdateUserStatus),
		Permission(PermissionUpdateUserServiceRoles),
		Permission(PermissionRequirePasswordChange),
		Permission(PermissionListAllAccounts),
	}

	// IdentityPermissions contains all identity-related permissions.
	IdentityPermissions = append(append([]Permission{}, IdentityAccountPermissions...), IdentityOperatorPermissions...)
)

// The self-service grants, which are this application's rather than platform's.
//
// platform leaves eight of its RPCs out of identitygrpc.Permissions() — the ones whose
// subject is the caller themselves. That is the right default for a directory: a grant on
// the method cannot say "only about yourself", so platform declines to invent one and
// leaves the decision to the consumer.
//
// This application's interceptor refuses a method it has no entry for, which makes the
// omission a closed door rather than an open one. So the eight are declared here, under
// names this application owns, and granted to an account member — which is everybody, since
// registration mints an account. It is the same arrangement waitlists needed for Join and
// Withdraw, and for the same reason.
const (
	// UpdateOwnProfilePermission gates a user editing their own names.
	UpdateOwnProfilePermission Permission = "identity.self.update_profile"
	// RecordOwnAgreementPermission gates a user accepting the terms or the privacy policy.
	RecordOwnAgreementPermission Permission = "identity.self.record_agreement"
	// ReadOwnPrincipalPermission gates a user reading themselves and their memberships.
	ReadOwnPrincipalPermission Permission = "identity.self.read_principal"
	// SetOwnDefaultAccountPermission gates a user choosing which account they land in.
	SetOwnDefaultAccountPermission Permission = "identity.self.set_default_account"
	// AnswerOwnInvitationsPermission gates accepting or rejecting an invitation.
	//
	// Whoever holds the link may answer it, which is what the token on the request
	// establishes — so the grant says a member may answer invitations at all, and the
	// token says which one.
	AnswerOwnInvitationsPermission Permission = "identity.self.answer_invitations"
	// ReadOwnInvitationsPermission gates the two invitation lists keyed on the caller:
	// what they sent, and what they were sent.
	ReadOwnInvitationsPermission Permission = "identity.self.read_invitations"
)

// IdentitySelfPermissions is what every member holds over themselves.
var IdentitySelfPermissions = []Permission{
	// Reading an account, its roster and a user's memberships. It is a member's grant
	// because the authorizer is what narrows it: platform permits an account the caller
	// holds a live membership in, and a user they share one with. An administrator's
	// grant would make a member unable to see the household they are in.
	//
	// PermissionReadUsers is deliberately not here. It gates the two directory-wide
	// reads as well as the keyed one — listing every user and searching them — which is
	// an operator's question and not a member's.
	Permission(PermissionReadAccounts),
	UpdateOwnProfilePermission,
	RecordOwnAgreementPermission,
	ReadOwnPrincipalPermission,
	SetOwnDefaultAccountPermission,
	AnswerOwnInvitationsPermission,
	ReadOwnInvitationsPermission,
}
