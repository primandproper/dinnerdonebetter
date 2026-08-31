package push

import (
	notificationsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/manager"

	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"

	"github.com/samber/do/v2"
)

// RegisterFanout registers the push Fanout with the injector.
//
// Prerequisites: the notifications data manager, a push sender, a logger and a metrics provider.
// Every process that pushes has all four already — the fanout is the part they were each about
// to write.
func RegisterFanout(i do.Injector) {
	do.Provide[*Fanout](i, func(i do.Injector) (*Fanout, error) {
		return NewFanout(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[notificationsmanager.NotificationsDataManager](i),
			do.MustInvoke[platformnotifications.PushNotificationSender](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
