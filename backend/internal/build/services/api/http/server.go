package api

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v7/healthcheck"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	"github.com/primandproper/platform-go/v7/routing"
	routingcfg "github.com/primandproper/platform-go/v7/routing/config"

	"github.com/samber/do/v2"
)

// RegisterAPIRouter registers the API router provider with the injector.
func RegisterAPIRouter(i do.Injector) {
	do.Provide[*routing.Router](i, func(i do.Injector) (*routing.Router, error) {
		return ProvideAPIRouter(
			*do.MustInvoke[*routingcfg.Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[auth.AuthDataService](i),
			do.MustInvoke[*paymentswebhook.WebhookHandler](i),
			do.MustInvoke[healthcheck.Registry](i),
		)
	})
}
