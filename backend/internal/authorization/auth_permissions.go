package authorization

import (
	signingrpc "github.com/primandproper/platform-go/v14/authentication/signin/grpc"
)

// What is left here after the identity adoption: the two authority questions that
// are this application's rather than the directory's.
//
// Reading, searching, archiving a user and changing their status all moved to
// identity_permissions.go, under platform's names, because platform's identity
// service is what enforces them now. Impersonation and session management did not
// move: neither is an identity operation — one is an audit-visible act this
// application performs with its own tokens, the other is the session store's —
// and platform's directory has no RPC that asks for either.
const (
	// ImpersonateUserPermission is a service admin permission.
	ImpersonateUserPermission Permission = "imitate.user"
	// ManageUserSessionsPermission is a service admin permission.
	ManageUserSessionsPermission Permission = "manage.user_sessions"

	// ReadAnySignInsPermission allows listing somebody else's sign-ins through platform's
	// SignInAdministrationService, and EndAnySignInsPermission ending them. They are the
	// operator half of platform's sign-in, and the counterpart there of
	// ManageUserSessionsPermission over AuthService's sessions.
	ReadAnySignInsPermission = signingrpc.PermissionReadAnySignIns
	EndAnySignInsPermission  = signingrpc.PermissionEndAnySignIns
)

var (
	// AuthPermissions contains all authentication-related permissions.
	AuthPermissions = []Permission{
		ImpersonateUserPermission,
		ManageUserSessionsPermission,
		ReadAnySignInsPermission,
		EndAnySignInsPermission,
	}
)
