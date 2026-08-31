package grpcapi

import (
	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/handlers/authentication"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"
	identitycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/config"
	mealplanningcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/config"
	oauthcfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/config"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	uploadedmediacfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/config"

	analyticscfg "github.com/primandproper/platform-go/v13/analytics/config"
	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"
	tokenscfg "github.com/primandproper/platform-go/v13/authentication/tokens/config"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	emailcfg "github.com/primandproper/platform-go/v13/email/config"
	"github.com/primandproper/platform-go/v13/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v13/featureflags/config"
	httpclientcfg "github.com/primandproper/platform-go/v13/httpclient"
	msgconfig "github.com/primandproper/platform-go/v13/messagequeue/config"
	meteringcfg "github.com/primandproper/platform-go/v13/metering/config"
	"github.com/primandproper/platform-go/v13/observability"
	operationscfg "github.com/primandproper/platform-go/v13/operations/config"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v13/search/text/config"
	"github.com/primandproper/platform-go/v13/server/grpc"
	"github.com/primandproper/platform-go/v13/server/http"
	webhookscfg "github.com/primandproper/platform-go/v13/webhooks/config"

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
	do.Provide[*operationscfg.Config](i, func(i do.Injector) (*operationscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Operations, nil
	})
	do.Provide[*featureflagscfg.Config](i, func(i do.Injector) (*featureflagscfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.FeatureFlags, nil
	})
	do.Provide[*httpclientcfg.Config](i, func(i do.Injector) (*httpclientcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.HTTPClient, nil
	})
	do.Provide[encoding.Config](i, func(i do.Injector) (encoding.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return cfg.Encoding, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Events, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Observability, nil
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
	do.Provide[*grpc.Config](i, func(i do.Injector) (*grpc.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.GRPCServer, nil
	})
	do.Provide[*dbcfg.Config](i, func(i do.Injector) (*dbcfg.Config, error) {
		cfg := do.MustInvoke[*config.APIServiceConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		return &do.MustInvoke[*dbcfg.Config](i).Config, nil
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
	do.Provide[*authcfg.SessionsConfig](i, func(i do.Injector) (*authcfg.SessionsConfig, error) {
		return &do.MustInvoke[*authcfg.Config](i).Sessions, nil
	})
	do.Provide[*oauth2servercfg.Config](i, func(i do.Injector) (*oauth2servercfg.Config, error) {
		cfg := do.MustInvoke[*authentication.Config](i)
		return &cfg.OAuth2, nil
	})

	// From ServicesConfig
	do.Provide[*identitycfg.Config](i, func(i do.Injector) (*identitycfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Users, nil
	})
	do.Provide[*dataprivacycfg.Config](i, func(i do.Injector) (*dataprivacycfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.DataPrivacy, nil
	})
	do.Provide[*mealplanningcfg.Config](i, func(i do.Injector) (*mealplanningcfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.MealPlanning, nil
	})
	do.Provide[*oauthcfg.Config](i, func(i do.Injector) (*oauthcfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.OAuth2Clients, nil
	})
	do.Provide[*uploadedmediacfg.Config](i, func(i do.Injector) (*uploadedmediacfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.UploadedMedia, nil
	})
	do.Provide[*paymentscfg.Config](i, func(i do.Injector) (*paymentscfg.Config, error) {
		svc := do.MustInvoke[*config.ServicesConfig](i)
		return &svc.Payments, nil
	})
}
