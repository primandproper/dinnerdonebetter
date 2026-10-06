package authentication

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients/authserver"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	signingrpc "github.com/primandproper/platform-go/v15/authentication/signin/grpc"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/primitives-go/v2/authentication/tokens/jwt"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/random"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeDirectory answers GetPrincipal from what it holds, refusing as the real directory would.
type fakeDirectory struct {
	err        error
	principals map[string]*platformidentity.Principal
}

func (d *fakeDirectory) GetPrincipal(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*platformidentity.Principal, error) {
	if d.err != nil {
		return nil, d.err
	}

	principal, ok := d.principals[userID]
	if !ok {
		return nil, platformidentity.ErrUserNotFound
	}

	if !principal.User.AccountStatus.AdmitsSignIn() {
		return nil, platformidentity.ErrSignInNotAdmitted
	}

	return principal, nil
}

// fakeSignIns answers CheckSignIn with ended for every family it was told has ended.
type fakeSignIns struct {
	ended map[string]bool
}

func (f *fakeSignIns) CheckSignIn(_ context.Context, _ tenancy.Scope, familyID, _ string) error {
	if f.ended[familyID] {
		return signin.ErrSignInEnded
	}

	return nil
}

type resolverHarness struct {
	signer    *jwt.Signer
	directory *fakeDirectory
	signIns   *fakeSignIns
	resolver  authserver.ScopedSubjectResolver
}

func newResolverHarness(t *testing.T) *resolverHarness {
	t.Helper()

	key, err := random.GenerateRawBytes(t.Context(), 32)
	require.NoError(t, err)

	signer, err := jwt.NewSigner(identifiers.New(), identifiers.New(), key)
	require.NoError(t, err)

	h := &resolverHarness{
		signer:    signer,
		directory: &fakeDirectory{principals: map[string]*platformidentity.Principal{}},
		signIns:   &fakeSignIns{ended: map[string]bool{}},
	}

	extractor, err := signingrpc.NewPrincipalExtractor(signer,
		&databasemock.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }},
		h.directory,
		signingrpc.WithSignInCheck(h.signIns),
		signingrpc.WithoutPasswordChangeGate(),
	)
	require.NoError(t, err)

	h.resolver = NewSessionResolver(extractor)

	return h
}

func (h *resolverHarness) addUser() *platformidentity.Principal {
	user := identityfakes.BuildFakeUser()
	user.ServiceRoles = []string{authorization.ServiceUserRoleName}

	principal := &platformidentity.Principal{User: user, ActiveAccountID: identifiers.New()}
	h.directory.principals[user.ID] = principal

	return principal
}

// signIn mints the token signin would, and returns it with the login it names.
func (h *resolverHarness) signIn(t *testing.T, principal *platformidentity.Principal, actorID string) (token, familyID string) {
	t.Helper()

	input := &signin.ClaimsInput{Principal: principal, FamilyID: identifiers.New(), ActorID: actorID}
	if actorID != "" {
		input.ActorScope = tenancy.Global()
	}

	claims, err := signin.DefaultClaims(t.Context(), input)
	require.NoError(t, err)

	token, _, err = h.signer.IssueToken(t.Context(), principal.User.ID, time.Hour, claims)
	require.NoError(t, err)

	return token, input.FamilyID
}

func authorizeRequest(t *testing.T, token string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", http.NoBody)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return req
}

func TestSessionResolver_ResolveScopedSubject(T *testing.T) {
	T.Parallel()

	T.Run("resolves a live sign-in to its user and account", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)
		user := h.addUser()
		token, _ := h.signIn(t, user, "")

		subject, scope, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, token))
		require.NoError(t, err)
		require.NotNil(t, subject)

		assert.Equal(t, user.User.ID, subject.ID)
		assert.Equal(t, user.ActiveAccountID, subject.Claims[authserver.ClaimAccountID])
		assert.True(t, scope.IsGlobal())
	})

	T.Run("leaves a request with no bearer token to the form", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, ""))
		require.NoError(t, err)
		assert.Nil(t, subject)
	})

	T.Run("refuses a signed-out session", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)
		token, familyID := h.signIn(t, h.addUser(), "")
		h.signIns.ended[familyID] = true

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, token))
		require.NoError(t, err)
		assert.Nil(t, subject)
	})

	T.Run("refuses a banned user", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)
		user := h.addUser()
		token, _ := h.signIn(t, user, "")

		// Signed in, and then banned: the token still names an account, which is the case the
		// authenticator this replaced let through.
		user.User.AccountStatus = platformidentity.StatusBanned

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, token))
		require.NoError(t, err)
		assert.Nil(t, subject)
	})

	T.Run("refuses an impersonation token", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)
		operator := h.addUser()
		token, _ := h.signIn(t, h.addUser(), operator.User.ID)

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, token))
		require.NoError(t, err)
		assert.Nil(t, subject)
	})

	T.Run("refuses a token nobody signed", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, identifiers.New()))
		require.NoError(t, err)
		assert.Nil(t, subject)
	})

	T.Run("ends the attempt when the directory cannot be read", func(t *testing.T) {
		t.Parallel()

		h := newResolverHarness(t)
		token, _ := h.signIn(t, h.addUser(), "")
		h.directory.err = errors.New("blah")

		subject, _, err := h.resolver.ResolveScopedSubject(t.Context(), authorizeRequest(t, token))
		require.Error(t, err)
		assert.Nil(t, subject)
	})
}

func TestBearerTokenFrom(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		token := identifiers.New()

		assert.Equal(t, token, bearerTokenFrom(authorizeRequest(t, token)))
	})

	T.Run("with no header", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, bearerTokenFrom(authorizeRequest(t, "")))
	})

	T.Run("with a non-bearer scheme", func(t *testing.T) {
		t.Parallel()

		req := authorizeRequest(t, "")
		req.Header.Set("Authorization", "Basic "+identifiers.New())

		assert.Empty(t, bearerTokenFrom(req))
	})
}
