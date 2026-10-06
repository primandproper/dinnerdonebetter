package notificationsstore

import (
	"context"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsStore registers the inbox and the device registry with
// the injector. Both resolve to one store.
func RegisterNotificationsStore(i do.Injector) {
	do.Provide[platformnotifications.Inbox](i, func(i do.Injector) (platformnotifications.Inbox, error) {
		return newStore(i)
	})

	do.Provide[platformnotifications.Registry](i, func(i do.Injector) (platformnotifications.Registry, error) {
		return newStore(i)
	})
}

func newStore(i do.Injector) (platformnotifications.Store, error) {
	return provideStore(
		do.MustInvoke[context.Context](i),
		do.MustInvoke[logging.Logger](i),
		do.MustInvoke[tracing.Provider](i),
		do.MustInvoke[metrics.Provider](i),
		do.MustInvoke[*platformrecording.Recorder](i),
		do.MustInvoke[database.Client](i),
	)
}
