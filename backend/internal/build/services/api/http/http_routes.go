package api

import (
	"net/http"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v7/encoding"
	"github.com/primandproper/platform-go/v7/healthcheck"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	"github.com/primandproper/platform-go/v7/routing"
	routingcfg "github.com/primandproper/platform-go/v7/routing/config"
	"github.com/primandproper/platform-go/v7/version"
)

func ProvideAPIRouter(
	routingConfig routingcfg.Config,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	authService auth.AuthDataService,
	paymentsWebhookHandler *paymentswebhook.WebhookHandler,
	healthRegistry healthcheck.Registry,
) (*routing.Router, error) {
	encoder := encoding.NewServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON)

	router, err := routingcfg.NewRouter(&routingConfig, encoder, logger, tracerProvider, metricsProvider)
	if err != nil {
		return nil, err
	}

	router.Group("/_ops_", func(metaRouter *routing.Router) {
		// Expose a liveness check on /live
		metaRouter.Handle(http.MethodGet, "/live", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			res.WriteHeader(http.StatusOK)
		}))

		// Expose a readiness check on /ready
		metaRouter.Handle(http.MethodGet, "/ready", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			result := healthRegistry.CheckAll(req.Context())
			status := http.StatusOK
			if result.Status != healthcheck.StatusUp {
				status = http.StatusServiceUnavailable
			}
			encoder.EncodeResponseWithStatus(req.Context(), res, result, status)
		}))

		metaRouter.Handle(http.MethodGet, "/version", http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
			encoder.EncodeResponseWithStatus(req.Context(), res, version.Get(), http.StatusOK)
		}))
	})

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
