package comments

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "comments_db_client"
)

// repository is platform's comment store with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the seven
// methods this package adds nothing to — every read, and both bulk deletes —
// are the platform's own rather than seven forwarding stubs that could drift
// from it.
type repository struct {
	platformcomments.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// ProvideCommentsRepository provides a new comment store.
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
	store, err := platformcomments.NewSQLStore(
		client,
		platformcomments.WithTablePrefix(ddbcomments.TablePrefix),
		platformcomments.WithTargets(targets),
		platformcomments.WithStoreLogger(logger),
		platformcomments.WithStoreTracerProvider(tracerProvider),
		platformcomments.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the comments store")
	}

	return &repository{
		Store:             store,
		client:            client,
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
	}, nil
}
