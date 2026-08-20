package webauthn

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"

	platformwebauthn "github.com/primandproper/platform-go/v12/authentication/webauthn"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sharedSessionStore stands in for the ceremony table: state written by one relying party is
// visible to another, and Consume is one-shot. It is what a second replica sees.
type sharedSessionStore struct {
	sessions map[string]*platformwebauthn.SessionData
	mu       sync.Mutex
}

func newSharedSessionStore() *sharedSessionStore {
	return &sharedSessionStore{sessions: map[string]*platformwebauthn.SessionData{}}
}

func (s *sharedSessionStore) Save(_ context.Context, session *platformwebauthn.SessionData, ttl time.Duration) error {
	if err := platformwebauthn.ValidateSession(session, ttl); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[session.Challenge] = session

	return nil
}

func (s *sharedSessionStore) Consume(_ context.Context, challenge string) (*platformwebauthn.SessionData, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	session, ok := s.sessions[challenge]
	if !ok {
		return nil, platformwebauthn.ErrSessionNotFound
	}

	delete(s.sessions, challenge)

	return session, nil
}

// buildTestService builds a service over the given store, so that a test can build two of them
// and prove a ceremony crosses between them.
func buildTestService(t *testing.T, store platformwebauthn.SessionStore) (*Service, *identitymock.RepositoryMock, *identity.User) {
	t.Helper()

	ctx := t.Context()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	relyingParty, err := platformwebauthn.NewRelyingParty(ctx, &platformwebauthn.Config{
		RPID:            "localhost",
		RPDisplayName:   "Testing",
		RPOrigins:       []string{"http://localhost:8080"},
		CeremonyTimeout: 2 * time.Minute,
	}, store, platformwebauthn.WithLogger(logger), platformwebauthn.WithTracerProvider(tracerProvider))
	require.NoError(t, err)

	user := identityfakes.BuildFakeUser()

	credentials := []*identity.WebAuthnCredential{
		{
			ID:            t.Name(),
			BelongsToUser: user.ID,
			CredentialID:  []byte("credential-id"),
			PublicKey:     []byte("public-key"),
		},
	}

	repo := &identitymock.RepositoryMock{
		GetWebAuthnCredentialsForUserFunc: func(context.Context, string) ([]*identity.WebAuthnCredential, error) {
			return credentials, nil
		},
	}

	service, err := NewService(logger, tracerProvider, relyingParty, repo, &staticUserStore{user: user})
	require.NoError(t, err)

	return service, repo, user
}

// staticUserStore answers with one user, by ID or by the username that user carries.
type staticUserStore struct {
	user *identity.User
}

func (s *staticUserStore) GetUserByID(_ context.Context, userID string) (*identity.User, error) {
	if s.user != nil && s.user.ID == userID {
		return s.user, nil
	}

	return nil, nil
}

func (s *staticUserStore) GetUserByUsername(_ context.Context, username string) (*identity.User, error) {
	if s.user != nil && s.user.Username == username {
		return s.user, nil
	}

	return nil, nil
}

func TestNewService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		service, _, _ := buildTestService(t, newSharedSessionStore())
		assert.NotNil(t, service)
	})

	T.Run("without a relying party", func(t *testing.T) {
		t.Parallel()

		service, err := NewService(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, nil, nil)

		assert.Nil(t, service)
		assert.ErrorIs(t, err, ErrNilRelyingParty)
	})
}

func TestService_BeginRegistrationOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, user := buildTestService(t, newSharedSessionStore())

		creation, challenge, err := service.BeginRegistrationOptions(ctx, user.ID)
		require.NoError(t, err)

		require.NotNil(t, creation)
		assert.Equal(t, creation.Response.Challenge.String(), challenge)
		assert.Equal(t, "localhost", creation.Response.RelyingParty.ID)
	})

	// The point of the durable store: the challenge is in the store rather than in the
	// process, so the replica that answers the ceremony need not be the one that issued it.
	T.Run("leaves the ceremony where another replica can consume it", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		store := newSharedSessionStore()

		issuer, _, user := buildTestService(t, store)

		_, challenge, err := issuer.BeginRegistrationOptions(ctx, user.ID)
		require.NoError(t, err)

		session, err := store.Consume(ctx, challenge)
		require.NoError(t, err)
		assert.Equal(t, challenge, session.Challenge)

		// And exactly once — a replayed answer finds nothing the second time.
		session, err = store.Consume(ctx, challenge)
		assert.Nil(t, session)
		assert.ErrorIs(t, err, platformwebauthn.ErrSessionNotFound)
	})

	T.Run("with unknown user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t, newSharedSessionStore())

		creation, challenge, err := service.BeginRegistrationOptions(ctx, "nonexistent")

		assert.Nil(t, creation)
		assert.Empty(t, challenge)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_BeginAuthenticationOptions(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		store := newSharedSessionStore()
		service, _, user := buildTestService(t, store)

		assertion, challenge, err := service.BeginAuthenticationOptions(ctx, user.Username)
		require.NoError(t, err)

		require.NotNil(t, assertion)
		assert.Equal(t, assertion.Response.Challenge.String(), challenge)
		assert.Len(t, assertion.Response.AllowedCredentials, 1)

		session, err := store.Consume(ctx, challenge)
		require.NoError(t, err)
		assert.Equal(t, []byte(user.ID), session.UserID)
	})

	// An empty username is a discoverable login: the passkey names the user, so the stored
	// ceremony carries no user handle and lists no credentials.
	T.Run("without a username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		store := newSharedSessionStore()
		service, _, _ := buildTestService(t, store)

		assertion, challenge, err := service.BeginAuthenticationOptions(ctx, "")
		require.NoError(t, err)

		require.NotNil(t, assertion)
		assert.Empty(t, assertion.Response.AllowedCredentials)

		session, err := store.Consume(ctx, challenge)
		require.NoError(t, err)
		assert.Empty(t, session.UserID)
	})

	T.Run("with unknown user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t, newSharedSessionStore())

		assertion, challenge, err := service.BeginAuthenticationOptions(ctx, "nobody")

		assert.Nil(t, assertion)
		assert.Empty(t, challenge)
		assert.ErrorIs(t, err, ErrUserNotFound)
	})
}

func TestService_FinishRegistration(T *testing.T) {
	T.Parallel()

	T.Run("with unparseable attestation", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, user := buildTestService(t, newSharedSessionStore())

		err := service.FinishRegistration(ctx, user.ID, []byte("not json"), "challenge")
		assert.Error(t, err)
	})
}

func TestService_FinishAuthentication(T *testing.T) {
	T.Parallel()

	T.Run("with unparseable assertion", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		service, _, _ := buildTestService(t, newSharedSessionStore())

		result, err := service.FinishAuthentication(ctx, "", []byte("not json"), "challenge")

		assert.Nil(t, result)
		assert.Error(t, err)
	})
}

func TestEncodeTransports(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		encoded, err := encodeTransports([]protocol.AuthenticatorTransport{protocol.USB, protocol.Internal})

		require.NoError(t, err)
		assert.JSONEq(t, `["usb","internal"]`, encoded)
	})

	// An empty column rather than the string "null", which is what marshaling a nil slice
	// would store and what the transport parser would then quietly read back as nothing.
	T.Run("with no transports", func(t *testing.T) {
		t.Parallel()

		encoded, err := encodeTransports(nil)

		require.NoError(t, err)
		assert.Empty(t, encoded)
	})
}
