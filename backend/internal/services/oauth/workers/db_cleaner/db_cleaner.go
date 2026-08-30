package dbcleaner

import (
	"context"
	"errors"
	"time"

	"github.com/primandproper/platform-go/v13/authentication/oauth2server"
	"github.com/primandproper/platform-go/v13/authentication/passwordreset"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName = "db_cleaner"
)

// Job removes the dead records of the two authentication stores that keep expiring rows: the
// authorization server's codes and tokens, and the password reset tokens.
//
// It is a garbage collector rather than a security control. Every read either store performs
// already refuses an expired, revoked, or redeemed record, so a row this has not reached yet
// is unusable. What it stops is those tables growing — with every login on one side, and with
// every password anybody ever forgot on the other, including the requests nobody followed up,
// which are the ones no redemption ever removes.
//
// It runs here, as one scheduled sweep for the fleet, rather than as either store's own
// sweeper goroutine — which is why the deployed configurations leave SweepInterval at zero and
// why the password reset store is built without WithSweeper. A sweeper per replica would have
// every pod running the same full-table delete on its own timer.
//
// Both sweeps run even if the first one fails. They are independent tables, and a
// misconfiguration on one is not a reason to let the other grow; the errors are joined so the
// job still reports the failure.
type Job struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	handledRecordsCounter metrics.Int64Counter
	oauth2Store           oauth2server.Store
	passwordResetStore    *passwordreset.SQLStore
	clock                 func() time.Time
}

func NewDBCleaner(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	oauth2Store oauth2server.Store,
	passwordResetStore *passwordreset.SQLStore,
) (*Job, error) {
	handledRecordsCounter, err := metricsProvider.NewInt64Counter("db_cleaner.handled_records")
	if err != nil {
		return nil, err
	}

	return &Job{
		logger:                logging.NewNamedLogger(logger, serviceName),
		tracer:                tracing.NewNamedTracer(tracerProvider, serviceName),
		handledRecordsCounter: handledRecordsCounter,
		oauth2Store:           oauth2Store,
		passwordResetStore:    passwordResetStore,
		clock:                 time.Now,
	}, nil
}

func (j *Job) Do(ctx context.Context) error {
	ctx, span := j.tracer.StartSpan(ctx)
	defer span.End()

	return errors.Join(j.sweepOAuth2(ctx), j.sweepPasswordResetTokens(ctx))
}

func (j *Job) sweepOAuth2(ctx context.Context) error {
	deleted, err := j.oauth2Store.Sweep(ctx, j.clock())
	if err != nil {
		j.logger.Error("sweeping expired oauth2 records", err)
		return err
	}

	j.recordSwept(ctx, "oauth2", deleted)
	j.logger.WithValue("swept", deleted).Info("swept expired oauth2 records")

	return nil
}

// sweepPasswordResetTokens deletes reset rows whose deadlines have passed.
//
// It takes no clock argument, unlike the sweep above: the platform store compares against its
// own, in UTC, so that the boundary a user hits at the last second of a link's life is decided
// in one place rather than by whoever calls the sweep.
func (j *Job) sweepPasswordResetTokens(ctx context.Context) error {
	deleted, err := j.passwordResetStore.Sweep(ctx)
	if err != nil {
		j.logger.Error("sweeping expired password reset tokens", err)
		return err
	}

	j.recordSwept(ctx, "password_reset_tokens", deleted)
	j.logger.WithValue("swept", deleted).Info("swept expired password reset tokens")

	return nil
}

func (j *Job) recordSwept(ctx context.Context, table string, deleted int64) {
	j.handledRecordsCounter.Add(ctx, deleted, metric.WithAttributes(
		attribute.KeyValue{
			Key:   "db_table",
			Value: attribute.StringValue(table),
		},
	))
}
