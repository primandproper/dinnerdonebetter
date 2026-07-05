package managers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	mockauthn "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	authfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	platformtotp "github.com/primandproper/platform-go/v3/authentication/totp"
	mocktotp "github.com/primandproper/platform-go/v3/authentication/totp/mock"
	"github.com/primandproper/platform-go/v3/database/filtering"
	"github.com/primandproper/platform-go/v3/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v3/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"
	"github.com/primandproper/platform-go/v3/qrcodes"
	"github.com/primandproper/platform-go/v3/random"
	randommock "github.com/primandproper/platform-go/v3/random/mock"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// mockPasswordResetTokenDataManager is a test double for auth.PasswordResetTokenDataManager.
type mockPasswordResetTokenDataManager struct {
	mock.Mock
}

func (m *mockPasswordResetTokenDataManager) GetPasswordResetTokenByID(ctx context.Context, passwordResetTokenID string) (*auth.PasswordResetToken, error) {
	args := m.Called(ctx, passwordResetTokenID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.PasswordResetToken), args.Error(1)
}

func (m *mockPasswordResetTokenDataManager) GetPasswordResetTokenByToken(ctx context.Context, token string) (*auth.PasswordResetToken, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.PasswordResetToken), args.Error(1)
}

func (m *mockPasswordResetTokenDataManager) CreatePasswordResetToken(ctx context.Context, input *auth.PasswordResetTokenDatabaseCreationInput) (*auth.PasswordResetToken, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.PasswordResetToken), args.Error(1)
}

func (m *mockPasswordResetTokenDataManager) RedeemPasswordResetToken(ctx context.Context, passwordResetTokenID string) error {
	return m.Called(ctx, passwordResetTokenID).Error(0)
}

type mockUserSessionDataManager struct {
	mock.Mock
}

func (m *mockUserSessionDataManager) CreateUserSession(ctx context.Context, input *auth.UserSessionDatabaseCreationInput) (*auth.UserSession, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*auth.UserSession), args.Error(1)
}

func (m *mockUserSessionDataManager) GetUserSessionBySessionTokenID(ctx context.Context, sessionTokenID string) (*auth.UserSession, error) {
	args := m.Called(ctx, sessionTokenID)
	return args.Get(0).(*auth.UserSession), args.Error(1)
}

func (m *mockUserSessionDataManager) GetUserSessionByRefreshTokenID(ctx context.Context, refreshTokenID string) (*auth.UserSession, error) {
	args := m.Called(ctx, refreshTokenID)
	return args.Get(0).(*auth.UserSession), args.Error(1)
}

func (m *mockUserSessionDataManager) GetActiveSessionsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[auth.UserSession]), args.Error(1)
}

func (m *mockUserSessionDataManager) RevokeUserSession(ctx context.Context, sessionID, userID string) error {
	return m.Called(ctx, sessionID, userID).Error(0)
}

func (m *mockUserSessionDataManager) RevokeAllSessionsForUser(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockUserSessionDataManager) RevokeAllSessionsForUserExcept(ctx context.Context, userID, sessionID string) error {
	return m.Called(ctx, userID, sessionID).Error(0)
}

func (m *mockUserSessionDataManager) UpdateSessionTokenIDs(ctx context.Context, sessionID, newSessionTokenID, newRefreshTokenID string, newExpiresAt time.Time) error {
	return m.Called(ctx, sessionID, newSessionTokenID, newRefreshTokenID, newExpiresAt).Error(0)
}

func (m *mockUserSessionDataManager) TouchSessionLastActive(ctx context.Context, sessionTokenID string) error {
	return m.Called(ctx, sessionTokenID).Error(0)
}

func TestProvideAuthManager(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		queueCfg := &msgconfig.QueuesConfig{DataChangesTopicName: t.Name()}

		mpp := &mockpublishers.PublisherProviderMock{
			ProvidePublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
				return &mockpublishers.PublisherMock{}, nil
			},
		}

		m, err := ProvideAuthManager(
			ctx,
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&mockPasswordResetTokenDataManager{},
			&mockUserSessionDataManager{},
			&identitymock.RepositoryMock{},
			&mockauthn.Authenticator{},
			&mocktotp.VerifierMock{},
			mpp,
			random.NewGenerator(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider()),
			qrcodes.NewBuilder(qrcodes.Issuer("test"), tracingnoop.NewTracerProvider(), loggingnoop.NewLogger()),
			queueCfg,
		)

		require.NoError(t, err)
		assert.NotNil(t, m)
	})
}

func TestAuthManager_Self(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		expectedUser := identityfakes.BuildFakeUser()
		expectedUser.ID = userID

		userDataManager := &identitymock.RepositoryMock{}
		userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, userID).Return(expectedUser, nil)

		sessionData := &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		}
		sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
			return sessionData, nil
		}

		manager := &AuthManager{
			userDataManager:           userDataManager,
			sessionContextDataFetcher: sessionFetcher,
			logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
			tracer:                    tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.Self(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, expectedUser.Username, result.Username)
		mock.AssertExpectationsForObjects(t, userDataManager)
	})
}

func TestAuthManager_CheckUserPermissions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		accountID := identityfakes.BuildFakeID()

		sessionData := &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID:             userID,
				ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
			},
			ActiveAccountID: accountID,
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{
				accountID: authorization.NewAccountRolePermissionChecker(nil),
			},
		}
		sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
			return sessionData, nil
		}

		manager := &AuthManager{
			sessionContextDataFetcher: sessionFetcher,
			logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
			tracer:                    tracing.NewTracerForTest("auth_manager"),
		}

		input := &auth.UserPermissionsRequestInput{
			Permissions: []string{"meal_planning:read"},
		}

		result, err := manager.CheckUserPermissions(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.NotNil(t, result.Permissions)
	})

	t.Run("session fetch error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
			return nil, errors.New("session error")
		}

		manager := &AuthManager{
			sessionContextDataFetcher: sessionFetcher,
			logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
			tracer:                    tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.CheckUserPermissions(ctx, &auth.UserPermissionsRequestInput{Permissions: []string{"read"}})

		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestProvideAuthManager_NilConfig(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	mpp := &mockpublishers.PublisherProviderMock{}

	m, err := ProvideAuthManager(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&mockPasswordResetTokenDataManager{},
		&mockUserSessionDataManager{},
		&identitymock.RepositoryMock{},
		&mockauthn.Authenticator{},
		&mocktotp.VerifierMock{},
		mpp,
		random.NewGenerator(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider()),
		qrcodes.NewBuilder(qrcodes.Issuer("test"), tracingnoop.NewTracerProvider(), loggingnoop.NewLogger()),
		nil, // nil config
	)

	assert.Error(t, err)
	assert.Nil(t, m)
}

func TestAuthManager_Self_SessionError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
		return nil, errors.New("session error")
	}

	manager := &AuthManager{
		sessionContextDataFetcher: sessionFetcher,
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.Self(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthManager_Self_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := identityfakes.BuildFakeID()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, userID).Return((*identity.User)(nil), sql.ErrNoRows)

	sessionFetcher := func(context.Context) (*sessions.ContextData, error) {
		return &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}, nil
	}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: sessionFetcher,
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.Self(ctx)

	assert.Error(t, err)
	assert.Nil(t, result)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_TOTPSecretVerification_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	key, err := totp.Generate(totp.GenerateOpts{Issuer: "test", AccountName: "user"})
	require.NoError(t, err)

	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecret = key.Secret()
	user.TwoFactorSecretVerifiedAt = nil

	token, err := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	require.NoError(t, err)

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserWithUnverifiedTwoFactorSecret), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.MarkUserTwoFactorSecretAsVerified), testutils.ContextMatcher, user.ID).Return(nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: token}
	err = manager.TOTPSecretVerification(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_TOTPSecretVerification_InvalidInput(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	manager := &AuthManager{
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.TOTPSecretVerification(ctx, &auth.TOTPSecretVerificationInput{UserID: "", TOTPToken: "123"})

	assert.Error(t, err)
}

func TestAuthManager_TOTPSecretVerification_AlreadyVerified(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	verifiedAt := time.Now()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = &verifiedAt

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserWithUnverifiedTwoFactorSecret), testutils.ContextMatcher, user.ID).Return(user, nil)

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: "123456"}
	err := manager.TOTPSecretVerification(ctx, input)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already verified")
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_RequestUsernameReminder_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeUsernameReminderRequestInput()
	input.EmailAddress = user.EmailAddress

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmail), testutils.ContextMatcher, input.EmailAddress).Return(user, nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_RequestUsernameReminder_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakeUsernameReminderRequestInput()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmail), testutils.ContextMatcher, input.EmailAddress).Return((*identity.User)(nil), sql.ErrNoRows)

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_CreatePasswordResetToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakePasswordResetTokenCreationRequestInput()
	input.EmailAddress = user.EmailAddress

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmail), testutils.ContextMatcher, input.EmailAddress).Return(user, nil)

	secretGen := &randommock.GeneratorMock{
		GenerateBase32EncodedStringFunc: func(_ context.Context, _ int) (string, error) {
			return "faketoken123", nil
		},
	}

	prtManager := &mockPasswordResetTokenDataManager{}
	createdToken := authfakes.BuildFakePasswordResetToken()
	createdToken.BelongsToUser = user.ID
	prtManager.On(reflection.GetMethodName(prtManager.CreatePasswordResetToken), testutils.ContextMatcher, mock.Anything).Return(createdToken, nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		userDataManager:               userDataManager,
		passwordResetTokenDataManager: prtManager,
		secretGenerator:               secretGen,
		dataChangesPublisher:          publisher,
		sessionContextDataFetcher:     func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                        loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                        tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager, prtManager)
}

func TestAuthManager_CreatePasswordResetToken_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenCreationRequestInput()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmail), testutils.ContextMatcher, input.EmailAddress).Return((*identity.User)(nil), sql.ErrNoRows)

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	// Returns success without sending email to avoid email enumeration.
	assert.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_RequestEmailVerificationEmail_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := identityfakes.BuildFakeID()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetEmailAddressVerificationTokenForUser), testutils.ContextMatcher, userID).Return("verification-token-123", nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestEmailVerificationEmail(ctx)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_VerifyUserEmailAddress_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeEmailAddressVerificationRequestInput()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmailAddressVerificationToken), testutils.ContextMatcher, input.Token).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.MarkUserEmailAddressAsVerified), testutils.ContextMatcher, user.ID, input.Token).Return(nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_VerifyUserEmailAddressByToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := "verification-token"

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmailAddressVerificationToken), testutils.ContextMatcher, token).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.MarkUserEmailAddressAsVerified), testutils.ContextMatcher, user.ID, token).Return(nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		userDataManager:      userDataManager,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_VerifyUserEmailAddressByToken_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	token := "invalid-token"

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmailAddressVerificationToken), testutils.ContextMatcher, token).Return((*identity.User)(nil), sql.ErrNoRows)

	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_UpdatePassword_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	password := authfakes.BuildFakePasswordUpdateInput()
	password.CurrentPassword = "current"
	password.NewPassword = "Abcdefghij123!@#$%^&*()"
	password.TOTPToken = ""

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.UpdateUserPassword), testutils.ContextMatcher, user.ID, mock.AnythingOfType("string")).Return(nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.PasswordMatches), testutils.ContextMatcher, user.HashedPassword, "current").Return(true, nil)
	authenticator.On(reflection.GetMethodName(authenticator.HashPassword), testutils.ContextMatcher, "Abcdefghij123!@#$%^&*()").Return("hashed", nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		authenticator:             authenticator,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
}

func TestAuthManager_UpdateUserEmailAddress_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	input := authfakes.BuildFakeUserEmailAddressUpdateInput()
	input.CurrentPassword = "current"

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.UpdateUserEmailAddress), testutils.ContextMatcher, user.ID, input.NewEmailAddress).Return(nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.PasswordMatches), testutils.ContextMatcher, user.HashedPassword, "current").Return(true, nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		authenticator:             authenticator,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserEmailAddress(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
}

func TestAuthManager_UpdateUserUsername_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	input := authfakes.BuildFakeUsernameUpdateInput()
	input.CurrentPassword = "current"

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.UpdateUserUsername), testutils.ContextMatcher, user.ID, input.NewUsername).Return(nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.PasswordMatches), testutils.ContextMatcher, user.HashedPassword, "current").Return(true, nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		authenticator:             authenticator,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserUsername(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
}

func TestAuthManager_PasswordResetTokenRedemption_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := authfakes.BuildFakePasswordResetToken()
	token.BelongsToUser = user.ID
	token.Token = "reset-token-123"
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.Token = token.Token
	input.NewPassword = "Abcdefghij123!@#$%^&*()"

	prtManager := &mockPasswordResetTokenDataManager{}
	prtManager.On(reflection.GetMethodName(prtManager.GetPasswordResetTokenByToken), testutils.ContextMatcher, token.Token).Return(token, nil)
	prtManager.On(reflection.GetMethodName(prtManager.RedeemPasswordResetToken), testutils.ContextMatcher, token.ID).Return(nil)

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.UpdateUserPassword), testutils.ContextMatcher, user.ID, mock.AnythingOfType("string")).Return(nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.HashPassword), testutils.ContextMatcher, "Abcdefghij123!@#$%^&*()").Return("hashed", nil)

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	manager := &AuthManager{
		passwordResetTokenDataManager: prtManager,
		userDataManager:               userDataManager,
		authenticator:                 authenticator,
		dataChangesPublisher:          publisher,
		sessionContextDataFetcher:     func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                        loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                        tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	require.NoError(t, err)
	mock.AssertExpectationsForObjects(t, prtManager, userDataManager, authenticator)
}

func TestAuthManager_NewTOTPSecret_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	verifiedAt := time.Now()
	user.TwoFactorSecretVerifiedAt = &verifiedAt
	input := authfakes.BuildFakeTOTPSecretRefreshInput()
	input.CurrentPassword = "current"
	token, _ := totp.GenerateCode(user.TwoFactorSecret, time.Now().UTC())
	input.TOTPToken = token

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)
	userDataManager.On(reflection.GetMethodName(userDataManager.MarkUserTwoFactorSecretAsUnverified), testutils.ContextMatcher, user.ID, mock.AnythingOfType("string")).Return(nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.PasswordMatches), testutils.ContextMatcher, user.HashedPassword, "current").Return(true, nil)

	totpVerifier := &mocktotp.VerifierMock{
		VerifyFunc: func(_ context.Context, secret, code string) error {
			if secret == user.TwoFactorSecret && code == token {
				return nil
			}
			return platformtotp.ErrInvalidCode
		},
	}

	secretGen := &randommock.GeneratorMock{
		GenerateBase32EncodedStringFunc: func(_ context.Context, _ int) (string, error) {
			return "newsecretencoded", nil
		},
	}

	qrBuilder := qrcodes.NewBuilder(qrcodes.Issuer("test"), tracingnoop.NewTracerProvider(), loggingnoop.NewLogger())

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		authenticator:             authenticator,
		totpVerifier:              totpVerifier,
		secretGenerator:           secretGen,
		qrCodeBuilder:             qrBuilder,
		dataChangesPublisher:      publisher,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.NewTOTPSecret(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newsecretencoded", result.TwoFactorSecret)
	assert.NotEmpty(t, result.TwoFactorQRCode)
	mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
}

func TestAuthManager_PasswordResetTokenRedemption_TokenNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()

	prtManager := &mockPasswordResetTokenDataManager{}
	prtManager.On(reflection.GetMethodName(prtManager.GetPasswordResetTokenByToken), testutils.ContextMatcher, input.Token).Return((*auth.PasswordResetToken)(nil), sql.ErrNoRows)

	manager := &AuthManager{
		passwordResetTokenDataManager: prtManager,
		sessionContextDataFetcher:     func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                        loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                        tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, prtManager)
}

func TestAuthManager_PasswordResetTokenRedemption_InvalidPassword(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := authfakes.BuildFakePasswordResetToken()
	token.BelongsToUser = user.ID
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.Token = token.Token
	input.NewPassword = "a" // too weak for entropy 60

	prtManager := &mockPasswordResetTokenDataManager{}
	prtManager.On(reflection.GetMethodName(prtManager.GetPasswordResetTokenByToken), testutils.ContextMatcher, token.Token).Return(token, nil)

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)

	manager := &AuthManager{
		passwordResetTokenDataManager: prtManager,
		userDataManager:               userDataManager,
		sessionContextDataFetcher:     func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                        loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                        tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, prtManager, userDataManager)
}

func TestAuthManager_VerifyUserEmailAddress_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakeEmailAddressVerificationRequestInput()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUserByEmailAddressVerificationToken), testutils.ContextMatcher, input.Token).Return((*identity.User)(nil), sql.ErrNoRows)

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return &sessions.ContextData{}, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_UpdatePassword_InvalidNewPassword(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	password := authfakes.BuildFakePasswordUpdateInput()
	password.CurrentPassword = "current"
	password.NewPassword = "a" // too weak for entropy 60
	password.TOTPToken = ""

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, user.ID).Return(user, nil)

	authenticator := &mockauthn.Authenticator{}
	authenticator.On(reflection.GetMethodName(authenticator.PasswordMatches), testutils.ContextMatcher, user.HashedPassword, "current").Return(true, nil)

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		authenticator:             authenticator,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	assert.Error(t, err)
	mock.AssertExpectationsForObjects(t, userDataManager, authenticator)
}

func TestAuthManager_NewTOTPSecret_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := identityfakes.BuildFakeID()
	input := authfakes.BuildFakeTOTPSecretRefreshInput()

	userDataManager := &identitymock.RepositoryMock{}
	userDataManager.On(reflection.GetMethodName(userDataManager.GetUser), testutils.ContextMatcher, userID).Return((*identity.User)(nil), sql.ErrNoRows)

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	manager := &AuthManager{
		userDataManager:           userDataManager,
		sessionContextDataFetcher: func(context.Context) (*sessions.ContextData, error) { return sessionData, nil },
		logger:                    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:                    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.NewTOTPSecret(ctx, input)

	assert.Error(t, err)
	assert.Nil(t, result)
	mock.AssertExpectationsForObjects(t, userDataManager)
}

func TestAuthManager_GetActiveSessionsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		filter := filtering.DefaultQueryFilter()

		expected := &filtering.QueryFilteredResult[auth.UserSession]{
			Data: []*auth.UserSession{
				{ID: identityfakes.BuildFakeID(), BelongsToUser: userID},
			},
		}

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.GetActiveSessionsForUser), testutils.ContextMatcher, userID, filter).Return(expected, nil)

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})

	t.Run("nil filter defaults", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()

		expected := &filtering.QueryFilteredResult[auth.UserSession]{
			Data: []*auth.UserSession{},
		}

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.GetActiveSessionsForUser), testutils.ContextMatcher, userID, filtering.DefaultQueryFilter()).Return(expected, nil)

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		filter := filtering.DefaultQueryFilter()

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.GetActiveSessionsForUser), testutils.ContextMatcher, userID, filter).Return((*filtering.QueryFilteredResult[auth.UserSession])(nil), errors.New("db error"))

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, filter)

		assert.Error(t, err)
		assert.Nil(t, result)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})
}

func TestAuthManager_RevokeSession(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := identityfakes.BuildFakeID()
		userID := identityfakes.BuildFakeID()

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.RevokeUserSession), testutils.ContextMatcher, sessionID, userID).Return(nil)

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeSession(ctx, sessionID, userID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := identityfakes.BuildFakeID()
		userID := identityfakes.BuildFakeID()

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.RevokeUserSession), testutils.ContextMatcher, sessionID, userID).Return(errors.New("db error"))

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeSession(ctx, sessionID, userID)

		assert.Error(t, err)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})
}

func TestAuthManager_RevokeAllSessionsForUserExcept(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		currentSessionID := identityfakes.BuildFakeID()

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.RevokeAllSessionsForUserExcept), testutils.ContextMatcher, userID, currentSessionID).Return(nil)

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := identityfakes.BuildFakeID()
		currentSessionID := identityfakes.BuildFakeID()

		sessionDM := &mockUserSessionDataManager{}
		sessionDM.On(reflection.GetMethodName(sessionDM.RevokeAllSessionsForUserExcept), testutils.ContextMatcher, userID, currentSessionID).Return(errors.New("db error"))

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID)

		assert.Error(t, err)
		mock.AssertExpectationsForObjects(t, sessionDM)
	})
}
