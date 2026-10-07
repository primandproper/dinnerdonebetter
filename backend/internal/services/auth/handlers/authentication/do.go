package authentication

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/authserver"
	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterAuthHTTPService registers the auth HTTP service providers with the injector.
func RegisterAuthHTTPService(i do.Injector) {
	// One authorization server for the process, resolved by both the HTTP handlers that
	// issue tokens and the gRPC interceptor that spends them. They must be the same
	// instance: Authenticate is a Store lookup, and a second server would be a second
	// Store with its own sweeper over the same rows.
	do.Provide[*oauth2server.Server](i, func(i do.Injector) (*oauth2server.Server, error) {
		logger := do.MustInvoke[logging.Logger](i)
		tracerProvider := do.MustInvoke[tracing.Provider](i)
		metricsProvider := do.MustInvoke[metrics.Provider](i)
		dbClient := do.MustInvoke[database.Client](i)
		clients := do.MustInvoke[platformoauth2clients.Store](i)
		signIn := do.MustInvoke[*signin.Service](i)

		authenticator, err := ProvideLoginFormAuthenticator(signIn, clients, dbClient, logger, tracerProvider, metricsProvider)
		if err != nil {
			return nil, err
		}

		// The sign-in half of the extractor every API request is resolved by. It is built here
		// rather than resolved, because the API's extractor accepts this server's access
		// tokens too and so cannot exist until this server does; this one only ever reads a
		// sign-in token, and needs neither the access tokens nor the password change gate.
		signIns, err := signingrpc.NewPrincipalExtractor(
			do.MustInvoke[tokens.Issuer](i),
			dbClient,
			do.MustInvoke[platformidentity.Store](i),
			signingrpc.WithSignInCheck(signIn),
			signingrpc.WithoutPasswordChangeGate(),
			signingrpc.WithExtractorLogger(logger),
			signingrpc.WithExtractorTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, err
		}

		resolver, err := authserver.NewGuardedResolver(NewSessionResolver(signIns), clients, dbClient,
			authserver.WithResolverLogger(logger),
			authserver.WithResolverTracerProvider(tracerProvider),
			authserver.WithResolverMetricsProvider(metricsProvider),
		)
		if err != nil {
			return nil, err
		}

		return ProvideOAuth2Server(
			do.MustInvoke[context.Context](i),
			logger,
			tracerProvider,
			metricsProvider,
			do.MustInvoke[*oauth2servercfg.Config](i),
			dbClient,
			authenticator,
			resolver,
			clients,
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
