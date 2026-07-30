package api

import (
	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	paymentscfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/config"

	analyticscfg "github.com/primandproper/platform-go/v8/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	emailcfg "github.com/primandproper/platform-go/v8/email/config"
	"github.com/primandproper/platform-go/v8/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v8/featureflags/config"
	msgconfig "github.com/primandproper/platform-go/v8/messagequeue/config"
	"github.com/primandproper/platform-go/v8/observability"
	routingcfg "github.com/primandproper/platform-go/v8/routing/config"
	"github.com/primandproper/platform-go/v8/server/http"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	// From APIServiceConfig
	do.Provide[*authcfg.Config](i, func(i do.Injector) (*authcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Auth, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Queues, nil
	})
	do.Provide[*emailcfg.Config](i, func(i do.Injector) (*emailcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Email, nil
	})
	do.Provide[*analyticscfg.Config](i, func(i do.Injector) (*analyticscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Analytics, nil
	})
	do.Provide[*featureflagscfg.Config](i, func(i do.Injector) (*featureflagscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.FeatureFlags, nil
	})
	do.Provide[encoding.Config](i, func(i do.Injector) (encoding.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.Encoding, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Events, nil
	})
	do.Provide[config.MetaSettings](i, func(i do.Injector) (config.MetaSettings, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.Meta, nil
	})
	do.Provide[*routingcfg.Config](i, func(i do.Injector) (*routingcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Routing, nil
	})
	do.Provide[http.Config](i, func(i do.Injector) (http.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.HTTPServer, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*config.ServicesConfig](i, func(i do.Injector) (*config.ServicesConfig, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Services, nil
	})

	// From ServicesConfig
	do.Provide[*authentication.Config](i, func(i do.Injector) (*authentication.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Auth, nil
	})
	do.Provide[*paymentscfg.Config](i, func(i do.Injector) (*paymentscfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Payments, nil
	})
	do.Provide[*authentication.OAuth2Config](i, func(i do.Injector) (*authentication.OAuth2Config, error) {
		cfg := do.MustInvoke[*authentication.Config](i)
		return &cfg.OAuth2, nil
	})
}
