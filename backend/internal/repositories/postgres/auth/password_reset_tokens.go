package auth

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

const (
	passwordResetO11yName = "password_reset_token_db_client"

	resourceTypePasswordResetTokens = "password_reset_tokens"
)

// passwordResetHooks is this repository's audit log, hung off the two password reset token
// writes worth recording.
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
// The platform calls each hook on the caller's transaction once the write has landed, so
// the token write and the entry describing it commit together, and a hook's error rolls
// the write back: a reset the log has no record of is precisely the reset an
// investigation needs. A hook is handed the stored token and never the issuance, so the
// secret that goes in the email has no route into the audit log.
//
// It implements the platform's Hooks outright rather than embedding NoopHooks, so a write
// platform adds later breaks this build until somebody decides whether it is recorded.
type passwordResetHooks struct {
	auditLogEntryRepo audit.Repository
	tracer            tracing.Tracer
	logger            logging.Logger
}

var _ passwordreset.Hooks = (*passwordResetHooks)(nil)

// ProvidePasswordResetTokenSQLStore builds the platform's password reset token store over
// this deployment's database, with no audit log around it.
//
// It is the store itself rather than the Store seam, because two of the things this
// deployment needs are not on the seam: Sweep, which the db-cleaner job runs for the fleet,
// and nothing else: the db-cleaner's container has no audit repository, so it gets the store
// with no hooks on it. Sweep calls none either way.
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
	return newPasswordResetTokenSQLStore(logger, tracerProvider, client)
}

// ProvidePasswordResetTokenStore builds the password reset token store the API server uses:
// the platform's, with this repository's audit log hung off the two writes worth recording.
func ProvidePasswordResetTokenStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) (passwordreset.Store, error) {
	return newPasswordResetTokenSQLStore(logger, tracerProvider, client, passwordreset.WithHooks(&passwordResetHooks{
		auditLogEntryRepo: auditLogEntryRepo,
		tracer:            tracing.NewNamedTracer(tracerProvider, passwordResetO11yName),
		logger:            logging.NewNamedLogger(logger, passwordResetO11yName),
	}))
}

// newPasswordResetTokenSQLStore is the one construction both providers share, so the
// audited store and the one the db-cleaner sweeps cannot disagree about the table.
func newPasswordResetTokenSQLStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	client database.Client,
	opts ...passwordreset.Option,
) (*passwordreset.SQLStore, error) {
	return passwordreset.NewSQLStore(
		&passwordreset.Config{TablePrefix: branding.TablePrefix},
		client,
		append([]passwordreset.Option{
			passwordreset.WithLogger(logging.NewNamedLogger(logger, passwordResetO11yName)),
			passwordreset.WithTracerProvider(tracerProvider),
		}, opts...)...,
	)
}

// AfterIssue records that a reset was asked for.
func (h *passwordResetHooks) AfterIssue(ctx context.Context, tx database.Tx, _ tenancy.Scope, token *passwordreset.Token) error {
	return h.record(ctx, tx, audit.NewEntry(token.UserID, "", resourceTypePasswordResetTokens, token.ID, platformaudit.EventCreated))
}

// AfterConsume records the redemption, and what it changed.
//
// Platform hands no before row, because a token Consume answers with was unredeemed a
// moment ago by construction — a redeemed one is refused before anything is written. So
// the before is the token with its redemption taken off, which is exactly what the row
// held.
func (h *passwordResetHooks) AfterConsume(ctx context.Context, tx database.Tx, _ tenancy.Scope, token *passwordreset.Token) error {
	before := *token
	before.RedeemedAt = nil

	changes, err := platformaudit.Diff(&before, token)
	if err != nil {
		return platformerrors.Wrap(err, "diffing the consumed password reset token")
	}

	entry := audit.NewEntry(token.UserID, "", resourceTypePasswordResetTokens, token.ID, platformaudit.EventUpdated)
	entry.Changes = changes

	return h.record(ctx, tx, entry)
}

// AfterRevokeForUser records nothing. It is only ever called immediately after a Consume
// this store has already recorded, in the same transaction, so its entry would say nothing
// the redemption's does not.
func (*passwordResetHooks) AfterRevokeForUser(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// AfterDeleteForUser records nothing, as the wrapper this replaced did not. It is the data
// privacy eraser's write, one table of many under a single erasure request, and an entry per
// table would name the subject in the log of the very request that asked for them to be
// forgotten.
func (*passwordResetHooks) AfterDeleteForUser(context.Context, database.Tx, tenancy.Scope, string, int64) error {
	return nil
}

// record writes one audit entry inside the transaction the token write ran in.
//
// A failure to record fails the write, and rolls it back. A reset the log has no record
// of is precisely the reset an investigation needs.
func (h *passwordResetHooks) record(ctx context.Context, tx database.Tx, entry *platformaudit.Entry) error {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	tracing.AttachToSpan(span, authkeys.PasswordResetTokenIDKey, entry.ResourceID)

	if err := h.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
		return observability.PrepareAndLogError(err, h.logger, span, "recording password reset token audit log entry")
	}

	return nil
}
