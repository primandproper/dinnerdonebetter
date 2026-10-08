package localdev

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/recordingspine"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"

	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// Spine builds the recording spine a process assembled by hand records and publishes through:
// the local dev server's seeders, the integration suites' fixtures, the one-shot tools.
//
// It exists because every such process used to build its repositories with a nil emitter, which
// the old emitter accepted and answered by announcing nothing — a seed or an import that wrote
// rows no index and no subscriber ever heard about. The spine is required now, and a process with
// no broker still writes its events to the outbox for the worker that has one.
func Spine(
	ctx context.Context,
	databaseClient database.Client,
	auditLog *auditlogentries.Log,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) (*recordingspine.Spine, error) {
	return recordingspine.New(ctx, databaseClient, auditLog.Recorder(), recordingspine.WithPillars(&observability.Pillars{
		Logger:          logger,
		TracerProvider:  tracerProvider,
		MetricsProvider: metricsnoop.NewMetricsProvider(),
	}))
}
