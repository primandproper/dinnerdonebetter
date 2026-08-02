package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"
)

// OAuthMethodPermissions is a named type for Wire dependency injection.
type OAuthMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the OAuth service's method permissions.
func ProvideMethodPermissions() OAuthMethodPermissions {
	return OAuthMethodPermissions{
		oauthsvc.OAuthService_CreateOAuth2Client_FullMethodName: {
			authorization.CreateOAuth2ClientsPermission,
		},
		oauthsvc.OAuthService_GetOAuth2Client_FullMethodName: {
			authorization.ReadOAuth2ClientsPermission,
		},
		oauthsvc.OAuthService_GetOAuth2Clients_FullMethodName: {
			authorization.ReadOAuth2ClientsPermission,
		},
		oauthsvc.OAuthService_ArchiveOAuth2Client_FullMethodName: {
			authorization.ArchiveOAuth2ClientsPermission,
		},
	}
}
