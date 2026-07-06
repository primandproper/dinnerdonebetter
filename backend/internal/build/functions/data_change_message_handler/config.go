package datachangemessagehandler

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	analyticscfg "github.com/primandproper/platform-go/v3/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v3/database/config"
	emailcfg "github.com/primandproper/platform-go/v3/email/config"
	"github.com/primandproper/platform-go/v3/encoding"
	httpclientcfg "github.com/primandproper/platform-go/v3/httpclient"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v3/notifications/mobile/config"
	"github.com/primandproper/platform-go/v3/observability"
	textsearchcfg "github.com/primandproper/platform-go/v3/search/text/config"
	"github.com/primandproper/platform-go/v3/uploads/objectstorage"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*objectstorage.Config](i, func(i do.Injector) (*objectstorage.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Storage, nil
	})
	do.Provide[*msgconfig.QueuesConfig](i, func(i do.Injector) (*msgconfig.QueuesConfig, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Queues, nil
	})
	do.Provide[*emailcfg.Config](i, func(i do.Injector) (*emailcfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Email, nil
	})
	do.Provide[*httpclientcfg.Config](i, func(i do.Injector) (*httpclientcfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return cfg.HTTPClient, nil
	})
	do.Provide[*analyticscfg.Config](i, func(i do.Injector) (*analyticscfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Analytics, nil
	})
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Search, nil
	})
	do.Provide[*msgconfig.Config](i, func(i do.Injector) (*msgconfig.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Events, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[encoding.Config](i, func(i do.Injector) (encoding.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return cfg.Encoding, nil
	})
	do.Provide[notificationscfg.Config](i, func(i do.Injector) (notificationscfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return cfg.PushNotifications, nil
	})
}
