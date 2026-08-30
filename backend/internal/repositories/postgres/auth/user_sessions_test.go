package auth

import (
	"context"
	"testing"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/database"
	mockdatabase "github.com/primandproper/platform-go/v13/database/mock"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/sessions"
	sessionsmock "github.com/primandproper/platform-go/v13/sessions/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exampleSessionPolicy is deliberately not the deployed one. What the integration test below
// needs from a policy is that nothing it establishes expires while it runs.
var exampleSessionPolicy = &authcfg.SessionsConfig{
	AbsoluteTimeout: time.Hour,
	IdleTimeout:     time.Hour,
	TouchInterval:   time.Minute,
}

// buildAuditedSessionStoreForTest wraps a store that does nothing but succeed, so the audit
// half can be exercised without a database.
func buildAuditedSessionStoreForTest(inner auth.SessionStore, auditRepo audit.Repository, db database.Client) *auditedUserSessionStore {
	return &auditedUserSessionStore{
		SessionStore:      inner,
		db:                db,
		auditLogEntryRepo: auditRepo,
		tracer:            tracing.NewTracerForTest("test"),
		logger:            loggingnoop.NewLogger(),
	}
}

func buildSessionStoreForTest(t *testing.T, db database.Client, auditRepo audit.Repository) auth.SessionStore {
	t.Helper()

	backend, err := ProvideUserSessionBackend(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), db)
	require.NoError(t, err)

	store, err := ProvideUserSessionStore(
		exampleSessionPolicy,
		backend,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditRepo,
		db,
	)
	require.NoError(t, err)

	return store
}

func TestQuerier_Integration_UserSessions(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.Writer())
	holder := auth.SessionHolder(user.ID)
	store := buildSessionStoreForTest(t, dbc, auditRepo)

	// establish
	first, err := store.NewFor(ctx, holder, sessions.Metadata{
		DeviceName:  "Jeffrey's laptop",
		IPAddress:   "192.168.1.1",
		UserAgent:   "TestBrowser/1.0",
		LoginMethod: auth.LoginMethodPassword,
	}, &auth.SessionPayload{SessionTokenID: "access-one", RefreshTokenID: "refresh-one"})
	require.NoError(t, err)
	require.NotEmpty(t, first.ID)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUserSessions, RelevantID: first.ID},
	})

	// The payload round-trips through the codec. It carries a blank `_ struct{}` field, per
	// this repository's struct convention, and a codec that choked on it would be an
	// authenticated request that could not find its own token IDs.
	read, err := store.Get(ctx, first.ID)
	require.NoError(t, err)
	require.NotNil(t, read.Data)
	assert.Equal(t, "access-one", read.Data.SessionTokenID)
	assert.Equal(t, "refresh-one", read.Data.RefreshTokenID)
	assert.Equal(t, "Jeffrey's laptop", read.Metadata.DeviceName)
	assert.Equal(t, auth.LoginMethodPassword, read.Metadata.LoginMethod)
	assert.Equal(t, holder, read.Holder)

	// rotation
	require.NoError(t, store.Save(ctx, first.ID, &auth.SessionPayload{SessionTokenID: "access-two", RefreshTokenID: "refresh-two"}))
	rotated, err := store.Get(ctx, first.ID)
	require.NoError(t, err)
	assert.Equal(t, "access-two", rotated.Data.SessionTokenID)

	// a second device
	second, err := store.NewFor(ctx, holder, sessions.Metadata{DeviceName: "iPhone", LoginMethod: auth.LoginMethodPasskey}, &auth.SessionPayload{})
	require.NoError(t, err)

	listed, err := store.List(ctx, holder, second.ID)
	require.NoError(t, err)
	require.Len(t, listed, 2)
	assert.Equal(t, second.ID, listed[0].ID)
	assert.True(t, listed[0].IsCurrent)
	assert.False(t, listed[1].IsCurrent)

	// sign out my other devices
	revoked, err := store.RevokeAllExcept(ctx, holder, second.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, revoked)

	// revocation is a delete: the session is gone, not flagged
	_, err = store.Get(ctx, first.ID)
	require.ErrorIs(t, err, sessions.ErrNotFound)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUserSessions, RelevantID: first.ID},
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUserSessions, RelevantID: second.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUserSessions, RelevantID: first.ID},
	})

	// a session that is not this holder's is answered as absent rather than refused
	require.ErrorIs(t, store.Revoke(ctx, auth.SessionHolder("somebody-else"), second.ID), sessions.ErrNotFound)

	// sign out everywhere
	revoked, err = store.RevokeAll(ctx, holder)
	require.NoError(t, err)
	assert.Equal(t, 1, revoked)

	listed, err = store.List(ctx, holder, "")
	require.NoError(t, err)
	assert.Empty(t, listed)
}

func TestProvideUserSessionStore(T *testing.T) {
	T.Parallel()

	T.Run("with nil database client", func(t *testing.T) {
		t.Parallel()

		backend, err := ProvideUserSessionBackend(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), nil)
		require.Error(t, err)
		assert.Nil(t, backend)
	})
}

func TestAuditedUserSessionStore_NewFor(T *testing.T) {
	T.Parallel()

	T.Run("with error establishing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		inner := &sessionsmock.StoreMock[auth.SessionPayload]{
			NewForFunc: func(context.Context, sessions.Holder, sessions.Metadata, *auth.SessionPayload) (*auth.UserSession, error) {
				return nil, expected
			},
		}

		store := buildAuditedSessionStoreForTest(inner, nil, &mockdatabase.ClientMock{})

		actual, err := store.NewFor(ctx, auth.SessionHolder(t.Name()), sessions.Metadata{}, &auth.SessionPayload{})
		require.ErrorIs(t, err, expected)
		assert.Nil(t, actual)
		assert.Len(t, inner.NewForCalls(), 1)
	})
}

func TestAuditedUserSessionStore_Revoke(T *testing.T) {
	T.Parallel()

	T.Run("with error revoking", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		inner := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeFunc: func(context.Context, sessions.Holder, string) error {
				return sessions.ErrNotFound
			},
		}

		store := buildAuditedSessionStoreForTest(inner, nil, &mockdatabase.ClientMock{})

		require.ErrorIs(t, store.Revoke(ctx, auth.SessionHolder(t.Name()), "session"), sessions.ErrNotFound)
		assert.Len(t, inner.RevokeCalls(), 1)
	})
}

// A bulk revocation that cannot read the set first does not revoke, because an audit log that
// missed a sign-out nobody can reconstruct is worse than a sign-out the user repeats.
func TestAuditedUserSessionStore_RevokeAll(T *testing.T) {
	T.Parallel()

	T.Run("with error listing", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		expected := platformerrors.New("blah")

		inner := &sessionsmock.StoreMock[auth.SessionPayload]{
			ListFunc: func(context.Context, sessions.Holder, string) ([]*auth.UserSession, error) {
				return nil, expected
			},
		}

		store := buildAuditedSessionStoreForTest(inner, nil, &mockdatabase.ClientMock{})

		revoked, err := store.RevokeAll(ctx, auth.SessionHolder(t.Name()))
		require.ErrorIs(t, err, expected)
		assert.Zero(t, revoked)
		assert.Empty(t, inner.RevokeAllExceptCalls())
	})
}
