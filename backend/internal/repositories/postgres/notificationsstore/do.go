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
// store as this application's notifications.Repository.
//
// It is what the manager, the push fanout and the privacy collector resolve, so
// that none of them changed when the store underneath them did. See Adapter.
func RegisterNotificationsRepository(i do.Injector) {
	do.Provide[ddbnotifications.Repository](i, func(i do.Injector) (ddbnotifications.Repository, error) {
		return NewAdapter(
			do.MustInvoke[platformnotifications.Inbox](i),
			do.MustInvoke[platformnotifications.Registry](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
