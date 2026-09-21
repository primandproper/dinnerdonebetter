package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	notificationscfg "github.com/primandproper/platform-go/v14/notifications/config"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterNotificationsStore registers the inbox and the device registry, each
// wrapped in this application's recording.
//
// The undecorated store is not registered under its own type. Everything here
// that writes a notification should write an audit entry with it, and a second
// registration would be a way to skip that by naming the wrong dependency.
func RegisterNotificationsStore(i do.Injector) {
	do.Provide[platformnotifications.Inbox](i, func(i do.Injector) (platformnotifications.Inbox, error) {
		store, err := newStore(i)
		if err != nil {
			return nil, err
		}

		return ProvideInbox(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*events.Emitter](i),
			store,
		)
	})

	do.Provide[platformnotifications.Registry](i, func(i do.Injector) (platformnotifications.Registry, error) {
		store, err := newStore(i)
		if err != nil {
			return nil, err
		}

		return ProvideRegistry(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[*events.Emitter](i),
			store,
		)
	})
}

// newStore builds the platform store at this application's table prefix.
//
// Built twice rather than shared, because the two halves are registered
// separately and a store holds no state worth sharing: it is a querier and a
// dialect, and both are settled from the same config either way.
func newStore(i do.Injector) (platformnotifications.Store, error) {
	return notificationscfg.NewStore(
		do.MustInvoke[context.Context](i),
		&notificationscfg.Config{},
		do.MustInvoke[database.Client](i),
		notificationscfg.WithLogger(do.MustInvoke[logging.Logger](i)),
		notificationscfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		notificationscfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		notificationscfg.WithStoreOptions(
			platformnotifications.WithTablePrefix(ddbnotifications.TablePrefix),
		),
	)
}

// RegisterNotificationsRepository registers the adapter that presents platform's
// store in this application's vocabulary. See Adapter.
//
// It registers under the concrete *Adapter rather than under
// notifications.Repository, which is the same shape the hand-written repository
// it replaced had: notifications.Repository resolves to the manager, which wraps
// whatever is registered here. Registering the adapter under the interface as
// well would be a second provider for one type, and samber/do refuses that — a
// panic at container build, so every binary that has a notifications chain.
func RegisterNotificationsRepository(i do.Injector) {
	do.Provide[*Adapter](i, func(i do.Injector) (*Adapter, error) {
		return NewAdapter(
			do.MustInvoke[platformnotifications.Inbox](i),
			do.MustInvoke[platformnotifications.Registry](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
