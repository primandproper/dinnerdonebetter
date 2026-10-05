package waitlists

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformwaitlists "github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "waitlists_db_client"
)

// ProvideWaitlistsRepository provides platform's waitlist store, with this
// application's recording hung off its writes.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (platformwaitlists.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	store, err := platformwaitlists.NewSQLStore(
		client,
		platformwaitlists.WithTablePrefix(branding.TablePrefix),
		platformwaitlists.WithHooks(&hooks{
			logger:   logging.NewNamedLogger(logger, o11yName),
			recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		}),
		platformwaitlists.WithStoreLogger(logger),
		platformwaitlists.WithStoreTracerProvider(tracerProvider),
		platformwaitlists.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the waitlists store")
	}

	return store, nil
}
