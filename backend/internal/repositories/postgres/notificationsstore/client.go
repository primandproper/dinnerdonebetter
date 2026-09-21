/*
Package notificationsstore is platform-go's notification store with this
application's recording around it.

It replaces internal/repositories/postgres/notifications, which owned two tables
of its own. The tables are platform's now — see renderNotificationsDDL — and what
is left here is the part that was never platform's: the audit entry and the data
change event every write owes, written as further statements in the transaction
that performed it.

The reads are embedded rather than forwarded. Eight of the fourteen methods add
nothing, and eight forwarding stubs are eight chances to drift from the store
they forward to.
*/
package notificationsstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	notificationscfg "github.com/primandproper/platform-go/v14/notifications/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "notifications_db_client"

	// resourceTypeUserNotifications is what an audit entry about an inbox row names.
	resourceTypeUserNotifications = "user_notifications"
	// resourceTypeUserDeviceTokens is what an audit entry about a handset names.
	resourceTypeUserDeviceTokens = "user_device_tokens"
)

// inbox is platform's inbox with this application's recording around its writes.
type inbox struct {
	platformnotifications.Inbox

	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

// registry is platform's device registry with the same around its writes.
type registry struct {
	platformnotifications.Registry

	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

// ProvideInbox wraps a platform inbox in this application's recording.
func ProvideInbox(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	_ metrics.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	store platformnotifications.Inbox,
) (platformnotifications.Inbox, error) {
	if store == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil notification inbox")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &inbox{
		Inbox:    store,
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}

// ProvideRegistry wraps a platform device registry in this application's recording.
func ProvideRegistry(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	_ metrics.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	store platformnotifications.Registry,
) (platformnotifications.Registry, error) {
	if store == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil device registry")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &registry{
		Registry: store,
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}

// ProvideAdapter builds the whole stack for a caller outside the injector: the
// platform store at this application's prefix, the recording around its writes,
// and the adapter presenting it as notifications.Repository.
//
// It exists for localdev and the integration harness, which build repositories
// directly rather than through samber/do. Keeping it here rather than repeating
// the three steps at each of them is what stops one of them wiring the
// undecorated store and losing the audit entries.
func ProvideAdapter(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
	db database.Client,
) (*Adapter, error) {
	store, err := notificationscfg.NewStore(ctx, &notificationscfg.Config{}, db,
		notificationscfg.WithLogger(logger),
		notificationscfg.WithTracerProvider(tracerProvider),
		notificationscfg.WithStoreOptions(
			platformnotifications.WithTablePrefix(ddbnotifications.TablePrefix),
		),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the notifications store")
	}

	decoratedInbox, err := ProvideInbox(logger, tracerProvider, metricsProvider, auditLogEntryRepo, eventEmitter, store)
	if err != nil {
		return nil, err
	}

	decoratedRegistry, err := ProvideRegistry(logger, tracerProvider, metricsProvider, auditLogEntryRepo, eventEmitter, store)
	if err != nil {
		return nil, err
	}

	return NewAdapter(decoratedInbox, decoratedRegistry, db)
}
