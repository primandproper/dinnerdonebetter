/*
Package notificationsstore is platform-go's notification store with this
application's recording hung off its writes.

It replaces internal/repositories/postgres/notifications, which owned two tables
of its own. The tables are platform's now — see renderNotificationsDDL — and what
is left here is the part that was never platform's: the audit entry and the data
change event every write owes.

Both are written from platform's notifications.Hooks, which the store calls on
the caller's transaction once each write has landed. A hook's error fails the
write, so the row, the entry and the event commit together or not at all.
*/
package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	notificationscfg "github.com/primandproper/platform-go/v15/notifications/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const o11yName = "notifications_db_client"

// ProvideStores builds the whole stack for a caller outside the injector: the
// platform store at this application's prefix, with the recording hung off the
// writes of both of its halves.
//
// It exists for localdev and the integration harness, which build repositories
// directly rather than through samber/do. Keeping it here rather than repeating
// the steps at each of them is what stops one of them wiring a store without
// hooks and losing the audit entries.
func ProvideStores(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	db database.Client,
) (platformnotifications.Inbox, platformnotifications.Registry, error) {
	store, err := provideStore(ctx, logger, tracerProvider, metricsProvider, auditLogEntryRepo, eventEmitter, db)
	if err != nil {
		return nil, nil, err
	}

	return store, store, nil
}

// provideStore builds the platform store at this application's table prefix,
// with this application's recording as its hooks.
func provideStore(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	db database.Client,
) (platformnotifications.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	store, err := notificationscfg.NewStore(ctx, &notificationscfg.Config{}, db,
		notificationscfg.WithLogger(logger),
		notificationscfg.WithTracerProvider(tracerProvider),
		notificationscfg.WithMetricsProvider(metricsProvider),
		notificationscfg.WithStoreOptions(
			platformnotifications.WithTablePrefix(branding.TablePrefix),
			platformnotifications.WithHooks(&hooks{
				logger:   logging.NewNamedLogger(logger, o11yName),
				recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
			}),
		),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the notifications store")
	}

	return store, nil
}
