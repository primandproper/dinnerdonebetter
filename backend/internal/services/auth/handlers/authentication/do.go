package authentication

import (
	"context"

	authn "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v4/authentication/tokens"
	"github.com/primandproper/platform-go/v4/authentication/totp"
	"github.com/primandproper/platform-go/v4/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"

	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/samber/do/v2"
)

// RegisterAuthHTTPService registers the auth HTTP service providers with the injector.
func RegisterAuthHTTPService(i do.Injector) {
	do.Provide[*manage.Manager](i, func(i do.Injector) (*manage.Manager, error) {
		return ProvideOAuth2ClientManager(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[*OAuth2Config](i),
			do.MustInvoke[oauth.Repository](i),
		), nil
	})

	do.Provide[*server.Server](i, func(i do.Injector) (*server.Server, error) {
		return ProvideOAuth2ServerImplementation(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
			do.MustInvoke[authn.Authenticator](i),
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[*manage.Manager](i),
		), nil
	})

	do.Provide[auth.AuthDataService](i, func(i do.Injector) (auth.AuthDataService, error) {
		return ProvideService(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[*Config](i),
			do.MustInvoke[authn.Authenticator](i),
			do.MustInvoke[totp.Verifier](i),
			do.MustInvoke[oauth.Repository](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
		)
	})
}
