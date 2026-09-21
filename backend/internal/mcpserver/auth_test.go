package mcpserver

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
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	identity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	oauth2memory "github.com/primandproper/primitives-go/v2/authentication/oauth2server/memory"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	totpmock "github.com/primandproper/primitives-go/v2/authentication/totp/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/auth"
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
	form.Set("username", username)
	form.Set("password", password)
	form.Set("totp_token", totpToken)

	req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/authorize", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	require.NoError(t, req.ParseForm())

	return req
}

func TestSubjectAuthenticator_AuthenticateSubject(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()
		exampleAccountID := identityfakes.BuildFakeAccount().ID

		directory := &identitymock.StoreMock{
			GetUserByUsernameFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
				assert.Equal(t, exampleUser.Username, username)
				return exampleUser, nil
			},
			GetPrincipalFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
				assert.Equal(t, exampleUser.ID, userID)

				return &identity.Principal{User: exampleUser, ActiveAccountID: exampleAccountID}, nil
			},
		}

		a := &subjectAuthenticator{
			directory: directory,
			db:        mockDBForTest(),
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return nil },
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		require.NoError(t, err)
		require.NotNil(t, subject)

		assert.Equal(t, exampleUser.ID, subject.ID)
		assert.Equal(t, exampleAccountID, subject.Claims[claimAccountID])
	})

	T.Run("with no such admin user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return nil, errors.New("blah")
				},
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)

		// ErrLoginFailed re-renders the form; anything else fails the authorization
		// request outright, which is the wrong answer to a human who can try again.
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, accessDeniedMessage, loginErr.Message)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()
		exampleUser.AccountStatus = identity.StatusBanned

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
				// The status is no longer checked here. GetPrincipal refuses one that
				// does not admit signing in, before it reads a membership, which is what
				// makes the check one thing rather than one per surface.
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

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)
		require.ErrorIs(t, err, identity.ErrSignInNotAdmitted)
	})

	// The door is admin-only, and it is a role check rather than a filtered lookup — so a
	// perfectly real user who holds no administrator role is refused with the message a
	// missing user gets, and not with a different one.
	T.Run("with a non-administrator", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := identityfakes.BuildFakeUser()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, accessDeniedMessage, loginErr.Message)
	})

	T.Run("with wrong password", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return false, nil },
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

		// The same message a missing user gets. Two different answers here make the
		// form an account enumeration oracle.
		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, accessDeniedMessage, loginErr.Message)
	})

	T.Run("with missing TOTP code", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return totp.ErrCodeRequired },
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, ""))
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)

		// This one names its own cause, because a human who forgot the code can fix it.
		var loginErr *oauth2server.LoginError
		require.ErrorAs(t, err, &loginErr)
		assert.Equal(t, "TOTP code is required.", loginErr.Message)
	})

	T.Run("with invalid TOTP code", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return totp.ErrInvalidCode },
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)
		require.ErrorIs(t, err, oauth2server.ErrLoginFailed)
	})

	T.Run("with unverified second factor skips TOTP", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()
		exampleUser.TwoFactorSecretVerifiedAt = nil
		exampleAccountID := identityfakes.BuildFakeAccount().ID

		verifier := &totpmock.VerifierMock{
			VerifyFunc: func(context.Context, string, string) error { return totp.ErrInvalidCode },
		}

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					return &identity.Principal{User: exampleUser, ActiveAccountID: exampleAccountID}, nil
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: verifier,
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, ""))
		require.NoError(t, err)
		require.NotNil(t, subject)
		assert.Empty(t, verifier.VerifyCalls())
	})

	T.Run("with error fetching default account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		exampleUser := buildFakeAdminForTest()

		a := &subjectAuthenticator{
			db: mockDBForTest(),
			directory: &identitymock.StoreMock{
				GetUserByUsernameFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*identity.User, error) {
					return exampleUser, nil
				},
				GetPrincipalFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string, string) (*identity.Principal, error) {
					return nil, errors.New("blah")
				},
			},
			authenticator: &authentication.AuthenticatorMock{
				PasswordMatchesFunc: func(context.Context, string, string) (bool, error) { return true, nil },
			},
			totpVerifier: &totpmock.VerifierMock{
				VerifyFunc: func(context.Context, string, string) error { return nil },
			},
		}

		subject, err := a.AuthenticateSubject(ctx, authorizeRequest(ctx, t, exampleUser.Username, examplePassword, exampleTOTPToken))
		assert.Nil(t, subject)
		require.Error(t, err)

		// Not a LoginError: the credentials were right, so re-rendering the form would
		// ask the human to fix a broken record by typing.
		require.NotErrorIs(t, err, oauth2server.ErrLoginFailed)
	})
}

const exampleResource = "http://localhost:8888"

// newTestAuthServer builds an authorization server over a memory store, and returns
// it alongside the store so a test can plant a token in it.
func newTestAuthServer(t *testing.T) (*oauth2server.Server, oauth2server.Store) {
	t.Helper()

	store := oauth2memory.NewStore()

	srv, err := oauth2server.NewServer(exampleResource, store, oauth2server.SubjectAuthenticatorFunc(
		func(context.Context, *http.Request) (*oauth2server.Subject, error) {
			return nil, oauth2server.ErrLoginFailed
		},
	))
	require.NoError(t, err)

	return srv, store
}

// plantAccessToken stores an access token and returns the bearer value for it. The
// store keys on the digest, so the value exists only here — which is the property
// that makes a database dump hold nothing redeemable.
func plantAccessToken(t *testing.T, store oauth2server.Store, audience []string) (bearer, userID, accountID string) {
	t.Helper()

	ctx := t.Context()
	bearer = "totally-opaque-access-token"
	exampleUser := buildFakeAdminForTest()
	exampleAccountID := identityfakes.BuildFakeAccount().ID

	require.NoError(t, store.CreateAccessToken(ctx, &oauth2server.AccessToken{
		IssuedAt:  time.Now().Add(-time.Minute),
		ExpiresAt: time.Now().Add(time.Hour),
		Hash:      oauth2server.Hash(bearer),
		ClientID:  "example-client",
		FamilyID:  "example-family",
		Subject: oauth2server.Subject{
			ID:     exampleUser.ID,
			Claims: map[string]string{claimAccountID: exampleAccountID},
		},
		Scopes:   []string{"mcp"},
		Audience: audience,
	}))

	return bearer, exampleUser.ID, exampleAccountID
}

func TestNewTokenVerifier(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		srv, store := newTestAuthServer(t)
		bearer, userID, accountID := plantAccessToken(t, store, []string{exampleResource})

		info, err := newTokenVerifier(srv, exampleResource)(ctx, bearer, nil)
		require.NoError(t, err)
		require.NotNil(t, info)

		assert.Equal(t, userID, info.UserID)
		assert.Equal(t, []string{"mcp"}, info.Scopes)

		// The account travels on the token, so a tool call costs no second lookup.
		assert.Equal(t, accountID, info.Extra[claimAccountID])
	})

	T.Run("with no audience", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		srv, store := newTestAuthServer(t)
		bearer, _, _ := plantAccessToken(t, store, nil)

		// A client that sends no RFC 8707 resource parameter gets a token naming no
		// resource. Refusing those would make every such client unable to sign in.
		info, err := newTokenVerifier(srv, exampleResource)(ctx, bearer, nil)
		require.NoError(t, err)
		assert.NotNil(t, info)
	})

	T.Run("with audience naming another resource server", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		srv, store := newTestAuthServer(t)
		bearer, _, _ := plantAccessToken(t, store, []string{"https://somewhere.else.example"})

		// This is the check no authorization server can make for us: the token is
		// perfectly valid, and it was minted to be spent somewhere else.
		info, err := newTokenVerifier(srv, exampleResource)(ctx, bearer, nil)
		assert.Nil(t, info)
		require.ErrorIs(t, err, auth.ErrInvalidToken)
		require.ErrorIs(t, err, errWrongAudience)
	})

	T.Run("with unknown token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		srv, _ := newTestAuthServer(t)

		info, err := newTokenVerifier(srv, exampleResource)(ctx, "nonsense", nil)
		assert.Nil(t, info)
		require.ErrorIs(t, err, auth.ErrInvalidToken)
	})

	T.Run("with expired token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		srv, store := newTestAuthServer(t)

		bearer := "expired-access-token"
		require.NoError(t, store.CreateAccessToken(ctx, &oauth2server.AccessToken{
			IssuedAt:  time.Now().Add(-2 * time.Hour),
			ExpiresAt: time.Now().Add(-time.Hour),
			Hash:      oauth2server.Hash(bearer),
			ClientID:  "example-client",
			Subject:   oauth2server.Subject{ID: identityfakes.BuildFakeUser().ID},
		}))

		// Expiry is the store's answer, not one this verifier repeats — a second
		// clock read here could disagree with the one the store used.
		info, err := newTokenVerifier(srv, exampleResource)(ctx, bearer, nil)
		assert.Nil(t, info)
		require.ErrorIs(t, err, auth.ErrInvalidToken)
	})
}

// buildFakeAdminForTest is a user the MCP login form admits: a real user who holds the
// service administrator role, which is what the door checks now that the lookup no longer
// filters non-administrators out inside the query.
func buildFakeAdminForTest() *identity.User {
	user := identityfakes.BuildFakeUser()
	user.ServiceRoles = []string{authorization.ServiceAdminRoleName}

	return user
}

// mockDBForTest is a client whose executors are nil, because the store above them is a
// mock and never sends a statement.
func mockDBForTest() *mockdatabase.ClientMock {
	return &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
	}
}
