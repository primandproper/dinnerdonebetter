package issuereports

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	platformissuereports "github.com/primandproper/platform-go/v13/issuereports"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "issue_reports_db_client"
)

// repository is platform's issue report store with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the seven
// methods this package adds nothing to — every read, and the erasure delete —
// are the platform's own rather than seven forwarding stubs that could drift
// from it.
type repository struct {
	platformissuereports.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// ProvideIssueReportsRepository provides a new issue report store.
func ProvideIssueReportsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (platformissuereports.Store, error) {
	store, err := platformissuereports.NewSQLStore(
		client,
		platformissuereports.WithTablePrefix(ddbissuereports.TablePrefix),
		platformissuereports.WithStoreLogger(logger),
		platformissuereports.WithStoreTracerProvider(tracerProvider),
		platformissuereports.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the issue reports store")
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
