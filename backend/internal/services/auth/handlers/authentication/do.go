package authentication

import (
	"context"

	authn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v12/authentication/oauth2server/config"
	"github.com/primandproper/platform-go/v12/authentication/tokens"
	"github.com/primandproper/platform-go/v12/authentication/totp"
	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthHTTPService registers the auth HTTP service providers with the injector.
func RegisterAuthHTTPService(i do.Injector) {
	// One authorization server for the process, resolved by both the HTTP handlers that
	// issue tokens and the gRPC interceptor that spends them. They must be the same
	// instance: Authenticate is a Store lookup, and a second server would be a second
	// Store with its own sweeper over the same rows.
	do.Provide[*oauth2server.Server](i, func(i do.Injector) (*oauth2server.Server, error) {
		return ProvideOAuth2Server(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*oauth2servercfg.Config](i),
			do.MustInvoke[database.Client](i),
			&subjectAuthenticator{
				identityRepo:  do.MustInvoke[identity.Repository](i),
				authenticator: do.MustInvoke[authn.Authenticator](i),
				totpVerifier:  do.MustInvoke[totp.Verifier](i),
				tokenIssuer:   do.MustInvoke[tokens.Issuer](i),
			},
			do.MustInvoke[oauth.Repository](i),
		)
	})

	do.Provide[auth.AuthDataService](i, func(i do.Injector) (auth.AuthDataService, error) {
		return ProvideService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[*oauth2server.Server](i),
			do.MustInvoke[tracing.Provider](i),
		)
	})
}
