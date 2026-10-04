package issuereports

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformissuereports "github.com/primandproper/platform-go/v14/issuereports"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "issue_reports_db_client"
)

// ProvideIssueReportsRepository provides platform's issue report store, with
// this application's recording hung off its writes.
func ProvideIssueReportsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (platformissuereports.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	store, err := platformissuereports.NewSQLStore(
		client,
		platformissuereports.WithTablePrefix(branding.TablePrefix),
		platformissuereports.WithHooks(&hooks{
			logger:   logging.NewNamedLogger(logger, o11yName),
			recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		}),
		platformissuereports.WithStoreLogger(logger),
		platformissuereports.WithStoreTracerProvider(tracerProvider),
		platformissuereports.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the issue reports store")
	}

	return store, nil
}
