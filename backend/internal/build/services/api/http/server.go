package api

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	paymentswebhook "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v11/healthcheck"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/routing"
	routingcfg "github.com/primandproper/platform-go/v11/routing/config"

	"github.com/samber/do/v2"
)

// RegisterAPIRouter registers the API router provider with the injector.
func RegisterAPIRouter(i do.Injector) {
	do.Provide[*routing.Router](i, func(i do.Injector) (*routing.Router, error) {
		return ProvideAPIRouter(
			do.MustInvoke[context.Context](i),
			*do.MustInvoke[*routingcfg.Config](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[auth.AuthDataService](i),
			do.MustInvoke[*paymentswebhook.WebhookHandler](i),
			do.MustInvoke[healthcheck.Registry](i),
		)
	})
}
