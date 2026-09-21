package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	identity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	"github.com/primandproper/platform-go/v14/sessions"
	sessionsmock "github.com/primandproper/platform-go/v14/sessions/mock"
	"github.com/primandproper/primitives-go/v2/authentication/tokens"
	mocktokens "github.com/primandproper/primitives-go/v2/authentication/tokens/mock"
	"github.com/primandproper/primitives-go/v2/authentication/totp"
	mocktotp "github.com/primandproper/primitives-go/v2/authentication/totp/mock"
	"github.com/primandproper/primitives-go/v2/database"
	mockdatabase "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	mockpublishers "github.com/primandproper/primitives-go/v2/messagequeue/mock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newClaimsMock builds a tokens.Claims-compatible mock.
// "sub" and "jti" are surfaced via Subject()/JTI(); extras are returned by Get/GetString.
func newClaimsMock(sub, jti string, extras map[string]string) *mocktokens.ClaimsMock {
	return &mocktokens.ClaimsMock{
		SubjectFunc: func() string { return sub },
		JTIFunc:     func() string { return jti },
		GetStringFunc: func(key string) (string, bool) {
			v, ok := extras[key]
			return v, ok
		},
		GetFunc: func(key string) (any, bool) {
			v, ok := extras[key]
			if !ok {
				return nil, false
			}
			return v, true
		},
		ExpiresAtFunc: func() time.Time { return time.Time{} },
	}
}

type managerTestMocks struct {
	tokenIssuer   *mocktokens.IssuerMock
	authenticator *AuthenticatorMock
	totpVerifier  *mocktotp.VerifierMock
	directory     *identitymock.StoreMock
	sessionStore  *sessionsmock.StoreMock[auth.SessionPayload]
	publisher     *mockpublishers.PublisherMock
}

// exampleSessionID is what the session store mock hands back from NewFor, and therefore what
// the `sid` claim on the tokens a login issues carries.
const exampleSessionID = "session-abc"

// newSessionStoreMock builds a session store that establishes and writes back without
// complaint, which is what every test that is not about the session store wants.
func newSessionStoreMock() *sessionsmock.StoreMock[auth.SessionPayload] {
	return &sessionsmock.StoreMock[auth.SessionPayload]{
		NewForFunc: func(_ context.Context, holder sessions.Holder, _ sessions.Metadata, _ *auth.SessionPayload) (*auth.UserSession, error) {
			return &auth.UserSession{ID: exampleSessionID, Holder: holder}, nil
		},
		SaveFunc: func(context.Context, string, *auth.SessionPayload) error {
			return nil
		},
	}
}

// helper to build a minimal manager for testing.
func buildTestManager(t *testing.T) (*manager, *managerTestMocks) {
	t.Helper()

	mocks := &managerTestMocks{
		tokenIssuer:   &mocktokens.IssuerMock{},
		authenticator: &AuthenticatorMock{},
		totpVerifier:  &mocktotp.VerifierMock{},
		directory:     &identitymock.StoreMock{},
		sessionStore:  newSessionStoreMock(),
		publisher: &mockpublishers.PublisherMock{
			PublishFunc:      func(_ context.Context, _ any, _ ...messagequeue.PublishOption) error { return nil },
			PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
		},
	}

	// The real sign-in service over the same mocks, rather than a mock of it.
	//
	// signin.Service is a concrete type and deliberately has no interface, so the choice
	// was to introduce one here or to build the real thing — and the real thing is what
	// these tests want anyway. Every argument it takes is already mocked: the identity
	// store is its Directory, AuthenticatorMock is its Authenticator, IssuerMock is its
	// TokenIssuer. What the subtests below assert is therefore that this package's door
	// wires platform's orchestration correctly, which is the thing an adoption can get
	// wrong; that the orchestration itself refuses a wrong password is platform's test.
	//
	// WithTransaction is the one call the mocks above do not cover. signin runs its hooks
	// inside a transaction, and the hooks here are the default no-ops, so the callback is
	// invoked with a nil Tx and does nothing with it.
	db := &mockdatabase.ClientMock{
		ReaderFunc: func() database.SQLQueryExecutor { return nil },
		WriterFunc: func() database.SQLQueryExecutor { return nil },
		WithTransactionFunc: func(ctx context.Context, fn func(database.Tx) error) error {
			return fn(nil)
		},
	}

	signInService, err := signin.NewService(
		db,
		mocks.directory,
		mocks.authenticator,
		mocks.tokenIssuer,
		signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
		signin.WithAdminServiceRoles(authorization.ServiceAdminRoleName),
		signin.WithTOTPVerifier(mocks.totpVerifier),
	)
	require.NoError(t, err)

	m := &manager{
		tokenIssuer:          mocks.tokenIssuer,
		signIn:               signInService,
		tracer:               tracing.NewNamedTracer(tracingnoop.NewTracerProvider(), "test"),
		logger:               loggingnoop.NewLogger(),
		dataChangesPublisher: mocks.publisher,
		directory:            mocks.directory,
		db:                   db,

		sessionStore:            mocks.sessionStore,
		maxAccessTokenLifetime:  15 * time.Minute,
		maxRefreshTokenLifetime: 24 * time.Hour,
	}

	return m, mocks
}

func buildExampleUser() *identity.User {
	return &identity.User{
		ID:             "user123",
		Username:       "testuser",
		HashedPassword: "hashedpassword",
		AccountStatus:  identity.StatusGood,
		EmailAddress:   "test@example.com",
		FirstName:      "Test",
		LastName:       "User",
	}
}

func Test_deriveDeviceName(T *testing.T) {
	T.Parallel()

	tests := []struct {
		name      string
		userAgent string
		expected  string
	}{
		{name: "empty user agent", userAgent: "", expected: "Unknown Device"},
		{name: "iPhone", userAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)", expected: "iPhone"},
		{name: "iPad", userAgent: "Mozilla/5.0 (iPad; CPU OS 16_0 like Mac OS X)", expected: "iPad"},
		{name: "Android", userAgent: "Mozilla/5.0 (Linux; Android 13; Pixel 7)", expected: "Android Device"},
		{name: "Mac via Macintosh", userAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)", expected: "Mac"},
		{name: "Mac via Mac OS", userAgent: "Mozilla/5.0 (compatible; Mac OS X 12_0)", expected: "Mac"},
		{name: "Windows", userAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64)", expected: "Windows PC"},
		{name: "Linux", userAgent: "Mozilla/5.0 (X11; Linux x86_64)", expected: "Linux"},
		{name: "unknown user agent", userAgent: "SomeCustomBot/1.0", expected: "Unknown Device"},
	}

	for _, tc := range tests {
		T.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := deriveDeviceName(tc.userAgent)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

// issueTokenFunc returns a moq IssueToken stub that alternates between (accessToken/accessJTI)
// on the first call and (refreshToken/refreshJTI) on the second.
func issueTokenFunc(accessToken, accessJTI, refreshToken, refreshJTI string) func(ctx context.Context, subject string, expiry time.Duration, extraClaims map[string]any) (string, string, error) {
	count := 0
	return func(_ context.Context, _ string, _ time.Duration, _ map[string]any) (string, string, error) {
		count++
		if count == 1 {
			return accessToken, accessJTI, nil
		}
		return refreshToken, refreshJTI, nil
	}
}

func TestManager_ProcessLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Empty(t, activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: "account123"}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		response, err := m.ProcessLogin(ctx, false, loginInput, &LoginMetadata{
			ClientIP:  "127.0.0.1",
			UserAgent: "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7)",
		})

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "access-token", response.AccessToken)
		assert.Equal(t, "refresh-token", response.RefreshToken)
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, "account123", response.AccountID)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.NewForCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
		assert.Len(t, mocks.tokenIssuer.IssueTokenCalls(), 2)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username:         "testuser",
			Password:         "validP@ssw0rd",
			DesiredAccountID: "specific-account",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "specific-account", activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "specific-account", response.AccountID)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.NewForCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "wrongP@ssw0rd",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		// PasswordMatches returns (false, nil) on a mismatch; validateLogin converts that to ErrPasswordDoesNotMatch.
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return false, nil
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = identity.StatusBanned

		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)

			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(context.Context, string, string) (bool, error) { return true, nil }

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		// A banned user must be rejected with an error and no token response, and after
		// the credentials are checked rather than before: a banned person presenting the
		// wrong password is told the password is wrong, not that they are banned.
		//
		// The sentinel is signin's rather than identity's, and the principal is never
		// read. The status check used to be a side effect of GetPrincipal refusing; signin
		// makes it a step of its own, between the password and the second factor, and
		// names the three statuses apart — suspended, terminated, unverified — where the
		// directory has one refusal for all of them.
		require.ErrorIs(t, err, signin.ErrUserBanned)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Empty(t, mocks.directory.GetPrincipalCalls())
	})

	T.Run("with nonexistent user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		loginInput := &auth.UserLoginInput{
			Username: "nouser",
			Password: "validP@ssw0rd",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return nil, errors.New("not found")
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
	})

	T.Run("with invalid login input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, _ := buildTestManager(t)

		loginInput := &auth.UserLoginInput{
			Username: "",
			Password: "",
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.Error(t, err)
		assert.Nil(t, response)
	})

	T.Run("with TOTP required but not provided", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		now := time.Now()
		user := buildExampleUser()
		user.TwoFactorSecretVerifiedAt = &now
		user.TwoFactorSecret = "ASECRET"

		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
			// TOTPToken intentionally left empty
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		// totp.Verify returns ErrCodeRequired when the code is empty.
		mocks.totpVerifier.VerifyFunc = func(_ context.Context, _, code string) error {
			if code == "" {
				return totp.ErrCodeRequired
			}
			return nil
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.ErrorIs(t, err, signin.ErrSecondFactorRequired)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)

		// The verifier is not consulted. A user who holds a proven second factor and
		// supplied no code is refused without one, which is the same answer the verifier
		// would have given and one fewer thing handed an empty string.
		assert.Empty(t, mocks.totpVerifier.VerifyCalls())
	})

	T.Run("with user not member of desired account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username:         "testuser",
			Password:         "validP@ssw0rd",
			DesiredAccountID: "other-account",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		// A named account the caller is not a live member of is refused by the read
		// itself, rather than answered with a principal claiming it.
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "other-account", activeAccountID)

			return nil, identity.ErrMembershipNotFound
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
	})

	// The administrator-only door refuses a real user who holds no administrator role, and
	// it refuses them *after* proving the password. That ordering is platform's and it is
	// the opposite of what this package used to do.
	//
	// Checking the role first is cheaper and is an enumeration oracle: an anonymous caller
	// who can distinguish "not an administrator" from "wrong password" can walk a list of
	// handles and learn which of them are operators, without holding a credential for any
	// of them. Paying for the hash first costs one comparison and closes it. The assertion
	// that the password was never checked has been inverted for that reason — it was
	// pinning the leak.
	T.Run("admin only refuses a user who holds no administrator role", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)

			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(context.Context, string, string) (bool, error) {
			return true, nil
		}

		response, err := m.ProcessLogin(ctx, true, loginInput, nil)

		require.ErrorIs(t, err, signin.ErrNotAnAdministrator)
		assert.Nil(t, response)

		// The password was proven and the role was not there, in that order. Nothing was
		// issued and no principal was resolved, which is where the refusal stops.
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Empty(t, mocks.directory.GetPrincipalCalls())
	})

	T.Run("admin only", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		// A proven second factor and a code for it, both of which the administrative door
		// requires whatever the service's policy is. This package used to check the
		// second factor only for a user who had proven one, on the administrative door as
		// well as the ordinary one — so an operator who never finished TOTP enrollment
		// could reach it with a password alone. docs/identity.md has said for a long time
		// that admin login "**requires** a valid TOTP token"; the code did not.
		now := time.Now()
		user := buildExampleUser()
		user.ServiceRoles = []string{authorization.ServiceAdminRoleName}
		user.TwoFactorSecret = "ASECRET"
		user.TwoFactorSecretVerifiedAt = &now

		loginInput := &auth.UserLoginInput{
			Username:  "testuser",
			Password:  "validP@ssw0rd",
			TOTPToken: "123456",
		}

		mocks.directory.GetUserByUsernameFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.totpVerifier.VerifyFunc = func(_ context.Context, secret, code string) error {
			assert.Equal(t, user.TwoFactorSecret, secret)
			assert.Equal(t, loginInput.TOTPToken, code)

			return nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Empty(t, activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: "account123"}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		response, err := m.ProcessLogin(ctx, true, loginInput, nil)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Len(t, mocks.directory.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.NewForCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})
}

func TestManager_ProcessPasskeyLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Empty(t, activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: "account123"}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionStore.NewForFunc = func(_ context.Context, holder sessions.Holder, metadata sessions.Metadata, _ *auth.SessionPayload) (*auth.UserSession, error) {
			assert.Equal(t, user.ID, holder.Principal)
			assert.Equal(t, auth.LoginMethodPasskey, metadata.LoginMethod)
			assert.Equal(t, "iPhone", metadata.DeviceName)
			assert.Equal(t, "10.0.0.1", metadata.IPAddress)
			return &auth.UserSession{ID: exampleSessionID, Holder: holder}, nil
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "", &LoginMetadata{
			ClientIP:  "10.0.0.1",
			UserAgent: "Mozilla/5.0 (iPhone; CPU iPhone OS 16_0 like Mac OS X)",
		})

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "access-token", response.AccessToken)
		assert.Equal(t, "refresh-token", response.RefreshToken)
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, "account123", response.AccountID)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.NewForCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "specific-account", activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "specific-account", nil)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "specific-account", response.AccountID)

		// One read. The user and the account they land in come back together, which is
		// what the three calls this path used to make have collapsed into.
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.NewForCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = identity.StatusBanned

		// The ban is the read's refusal rather than a field this path inspects: a status
		// that does not admit signing in is refused before any membership is read.
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			return nil, identity.ErrSignInNotAdmitted
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "", nil)

		require.ErrorIs(t, err, ErrUserBanned)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
	})

	T.Run("with nonexistent user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
			assert.Equal(t, "nonexistent", userID)

			return nil, errors.New("not found")
		}

		response, err := m.ProcessPasskeyLogin(ctx, "nonexistent", "", nil)

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
	})

	T.Run("with user not member of desired account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		// A named account the caller is not a live member of is refused by the read
		// itself, rather than answered with a principal claiming it.
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "other-account", activeAccountID)

			return nil, identity.ErrMembershipNotFound
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "other-account", nil)

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
	})
}

func TestManager_ExchangeTokenForUser(T *testing.T) {
	T.Parallel()

	T.Run("standard with session validation", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "refresh-jti-old", map[string]string{"account_id": "account123", "sid": exampleSessionID}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		mocks.sessionStore.GetFunc = func(_ context.Context, id string) (*auth.UserSession, error) {
			assert.Equal(t, exampleSessionID, id)
			return &auth.UserSession{
				ID:     exampleSessionID,
				Holder: auth.SessionHolder(user.ID),
				Data:   &auth.SessionPayload{RefreshTokenID: "refresh-jti-old"},
			}, nil
		}

		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Empty(t, activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: "account123"}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("new-access-token", "new-access-jti", "new-refresh-token", "new-refresh-jti")

		mocks.sessionStore.SaveFunc = func(_ context.Context, id string, data *auth.SessionPayload) error {
			assert.Equal(t, exampleSessionID, id)
			assert.Equal(t, "new-access-jti", data.SessionTokenID)
			assert.Equal(t, "new-refresh-jti", data.RefreshTokenID)
			return nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "new-access-token", response.AccessToken)
		assert.Equal(t, "new-refresh-token", response.RefreshToken)
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, "account123", response.AccountID)

		assert.Len(t, mocks.sessionStore.GetCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})

	T.Run("with revoked session", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "revoked-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "old-jti", map[string]string{"account_id": "account123", "sid": exampleSessionID}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		// A session the store cannot find is one that was revoked or has expired. Either
		// way the refresh token naming it is spent.
		mocks.sessionStore.GetFunc = func(_ context.Context, id string) (*auth.UserSession, error) {
			assert.Equal(t, exampleSessionID, id)
			return nil, sessions.ErrNotFound
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.GetCalls(), 1)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "refresh-jti", map[string]string{"account_id": "account123", "sid": exampleSessionID}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		mocks.sessionStore.GetFunc = func(_ context.Context, id string) (*auth.UserSession, error) {
			assert.Equal(t, exampleSessionID, id)
			return &auth.UserSession{
				ID:   exampleSessionID,
				Data: &auth.SessionPayload{RefreshTokenID: "refresh-jti"},
			}, nil
		}

		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "desired-account", activeAccountID)

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("new-access-token", "new-access-jti", "new-refresh-token", "new-refresh-jti")

		mocks.sessionStore.SaveFunc = func(_ context.Context, id string, data *auth.SessionPayload) error {
			assert.Equal(t, exampleSessionID, id)
			assert.Equal(t, "new-access-jti", data.SessionTokenID)
			assert.Equal(t, "new-refresh-jti", data.RefreshTokenID)
			return nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "desired-account")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "desired-account", response.AccountID)

		assert.Len(t, mocks.sessionStore.GetCalls(), 1)
		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Len(t, mocks.sessionStore.SaveCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = identity.StatusBanned
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "jti", map[string]string{"account_id": "account123"}), nil
		}
		// The ban refuses the read, before the session the token names is even looked
		// at: somebody who may not sign in should not have their session touched.
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			return nil, identity.ErrSignInNotAdmitted
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		// A banned user must be rejected with an error and no token response.
		require.ErrorIs(t, err, ErrUserBanned)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Empty(t, mocks.sessionStore.GetCalls())
	})

	T.Run("with a refresh token naming no session", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "sessionless-refresh-token"

		// No "sid" claim. Every token this application issues carries one, so a refresh
		// token without it names no session — and a refresh that minted tokens for no
		// session would mint a pair nothing can sign out.
		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "", map[string]string{"account_id": "account123"}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}
		mocks.sessionStore.GetFunc = func(_ context.Context, id string) (*auth.UserSession, error) {
			assert.Empty(t, id)
			return nil, sessions.ErrIDRequired
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Empty(t, mocks.tokenIssuer.IssueTokenCalls())
	})

	T.Run("with a superseded refresh token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "already-spent-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "spent-jti", map[string]string{"account_id": "account123", "sid": exampleSessionID}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		// The session is perfectly live; it has simply been issued a newer pair since,
		// which is what spending this refresh token once already did.
		mocks.sessionStore.GetFunc = func(_ context.Context, id string) (*auth.UserSession, error) {
			assert.Equal(t, exampleSessionID, id)
			return &auth.UserSession{
				ID:   exampleSessionID,
				Data: &auth.SessionPayload{RefreshTokenID: "newer-jti"},
			}, nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.Error(t, err)
		require.ErrorIs(t, err, ErrSessionSuperseded)
		assert.Nil(t, response)

		assert.Len(t, mocks.sessionStore.GetCalls(), 1)
		assert.Empty(t, mocks.tokenIssuer.IssueTokenCalls())
	})

	T.Run("with invalid refresh token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		refreshToken := "bad-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return nil, errors.New("invalid token")
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.Error(t, err)
		assert.Nil(t, response)
	})

	T.Run("with nonexistent user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock("nonexistent-user", "jti", map[string]string{"account_id": "account123"}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, _ string) (*identity.Principal, error) {
			assert.Equal(t, "nonexistent-user", userID)

			return nil, errors.New("not found")
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
	})

	T.Run("with user not member of desired account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "jti", map[string]string{"account_id": "account123", "sid": exampleSessionID}), nil
		}
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)

			if activeAccountID == "" {
				activeAccountID = "account123"
			}

			return &identity.Principal{User: user, ActiveAccountID: activeAccountID}, nil
		}

		// A named account the caller is not a live member of is refused by the read
		// itself, rather than answered with a principal claiming it — and the refusal
		// lands before the session is read, for the reason the banned case gives.
		mocks.directory.GetPrincipalFunc = func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID, activeAccountID string) (*identity.Principal, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "wrong-account", activeAccountID)

			return nil, identity.ErrMembershipNotFound
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "wrong-account")

		require.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.directory.GetPrincipalCalls(), 1)
		assert.Empty(t, mocks.sessionStore.GetCalls())
	})
}
