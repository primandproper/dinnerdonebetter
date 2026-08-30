package dbcleaner

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName = "db_cleaner"
)

// Job removes the authorization server's dead records.
//
// It is a garbage collector rather than a security control: every read the store performs
// already refuses an expired or revoked record, so a row this has not reached yet is unusable.
// What it stops is the code and token tables growing with every login.
//
// It runs here, as one scheduled sweep for the fleet, rather than as the store's own sweeper
// goroutine — which is why the deployed configurations leave SweepInterval at zero. A sweeper
// per replica would have every pod running the same full-table delete on its own timer.
type Job struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	handledRecordsCounter metrics.Int64Counter
	store                 oauth2server.Store
	clock                 func() time.Time
}

func NewDBCleaner(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	store oauth2server.Store,
) (*Job, error) {
	handledRecordsCounter, err := metricsProvider.NewInt64Counter("db_cleaner.handled_records")
	if err != nil {
		return nil, err
	}

	return &Job{
		logger:                logging.NewNamedLogger(logger, serviceName),
		tracer:                tracing.NewNamedTracer(tracerProvider, serviceName),
		handledRecordsCounter: handledRecordsCounter,
		store:                 store,
		clock:                 time.Now,
	}, nil
}

func (j *Job) Do(ctx context.Context) error {
	ctx, span := j.tracer.StartSpan(ctx)
	defer span.End()

	deleted, err := j.store.Sweep(ctx, j.clock())
	if err != nil {
		j.logger.Error("sweeping expired oauth2 records", err)
		return err
	}

	j.handledRecordsCounter.Add(ctx, deleted, metric.WithAttributes(
		attribute.KeyValue{
			Key:   "db_table",
			Value: attribute.StringValue("oauth2"),
		},
	))

	j.logger.WithValue("swept", deleted).Info("swept expired oauth2 records")

	return nil
}
