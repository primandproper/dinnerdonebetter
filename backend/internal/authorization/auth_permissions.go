package authorization

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
)

var (
	// AuthPermissions contains all authentication-related permissions.
	AuthPermissions = []Permission{
		ImpersonateUserPermission,
		ManageUserSessionsPermission,
	}
)
