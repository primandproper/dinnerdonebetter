package dbcleaner

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/platform-go/v14/authentication/signin/refreshtokens"
	sessionsdatabase "github.com/primandproper/platform-go/v14/sessions/database"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	serviceName = "db_cleaner"
)

// Job removes the dead records of the four authentication stores that keep expiring rows:
// the authorization server's codes and tokens, the password reset tokens, the user sessions,
// and the refresh tokens platform's sign-in rotates.
//
// It is a garbage collector rather than a security control. Every read any of the four
// performs already refuses an expired, revoked, or redeemed record — a session past either
// of its deadlines is refused by the store's policy rather than by the row's absence — so a
// row this has not reached yet is unusable. What it stops is those tables growing: with
// every login on two of them, with every refresh on the fourth, and with every password
// anybody ever forgot on the third, including the requests nobody followed up, which are the
// ones no redemption ever removes.
//
// It runs here, as one scheduled sweep for the fleet, rather than as any store's own sweeper
// goroutine: a sweeper per replica would have every pod running the same full-table delete on
// its own timer.
//
// Three of the four stores are built without WithSweeper, so that intent takes effect. The
// authorization server's is not, and cannot be from where it is configured: its config
// documents a non-positive SweepInterval as "no sweeper", but EnsureDefaults rewrites the
// zero the deployed configurations set to ten minutes before it reaches WithSweeper — so
// that store sweeps per replica as well as here, and has since it was adopted. See
// platform-go#456. It is waste rather than a fault, since the second sweep finds what the
// first one left, which is nothing; the fix belongs upstream, and a local one would spell
// "off" as a negative duration nothing documents.
//
// Every sweep runs even if an earlier one fails. They are independent tables, and a
// misconfiguration on one is not a reason to let the others grow; the errors are joined so
// the job still reports the failure.
type Job struct {
	logger                logging.Logger
	tracer                tracing.Tracer
	handledRecordsCounter metrics.Int64Counter
	oauth2Store           oauth2server.Store
	passwordResetStore    *passwordreset.SQLStore
	refreshTokenStore     *refreshtokens.SQLStore
	sessionBackend        *sessionsdatabase.Backend[auth.SessionPayload]
}

func NewDBCleaner(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	oauth2Store oauth2server.Store,
	passwordResetStore *passwordreset.SQLStore,
	refreshTokenStore *refreshtokens.SQLStore,
	sessionBackend *sessionsdatabase.Backend[auth.SessionPayload],
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
		refreshTokenStore:     refreshTokenStore,
		sessionBackend:        sessionBackend,
	}, nil
}

func (j *Job) Do(ctx context.Context) error {
	ctx, span := j.tracer.StartSpan(ctx)
	defer span.End()

	return errors.Join(
		j.sweepOAuth2(ctx),
		j.sweepPasswordResetTokens(ctx),
		j.sweepRefreshTokens(ctx),
		j.sweepUserSessions(ctx),
	)
}

func (j *Job) sweepOAuth2(ctx context.Context) error {
	// v14 dropped the caller-supplied horizon: the store sweeps against the same
	// clock every deadline in its schema was stamped from and every read refuses
	// against, so a sweep reclaims exactly the rows it has already stopped
	// answering with. A third opinion from here could reap a row a request had
	// just read.
	deleted, err := j.oauth2Store.Sweep(ctx)
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

// sweepRefreshTokens deletes refresh token rows past their purge deadline.
//
// A refresh token is not deleted when it expires or is spent. It is kept for a retention
// window past that, because a spent token presented again is how the store detects a stolen
// one, and it can only recognize a token it still has a row for. purge_after is the end of
// that window, and it is the store's own clock that decides it, as with the two above.
func (j *Job) sweepRefreshTokens(ctx context.Context) error {
	deleted, err := j.refreshTokenStore.Sweep(ctx)
	if err != nil {
		j.logger.Error("sweeping purgeable refresh tokens", err)
		return err
	}

	j.recordSwept(ctx, "signin_refresh_tokens", deleted)
	j.logger.WithValue("swept", deleted).Info("swept purgeable refresh tokens")

	return nil
}

// sweepUserSessions deletes session rows whose deadlines have passed.
//
// Like the reset tokens above it takes no clock argument: the backend compares against its
// own, which is the clock expires_at was stamped from. Asking the database server for the
// time instead would put two clocks on the two sides of one comparison.
//
// A revoked session never reaches this. Revocation removes the row, which is the whole
// reason the platform's table has no revoked_at column for a sweep to have to interpret.
func (j *Job) sweepUserSessions(ctx context.Context) error {
	deleted, err := j.sessionBackend.Sweep(ctx)
	if err != nil {
		j.logger.Error("sweeping expired user sessions", err)
		return err
	}

	j.recordSwept(ctx, "sessions", deleted)
	j.logger.WithValue("swept", deleted).Info("swept expired user sessions")

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
