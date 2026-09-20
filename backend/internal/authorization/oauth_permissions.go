package authorization

import (
	oauth2clientsgrpc "github.com/primandproper/platform-go/v14/authentication/oauth2clients/grpc"
)

// The OAuth2 client permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go.
const (
	// CreateOAuth2ClientsPermission is a permission.
	CreateOAuth2ClientsPermission = oauth2clientsgrpc.PermissionCreateClients
	// ReadOAuth2ClientsPermission is a permission.
	ReadOAuth2ClientsPermission = oauth2clientsgrpc.PermissionReadClients
	// ArchiveOAuth2ClientsPermission is a permission.
	ArchiveOAuth2ClientsPermission = oauth2clientsgrpc.PermissionArchiveClients
)

var (
	// OAuthPermissions contains all OAuth-related permissions.
	OAuthPermissions = []Permission{
		CreateOAuth2ClientsPermission,
		ReadOAuth2ClientsPermission,
		ArchiveOAuth2ClientsPermission,
	}
)
