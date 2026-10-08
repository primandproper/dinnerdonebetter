package api

import (
	"context"

	paymentsbuild "github.com/primandproper/dinnerdonebetter/backend/internal/build/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	"github.com/primandproper/primitives-go/v2/healthcheck"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/ratelimiting"
	"github.com/primandproper/primitives-go/v2/routing"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"

	"github.com/samber/do/v2"
)

// RegisterAPIRouter registers the API router provider with the injector.
func RegisterAPIRouter(i do.Injector) {
	do.Provide[*routing.Router](i, func(i do.Injector) (*routing.Router, error) {
		authorizeThrottle, err := interceptors.NewAuthorizeFormThrottle(
			do.MustInvoke[ratelimiting.RateLimiter](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
		if err != nil {
			return nil, err
		}

		return ProvideAPIRouter(
			do.MustInvoke[context.Context](i),
			*do.MustInvoke[*routingcfg.Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[auth.AuthDataService](i),
			do.MustInvoke[*paymentsbuild.WebhookHandlers](i),
			do.MustInvoke[healthcheck.Registry](i),
			do.MustInvoke[*PlatformSurfaces](i),
			authorizeThrottle,
		)
	})
}
