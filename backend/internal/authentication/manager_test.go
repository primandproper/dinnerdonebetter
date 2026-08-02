package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"

	"github.com/primandproper/platform-go/v9/authentication/tokens"
	mocktokens "github.com/primandproper/platform-go/v9/authentication/tokens/mock"
	"github.com/primandproper/platform-go/v9/authentication/totp"
	mocktotp "github.com/primandproper/platform-go/v9/authentication/totp/mock"
	mockpublishers "github.com/primandproper/platform-go/v9/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

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
	tokenIssuer         *mocktokens.IssuerMock
	authenticator       *AuthenticatorMock
	totpVerifier        *mocktotp.VerifierMock
	userAuthDataManager *identitymock.RepositoryMock
	sessionDataManager  *authmock.UserSessionDataManagerMock
	publisher           *mockpublishers.PublisherMock
}

// helper to build a minimal manager for testing.
func buildTestManager(t *testing.T) (*manager, *managerTestMocks) {
	t.Helper()

	mocks := &managerTestMocks{
		tokenIssuer:         &mocktokens.IssuerMock{},
		authenticator:       &AuthenticatorMock{},
		totpVerifier:        &mocktotp.VerifierMock{},
		userAuthDataManager: &identitymock.RepositoryMock{},
		sessionDataManager:  &authmock.UserSessionDataManagerMock{},
		publisher: &mockpublishers.PublisherMock{
			PublishFunc:      func(_ context.Context, _ any) error { return nil },
			PublishAsyncFunc: func(_ context.Context, _ any) {},
		},
	}

	m := &manager{
		tokenIssuer:             mocks.tokenIssuer,
		authenticator:           mocks.authenticator,
		totpVerifier:            mocks.totpVerifier,
		tracer:                  tracing.NewNamedTracer(tracingnoop.NewTracerProvider(), "test"),
		logger:                  loggingnoop.NewLogger(),
		dataChangesPublisher:    mocks.publisher,
		userAuthDataManager:     mocks.userAuthDataManager,
		sessionDataManager:      mocks.sessionDataManager,
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
		AccountStatus:  string(identity.GoodStandingUserAccountStatus),
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

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.userAuthDataManager.GetDefaultAccountIDForUserFunc = func(_ context.Context, userID string) (string, error) {
			assert.Equal(t, user.ID, userID)
			return "account123", nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionDataManager.CreateUserSessionFunc = func(_ context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
			assert.NotNil(t, input)
			return &auth.UserSession{}, nil
		}

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

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.GetDefaultAccountIDForUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.CreateUserSessionCalls(), 1)
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

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "specific-account", accountID)
			return true, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionDataManager.CreateUserSessionFunc = func(_ context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
			assert.NotNil(t, input)
			return &auth.UserSession{}, nil
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "specific-account", response.AccountID)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.CreateUserSessionCalls(), 1)
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

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
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

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = string(identity.BannedUserAccountStatus)

		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
		}

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		// A banned user must be rejected with an error and no token response.
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserBanned)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
	})

	T.Run("with nonexistent user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		loginInput := &auth.UserLoginInput{
			Username: "nouser",
			Password: "validP@ssw0rd",
		}

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return nil, errors.New("not found")
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
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

		assert.Error(t, err)
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

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
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

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.totpVerifier.VerifyCalls(), 1)
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

		mocks.userAuthDataManager.GetUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "other-account", accountID)
			return false, nil
		}

		response, err := m.ProcessLogin(ctx, false, loginInput, nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
	})

	T.Run("admin only", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		loginInput := &auth.UserLoginInput{
			Username: "testuser",
			Password: "validP@ssw0rd",
		}

		mocks.userAuthDataManager.GetAdminUserByUsernameFunc = func(_ context.Context, username string) (*identity.User, error) {
			assert.Equal(t, loginInput.Username, username)
			return user, nil
		}
		mocks.authenticator.PasswordMatchesFunc = func(_ context.Context, hash, password string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, loginInput.Password, password)
			return true, nil
		}
		mocks.userAuthDataManager.GetDefaultAccountIDForUserFunc = func(_ context.Context, userID string) (string, error) {
			assert.Equal(t, user.ID, userID)
			return "account123", nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionDataManager.CreateUserSessionFunc = func(_ context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
			assert.NotNil(t, input)
			return &auth.UserSession{}, nil
		}

		response, err := m.ProcessLogin(ctx, true, loginInput, nil)

		require.NoError(t, err)
		require.NotNil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetAdminUserByUsernameCalls(), 1)
		assert.Len(t, mocks.authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.GetDefaultAccountIDForUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.CreateUserSessionCalls(), 1)
	})
}

func TestManager_ProcessPasskeyLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}
		mocks.userAuthDataManager.GetDefaultAccountIDForUserFunc = func(_ context.Context, userID string) (string, error) {
			assert.Equal(t, user.ID, userID)
			return "account123", nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionDataManager.CreateUserSessionFunc = func(_ context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
			assert.Equal(t, auth.LoginMethodPasskey, input.LoginMethod)
			return &auth.UserSession{}, nil
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

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.GetDefaultAccountIDForUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.CreateUserSessionCalls(), 1)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}
		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "specific-account", accountID)
			return true, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("access-token", "access-jti", "refresh-token", "refresh-jti")

		mocks.sessionDataManager.CreateUserSessionFunc = func(_ context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
			assert.NotNil(t, input)
			return &auth.UserSession{}, nil
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "specific-account", nil)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "specific-account", response.AccountID)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.CreateUserSessionCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = string(identity.BannedUserAccountStatus)

		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "", nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
	})

	T.Run("with nonexistent user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, "nonexistent", userID)
			return nil, errors.New("not found")
		}

		response, err := m.ProcessPasskeyLogin(ctx, "nonexistent", "", nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
	})

	T.Run("with user not member of desired account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()

		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}
		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "other-account", accountID)
			return false, nil
		}

		response, err := m.ProcessPasskeyLogin(ctx, user.ID, "other-account", nil)

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
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
			return newClaimsMock(user.ID, "refresh-jti-old", map[string]string{"account_id": "account123", "sid": "session-abc"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		mocks.sessionDataManager.GetUserSessionByRefreshTokenIDFunc = func(_ context.Context, refreshTokenID string) (*auth.UserSession, error) {
			assert.Equal(t, "refresh-jti-old", refreshTokenID)
			return &auth.UserSession{
				ID:             "session-abc",
				BelongsToUser:  user.ID,
				RefreshTokenID: "refresh-jti-old",
			}, nil
		}

		mocks.userAuthDataManager.GetDefaultAccountIDForUserFunc = func(_ context.Context, userID string) (string, error) {
			assert.Equal(t, user.ID, userID)
			return "account123", nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("new-access-token", "new-access-jti", "new-refresh-token", "new-refresh-jti")

		mocks.sessionDataManager.UpdateSessionTokenIDsFunc = func(_ context.Context, sessionID, newSessionTokenID, newRefreshTokenID string, newExpiresAt time.Time) error {
			assert.Equal(t, "session-abc", sessionID)
			assert.Equal(t, "new-access-jti", newSessionTokenID)
			assert.Equal(t, "new-refresh-jti", newRefreshTokenID)
			assert.False(t, newExpiresAt.IsZero())
			return nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "new-access-token", response.AccessToken)
		assert.Equal(t, "new-refresh-token", response.RefreshToken)
		assert.Equal(t, user.ID, response.UserID)
		assert.Equal(t, "account123", response.AccountID)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.GetUserSessionByRefreshTokenIDCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.GetDefaultAccountIDForUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.UpdateSessionTokenIDsCalls(), 1)
	})

	T.Run("with revoked session", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "revoked-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "old-jti", map[string]string{"account_id": "account123", "sid": "session-abc"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		// Session not found means it was revoked
		mocks.sessionDataManager.GetUserSessionByRefreshTokenIDFunc = func(_ context.Context, refreshTokenID string) (*auth.UserSession, error) {
			assert.Equal(t, "old-jti", refreshTokenID)
			return nil, errors.New("not found")
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.GetUserSessionByRefreshTokenIDCalls(), 1)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "refresh-jti", map[string]string{"account_id": "account123", "sid": "session-abc"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		mocks.sessionDataManager.GetUserSessionByRefreshTokenIDFunc = func(_ context.Context, refreshTokenID string) (*auth.UserSession, error) {
			assert.Equal(t, "refresh-jti", refreshTokenID)
			return &auth.UserSession{
				ID:             "session-abc",
				RefreshTokenID: "refresh-jti",
			}, nil
		}

		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "desired-account", accountID)
			return true, nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("new-access-token", "new-access-jti", "new-refresh-token", "new-refresh-jti")

		mocks.sessionDataManager.UpdateSessionTokenIDsFunc = func(_ context.Context, sessionID, newSessionTokenID, newRefreshTokenID string, newExpiresAt time.Time) error {
			assert.Equal(t, "session-abc", sessionID)
			assert.Equal(t, "new-access-jti", newSessionTokenID)
			assert.Equal(t, "new-refresh-jti", newRefreshTokenID)
			assert.False(t, newExpiresAt.IsZero())
			return nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "desired-account")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "desired-account", response.AccountID)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.GetUserSessionByRefreshTokenIDCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.UpdateSessionTokenIDsCalls(), 1)
	})

	T.Run("with banned user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		user.AccountStatus = string(identity.BannedUserAccountStatus)
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "jti", map[string]string{"account_id": "account123"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		// A banned user must be rejected with an error and no token response.
		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrUserBanned)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
	})

	T.Run("without JTI or session ID in token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "legacy-refresh-token"

		// Legacy token: JTI empty, no "sid" claim.
		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "", map[string]string{"account_id": "account123"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		mocks.userAuthDataManager.GetDefaultAccountIDForUserFunc = func(_ context.Context, userID string) (string, error) {
			assert.Equal(t, user.ID, userID)
			return "account123", nil
		}

		mocks.tokenIssuer.IssueTokenFunc = issueTokenFunc("new-access-token", "new-access-jti", "new-refresh-token", "new-refresh-jti")

		// UpdateSessionTokenIDs should NOT be called because sessionID is empty

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, "new-access-token", response.AccessToken)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.GetDefaultAccountIDForUserCalls(), 1)
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

		assert.Error(t, err)
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
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, "nonexistent-user", userID)
			return nil, errors.New("not found")
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "")

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
	})

	T.Run("with user not member of desired account", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		m, mocks := buildTestManager(t)

		user := buildExampleUser()
		refreshToken := "valid-refresh-token"

		mocks.tokenIssuer.ParseTokenFunc = func(_ context.Context, _ string) (tokens.Claims, error) {
			return newClaimsMock(user.ID, "jti", map[string]string{"account_id": "account123", "sid": "session-abc"}), nil
		}
		mocks.userAuthDataManager.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		}

		mocks.sessionDataManager.GetUserSessionByRefreshTokenIDFunc = func(_ context.Context, refreshTokenID string) (*auth.UserSession, error) {
			assert.Equal(t, "jti", refreshTokenID)
			return &auth.UserSession{
				ID:             "session-abc",
				RefreshTokenID: "jti",
			}, nil
		}

		mocks.userAuthDataManager.UserIsMemberOfAccountFunc = func(_ context.Context, userID, accountID string) (bool, error) {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, "wrong-account", accountID)
			return false, nil
		}

		response, err := m.ExchangeTokenForUser(ctx, refreshToken, "wrong-account")

		assert.Error(t, err)
		assert.Nil(t, response)

		assert.Len(t, mocks.userAuthDataManager.GetUserCalls(), 1)
		assert.Len(t, mocks.sessionDataManager.GetUserSessionByRefreshTokenIDCalls(), 1)
		assert.Len(t, mocks.userAuthDataManager.UserIsMemberOfAccountCalls(), 1)
	})
}
