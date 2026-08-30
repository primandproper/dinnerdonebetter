package auth

import (
	"context"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/authentication/passwordreset"
	passwordresetmock "github.com/primandproper/platform-go/v13/authentication/passwordreset/mock"
	"github.com/primandproper/platform-go/v13/database"
	mockdatabase "github.com/primandproper/platform-go/v13/database/mock"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleTokenLifetime = 30 * time.Minute

// buildAuditedStoreForTest wraps a store that does nothing but succeed, so the audit half
// can be exercised without a database.
func buildAuditedStoreForTest(inner passwordreset.Store, auditRepo audit.Repository, db database.Client) *auditedPasswordResetTokenStore {
	return &auditedPasswordResetTokenStore{
		Store:             inner,
		db:                db,
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
	issuance, err := store.Issue(ctx, tenancy.Global(), user.ID, exampleTokenLifetime)
	require.NoError(t, err)
	require.NotNil(t, issuance)
	assert.NotEmpty(t, issuance.Secret)
	assert.Equal(t, user.ID, issuance.Token.UserID)
	assert.Nil(t, issuance.Token.RedeemedAt)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePasswordResetTokens, RelevantID: issuance.Token.ID},
	})

	// the row holds a digest, not the token. This is the property the hand-written store
	// this replaced did not have, and the reason a database copy is no longer a password
	// reset for every account with an outstanding link.
	var stored string
	require.NoError(t, dbc.Reader().QueryRowContext(ctx,
		`SELECT token_digest FROM ddb_password_reset_tokens WHERE id = $1`, issuance.Token.ID).Scan(&stored))
	assert.NotEqual(t, issuance.Secret, stored)

	// verify does not spend it
	verified, err := store.Verify(ctx, tenancy.Global(), issuance.Secret)
	require.NoError(t, err)
	assert.Equal(t, issuance.Token.ID, verified.ID)
	assert.Nil(t, verified.RedeemedAt)

	// consume
	consumed, err := store.Consume(ctx, tenancy.Global(), issuance.Secret)
	require.NoError(t, err)
	assert.Equal(t, issuance.Token.ID, consumed.ID)
	assert.NotNil(t, consumed.RedeemedAt)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypePasswordResetTokens, RelevantID: issuance.Token.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypePasswordResetTokens, RelevantID: issuance.Token.ID},
	})

	// a token is spendable exactly once, and the store is what says so
	_, err = store.Consume(ctx, tenancy.Global(), issuance.Secret)
	require.ErrorIs(t, err, passwordreset.ErrTokenRedeemed)

	// revoking takes the outstanding links with it
	second, err := store.Issue(ctx, tenancy.Global(), user.ID, exampleTokenLifetime)
	require.NoError(t, err)

	revoked, err := store.RevokeForUser(ctx, tenancy.Global(), user.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(1), revoked)

	_, err = store.Verify(ctx, tenancy.Global(), second.Secret)
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

func TestAuditedPasswordResetTokenStore_Issue(T *testing.T) {
	T.Parallel()

	T.Run("with error issuing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		inner := &passwordresetmock.StoreMock{
			IssueFunc: func(context.Context, tenancy.Scope, string, time.Duration) (*passwordreset.Issuance, error) {
				return nil, expected
			},
		}

		store := buildAuditedStoreForTest(inner, nil, &mockdatabase.ClientMock{})

		actual, err := store.Issue(ctx, tenancy.Global(), t.Name(), exampleTokenLifetime)
		require.ErrorIs(t, err, expected)
		assert.Nil(t, actual)
		require.Len(t, inner.IssueCalls(), 1)
		assert.Equal(t, exampleTokenLifetime, inner.IssueCalls()[0].TTL)
	})
}

func TestAuditedPasswordResetTokenStore_Consume(T *testing.T) {
	T.Parallel()

	T.Run("with error consuming", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		inner := &passwordresetmock.StoreMock{
			ConsumeFunc: func(context.Context, tenancy.Scope, string) (*passwordreset.Token, error) {
				return nil, passwordreset.ErrTokenRedeemed
			},
		}

		store := buildAuditedStoreForTest(inner, nil, &mockdatabase.ClientMock{})

		actual, err := store.Consume(ctx, tenancy.Global(), t.Name())
		require.ErrorIs(t, err, passwordreset.ErrTokenRedeemed)
		assert.Nil(t, actual)
		assert.Len(t, inner.ConsumeCalls(), 1)
	})
}
