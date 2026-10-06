package localdev

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// errNoPlatformRecorder is an audit repository this package did not build: platform's recorder
// is behind the one auditlogentries builds, and nothing else can stand in for it.
var errNoPlatformRecorder = errors.New("the audit log repository exposes no platform recorder")

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
	auditRepo audit.Repository,
	logger logging.Logger,
	tracerProvider tracing.Provider,
) (*events.Emitter, error) {
	auditRecorder, ok := auditlogentries.RecorderFrom(auditRepo)
	if !ok {
		return nil, errNoPlatformRecorder
	}

	return events.New(ctx, databaseClient, auditRecorder, events.WithPillars(&observability.Pillars{
		Logger:          logger,
		TracerProvider:  tracerProvider,
		MetricsProvider: metricsnoop.NewMetricsProvider(),
	}))
}
