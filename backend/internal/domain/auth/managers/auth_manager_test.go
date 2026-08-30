package managers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	mockauthn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	authmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/mock"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v13/authentication/passwordreset"
	passwordresetmock "github.com/primandproper/platform-go/v13/authentication/passwordreset/mock"
	platformtotp "github.com/primandproper/platform-go/v13/authentication/totp"
	mocktotp "github.com/primandproper/platform-go/v13/authentication/totp/mock"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	"github.com/primandproper/platform-go/v13/messagequeue"
	mockpublishers "github.com/primandproper/platform-go/v13/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/qrcodes"
	"github.com/primandproper/platform-go/v13/random"
	randommock "github.com/primandproper/platform-go/v13/random/mock"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/pquerna/otp/totp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideAuthManager(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		queueCfg := &queuescfg.Config{DataChangesTopicName: t.Name()}

		mpp := &mockpublishers.PublisherProviderMock{
			NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
				return &mockpublishers.PublisherMock{}, nil
			},
		}

		m, err := ProvideAuthManager(
			ctx,
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&passwordresetmock.StoreMock{},
			&authmock.UserSessionDataManagerMock{},
			&identitymock.RepositoryMock{},
			&mockauthn.AuthenticatorMock{},
			&mocktotp.VerifierMock{},
			mpp,
			random.NewGenerator(random.WithLogger(loggingnoop.NewLogger()), random.WithTracerProvider(tracingnoop.NewTracerProvider())),
			qrcodes.NewBuilder(qrcodes.Issuer("test"), qrcodes.WithTracerProvider(tracingnoop.NewTracerProvider()), qrcodes.WithLogger(loggingnoop.NewLogger())),
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
		userID := fake.BuildFakeID()
		expectedUser := identityfakes.BuildFakeUser()
		expectedUser.ID = userID

		userDataManager := &identitymock.RepositoryMock{
			GetUserFunc: func(_ context.Context, actualUserID string) (*identity.User, error) {
				assert.Equal(t, userID, actualUserID)
				return expectedUser, nil
			},
		}

		sessionData := &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		}
		ctx = sessions.AttachToContext(ctx, sessionData)

		manager := &AuthManager{
			userDataManager: userDataManager,
			logger:          loggingnoop.NewLogger().WithName("auth_manager"),
			tracer:          tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.Self(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, expectedUser.Username, result.Username)
		assert.Len(t, userDataManager.GetUserCalls(), 1)
	})
}

func TestAuthManager_CheckUserPermissions(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		accountID := fake.BuildFakeID()

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
		ctx = sessions.AttachToContext(ctx, sessionData)

		manager := &AuthManager{
			logger: loggingnoop.NewLogger().WithName("auth_manager"),
			tracer: tracing.NewTracerForTest("auth_manager"),
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

		manager := &AuthManager{
			logger: loggingnoop.NewLogger().WithName("auth_manager"),
			tracer: tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.CheckUserPermissions(ctx, &auth.UserPermissionsRequestInput{Permissions: []string{"read"}})

		require.Error(t, err)
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
		&passwordresetmock.StoreMock{},
		&authmock.UserSessionDataManagerMock{},
		&identitymock.RepositoryMock{},
		&mockauthn.AuthenticatorMock{},
		&mocktotp.VerifierMock{},
		mpp,
		random.NewGenerator(random.WithLogger(loggingnoop.NewLogger()), random.WithTracerProvider(tracingnoop.NewTracerProvider())),
		qrcodes.NewBuilder(qrcodes.Issuer("test"), qrcodes.WithTracerProvider(tracingnoop.NewTracerProvider()), qrcodes.WithLogger(loggingnoop.NewLogger())),
		nil, // nil config
	)

	require.Error(t, err)
	assert.Nil(t, m)
}

func TestAuthManager_Self_SessionError(t *testing.T) {
	t.Parallel()

	ctx := t.Context()

	manager := &AuthManager{
		logger: loggingnoop.NewLogger().WithName("auth_manager"),
		tracer: tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.Self(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
}

func TestAuthManager_Self_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := fake.BuildFakeID()

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, actualUserID string) (*identity.User, error) {
			assert.Equal(t, userID, actualUserID)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}})

	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.Self(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
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

	userDataManager := &identitymock.RepositoryMock{
		GetUserWithUnverifiedTwoFactorSecretFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		MarkUserTwoFactorSecretAsVerifiedFunc: func(_ context.Context, userID string) error {
			assert.Equal(t, user.ID, userID)
			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	totpVerifier := &mocktotp.VerifierMock{
		VerifyFunc: func(_ context.Context, secret, code string) error {
			if secret == user.TwoFactorSecret && code == token {
				return nil
			}
			return platformtotp.ErrInvalidCode
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager:      userDataManager,
		totpVerifier:         totpVerifier,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: token}
	err = manager.TOTPSecretVerification(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserWithUnverifiedTwoFactorSecretCalls(), 1)
	assert.Len(t, userDataManager.MarkUserTwoFactorSecretAsVerifiedCalls(), 1)
}

func TestAuthManager_TOTPSecretVerification_InvalidInput(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		logger: loggingnoop.NewLogger().WithName("auth_manager"),
		tracer: tracing.NewTracerForTest("auth_manager"),
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

	userDataManager := &identitymock.RepositoryMock{
		GetUserWithUnverifiedTwoFactorSecretFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: "123456"}
	err := manager.TOTPSecretVerification(ctx, input)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "already verified")
	assert.Len(t, userDataManager.GetUserWithUnverifiedTwoFactorSecretCalls(), 1)
}

func TestAuthManager_RequestUsernameReminder_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeUsernameReminderRequestInput()
	input.EmailAddress = user.EmailAddress

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailFunc: func(_ context.Context, email string) (*identity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return user, nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager:      userDataManager,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailCalls(), 1)
}

func TestAuthManager_RequestUsernameReminder_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakeUsernameReminderRequestInput()

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailFunc: func(_ context.Context, email string) (*identity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	// A missing user must not leak existence: the flow returns success without sending a reminder.
	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailCalls(), 1)
}

func TestAuthManager_CreatePasswordResetToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakePasswordResetTokenCreationRequestInput()
	input.EmailAddress = user.EmailAddress

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailFunc: func(_ context.Context, email string) (*identity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return user, nil
		},
	}

	issuance := &passwordreset.Issuance{
		Token:  &passwordreset.Token{ID: fake.BuildFakeID(), UserID: user.ID},
		Secret: fake.BuildFakeString(),
	}

	tokenStore := &passwordresetmock.StoreMock{
		IssueFunc: func(_ context.Context, scope tenancy.Scope, userID string, ttl time.Duration) (*passwordreset.Issuance, error) {
			assert.Equal(t, tenancy.Global(), scope)
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, passwordResetTokenLifetime, ttl)
			return issuance, nil
		},
	}

	var published []*audit.DataChangeMessage
	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, msg any, _ ...messagequeue.PublishOption) {
			published = append(published, msg.(*audit.DataChangeMessage))
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager:      userDataManager,
		passwordResetTokens:  tokenStore,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailCalls(), 1)
	assert.Len(t, tokenStore.IssueCalls(), 1)

	// The secret rides on the message because the store keeps only a digest of it, and the
	// email handler has nowhere else to get it.
	require.Len(t, published, 1)
	assert.Equal(t, issuance.Secret, published[0].Context[authkeys.PasswordResetTokenSecretKey])
	assert.Equal(t, issuance.Token.ID, published[0].Context[authkeys.PasswordResetTokenIDKey])
}

func TestAuthManager_CreatePasswordResetToken_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenCreationRequestInput()

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailFunc: func(_ context.Context, email string) (*identity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	// Returns success without sending email to avoid email enumeration.
	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailCalls(), 1)
}

func TestAuthManager_RequestEmailVerificationEmail_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := fake.BuildFakeID()

	userDataManager := &identitymock.RepositoryMock{
		GetEmailAddressVerificationTokenForUserFunc: func(_ context.Context, actualUserID string) (string, error) {
			assert.Equal(t, userID, actualUserID)
			return "verification-token-123", nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager:      userDataManager,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestEmailVerificationEmail(ctx)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetEmailAddressVerificationTokenForUserCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddress_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeEmailAddressVerificationRequestInput()

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailAddressVerificationTokenFunc: func(_ context.Context, token string) (*identity.User, error) {
			assert.Equal(t, input.Token, token)
			return user, nil
		},
		MarkUserEmailAddressAsVerifiedFunc: func(_ context.Context, userID, token string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, input.Token, token)
			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager:      userDataManager,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailAddressVerificationTokenCalls(), 1)
	assert.Len(t, userDataManager.MarkUserEmailAddressAsVerifiedCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddressByToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := "verification-token"

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailAddressVerificationTokenFunc: func(_ context.Context, actualToken string) (*identity.User, error) {
			assert.Equal(t, token, actualToken)
			return user, nil
		},
		MarkUserEmailAddressAsVerifiedFunc: func(_ context.Context, userID, actualToken string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, token, actualToken)
			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	manager := &AuthManager{
		userDataManager:      userDataManager,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserByEmailAddressVerificationTokenCalls(), 1)
	assert.Len(t, userDataManager.MarkUserEmailAddressAsVerifiedCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddressByToken_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	token := "invalid-token"

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailAddressVerificationTokenFunc: func(_ context.Context, actualToken string) (*identity.User, error) {
			assert.Equal(t, token, actualToken)
			return nil, sql.ErrNoRows
		},
	}

	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	require.Error(t, err)
	assert.Len(t, userDataManager.GetUserByEmailAddressVerificationTokenCalls(), 1)
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

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserPasswordFunc: func(_ context.Context, userID, newHash string) error {
			assert.Equal(t, user.ID, userID)
			assert.NotEmpty(t, newHash)
			return nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		PasswordMatchesFunc: func(_ context.Context, hash, plaintext string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, "current", plaintext)
			return true, nil
		},
		HashPasswordFunc: func(_ context.Context, plaintext string) (string, error) {
			assert.Equal(t, "Abcdefghij123!@#$%^&*()", plaintext)
			return "hashed", nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager:      userDataManager,
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, userDataManager.UpdateUserPasswordCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
	assert.Len(t, authenticator.HashPasswordCalls(), 1)
}

func TestAuthManager_UpdateUserEmailAddress_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	input := authfakes.BuildFakeUserEmailAddressUpdateInput()
	input.CurrentPassword = "current"

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserEmailAddressFunc: func(_ context.Context, userID, newEmailAddress string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, input.NewEmailAddress, newEmailAddress)
			return nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		PasswordMatchesFunc: func(_ context.Context, hash, plaintext string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, "current", plaintext)
			return true, nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager:      userDataManager,
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserEmailAddress(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, userDataManager.UpdateUserEmailAddressCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_UpdateUserUsername_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	input := authfakes.BuildFakeUsernameUpdateInput()
	input.CurrentPassword = "current"

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserUsernameFunc: func(_ context.Context, userID, newUsername string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, input.NewUsername, newUsername)
			return nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		PasswordMatchesFunc: func(_ context.Context, hash, plaintext string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, "current", plaintext)
			return true, nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager:      userDataManager,
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserUsername(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, userDataManager.UpdateUserUsernameCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_PasswordResetTokenRedemption_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := &passwordreset.Token{ID: fake.BuildFakeID(), UserID: user.ID}
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "Abcdefghij123!@#$%^&*()"

	tokenStore := &passwordresetmock.StoreMock{
		ConsumeFunc: func(_ context.Context, scope tenancy.Scope, secret string) (*passwordreset.Token, error) {
			assert.Equal(t, tenancy.Global(), scope)
			assert.Equal(t, input.Token, secret)
			return token, nil
		},
		RevokeForUserFunc: func(_ context.Context, _ tenancy.Scope, userID string) (int64, error) {
			assert.Equal(t, user.ID, userID)
			return 0, nil
		},
	}

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserPasswordFunc: func(_ context.Context, userID, newHash string) error {
			assert.Equal(t, user.ID, userID)
			assert.NotEmpty(t, newHash)
			return nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		HashPasswordFunc: func(_ context.Context, plaintext string) (string, error) {
			assert.Equal(t, "Abcdefghij123!@#$%^&*()", plaintext)
			return "hashed", nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		passwordResetTokens:  tokenStore,
		userDataManager:      userDataManager,
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	require.NoError(t, err)
	assert.Len(t, tokenStore.ConsumeCalls(), 1)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, userDataManager.UpdateUserPasswordCalls(), 1)
	assert.Len(t, authenticator.HashPasswordCalls(), 1)
	// A completed reset takes the user's other outstanding links with it.
	assert.Len(t, tokenStore.RevokeForUserCalls(), 1)
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

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		MarkUserTwoFactorSecretAsUnverifiedFunc: func(_ context.Context, userID, newSecret string) error {
			assert.Equal(t, user.ID, userID)
			assert.NotEmpty(t, newSecret)
			return nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		PasswordMatchesFunc: func(_ context.Context, hash, plaintext string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, "current", plaintext)
			return true, nil
		},
	}

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

	qrBuilder := qrcodes.NewBuilder(qrcodes.Issuer("test"), qrcodes.WithTracerProvider(tracingnoop.NewTracerProvider()), qrcodes.WithLogger(loggingnoop.NewLogger()))

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager:      userDataManager,
		authenticator:        authenticator,
		totpVerifier:         totpVerifier,
		secretGenerator:      secretGen,
		qrCodeBuilder:        qrBuilder,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.NewTOTPSecret(ctx, input)

	require.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "newsecretencoded", result.TwoFactorSecret)
	assert.NotEmpty(t, result.TwoFactorQRCode)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, userDataManager.MarkUserTwoFactorSecretAsUnverifiedCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_PasswordResetTokenRedemption_TokenNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "Abcdefghij123!@#$%^&*()"

	tokenStore := &passwordresetmock.StoreMock{
		ConsumeFunc: func(_ context.Context, _ tenancy.Scope, secret string) (*passwordreset.Token, error) {
			assert.Equal(t, input.Token, secret)
			return nil, passwordreset.ErrTokenNotFound
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		passwordResetTokens: tokenStore,
		logger:              loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:              tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	require.Error(t, err)
	assert.Len(t, tokenStore.ConsumeCalls(), 1)
}

func TestAuthManager_PasswordResetTokenRedemption_TokenAlreadyRedeemed(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "Abcdefghij123!@#$%^&*()"

	userDataManager := &identitymock.RepositoryMock{}

	tokenStore := &passwordresetmock.StoreMock{
		ConsumeFunc: func(_ context.Context, _ tenancy.Scope, _ string) (*passwordreset.Token, error) {
			return nil, passwordreset.ErrTokenRedeemed
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		passwordResetTokens: tokenStore,
		userDataManager:     userDataManager,
		logger:              loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:              tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	// A token spent once is refused by the store, and nothing downstream of it runs: the
	// password is never written for a link somebody else already answered.
	require.ErrorIs(t, err, passwordreset.ErrTokenRedeemed)
	assert.Empty(t, userDataManager.UpdateUserPasswordCalls())
}

func TestAuthManager_PasswordResetTokenRedemption_InvalidPassword(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "a" // too weak for entropy 60

	tokenStore := &passwordresetmock.StoreMock{}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		passwordResetTokens: tokenStore,
		logger:              loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:              tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	require.Error(t, err)
	// The password is vetted before the token is spent, so a rejected password does not
	// cost the user their link.
	assert.Empty(t, tokenStore.ConsumeCalls())
}

func TestAuthManager_VerifyUserEmailAddress_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakeEmailAddressVerificationRequestInput()

	userDataManager := &identitymock.RepositoryMock{
		GetUserByEmailAddressVerificationTokenFunc: func(_ context.Context, token string) (*identity.User, error) {
			assert.Equal(t, input.Token, token)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	require.Error(t, err)
	assert.Len(t, userDataManager.GetUserByEmailAddressVerificationTokenCalls(), 1)
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

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
	}

	authenticator := &mockauthn.AuthenticatorMock{
		PasswordMatchesFunc: func(_ context.Context, hash, plaintext string) (bool, error) {
			assert.Equal(t, user.HashedPassword, hash)
			assert.Equal(t, "current", plaintext)
			return true, nil
		},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: user.ID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager: userDataManager,
		authenticator:   authenticator,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	require.Error(t, err)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_NewTOTPSecret_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := fake.BuildFakeID()
	input := authfakes.BuildFakeTOTPSecretRefreshInput()

	userDataManager := &identitymock.RepositoryMock{
		GetUserFunc: func(_ context.Context, actualUserID string) (*identity.User, error) {
			assert.Equal(t, userID, actualUserID)
			return nil, sql.ErrNoRows
		},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		userDataManager: userDataManager,
		logger:          loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:          tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.NewTOTPSecret(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Len(t, userDataManager.GetUserCalls(), 1)
}

func TestAuthManager_GetActiveSessionsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		filter := filtering.DefaultQueryFilter()

		expected := &filtering.QueryFilteredResult[auth.UserSession]{
			Data: []*auth.UserSession{
				{ID: fake.BuildFakeID(), BelongsToUser: userID},
			},
		}

		sessionDM := &authmock.UserSessionDataManagerMock{
			GetActiveSessionsForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, filter, actualFilter)
				return expected, nil
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, sessionDM.GetActiveSessionsForUserCalls(), 1)
	})

	t.Run("nil filter defaults", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()

		expected := &filtering.QueryFilteredResult[auth.UserSession]{
			Data: []*auth.UserSession{},
		}

		sessionDM := &authmock.UserSessionDataManagerMock{
			GetActiveSessionsForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, filtering.DefaultQueryFilter(), actualFilter)
				return expected, nil
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, nil)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, sessionDM.GetActiveSessionsForUserCalls(), 1)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		filter := filtering.DefaultQueryFilter()

		sessionDM := &authmock.UserSessionDataManagerMock{
			GetActiveSessionsForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[auth.UserSession], error) {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, filter, actualFilter)
				return nil, errors.New("db error")
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, filter)

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Len(t, sessionDM.GetActiveSessionsForUserCalls(), 1)
	})
}

func TestAuthManager_RevokeSession(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := fake.BuildFakeID()
		userID := fake.BuildFakeID()

		sessionDM := &authmock.UserSessionDataManagerMock{
			RevokeUserSessionFunc: func(_ context.Context, actualSessionID, actualUserID string) error {
				assert.Equal(t, sessionID, actualSessionID)
				assert.Equal(t, userID, actualUserID)
				return nil
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeSession(ctx, sessionID, userID)

		require.NoError(t, err)
		assert.Len(t, sessionDM.RevokeUserSessionCalls(), 1)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := fake.BuildFakeID()
		userID := fake.BuildFakeID()

		sessionDM := &authmock.UserSessionDataManagerMock{
			RevokeUserSessionFunc: func(_ context.Context, actualSessionID, actualUserID string) error {
				assert.Equal(t, sessionID, actualSessionID)
				assert.Equal(t, userID, actualUserID)
				return errors.New("db error")
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeSession(ctx, sessionID, userID)

		require.Error(t, err)
		assert.Len(t, sessionDM.RevokeUserSessionCalls(), 1)
	})
}

func TestAuthManager_RevokeAllSessionsForUserExcept(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		currentSessionID := fake.BuildFakeID()

		sessionDM := &authmock.UserSessionDataManagerMock{
			RevokeAllSessionsForUserExceptFunc: func(_ context.Context, actualUserID, sessionID string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, currentSessionID, sessionID)
				return nil
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID)

		require.NoError(t, err)
		assert.Len(t, sessionDM.RevokeAllSessionsForUserExceptCalls(), 1)
	})

	t.Run("error from data manager", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		currentSessionID := fake.BuildFakeID()

		sessionDM := &authmock.UserSessionDataManagerMock{
			RevokeAllSessionsForUserExceptFunc: func(_ context.Context, actualUserID, sessionID string) error {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, currentSessionID, sessionID)
				return errors.New("db error")
			},
		}

		manager := &AuthManager{
			sessionDataManager: sessionDM,
			tracer:             tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID)

		require.Error(t, err)
		assert.Len(t, sessionDM.RevokeAllSessionsForUserExceptCalls(), 1)
	})
}
