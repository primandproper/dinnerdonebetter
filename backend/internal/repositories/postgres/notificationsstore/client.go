/*
Package notificationsstore is platform-go's notification store, assembled for
this application.

It replaces internal/repositories/postgres/notifications, which owned two tables
of its own. The tables are platform's now — see renderNotificationsDDL — and so is
the recording: notifications.RecordingHooks writes the audit entry and the event
every inbox and registry write owes, on the write's transaction, through the
recording.Recorder this process built. What this package still decides is the
table prefix, which has to match the one the migration was rendered with, and the
observability names.
*/
package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformnotifications "github.com/primandproper/platform-go/v15/notifications"
	notificationscfg "github.com/primandproper/platform-go/v15/notifications/config"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const o11yName = "notifications_db_client"

// ProvideStores builds the inbox and the device registry, which are the same
// store answering two interfaces: the consumers that write notifications and
// the consumers that register handsets take the half they use.
func ProvideStores(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	recorder *platformrecording.Recorder,
	db database.Client,
) (platformnotifications.Inbox, platformnotifications.Registry, error) {
	store, err := provideStore(ctx, logger, tracerProvider, metricsProvider, recorder, db)
	if err != nil {
		return nil, nil, err
	}

	return store, store, nil
}

func provideStore(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	recorder *platformrecording.Recorder,
	db database.Client,
) (platformnotifications.Store, error) {
	hooks, err := platformnotifications.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the notifications store's recording hooks")
	}

	store, err := notificationscfg.NewStore(ctx, &notificationscfg.Config{}, db,
		notificationscfg.WithLogger(logging.NewNamedLogger(logger, o11yName)),
		notificationscfg.WithTracerProvider(tracerProvider),
		notificationscfg.WithMetricsProvider(metricsProvider),
		notificationscfg.WithStoreOptions(
			platformnotifications.WithTablePrefix(branding.TablePrefix),
			platformnotifications.WithHooks(hooks),
		),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the notifications store")
	}

	return store, nil
}
