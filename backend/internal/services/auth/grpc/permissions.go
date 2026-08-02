package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
)

// AuthMethodPermissions is a named type for Wire dependency injection.
type AuthMethodPermissions map[string][]authorization.Permission

var noPerms = []authorization.Permission{}

// ProvideMethodPermissions returns a Wire provider for the auth service's method permissions.
func ProvideMethodPermissions() AuthMethodPermissions {
	return AuthMethodPermissions{
		// Methods that don't require specific permissions (authenticated user only)
		authsvc.AuthService_EvaluateBooleanFeatureFlag_FullMethodName:    noPerms,
		authsvc.AuthService_EvaluateInt64FeatureFlag_FullMethodName:      noPerms,
		authsvc.AuthService_EvaluateStringFeatureFlag_FullMethodName:     noPerms,
		authsvc.AuthService_CheckPermissions_FullMethodName:              noPerms,
		authsvc.AuthService_GetAuthStatus_FullMethodName:                 noPerms,
		authsvc.AuthService_GetActiveAccount_FullMethodName:              noPerms,
		authsvc.AuthService_UpdatePassword_FullMethodName:                noPerms,
		authsvc.AuthService_RefreshTOTPSecret_FullMethodName:             noPerms,
		authsvc.AuthService_VerifyTOTPSecret_FullMethodName:              noPerms,
		authsvc.AuthService_RequestPasswordResetToken_FullMethodName:     noPerms,
		authsvc.AuthService_RedeemPasswordResetToken_FullMethodName:      noPerms,
		authsvc.AuthService_ExchangeToken_FullMethodName:                 noPerms,
		authsvc.AuthService_AdminLoginForToken_FullMethodName:            noPerms,
		authsvc.AuthService_GetSelf_FullMethodName:                       noPerms,
		authsvc.AuthService_LoginForToken_FullMethodName:                 noPerms,
		authsvc.AuthService_RequestEmailVerificationEmail_FullMethodName: noPerms,
		authsvc.AuthService_RequestUsernameReminder_FullMethodName:       noPerms,
		authsvc.AuthService_VerifyEmailAddress_FullMethodName:            noPerms,
		authsvc.AuthService_BeginPasskeyRegistration_FullMethodName:      noPerms,
		authsvc.AuthService_FinishPasskeyRegistration_FullMethodName:     noPerms,
		authsvc.AuthService_ListPasskeys_FullMethodName:                  noPerms,
		authsvc.AuthService_ArchivePasskey_FullMethodName:                noPerms,
		authsvc.AuthService_ListActiveSessions_FullMethodName:            noPerms,
		authsvc.AuthService_RevokeSession_FullMethodName:                 noPerms,
		authsvc.AuthService_RevokeAllOtherSessions_FullMethodName:        noPerms,
		authsvc.AuthService_RevokeCurrentSession_FullMethodName:          noPerms,
		authsvc.AuthService_AdminListSessionsForUser_FullMethodName:      {authorization.ManageUserSessionsPermission},
		authsvc.AuthService_AdminRevokeUserSession_FullMethodName:        {authorization.ManageUserSessionsPermission},
		authsvc.AuthService_AdminRevokeAllUserSessions_FullMethodName:    {authorization.ManageUserSessionsPermission},
	}
}
