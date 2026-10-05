package auth

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/mock"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleTokenLifetime = 30 * time.Minute

// buildHooksForTest builds the recording hooks over an audit repository a test controls, so
// what they record can be exercised without a database.
func buildHooksForTest(auditRepo audit.Repository) *passwordResetHooks {
	return &passwordResetHooks{
		auditLogEntryRepo: auditRepo,
		tracer:            tracing.NewTracerForTest("test"),
		logger:            loggingnoop.NewLogger(),
	}
}

func TestQuerier_Integration_PasswordResetTokens(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())

	store, err := ProvidePasswordResetTokenStore(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditRepo, dbc)
	require.NoError(t, err)

	// issue
	issuance, err := issueT(ctx, dbc, store, user.ID, exampleTokenLifetime)
	require.NoError(t, err)
	require.NotNil(t, issuance)
	assert.NotEmpty(t, issuance.Secret)
	assert.Equal(t, user.ID, issuance.Token.UserID)
	assert.Nil(t, issuance.Token.RedeemedAt)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, dbc, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: resourceTypePasswordResetTokens, ResourceID: issuance.Token.ID},
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

	pgtesting.AssertAuditLogContainsForUser(t, ctx, dbc, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: resourceTypePasswordResetTokens, ResourceID: issuance.Token.ID},
		{EventType: platformaudit.EventUpdated, ResourceType: resourceTypePasswordResetTokens, ResourceID: issuance.Token.ID},
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

	T.Run("with nil database client", func(t *testing.T) {
		t.Parallel()

		actual, err := ProvidePasswordResetTokenStore(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestPasswordResetHooks_AfterIssue(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		token := &passwordreset.Token{ID: "token", UserID: "user", ExpiresAt: time.Now().Add(exampleTokenLifetime)}

		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error { return nil },
		}

		// database.NewTxForTesting exists for exactly this: the marker method on database.Tx
		// is unexported, so a test double cannot implement one. Nothing is ever sent on this
		// transaction — the audit repository is mocked.
		require.NoError(t, buildHooksForTest(auditRepo).AfterIssue(ctx, database.NewTxForTesting(nil), tenancy.Global(), token))

		require.Len(t, auditRepo.RecordCalls(), 1)
		require.Len(t, auditRepo.RecordCalls()[0].Entries, 1)
		entry := auditRepo.RecordCalls()[0].Entries[0]
		assert.Equal(t, platformaudit.EventCreated, entry.EventType)
		assert.Equal(t, resourceTypePasswordResetTokens, entry.ResourceType)
		assert.Equal(t, token.ID, entry.ResourceID)
		assert.Equal(t, token.UserID, entry.Actor.ID)
		assert.Empty(t, entry.Changes)
	})

	T.Run("with error recording", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error { return expected },
		}

		err := buildHooksForTest(auditRepo).AfterIssue(ctx, database.NewTxForTesting(nil), tenancy.Global(), &passwordreset.Token{ID: "token", UserID: "user"})
		require.ErrorIs(t, err, expected)
	})
}

func TestPasswordResetHooks_AfterConsume(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		redeemedAt := time.Now()
		token := &passwordreset.Token{ID: "token", UserID: "user", RedeemedAt: &redeemedAt}

		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error { return nil },
		}

		require.NoError(t, buildHooksForTest(auditRepo).AfterConsume(ctx, database.NewTxForTesting(nil), tenancy.Global(), token))

		require.Len(t, auditRepo.RecordCalls(), 1)
		require.Len(t, auditRepo.RecordCalls()[0].Entries, 1)
		entry := auditRepo.RecordCalls()[0].Entries[0]
		assert.Equal(t, platformaudit.EventUpdated, entry.EventType)
		assert.Equal(t, resourceTypePasswordResetTokens, entry.ResourceType)
		assert.Equal(t, token.ID, entry.ResourceID)
		assert.Equal(t, token.UserID, entry.Actor.ID)

		// No diff: platform hands no before row, and inventing one would only restate
		// the redemption the entry already records.
		assert.Empty(t, entry.Changes)
	})

	T.Run("with error recording", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")
		redeemedAt := time.Now()

		auditRepo := &auditmock.RepositoryMock{
			RecordFunc: func(context.Context, database.Tx, ...*platformaudit.Entry) error { return expected },
		}

		err := buildHooksForTest(auditRepo).AfterConsume(ctx, database.NewTxForTesting(nil), tenancy.Global(), &passwordreset.Token{ID: "token", UserID: "user", RedeemedAt: &redeemedAt})
		require.ErrorIs(t, err, expected)
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
