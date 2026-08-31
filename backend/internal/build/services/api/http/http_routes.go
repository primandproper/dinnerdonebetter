package api

import (
	"context"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	"github.com/primandproper/platform-go/v13/encoding"
	"github.com/primandproper/platform-go/v13/healthcheck"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/routing"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"
	"github.com/primandproper/platform-go/v13/version"
)

// maxRequestBodyBytes bounds the request body of every route this router serves, the raw ones
// included: a Handle route has no binding step to enforce a bound in, so the Router applies its
// default to those as middleware instead.
//
// Nothing between the socket and a handler's read forms an opinion about how much to take: net/http
// bounds headers and not bodies, so without this the ceiling is whatever a client cares to send. The
// number is the platform's own default for an unparsed body, and it is generous for what this router
// actually serves — an OAuth2 form post and a payment provider's webhook event, neither of which is
// within three orders of magnitude of it. The bulk API is gRPC and is bounded there.
const maxRequestBodyBytes = 1 << 20 // 1 MiB

func ProvideAPIRouter(
	ctx context.Context,
	routingConfig routingcfg.Config,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	authService auth.AuthDataService,
	paymentsWebhookHandler *paymentswebhook.WebhookHandler,
	healthRegistry healthcheck.Registry,
) (*routing.Router, error) {
	encoder := encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger), encoding.WithTracerProvider(tracerProvider))

	router, err := routingcfg.NewRouter(ctx, &routingConfig, encoder,
		routingcfg.WithLogger(logger),
		routingcfg.WithTracerProvider(tracerProvider),
		routingcfg.WithMetricsProvider(metricsProvider),
		routingcfg.WithRouterOptions(routing.WithDefaultMaxRequestBody(maxRequestBodyBytes)),
	)
	if err != nil {
		return nil, err
	}

	registerOpsRoutes(router, healthRegistry)

	// The paths are the ones oauth2server publishes in its discovery document — which is derived
	// from the issuer, so mounting them anywhere else would mean advertising addresses that do
	// not answer. POST /authorize is what a first-party client uses: it carries the session JWT
	// in an Authorization header and gets the redirect back, where a browser GETs the same URL
	// and is shown a login form. RFC 7591 registration is deliberately not routed.
	router.Handle(http.MethodGet, oauth2server.PathAuthorizationServerMetadata, http.HandlerFunc(authService.AuthorizationServerMetadataHandler))
	router.Handle(http.MethodGet, oauth2server.PathAuthorize, http.HandlerFunc(authService.AuthorizeHandler))
	router.Handle(http.MethodPost, oauth2server.PathAuthorize, http.HandlerFunc(authService.AuthorizeHandler))
	router.Handle(http.MethodPost, oauth2server.PathToken, http.HandlerFunc(authService.TokenHandler))
	router.Handle(http.MethodPost, oauth2server.PathRevoke, http.HandlerFunc(authService.RevokeHandler))

	router.Group("/api/payments/webhooks", func(paymentsRouter *routing.Router) {
		paymentsRouter.Handle(http.MethodPost, "/{provider}", http.HandlerFunc(paymentsWebhookHandler.Handle))
	})

	if err = router.Err(); err != nil {
		return nil, err
	}

	return router, nil
}

// registerOpsRoutes mounts the probe and version endpoints under /_ops_.
//
// They are typed routes rather than raw handlers, which for /ready is what routing.Result exists
// for: the endpoint answers one body shape with either 200 or 503, and the status rides the
// return rather than a ResponseWriter the handler had to be handed. Modeling "unhealthy" as a
// returned error instead would say the handler failed and log a fault every time a probe found
// the service down — which is the one moment the logs are worth reading.
//
// Enveloping is off on all three. These are the same bytes they have always been: a liveness
// probe reading a status code, a readiness probe a kubelet compares against nothing, and a
// version document deploy tooling parses.
func registerOpsRoutes(router *routing.Router, healthRegistry healthcheck.Registry) {
	router.Group("/_ops_", func(metaRouter *routing.Router) {
		// Liveness: the process is running and serving. It deliberately checks nothing else,
		// so a dependency being down restarts no pods.
		routing.Get(metaRouter, "/live", func(context.Context, routing.Empty) (routing.Empty, error) {
			return routing.Empty{}, nil
		}, routing.WithEnvelope(false))

		// Readiness: every registered component reported up.
		routing.Get(metaRouter, "/ready", func(ctx context.Context, _ routing.Empty) (routing.Result[*healthcheck.Result], error) {
			result := healthRegistry.CheckAll(ctx)

			status := http.StatusOK
			if result.Status != healthcheck.StatusUp {
				status = http.StatusServiceUnavailable
			}

			return routing.Result[*healthcheck.Result]{Value: result, Status: status}, nil
		},
			routing.WithEnvelope(false),
			// The 503 is declared because nothing can infer it: the status is chosen per
			// response, and the reflected type says only that a Result was returned.
			routing.WithAdditionalResponse(http.StatusServiceUnavailable, new(healthcheck.Result), "one or more components are down"),
		)

		routing.Get(metaRouter, "/version", func(context.Context, routing.Empty) (version.Info, error) {
			return version.Get(), nil
		}, routing.WithEnvelope(false))
	})
}
