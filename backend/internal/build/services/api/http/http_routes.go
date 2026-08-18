package api

import (
	"context"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v11/encoding"
	"github.com/primandproper/platform-go/v11/healthcheck"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/routing"
	routingcfg "github.com/primandproper/platform-go/v11/routing/config"
	"github.com/primandproper/platform-go/v11/version"
)

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
	)
	if err != nil {
		return nil, err
	}

	registerOpsRoutes(router, healthRegistry)

	router.Group("/oauth2", func(userRouter *routing.Router) {
		userRouter.Handle(http.MethodGet, "/authorize", http.HandlerFunc(authService.AuthorizeHandler))
		userRouter.Handle(http.MethodPost, "/token", http.HandlerFunc(authService.TokenHandler))
		userRouter.Handle(http.MethodPost, "/revoke", http.HandlerFunc(authService.RevokeHandler))
	})

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
