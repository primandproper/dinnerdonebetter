package config

import (
	authcfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/config"

	analyticscfg "github.com/primandproper/platform-go/v5/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v5/database/config"
	emailcfg "github.com/primandproper/platform-go/v5/email/config"
	"github.com/primandproper/platform-go/v5/encoding"
	featureflagscfg "github.com/primandproper/platform-go/v5/featureflags/config"
	httpclientcfg "github.com/primandproper/platform-go/v5/httpclient"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v5/notifications/mobile/config"
	"github.com/primandproper/platform-go/v5/observability"
	routingcfg "github.com/primandproper/platform-go/v5/routing/config"
	textsearchcfg "github.com/primandproper/platform-go/v5/search/text/config"
	"github.com/primandproper/platform-go/v5/server/grpc"
	"github.com/primandproper/platform-go/v5/server/http"

	"github.com/samber/do/v2"
)

func ProvideHTTPServerConfigFromAPIServiceConfig(cfg *APIServiceConfig) http.Config {
	return cfg.HTTPServer
}

// RegisterAPIServiceConfigs registers all APIServiceConfig sub-fields with the injector.
func RegisterAPIServiceConfigs(i do.Injector) {
	do.Provide[*authcfg.Config](i, func(i do.Injector) (*authcfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Auth, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[emailcfg.Config](i, func(i do.Injector) (emailcfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.Email, nil
	})
	do.Provide[analyticscfg.Config](i, func(i do.Injector) (analyticscfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.Analytics, nil
	})
	do.Provide[featureflagscfg.Config](i, func(i do.Injector) (featureflagscfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.FeatureFlags, nil
	})
	do.Provide[encoding.Config](i, func(i do.Injector) (encoding.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.Encoding, nil
	})
	do.Provide[routingcfg.Config](i, func(i do.Injector) (routingcfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.Routing, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[MetaSettings](i, func(i do.Injector) (MetaSettings, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.Meta, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Events, nil
	})
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Queues, nil
	})
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.TextSearch, nil
	})
	do.Provide[*ServicesConfig](i, func(i do.Injector) (*ServicesConfig, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return &cfg.Services, nil
	})
	do.Provide[http.Config](i, func(i do.Injector) (http.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.HTTPServer, nil
	})
	do.Provide[grpc.Config](i, func(i do.Injector) (grpc.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.GRPCServer, nil
	})
	do.Provide[*httpclientcfg.Config](i, func(i do.Injector) (*httpclientcfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.HTTPClient, nil
	})
	do.Provide[notificationscfg.Config](i, func(i do.Injector) (notificationscfg.Config, error) {
		cfg := do.MustInvoke[*APIServiceConfig](i)
		return cfg.PushNotifications, nil
	})
}
