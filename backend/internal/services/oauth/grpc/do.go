package grpc

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth/manager"
	oauthsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOAuthService registers the OAuth gRPC service with the injector.
func RegisterOAuthService(i do.Injector) {
	do.Provide[OAuthMethodPermissions](i, func(i do.Injector) (OAuthMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[oauthsvc.OAuthServiceServer](i, func(i do.Injector) (oauthsvc.OAuthServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[manager.OAuth2Manager](i),
		), nil
	})
}
