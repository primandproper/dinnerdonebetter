package api

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v4/encoding"
	"github.com/primandproper/platform-go/v4/healthcheck"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/metrics"
	"github.com/primandproper/platform-go/v4/observability/tracing"
	"github.com/primandproper/platform-go/v4/routing"
	routingcfg "github.com/primandproper/platform-go/v4/routing/config"
	"github.com/primandproper/platform-go/v4/version"
)

func ProvideAPIRouter(
	routingConfig routingcfg.Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	authService auth.AuthDataService,
	paymentsWebhookHandler *paymentswebhook.WebhookHandler,
	healthRegistry healthcheck.Registry,
) (routing.Router, error) {
	router, err := routingConfig.NewRouter(logger, tracerProvider, metricsProvider)
	if err != nil {
		return nil, err
	}

	encoder := encoding.NewServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON)

	router.Route("/_ops_", func(metaRouter routing.Router) {
		// Expose a liveness check on /live
		metaRouter.Get("/live", func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		})

		// Expose a readiness check on /ready
		metaRouter.Get("/ready", func(res http.ResponseWriter, req *http.Request) {
			result := healthRegistry.CheckAll(req.Context())
			status := http.StatusOK
			if result.Status != healthcheck.StatusUp {
				status = http.StatusServiceUnavailable
			}
			encoder.EncodeResponseWithStatus(req.Context(), res, result, status)
		})

		metaRouter.Get("/version", func(res http.ResponseWriter, req *http.Request) {
			encoder.EncodeResponseWithStatus(req.Context(), res, version.Get(), http.StatusOK)
		})
	})

	router.Route("/oauth2", func(userRouter routing.Router) {
		userRouter.Get("/authorize", authService.AuthorizeHandler)
		userRouter.Post("/token", authService.TokenHandler)
		userRouter.Post("/revoke", authService.RevokeHandler)
	})

	router.Route("/api/payments/webhooks", func(paymentsRouter routing.Router) {
		paymentsRouter.Post("/{provider}", paymentsWebhookHandler.Handle)
	})

	return router, nil
}
