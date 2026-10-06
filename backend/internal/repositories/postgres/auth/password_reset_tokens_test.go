package auth

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/primitives-go/v2/database"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleTokenLifetime = 30 * time.Minute

func TestQuerier_Integration_PasswordResetTokens(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	auditRecorder, ok := auditlogentries.RecorderFrom(auditRepo)
	require.True(t, ok)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())

	store, err := ProvidePasswordResetTokenStore(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(),
		pgtesting.NewRecorderForTest(t, ctx, dbc, auditRecorder), dbc)
	require.NoError(t, err)

	// issue
	issuance, err := issueT(ctx, dbc, store, user.ID, exampleTokenLifetime)
	require.NoError(t, err)
	require.NotNil(t, issuance)
	assert.NotEmpty(t, issuance.Secret)
	assert.Equal(t, user.ID, issuance.Token.UserID)
	assert.Nil(t, issuance.Token.RedeemedAt)

	// A reset is asked for by somebody who cannot sign in, so there is no principal on the
	// context and the entry names nobody as its actor. It is filed on the user's own chain all
	// the same — platform's hooks name the token's user as the entry's subject — which is the
	// chain "was a link issued for this account" is answered from.
	pgtesting.AssertAuditLogContains(t, ctx, dbc, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: passwordreset.ResourceTypeToken, ResourceID: issuance.Token.ID},
	})

	// the row holds a digest, not the token. This is the property the hand-written store
	// this replaced did not have, and the reason a database copy is no longer a password
	// reset for every account with an outstanding link.
	var stored string
	require.NoError(t, dbc.Reader().QueryRowContext(ctx,
		`SELECT token_digest FROM ddb_password_reset_tokens WHERE id = $1`, issuance.Token.ID).Scan(&stored))
	assert.NotEqual(t, issuance.Secret, stored)

	// verify does not spend it
	verified, err := store.Verify(ctx, dbc.Reader(), tenancy.Global(), issuance.Secret)
	require.NoError(t, err)
	assert.Equal(t, issuance.Token.ID, verified.ID)
	assert.Nil(t, verified.RedeemedAt)

	// consume
	consumed, err := consumeT(ctx, dbc, store, issuance.Secret)
	require.NoError(t, err)
	assert.Equal(t, issuance.Token.ID, consumed.ID)
	assert.NotNil(t, consumed.RedeemedAt)

	pgtesting.AssertAuditLogContains(t, ctx, dbc, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: passwordreset.ResourceTypeToken, ResourceID: issuance.Token.ID},
		{EventType: platformaudit.EventUpdated, ResourceType: passwordreset.ResourceTypeToken, ResourceID: issuance.Token.ID},
	})

	// a token is spendable exactly once, and the store is what says so
	_, err = consumeT(ctx, dbc, store, issuance.Secret)
	require.ErrorIs(t, err, passwordreset.ErrTokenRedeemed)

	// revoking takes the outstanding links with it
	second, err := issueT(ctx, dbc, store, user.ID, exampleTokenLifetime)
	require.NoError(t, err)

	revoked, err := revokeForUserT(ctx, dbc, store, user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), revoked)

	_, err = store.Verify(ctx, dbc.Reader(), tenancy.Global(), second.Secret)
	require.ErrorIs(t, err, passwordreset.ErrTokenNotFound)
}

func TestProvidePasswordResetTokenStore(T *testing.T) {
	T.Parallel()

	T.Run("with nil recorder", func(t *testing.T) {
		t.Parallel()

		// Refused before the store is built: a token store that recorded nothing would be
		// the one an investigation finds empty.
		actual, err := ProvidePasswordResetTokenStore(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, nil)
		require.ErrorIs(t, err, passwordreset.ErrNilRecorder)
		assert.Nil(t, actual)
	})
}

// issueT mints one token on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test that wants one
// row written supplies the transaction the production caller would — and, here, the transaction
// that carries the audit entry alongside it.
func issueT(ctx context.Context, db database.Client, store passwordreset.Store, userID string, ttl time.Duration) (*passwordreset.Issuance, error) {
	return writeT(ctx, db, func(tx database.Tx) (*passwordreset.Issuance, error) {
		return store.Issue(ctx, tx, tenancy.Global(), userID, ttl)
	})
}

// consumeT spends one token on a transaction of its own.
func consumeT(ctx context.Context, db database.Client, store passwordreset.Store, secret string) (*passwordreset.Token, error) {
	return writeT(ctx, db, func(tx database.Tx) (*passwordreset.Token, error) {
		return store.Consume(ctx, tx, tenancy.Global(), secret)
	})
}

// revokeForUserT takes one user's outstanding links back, on a transaction of its own.
func revokeForUserT(ctx context.Context, db database.Client, store passwordreset.Store, userID string) (int64, error) {
	return writeT(ctx, db, func(tx database.Tx) (int64, error) {
		return store.RevokeForUser(ctx, tx, tenancy.Global(), userID)
	})
}

func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
