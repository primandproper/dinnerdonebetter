package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
)

// IdentityMethodPermissions is a named type for Wire dependency injection.
type IdentityMethodPermissions map[string][]authorization.Permission

var noPerms = []authorization.Permission{}

// ProvideMethodPermissions returns a Wire provider for the identity service's method permissions.
func ProvideMethodPermissions() IdentityMethodPermissions {
	return IdentityMethodPermissions{
		identitysvc.IdentityService_AdminSetPasswordChangeRequired_FullMethodName: {
			authorization.UpdateUserStatusPermission,
		},
		identitysvc.IdentityService_AdminUpdateUserStatus_FullMethodName: {
			authorization.UpdateUserStatusPermission,
		},
		identitysvc.IdentityService_UpdateAccount_FullMethodName: {
			authorization.UpdateAccountPermission,
		},
		identitysvc.IdentityService_ArchiveAccount_FullMethodName: {
			authorization.ArchiveAccountPermission,
		},
		identitysvc.IdentityService_CreateAccountInvitation_FullMethodName: {
			authorization.InviteUserToAccountPermission,
		},
		identitysvc.IdentityService_CancelAccountInvitation_FullMethodName: {
			authorization.InviteUserToAccountPermission,
		},
		identitysvc.IdentityService_TransferAccountOwnership_FullMethodName: {
			authorization.TransferAccountPermission,
		},
		identitysvc.IdentityService_UpdateAccountMemberPermissions_FullMethodName: {
			authorization.ModifyMemberPermissionsForAccountPermission,
		},
		identitysvc.IdentityService_ArchiveUserMembership_FullMethodName: {
			authorization.RemoveMemberAccountPermission,
		},
		identitysvc.IdentityService_GetUser_FullMethodName: {
			authorization.ReadUserPermission,
		},
		identitysvc.IdentityService_GetUsers_FullMethodName: {
			authorization.ReadUserPermission,
		},
		identitysvc.IdentityService_SearchForUsers_FullMethodName: {
			authorization.ReadUserPermission,
		},
		identitysvc.IdentityService_ArchiveUser_FullMethodName: {
			authorization.ArchiveUserPermission,
		},
		identitysvc.IdentityService_GetAccountsForUser_FullMethodName: {
			authorization.ReadUserDataPermission,
		},
		identitysvc.IdentityService_GetUsersForAccount_FullMethodName: {
			authorization.ReadUserDataPermission,
		},
		// Methods that don't require specific permissions (authenticated user only)
		identitysvc.IdentityService_RejectAccountInvitation_FullMethodName:       noPerms,
		identitysvc.IdentityService_AcceptAccountInvitation_FullMethodName:       noPerms,
		identitysvc.IdentityService_GetReceivedAccountInvitations_FullMethodName: noPerms,
		identitysvc.IdentityService_GetSentAccountInvitations_FullMethodName:     noPerms,
		identitysvc.IdentityService_GetAccountInvitation_FullMethodName:          noPerms,
		identitysvc.IdentityService_SetDefaultAccount_FullMethodName:             noPerms,
		identitysvc.IdentityService_CreateAccount_FullMethodName:                 noPerms,
		identitysvc.IdentityService_GetAccount_FullMethodName:                    noPerms,
		identitysvc.IdentityService_GetAccounts_FullMethodName:                   noPerms,
		identitysvc.IdentityService_UploadUserAvatar_FullMethodName:              noPerms,
		identitysvc.IdentityService_UpdateUserDetails_FullMethodName:             noPerms,
		identitysvc.IdentityService_UpdateUserEmailAddress_FullMethodName:        noPerms,
		identitysvc.IdentityService_UpdateUserUsername_FullMethodName:            noPerms,
	}
}
