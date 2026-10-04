package comments

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformcomments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "comments_db_client"
)

// ProvideCommentsRepository provides platform's comment store, with this
// application's recording hung off its writes.
//
// The target catalog is a parameter because it is the one thing platform refuses
// to guess at, and a store built without one accepts no writes at all. It is
// assembled in internal/build/comments, where the domains that own the things
// being commented on are already in hand.
func ProvideCommentsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
	targets platformcomments.Targets,
) (platformcomments.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	store, err := platformcomments.NewSQLStore(
		client,
		&hooks{
			logger:   logging.NewNamedLogger(logger, o11yName),
			recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		},
		platformcomments.WithTablePrefix(branding.TablePrefix),
		platformcomments.WithTargets(targets),
		platformcomments.WithStoreLogger(logger),
		platformcomments.WithStoreTracerProvider(tracerProvider),
		platformcomments.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the comments store")
	}

	return store, nil
}
