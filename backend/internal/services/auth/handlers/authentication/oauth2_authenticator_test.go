package authentication

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	mockauthn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	identity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	tokensmock "github.com/primandproper/primitives-go/v2/authentication/tokens/mock"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	totpmock "github.com/primandproper/primitives-go/v2/authentication/totp/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// formRequest builds the POST /authorize a login form produces.
func formRequest(t *testing.T, values url.Values) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", strings.NewReader(values.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	return req
}

// bearerRequest builds the POST /authorize a first-party client produces: a session JWT in the
// header and no form fields at all.
func bearerRequest(t *testing.T, token string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", http.NoBody)
	req.Header.Set("Authorization", "Bearer "+token)

	return req
}

func TestSubjectAuthenticator_BearerPath(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		accountID := identityfakes.BuildFakeAccount().ID

		a := &subjectAuthenticator{
			tokenIssuer: stubIssuer(user.ID, nil),
			db:          mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetPrincipalFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
					assert.Equal(t, user.ID, userID)

					return &identity.Principal{User: user, ActiveAccountID: accountID}, nil
				},
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), bearerRequest(t, "a.b.c"))
		require.NoError(t, err)

		assert.Equal(t, user.ID, subject.ID)
		assert.Equal(t, accountID, subject.Claims[ClaimAccountID])
	})

	T.Run("with an account named in the token", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		accountID := identityfakes.BuildFakeAccount().ID

		// LoginForToken with a DesiredAccountID mints a JWT naming the account. Honoring it is
		// what keeps the OAuth2 token on the account the user chose rather than silently on
		// their default — so the repository must not be consulted at all here.
		a := &subjectAuthenticator{
			tokenIssuer: stubIssuerWithClaims(user.ID, map[string]string{ClaimAccountID: accountID}, nil),
			db:          mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					t.Fatal("the default account was resolved for a token that already named one")

					return nil, nil
				},
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), bearerRequest(t, "a.b.c"))
		require.NoError(t, err)

		assert.Equal(t, accountID, subject.Claims[ClaimAccountID])
	})

	T.Run("with an unparseable token", func(t *testing.T) {
		t.Parallel()

		// A LoginError rather than a hard failure: an expired session is the ordinary reason to
		// arrive with a bad token, and the answer is the login form.
		a := &subjectAuthenticator{tokenIssuer: stubIssuer("", errors.New("expired"))}

		subject, err := a.AuthenticateSubject(t.Context(), bearerRequest(t, "a.b.c"))

		assert.Nil(t, subject)
		assert.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})

	T.Run("with a token naming no subject", func(t *testing.T) {
		t.Parallel()

		a := &subjectAuthenticator{tokenIssuer: stubIssuer("", nil)}

		subject, err := a.AuthenticateSubject(t.Context(), bearerRequest(t, "a.b.c"))

		assert.Nil(t, subject)
		assert.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})
}

func TestSubjectAuthenticator_CredentialPath(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		user.TwoFactorSecretVerifiedAt = nil
		accountID := identityfakes.BuildFakeAccount().ID

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
					assert.Equal(t, user.Username, username)
					return user, nil
				},
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					return &identity.Principal{User: user, ActiveAccountID: accountID}, nil
				},
			},
			authenticator: &mockauthn.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{
			"username": {user.Username},
			"password": {"correct horse battery staple"},
		}))
		require.NoError(t, err)

		assert.Equal(t, user.ID, subject.ID)
		assert.Equal(t, accountID, subject.Claims[ClaimAccountID])
	})

	T.Run("with a wrong password", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return user, nil
				},
			},
			authenticator: &mockauthn.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return false, nil },
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{
			"username": {user.Username},
			"password": {"wrong"},
		}))

		assert.Nil(t, subject)
		assert.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})

	T.Run("with an unknown username", func(t *testing.T) {
		t.Parallel()

		// The same message as a wrong password, deliberately: distinguishable answers make a
		// public form an account enumeration oracle.
		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return nil, errors.New("no such user")
				},
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{"username": {"nobody"}}))

		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, loginFailedMessage, loginErr.Message)
	})

	T.Run("with a banned account", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		user.AccountStatus = identity.StatusBanned

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return user, nil
				},
				// The status is no longer checked in this authenticator. GetPrincipal
				// refuses one that does not admit signing in, before it reads a
				// membership — so the refusal arrives from the read every authenticated
				// request already makes rather than from a check each surface remembers.
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					return nil, identity.ErrSignInNotAdmitted
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return nil },
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{"username": {user.Username}}))

		assert.Nil(t, subject)
		assert.ErrorIs(t, err, identity.ErrSignInNotAdmitted)
	})

	T.Run("with a missing TOTP code", func(t *testing.T) {
		t.Parallel()

		user := identityfakes.BuildFakeUser()
		verifiedAt := time.Now()
		user.TwoFactorSecretVerifiedAt = &verifiedAt

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return user, nil
				},
			},
			authenticator: &mockauthn.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return totp.ErrCodeRequired },
			},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{
			"username": {user.Username},
			"password": {"correct horse battery staple"},
		}))

		assert.Nil(t, subject)

		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, "TOTP code is required.", loginErr.Message)
	})

	T.Run("with no resolvable default account", func(t *testing.T) {
		t.Parallel()

		// Not a LoginError. The credentials were right and the record is broken; re-rendering
		// the form would ask the human to fix it by typing.
		user := identityfakes.BuildFakeUser()
		user.TwoFactorSecretVerifiedAt = nil

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return user, nil
				},
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					return nil, errors.New("no memberships")
				},
			},
			authenticator: &mockauthn.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{},
		}

		subject, err := a.AuthenticateSubject(t.Context(), formRequest(t, url.Values{
			"username": {user.Username},
			"password": {"correct horse battery staple"},
		}))

		assert.Nil(t, subject)
		require.Error(t, err)
		assert.NotErrorIs(t, err, oauth2server.ErrLoginFailed)
	})
}

func TestBearerTokenFrom(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "abc123", bearerTokenFrom(bearerRequest(t, "abc123")))
	})

	T.Run("with no header", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, bearerTokenFrom(httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", http.NoBody)))
	})

	T.Run("with a non-bearer scheme", func(t *testing.T) {
		t.Parallel()

		// Basic credentials at /authorize are the client's, not the resource owner's. Reading
		// them as a session token would be authenticating the wrong party.
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/authorize", http.NoBody)
		req.Header.Set("Authorization", "Basic Zm9vOmJhcg==")

		assert.Empty(t, bearerTokenFrom(req))
	})
}

// stubIssuer returns a tokens.Issuer whose ParseToken answers with the given subject and claims,
// or fails with err.
func stubIssuer(subject string, err error) tokens.Issuer {
	return stubIssuerWithClaims(subject, nil, err)
}

func stubIssuerWithClaims(subject string, extra map[string]string, err error) tokens.Issuer {
	return &tokensmock.IssuerMock{
		ParseTokenFunc: func(context.Context, string) (tokens.Claims, error) {
			if err != nil {
				return nil, err
			}

			return &stubClaims{subject: subject, extra: extra}, nil
		},
	}
}

// stubClaims is the parsed shape of a session JWT, with only the two fields this authenticator
// reads populated.
type stubClaims struct {
	extra   map[string]string
	subject string
}

var _ tokens.Claims = (*stubClaims)(nil)

func (c *stubClaims) Subject() string      { return c.subject }
func (c *stubClaims) JTI() string          { return "" }
func (c *stubClaims) ExpiresAt() time.Time { return time.Time{} }
func (c *stubClaims) Get(key string) (any, bool) {
	value, ok := c.extra[key]
	return value, ok
}

func (c *stubClaims) GetString(key string) (string, bool) {
	value, ok := c.extra[key]
	return value, ok
}

// mockDBForTest is a client whose executors are nil, because the store above them is a
// mock and never sends a statement.
func mockDBForTest() *mockdatabase.ClientMock {
	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
	}
}
