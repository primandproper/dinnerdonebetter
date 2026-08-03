package datachangemessagehandler

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	dataprivacycfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/config"

	analyticscfg "github.com/primandproper/platform-go/v9/analytics/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	emailcfg "github.com/primandproper/platform-go/v9/email/config"
	"github.com/primandproper/platform-go/v9/encoding"
	httpclientcfg "github.com/primandproper/platform-go/v9/httpclient"
	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	notificationscfg "github.com/primandproper/platform-go/v9/notifications/mobile/config"
	"github.com/primandproper/platform-go/v9/observability"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*dataprivacycfg.Config](i, func(i do.Injector) (*dataprivacycfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.DataPrivacy, nil
	})
	do.Provide[*queuescfg.Config](i, func(i do.Injector) (*queuescfg.Config, error) {
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
	do.Provide[*dbcfg.Config](i, func(i do.Injector) (*dbcfg.Config, error) {
		cfg := do.MustInvoke[*config.AsyncMessageHandlerConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		return &do.MustInvoke[*dbcfg.Config](i).Config, nil
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
