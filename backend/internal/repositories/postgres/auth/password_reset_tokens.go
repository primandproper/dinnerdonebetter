package auth

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const passwordResetO11yName = "password_reset_token_db_client"

// The password reset token store is platform's, whole, and so is its recording.
//
// The store keeps the secret as a digest, makes single use the affected-row count of a guarded
// UPDATE inside one transaction rather than a decision the caller makes on a read, and refuses
// an expired row whether or not anything has swept it. None of that is worth reimplementing
// here, and the version that used to live in this file got the second of the three wrong.
//
// What it records is platform's too: a reset issued and a reset spent are both facts an
// investigation asks for months later — "was a link issued for this account before the
// takeover, and was it used?" — and the row itself cannot answer the first, because the sweeper
// deletes it once it expires. passwordreset.RecordingHooks writes each as an audit entry filed
// under the user and an event on the outbox, on the caller's transaction once the write has
// landed, so the token write and the entry describing it commit together and a hook's error
// rolls the write back. A hook is handed the stored token and never the issuance, so the secret
// that goes in the email has no route into the audit log or onto the broker.
//
// What is this application's is the table prefix, and which processes get which store.

// ProvidePasswordResetTokenSQLStore builds the platform's password reset token store over
// this deployment's database, recording nothing.
//
// It is the store itself rather than the Store seam, because the one thing this deployment
// needs that is not on the seam is Sweep, which the db-cleaner job runs for the fleet — and
// that job's container has no recorder, so it gets the store with no hooks on it. Sweep
// calls none either way.
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
// the platform's, with platform's recording of the two writes worth recording.
func ProvidePasswordResetTokenStore(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	recorder *platformrecording.Recorder,
	client database.Client,
) (passwordreset.Store, error) {
	hooks, err := passwordreset.NewRecordingHooks(recorder)
	if err != nil {
		return nil, err
	}

	return newPasswordResetTokenSQLStore(logger, tracerProvider, client, passwordreset.WithHooks(hooks))
}

// newPasswordResetTokenSQLStore is the one construction both providers share, so the
// recorded store and the one the db-cleaner sweeps cannot disagree about the table.
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
