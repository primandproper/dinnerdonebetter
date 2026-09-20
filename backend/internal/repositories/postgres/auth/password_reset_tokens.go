package auth

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	passwordResetO11yName = "password_reset_token_db_client"

	resourceTypePasswordResetTokens = "password_reset_tokens"
)

var _ passwordreset.Store = (*auditedPasswordResetTokenStore)(nil)

// auditedPasswordResetTokenStore is platform-go's password reset token store with this
// repository's audit log wrapped around the two events worth recording.
//
// The store itself is the platform's, whole: the secret is stored as a digest, single use
// is the affected-row count of a guarded UPDATE inside one transaction rather than a
// decision the caller makes on a read, and an expired row is refused whether or not
// anything has swept it. None of that is worth reimplementing here, and the version that
// used to live in this file got the second of the three wrong.
//
// What the platform has no opinion about is who wants a record of it. A reset issued and
// a reset spent are both facts an investigation asks for months later — "was a link
// issued for this account before the takeover, and was it used?" — and the row itself
// cannot answer the first, because the sweeper deletes it once it expires.
//
// Verify and RevokeForUser are the embedded store's, unrecorded, and each for its own
// reason. Verify spends nothing and happens on every page load of the reset form, so an
// entry per call would bury the two that matter. RevokeForUser is only ever called
// immediately after a Consume this store has already recorded, so its entry would say
// nothing the redemption's does not.
type auditedPasswordResetTokenStore struct {
	passwordreset.Store

	db                database.Client
	auditLogEntryRepo audit.Repository
	tracer            tracing.Tracer
	logger            logging.Logger
}

// ProvidePasswordResetTokenSQLStore builds the platform's password reset token store over
// this deployment's database, with no audit log around it.
//
// It is the store itself rather than the Store seam, because two of the things this
// deployment needs are not on the seam: Sweep, which the db-cleaner job runs for the fleet,
// and the concrete type the audited wrapper below is built from.
//
// No sweeper goroutine is started. That is the same call the authorization server's tables
// make — one scheduled sweep for the deployment rather than one per replica, each running
// the same full-table delete on its own timer. See services/oauth/workers/db_cleaner.
//
// Reads and writes go through the write pool, which is the platform store's choice rather
// than one made here: a reset row is written by the request that asks for a link and read
// by the request that follows it seconds later, and replica lag turns that into a link
// that is "not found" and then works on reload.
func ProvidePasswordResetTokenSQLStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	client database.Client,
) (*passwordreset.SQLStore, error) {
	return passwordreset.NewSQLStore(
		&passwordreset.Config{TablePrefix: auth.TablePrefix},
		client,
		passwordreset.WithLogger(logging.NewNamedLogger(logger, passwordResetO11yName)),
		passwordreset.WithTracerProvider(tracerProvider),
	)
}

// ProvidePasswordResetTokenStore builds the password reset token store the API server uses:
// the platform's, with this repository's audit log around the two events worth recording.
func ProvidePasswordResetTokenStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) (passwordreset.Store, error) {
	store, err := ProvidePasswordResetTokenSQLStore(logger, tracerProvider, client)
	if err != nil {
		return nil, err
	}

	return &auditedPasswordResetTokenStore{
		Store:             store,
		db:                client,
		auditLogEntryRepo: auditLogEntryRepo,
		tracer:            tracing.NewNamedTracer(tracerProvider, passwordResetO11yName),
		logger:            logging.NewNamedLogger(logger, passwordResetO11yName),
	}, nil
}

// Issue mints a token and records that a reset was asked for.
func (s *auditedPasswordResetTokenStore) Issue(ctx context.Context, tx database.Tx, scope tenancy.Scope, userID string, ttl time.Duration) (*passwordreset.Issuance, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	issuance, err := s.Store.Issue(ctx, tx, scope, userID, ttl)
	if err != nil {
		return nil, err
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, issuance.Token.ID)

	if err = s.record(ctx, tx, span, issuance.Token, audit.AuditLogEventTypeCreated); err != nil {
		return nil, err
	}

	return issuance, nil
}

// Consume spends a token and records the redemption.
func (s *auditedPasswordResetTokenStore) Consume(ctx context.Context, tx database.Tx, scope tenancy.Scope, secret string) (*passwordreset.Token, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	token, err := s.Store.Consume(ctx, tx, scope, secret)
	if err != nil {
		return nil, err
	}
	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, token.ID)

	if err = s.record(ctx, tx, span, token, audit.AuditLogEventTypeUpdated); err != nil {
		return nil, err
	}

	return token, nil
}

// record writes one audit entry for a token, inside the caller's transaction.
//
// It used to open one of its own, and had to: the write it describes committed inside the
// platform store — a redemption was one transaction there by design — and nothing outside
// that package could join it. The pair was therefore not atomic, and the gap had a
// direction: a crash between them lost the entry, never the redemption.
//
// platform-go v14 removed the reason. Issue and Consume take the caller's executor, so
// the token write and the entry describing it are now one transaction, and an entry
// cannot be lost by a crash that keeps the redemption. The gap is closed rather than
// merely pointed the right way.
//
// A failure to record still fails the operation, and now rolls it back. A reset the log
// has no record of is precisely the reset an investigation needs.
func (s *auditedPasswordResetTokenStore) record(ctx context.Context, tx database.Tx, span tracing.Span, token *passwordreset.Token, eventType string) error {
	if err := s.auditLogEntryRepo.Record(ctx, tx, &audit.AuditLogEntry{
		ResourceType:  resourceTypePasswordResetTokens,
		RelevantID:    token.ID,
		EventType:     eventType,
		BelongsToUser: token.UserID,
	}); err != nil {
		return observability.PrepareAndLogError(err, s.logger, span, "recording password reset token audit log entry")
	}

	return nil
}
