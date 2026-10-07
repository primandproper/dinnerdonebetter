package authentication

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	oauth2clientsmock "github.com/primandproper/platform-go/v15/authentication/oauth2clients/mock"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// formHarness is the API server's login form over a real sign-in service, with the directory and
// the hasher faked. The hasher counts, because a hash is what a sign-in's time is spent on.
type formHarness struct {
	users         map[string]*platformidentity.User
	authenticator oauth2server.SubjectAuthenticator
	password      string
	hashes        atomic.Int32
}

func newFormHarness(t *testing.T) *formHarness {
	t.Helper()

	h := &formHarness{users: map[string]*platformidentity.User{}, password: identifiers.New()}

	directory := &identitymock.StoreMock{
		GetUserByUsernameFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*platformidentity.User, error) {
			user, ok := h.users[username]
			if !ok {
				return nil, platformidentity.ErrUserNotFound
			}

			return user, nil
		},
		GetPrincipalFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*platformidentity.Principal, error) {
			for _, user := range h.users {
				if user.ID == userID {
					return &platformidentity.Principal{User: user, ActiveAccountID: identifiers.New()}, nil
				}
			}

			return nil, platformidentity.ErrUserNotFound
		},
	}

	hasher := &authentication.AuthenticatorMock{
		HashPasswordFunc: func(context.Context, string) (string, error) {
			h.hashes.Add(1)
			return identifiers.New(), nil
		},
		PasswordMatchesFunc: func(_ context.Context, _, password string) (bool, error) {
			h.hashes.Add(1)
			return password == h.password, nil
		},
	}

	db := &databasemock.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
		WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
			return fn(nil)
		},
	}

	signIn, err := signin.NewService(db, directory, hasher, tokens.NewNoopTokenIssuer(),
		[]string{authorization.AccountAdminRoleName},
		signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
	)
	require.NoError(t, err)

	h.authenticator, err = ProvideLoginFormAuthenticator(signIn, &oauth2clientsmock.StoreMock{}, db, loggingnoop.NewLogger(), nil, nil)
	require.NoError(t, err)

	return h
}

func (h *formHarness) addUser() *platformidentity.User {
	user := identityfakes.BuildFakeUser()
	user.HashedPassword = identifiers.New()
	user.TwoFactorSecretVerifiedAt = nil
	h.users[user.Username] = user

	return user
}

func (h *formHarness) submit(t *testing.T, username, password string) (*oauth2server.Subject, error) {
	t.Helper()

	form := url.Values{}
	form.Set(oauth2server.FieldUsername, username)
	form.Set(oauth2server.FieldPassword, password)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	h.hashes.Store(0)

	return h.authenticator.AuthenticateSubject(t.Context(), req)
}

func TestProvideLoginFormAuthenticator(T *testing.T) {
	T.Parallel()

	T.Run("signs somebody in", func(t *testing.T) {
		t.Parallel()

		h := newFormHarness(t)
		user := h.addUser()

		subject, err := h.submit(t, user.Username, h.password)
		require.NoError(t, err)
		require.NotNil(t, subject)

		assert.Equal(t, user.ID, subject.ID)
	})

	T.Run("an unknown user and a wrong password cost the same", func(t *testing.T) {
		t.Parallel()

		h := newFormHarness(t)
		user := h.addUser()

		subject, err := h.submit(t, identifiers.New(), h.password)
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)
		assert.Equal(t, int32(1), h.hashes.Load(), "an unknown user")

		subject, err = h.submit(t, user.Username, identifiers.New())
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)
		assert.Equal(t, int32(1), h.hashes.Load(), "a wrong password")
	})

	T.Run("refuses a banned user after proving the password", func(t *testing.T) {
		t.Parallel()

		h := newFormHarness(t)
		user := h.addUser()
		user.AccountStatus = platformidentity.StatusBanned

		subject, err := h.submit(t, user.Username, h.password)
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})
}
