package datachangemessagehandler

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	notificationscfg "github.com/primandproper/primitives-go/v2/notifications/mobile/config"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers the config sub-fields this application's own registrations read.
//
// Everything platform builds reads its block from the injector already: service.Register put each
// one there beside the registration that consumes it.
func RegisterConfigs(i do.Injector) {
	do.Provide[*queuescfg.Config](i, func(i do.Injector) (*queuescfg.Config, error) {
		return &do.MustInvoke[*config.AsyncMessageHandlerConfig](i).Queues, nil
	})
	do.Provide[*textsearchcfg.Config](i, func(i do.Injector) (*textsearchcfg.Config, error) {
		return &do.MustInvoke[*config.AsyncMessageHandlerConfig](i).Search, nil
	})
	// A pointer, which is what RegisterPushSender resolves: NewPushSender applies its
	// defaults to what it is handed, and a value copy would discard them.
	do.Provide[*notificationscfg.Config](i, func(i do.Injector) (*notificationscfg.Config, error) {
		return &do.MustInvoke[*config.AsyncMessageHandlerConfig](i).PushNotifications, nil
	})
}
