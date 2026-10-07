package grpcapi

import (
	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	identitycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/config"
	mealplanningcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/config"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	entitlementscfg "github.com/primandproper/platform-go/v15/entitlements/config"
	linkscfg "github.com/primandproper/platform-go/v15/links/config"
	meteringcfg "github.com/primandproper/platform-go/v15/metering/config"
	operationscfg "github.com/primandproper/platform-go/v15/operations/config"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	analyticscfg "github.com/primandproper/primitives-go/v2/analytics/config"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"
	emailcfg "github.com/primandproper/primitives-go/v2/email/config"
	ratelimitingcfg "github.com/primandproper/primitives-go/v2/ratelimiting/config"
	routingcfg "github.com/primandproper/primitives-go/v2/routing/config"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"
	uploadscfg "github.com/primandproper/primitives-go/v2/uploads/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	// From APIServiceConfig
	do.Provide[*authcfg.Config](i, func(i do.Injector) (*authcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Auth, nil
	})
	do.Provide[*queuescfg.Config](i, func(i do.Injector) (*queuescfg.Config, error) {
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
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.TextSearch, nil
	})
	do.Provide[*webhookscfg.Config](i, func(i do.Injector) (*webhookscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Webhooks, nil
	})
	do.Provide[*meteringcfg.Config](i, func(i do.Injector) (*meteringcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Metering, nil
	})
	do.Provide[*entitlementscfg.Config](i, func(i do.Injector) (*entitlementscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Entitlements, nil
	})
	do.Provide[*operationscfg.Config](i, func(i do.Injector) (*operationscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Operations, nil
	})
	do.Provide[*linkscfg.Config](i, func(i do.Injector) (*linkscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Links, nil
	})
	do.Provide[config.MetaSettings](i, func(i do.Injector) (config.MetaSettings, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.Meta, nil
	})
	do.Provide[*routingcfg.Config](i, func(i do.Injector) (*routingcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Routing, nil
	})
	do.Provide[*config.ServicesConfig](i, func(i do.Injector) (*config.ServicesConfig, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Services, nil
	})

	// From authentication.Config (nested under ServicesConfig.Auth)
	do.Provide[*authentication.Config](i, func(i do.Injector) (*authentication.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Auth, nil
	})
	do.Provide[*authcfg.TokensConfig](i, func(i do.Injector) (*authcfg.TokensConfig, error) {
		cfg := do.MustInvoke[*authentication.Config](i)
		return &cfg.Tokens, nil
	})
	do.Provide[*tokenscfg.Config](i, func(i do.Injector) (*tokenscfg.Config, error) {
		return &do.MustInvoke[*authcfg.TokensConfig](i).Config, nil
	})
	do.Provide[*oauth2servercfg.Config](i, func(i do.Injector) (*oauth2servercfg.Config, error) {
		cfg := do.MustInvoke[*authentication.Config](i)
		return &cfg.OAuth2, nil
	})
	do.Provide[*ratelimitingcfg.Config](i, func(i do.Injector) (*ratelimitingcfg.Config, error) {
		cfg := do.MustInvoke[*authentication.Config](i)
		return &cfg.RateLimiting, nil
	})

	// From ServicesConfig
	do.Provide[*identitycfg.Config](i, func(i do.Injector) (*identitycfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Users, nil
	})
	do.Provide[*mealplanningcfg.Config](i, func(i do.Injector) (*mealplanningcfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.MealPlanning, nil
	})
	do.Provide[*uploadscfg.Config](i, func(i do.Injector) (*uploadscfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.UploadedMedia, nil
	})
	do.Provide[*paymentscfg.Config](i, func(i do.Injector) (*paymentscfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Payments, nil
	})
}
