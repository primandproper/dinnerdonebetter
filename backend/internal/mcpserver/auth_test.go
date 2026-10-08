package mcpserver

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	identity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	totpmock "github.com/primandproper/primitives-go/v2/authentication/totp/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/pointer"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	examplePassword  = "correct horse battery staple"
	exampleTOTPToken = "123456"
)

// authorizeRequest builds the POST /authorize a login form submission is, with the
// form already parsed — which is what the authorization server hands a
// SubjectAuthenticator.
func authorizeRequest(ctx context.Context, t *testing.T, username, password, totpToken string) *http.Request {
	t.Helper()

	form := url.Values{}
	form.Set(oauth2server.FieldUsername, username)
	form.Set(oauth2server.FieldPassword, password)
	form.Set(oauth2server.FieldTOTPCode, totpToken)

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())

	return req
}

// loginHarness is the MCP login over a real sign-in service, with the directory and the hasher
// faked. The hasher counts, because the number of hashes a refusal costs is what its timing is:
// one argon2 run dwarfs everything else a sign-in does.
type loginHarness struct {
	readErr error
	users   map[string]*identity.User
	a       *subjectAuthenticator
	hashes  atomic.Int32
}

func newLoginHarness(t *testing.T) *loginHarness {
	t.Helper()

	h := &loginHarness{users: map[string]*identity.User{}}

	directory := &identitymock.StoreMock{
		GetUserByUsernameFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			if h.readErr != nil {
				return nil, h.readErr
			}

			user, ok := h.users[username]
			if !ok {
				return nil, identity.ErrUserNotFound
			}

			return user, nil
		},
		GetPrincipalFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
			for _, user := range h.users {
				if user.ID == userID {
					return &identity.Principal{User: user, ActiveAccountID: user.ID + "_account"}, nil
				}
			}

			return nil, identity.ErrUserNotFound
		},
	}

	hasher := &authentication.AuthenticatorMock{
		HashPasswordFunc: func(context.Context, string) (string, error) {
			h.hashes.Add(1)
			return identifiers.New(), nil
		},
		PasswordMatchesFunc: func(_ context.Context, _, password string) (bool, error) {
			h.hashes.Add(1)
			return password == examplePassword, nil
		},
	}

	totpVerifier := &totpmock.VerifierMock{
		VerifyFunc: func(_ context.Context, _, code string) error {
			switch code {
			case exampleTOTPToken:
				return nil
			case "":
				return totp.ErrCodeRequired
			default:
				return totp.ErrInvalidCode
			}
		},
	}

	signIn, err := NewAdminSignIn(mockDBForTest(), directory, hasher, totpVerifier, nil, nil)
	require.NoError(t, err)

	h.a = &subjectAuthenticator{signIn: signIn}

	return h
}

// addUser puts somebody in the directory with a proven second factor, holding the service role
// named.
func (h *loginHarness) addUser(serviceRole string) *identity.User {
	user := identityfakes.BuildFakeUser()
	user.ServiceRoles = []string{serviceRole}
	user.HashedPassword = identifiers.New()
	user.TwoFactorSecretVerifiedAt = pointer.To(time.Now().Add(-time.Hour))
	h.users[user.Username] = user

	return user
}

func (h *loginHarness) authenticate(t *testing.T, username, password, code string) (*oauth2server.Subject, error) {
	t.Helper()

	h.hashes.Store(0)

	return h.a.AuthenticateSubject(t.Context(), authorizeRequest(t.Context(), t, username, password, code))
}

// requireAccessDenied asserts the form is re-rendered with the uninformative message, which is
// what every refusal before the second factor says.
func requireAccessDenied(t *testing.T, subject *oauth2server.Subject, err error) {
	t.Helper()

	assert.Nil(t, subject)
	require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

	var loginErr *oauth2server.LoginError
	require.ErrorAs(t, err, &loginErr)
	assert.Equal(t, accessDeniedMessage, loginErr.Message)
}

func TestSubjectAuthenticator_AuthenticateSubject(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)

		subject, err := h.authenticate(t, admin.Username, examplePassword, exampleTOTPToken)
		require.NoError(t, err)
		require.NotNil(t, subject)

		assert.Equal(t, admin.ID, subject.ID)
		assert.Equal(t, admin.ID+"_account", subject.Claims[claimAccountID])
	})

	T.Run("an unknown user, a non-administrator and a wrong password cost the same", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)
		member := h.addUser(authorization.ServiceUserRoleName)

		subject, err := h.authenticate(t, identifiers.New(), examplePassword, exampleTOTPToken)
		requireAccessDenied(t, subject, err)
		assert.Equal(t, int32(1), h.hashes.Load(), "an unknown user")

		subject, err = h.authenticate(t, member.Username, examplePassword, exampleTOTPToken)
		requireAccessDenied(t, subject, err)
		assert.Equal(t, int32(1), h.hashes.Load(), "a non-administrator")

		subject, err = h.authenticate(t, member.Username, "wrong", exampleTOTPToken)
		requireAccessDenied(t, subject, err)
		assert.Equal(t, int32(1), h.hashes.Load(), "a non-administrator's wrong password")

		subject, err = h.authenticate(t, admin.Username, "wrong", exampleTOTPToken)
		requireAccessDenied(t, subject, err)
		assert.Equal(t, int32(1), h.hashes.Load(), "an administrator's wrong password")
	})

	T.Run("with a banned administrator", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)
		admin.AccountStatus = identity.StatusBanned

		subject, err := h.authenticate(t, admin.Username, examplePassword, exampleTOTPToken)
		requireAccessDenied(t, subject, err)
	})

	T.Run("with a data administrator", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		dataAdmin := h.addUser(authorization.ServiceDataAdminRoleName)

		subject, err := h.authenticate(t, dataAdmin.Username, examplePassword, exampleTOTPToken)
		requireAccessDenied(t, subject, err)
	})

	T.Run("with a missing TOTP code", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)

		subject, err := h.authenticate(t, admin.Username, examplePassword, "")
		assert.Nil(t, subject)

		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, "TOTP code is required.", loginErr.Message)
	})

	T.Run("with an invalid TOTP code", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)

		subject, err := h.authenticate(t, admin.Username, examplePassword, "000000")
		requireAccessDenied(t, subject, err)
	})

	T.Run("with an administrator who has not proven a second factor", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)
		admin.TwoFactorSecretVerifiedAt = nil

		// The administrative door demands one whatever the account's own policy says.
		subject, err := h.authenticate(t, admin.Username, examplePassword, "")
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})

	T.Run("with a directory that cannot be read", func(t *testing.T) {
		t.Parallel()

		h := newLoginHarness(t)
		admin := h.addUser(authorization.ServiceAdminRoleName)
		h.readErr = errors.New("blah")

		subject, err := h.authenticate(t, admin.Username, examplePassword, exampleTOTPToken)
		assert.Nil(t, subject)
		require.Error(t, err)

		// Not a LoginError: re-rendering the form would ask the human to fix an outage by
		// typing.
		require.NotErrorIs(t, err, oauth2server.ErrLoginFailed)
	})
}

const exampleResource = "http://localhost:8888"

// mockDBForTest is a client whose executors are nil, because the store above them is a
// mock and never sends a statement. A transaction runs its function with no Tx, which is all
// the sign-in service's no-op hooks need of one.
func mockDBForTest() *mockdatabase.ClientMock {
	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
		WithTransactionFunc: func(_ context.Context, fn func(database.Tx) error) error {
			return fn(nil)
		},
	}
}
