package interceptors

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identitybuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/identity"

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
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	ratelimitingcfg "github.com/primandproper/primitives-go/v2/ratelimiting/config"

	"github.com/samber/do/v2"
)

// RegisterAuthInterceptor registers the auth interceptor, the extractor it resolves callers
// through, and the limiter the anonymous doors are throttled by, with the injector.
func RegisterAuthInterceptor(i do.Injector) {
	// One limiter for the process, which the gRPC doors and the OAuth2 login form both spend
	// from, each under keys of its own.
	do.Provide[ratelimiting.RateLimiter](i, func(i do.Injector) (ratelimiting.RateLimiter, error) {
		return ratelimitingcfg.NewRateLimiter(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*ratelimitingcfg.Config](i),
			ratelimitingcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			ratelimitingcfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			ratelimitingcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*signingrpc.PrincipalExtractor](i, func(i do.Injector) (*signingrpc.PrincipalExtractor, error) {
		verifier, err := ProvideAccessTokenVerifier(
			do.MustInvoke[*oauth2server.Server](i),
			do.MustInvoke[*oauth2servercfg.Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
		if err != nil {
			return nil, err
		}

		return ProvidePrincipalExtractor(
			do.MustInvoke[tokens.Issuer](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformidentity.Store](i),
			do.MustInvoke[*signin.Service](i),
			verifier,
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
		)
	})

	do.Provide[*AuthInterceptor](i, func(i do.Injector) (*AuthInterceptor, error) {
		return ProvideAuthInterceptor(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[*signingrpc.PrincipalExtractor](i),
			do.MustInvoke[*identitybuild.SessionBuilder](i),
			do.MustInvoke[MethodPermissionsMap](i),
		)
	})
}

// ProvidePrincipalExtractor builds platform's sign-in extractor, which is who every request on
// this server is resolved as.
//
// A bearer is tried as a sign-in token first and as an OAuth2 access token second. The order is
// the extractor's, and it is the right one for this server: a sign-in token is a signature check
// before it is a read, and every first-party client holds one, where an access token is a store
// lookup that only a third-party client presents. Trying the opaque token first, as this server
// used to, spent a database read on every first-party request to learn that it was not one.
//
// Service roles ride only on a token minted through the administrative door. Any other token —
// LoginForToken, a refresh of one, a passkey sign-in, an OAuth2 access token — keeps what
// authorization.OrdinaryServiceRoles keeps, which is the person and not the operator.
//
// The access-token half is oauth2server.Verifier, which is what decides a token is live, minted
// for this resource, and holding any scope a method needs. Its subject is resolved in the same
// directory, as an ordinary door's — a third-party client acting for an operator has skipped the
// administrative door, and is the last credential operator grants belong on.
//
// The password-change gate is the extractor's own, built from the identity store, and it allows
// platform's PasswordChangeMethods: everything a person told to change their password needs to
// learn so, make the change, or end a login they no longer trust.
func ProvidePrincipalExtractor(
	tokenIssuer tokens.Issuer,
	client database.Client,
	directory signingrpc.PrincipalDirectory,
	signIns signingrpc.SignInChecker,
	accessTokens *oauth2server.Verifier,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) (*signingrpc.PrincipalExtractor, error) {
	return signingrpc.NewPrincipalExtractor(tokenIssuer, client, directory,
		signingrpc.WithSignInCheck(signIns),
		signingrpc.WithOrdinaryServiceRoles(authorization.OrdinaryServiceRoles),
		signingrpc.WithAccessTokens(accessTokens),
		signingrpc.WithExtractorLogger(logger),
		signingrpc.WithExtractorTracerProvider(tracerProvider),
	)
}

// ProvideAccessTokenVerifier builds the resource-server half of the OAuth2 server this process
// runs: what an access token presented to the API is checked against.
//
// Verifier.Verify is primitives-go#41's answer to the audience check this server used to make for
// itself. It refuses a token with no audience, which the local check let through: a token minted
// with no resource indicator is one RFC 8707 exists to stop, because any sibling resource server
// sharing this store — the MCP server does — would accept it too. A client asking for a token for
// this API names it in the resource parameter.
func ProvideAccessTokenVerifier(
	srv *oauth2server.Server,
	cfg *oauth2servercfg.Config,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
) (*oauth2server.Verifier, error) {
	metadata, err := oauth2server.NewResourceMetadata(ResourceIdentifier(cfg), []string{cfg.Issuer})
	if err != nil {
		return nil, err
	}

	return oauth2server.NewVerifier(metadata, srv,
		oauth2server.WithVerifierLogger(logger),
		oauth2server.WithVerifierTracerProvider(tracerProvider),
		oauth2server.WithVerifierMetricsProvider(metricsProvider),
	)
}

// ResourceIdentifier is the RFC 8707 name this server answers to.
//
// The first configured resource, falling back to the issuer: a deployment where the
// authorization server and the resource server are the same process — which this one is —
// names itself once, and Resources exists for the case where they are not.
func ResourceIdentifier(cfg *oauth2servercfg.Config) string {
	if len(cfg.Resources) > 0 {
		return cfg.Resources[0]
	}

	return cfg.Issuer
}
