package managers

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	mockauthn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	authfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/fakes"
	authkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/keys"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v14/authentication/passwordreset"
	passwordresetmock "github.com/primandproper/platform-go/v14/authentication/passwordreset/mock"
	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	platformsessions "github.com/primandproper/platform-go/v14/sessions"
	sessionsmock "github.com/primandproper/platform-go/v14/sessions/mock"
	mocktokens "github.com/primandproper/primitives-go/v2/authentication/tokens/mock"
	platformtotp "github.com/primandproper/primitives-go/v2/authentication/totp"
	mocktotp "github.com/primandproper/primitives-go/v2/authentication/totp/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/messagequeue"
	mockpublishers "github.com/primandproper/primitives-go/v2/messagequeue/mock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/qrcodes"
	"github.com/primandproper/primitives-go/v2/random"
	randommock "github.com/primandproper/primitives-go/v2/random/mock"
	"github.com/primandproper/primitives-go/v2/tenancy"

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
			testutils.MockDatabaseClient(),
			&passwordresetmock.StoreMock{},
			&sessionsmock.StoreMock[auth.SessionPayload]{},
			directoryForTest(t, &identitymock.StoreMock{}),
			&identitymock.StoreMock{},
			signInForTest(t, &identitymock.StoreMock{}, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
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

		userStore := &identitymock.StoreMock{
			GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualUserID string) (*platformidentity.User, error) {
				assert.Equal(t, userID, actualUserID)
				return expectedUser, nil
			},
		}

		sessionData := &sessions.ContextData{
			Requester: sessions.RequesterInfo{UserID: userID},
		}
		ctx = sessions.AttachToContext(ctx, sessionData)

		manager := &AuthManager{
			db:        testutils.MockDatabaseClient(),
			users:     userStore,
			directory: directoryForTest(t, userStore),
			signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
			logger:    loggingnoop.NewLogger().WithName("auth_manager"),
			tracer:    tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.Self(ctx)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, userID, result.ID)
		assert.Equal(t, expectedUser.Username, result.Username)
		assert.Len(t, userStore.GetUserCalls(), 1)
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
			db:     testutils.MockDatabaseClient(),
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
			db:     testutils.MockDatabaseClient(),
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
		testutils.MockDatabaseClient(),
		&passwordresetmock.StoreMock{},
		&sessionsmock.StoreMock[auth.SessionPayload]{},
		directoryForTest(t, &identitymock.StoreMock{}),
		&identitymock.StoreMock{},
		signInForTest(t, &identitymock.StoreMock{}, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
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
		db:     testutils.MockDatabaseClient(),
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualUserID string) (*platformidentity.User, error) {
			assert.Equal(t, userID, actualUserID)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}})

	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.Self(ctx)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Len(t, userStore.GetUserCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		MarkUserTwoFactorSecretVerifiedFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, totpVerifier),
		totpVerifier:         totpVerifier,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: token}
	err = manager.TOTPSecretVerification(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userStore.GetUserCalls(), 1)
	assert.Len(t, userStore.MarkUserTwoFactorSecretVerifiedCalls(), 1)
}

func TestAuthManager_TOTPSecretVerification_InvalidInput(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:     testutils.MockDatabaseClient(),
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
	}

	userStore.MarkUserTwoFactorSecretVerifiedFunc = func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
		assert.Equal(t, user.ID, userID)

		return user, nil
	}

	verifier := &mocktotp.VerifierMock{
		VerifyFunc: func(context.Context, string, string) error { return nil },
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, verifier),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
		dataChangesPublisher: &mockpublishers.PublisherMock{
			PublishAsyncFunc: func(context.Context, any, ...messagequeue.PublishOption) {},
		},
	}

	input := &auth.TOTPSecretVerificationInput{UserID: user.ID, TOTPToken: "123456"}
	err := manager.TOTPSecretVerification(ctx, input)

	// Verifying an already-proven secret with a valid code now succeeds rather than
	// refusing, and the refusal is not missed.
	//
	// This used to answer "two factor secret already verified" before it looked at the
	// code, which made the endpoint a way to ask whether any user id had proven a second
	// factor — it takes the subject from the request body rather than from the session, so
	// the id did not have to be the caller's. platform checks the code first, so there is
	// no answer to read without one. Re-stamping a verification that already happened is
	// the harmless half of that trade.
	require.NoError(t, err)
	assert.Len(t, userStore.GetUserCalls(), 1)
	assert.Len(t, userStore.MarkUserTwoFactorSecretVerifiedCalls(), 1)
}

func TestAuthManager_RequestUsernameReminder_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeUsernameReminderRequestInput()
	input.EmailAddress = user.EmailAddress

	userStore := &identitymock.StoreMock{
		GetUserByEmailAddressFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, email string) (*platformidentity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return user, nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userStore.GetUserByEmailAddressCalls(), 1)
}

func TestAuthManager_RequestUsernameReminder_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakeUsernameReminderRequestInput()

	userStore := &identitymock.StoreMock{
		GetUserByEmailAddressFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, email string) (*platformidentity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestUsernameReminder(ctx, input)

	// A missing user must not leak existence: the flow returns success without sending a reminder.
	require.NoError(t, err)
	assert.Len(t, userStore.GetUserByEmailAddressCalls(), 1)
}

func TestAuthManager_CreatePasswordResetToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakePasswordResetTokenCreationRequestInput()
	input.EmailAddress = user.EmailAddress

	userStore := &identitymock.StoreMock{
		GetUserByEmailAddressFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, email string) (*platformidentity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return user, nil
		},
	}

	issuance := &passwordreset.Issuance{
		Token:  &passwordreset.Token{ID: fake.BuildFakeID(), UserID: user.ID},
		Secret: fake.BuildFakeString(),
	}

	tokenStore := &passwordresetmock.StoreMock{
		IssueFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, userID string, ttl time.Duration) (*passwordreset.Issuance, error) {
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		passwordResetTokens:  tokenStore,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	require.NoError(t, err)
	assert.Len(t, userStore.GetUserByEmailAddressCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserByEmailAddressFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, email string) (*platformidentity.User, error) {
			assert.Equal(t, input.EmailAddress, email)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.CreatePasswordResetToken(ctx, input)

	// Returns success without sending email to avoid email enumeration.
	require.NoError(t, err)
	assert.Len(t, userStore.GetUserByEmailAddressCalls(), 1)
}

func TestAuthManager_RequestEmailVerificationEmail_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := fake.BuildFakeID()

	userStore := &identitymock.StoreMock{
		// The service re-reads the user it wrote, so the read has to answer too.
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualUserID string) (*platformidentity.User, error) {
			return &platformidentity.User{ID: actualUserID}, nil
		},
		// The token is minted and stored rather than read back: the column holds a
		// digest, and no read fills the secret in.
		SetUserEmailAddressVerificationTokenFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, actualUserID, token string, expiresAt time.Time) error {
			assert.Equal(t, userID, actualUserID)
			assert.NotEmpty(t, token)

			// A deadline, and one in the future. platform refuses a zero one, so the
			// assertion that matters is not that the column is set but that this
			// application computed something usable: a link minted with a deadline
			// already past is a mail nobody can act on.
			assert.False(t, expiresAt.IsZero())
			assert.True(t, expiresAt.After(time.Now()))

			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		secretGenerator:      secretGeneratorForTest(),
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.RequestEmailVerificationEmail(ctx)

	require.NoError(t, err)
	assert.Len(t, userStore.SetUserEmailAddressVerificationTokenCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddress_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	input := authfakes.BuildFakeEmailAddressVerificationRequestInput()

	userStore := &identitymock.StoreMock{
		// The service re-reads the user it wrote, on the transaction that wrote it,
		// so a mock that stubs only the write is a mock the service cannot finish.
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, _ string) (*platformidentity.User, error) {
			return user, nil
		},
		GetUserByEmailVerificationTokenFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, token string) (*platformidentity.User, error) {
			assert.Equal(t, input.Token, token)
			return user, nil
		},
		MarkUserEmailAddressVerifiedFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID, actualToken string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, input.Token, actualToken)

			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	require.NoError(t, err)
	// Two reads by the token: this manager's, for the user id its event names, and
	// signin's own inside VerifyEmailAddress. The second is the price of an event that
	// names somebody — VerifyEmailAddress answers with an error and nothing else, and
	// by the time it returns the link is spent and the column that found the user is
	// cleared, so the read cannot happen afterwards.
	assert.Len(t, userStore.GetUserByEmailVerificationTokenCalls(), 2)
	assert.Len(t, userStore.MarkUserEmailAddressVerifiedCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddressByToken_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	token := "verification-token"

	userStore := &identitymock.StoreMock{
		// The service re-reads the user it wrote, on the transaction that wrote it,
		// so a mock that stubs only the write is a mock the service cannot finish.
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, _ string) (*platformidentity.User, error) {
			return user, nil
		},
		GetUserByEmailVerificationTokenFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualToken string) (*platformidentity.User, error) {
			assert.Equal(t, token, actualToken)
			return user, nil
		},
		MarkUserEmailAddressVerifiedFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID, actualToken string) error {
			assert.Equal(t, user.ID, userID)
			assert.Equal(t, token, actualToken)
			return nil
		},
	}

	publisher := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any, _ ...messagequeue.PublishOption) {},
	}

	manager := &AuthManager{
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	require.NoError(t, err)
	// Two reads by the token: this manager's, for the user id its event names, and
	// signin's own inside VerifyEmailAddress. The second is the price of an event that
	// names somebody — VerifyEmailAddress answers with an error and nothing else, and
	// by the time it returns the link is spent and the column that found the user is
	// cleared, so the read cannot happen afterwards.
	assert.Len(t, userStore.GetUserByEmailVerificationTokenCalls(), 2)
	assert.Len(t, userStore.MarkUserEmailAddressVerifiedCalls(), 1)
}

func TestAuthManager_VerifyUserEmailAddressByToken_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	token := "invalid-token"

	userStore := &identitymock.StoreMock{
		GetUserByEmailVerificationTokenFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualToken string) (*platformidentity.User, error) {
			assert.Equal(t, token, actualToken)
			return nil, sql.ErrNoRows
		},
	}

	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddressByToken(ctx, token)

	require.Error(t, err)
	assert.Len(t, userStore.GetUserByEmailVerificationTokenCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserPasswordFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID, newHash string) error {
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, authenticator, &mocktotp.VerifierMock{}),
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	require.NoError(t, err)
	// One read. This manager used to make its own before handing the write to the
	// directory service, which read again on either side of it; signin reads once, proves
	// the password against what it read, and writes.
	assert.Len(t, userStore.GetUserCalls(), 1)
	assert.Len(t, userStore.UpdateUserPasswordCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		// The service reads the user, applies the update to it and writes the whole row
		// back, so what lands here is the user as it will be rather than the diff.
		UpdateUserFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, updated *platformidentity.User) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, updated.ID)
			assert.Equal(t, input.NewEmailAddress, updated.EmailAddress)

			return updated, nil
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, authenticator, &mocktotp.VerifierMock{}),
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserEmailAddress(ctx, input)

	require.NoError(t, err)
	// This manager's credential check and the service's own read before it applies the update.
	assert.Len(t, userStore.GetUserCalls(), 2)
	assert.Len(t, userStore.UpdateUserCalls(), 1)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_UpdateUserUsername_Success(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	user := identityfakes.BuildFakeUser()
	user.TwoFactorSecretVerifiedAt = nil
	input := authfakes.BuildFakeUsernameUpdateInput()
	input.CurrentPassword = "current"

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, updated *platformidentity.User) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, updated.ID)

			// The folded spelling, not the one submitted. The directory lowers a handle
			// on write and on every lookup, so two users cannot differ by case alone —
			// and what the column receives is the fold rather than what was typed.
			assert.Equal(t, platformidentity.FoldHandle(input.NewUsername), updated.Username)

			return updated, nil
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, authenticator, &mocktotp.VerifierMock{}),
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdateUserUsername(ctx, input)

	require.NoError(t, err)
	// This manager's credential check and the service's own read before it applies the update.
	assert.Len(t, userStore.GetUserCalls(), 2)
	assert.Len(t, userStore.UpdateUserCalls(), 1)
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
		ConsumeFunc: func(_ context.Context, _ database.Tx, scope tenancy.Scope, secret string) (*passwordreset.Token, error) {
			assert.Equal(t, tenancy.Global(), scope)
			assert.Equal(t, input.Token, secret)
			return token, nil
		},
		RevokeForUserFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID string) (int64, error) {
			assert.Equal(t, user.ID, userID)
			return 0, nil
		},
	}

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserPasswordFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID, newHash string) error {
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
		db:                   testutils.MockDatabaseClient(),
		passwordResetTokens:  tokenStore,
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, authenticator, &mocktotp.VerifierMock{}),
		authenticator:        authenticator,
		dataChangesPublisher: publisher,
		logger:               loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:               tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	require.NoError(t, err)
	assert.Len(t, tokenStore.ConsumeCalls(), 1)
	// One read here and the service's two around its write — see UpdatePassword above.
	assert.Len(t, userStore.GetUserCalls(), 3)
	assert.Len(t, userStore.UpdateUserPasswordCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
			assert.Equal(t, user.ID, userID)
			return user, nil
		},
		UpdateUserTwoFactorSecretFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, userID, newSecret string) error {
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
		db:                   testutils.MockDatabaseClient(),
		users:                userStore,
		directory:            directoryForTest(t, userStore),
		signIn:               signInForTest(t, userStore, authenticator, totpVerifier, secretGen),
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
	assert.NotEmpty(t, result.TwoFactorQRCode)

	// The secret handed back is the secret written, rather than a fixed string this test
	// arranged for. signin mints it with a totp.Generator of its own — not the
	// random.Generator this manager still holds for registration — so pinning a literal
	// would be pinning which generator platform happens to use. What has to be true is
	// that the caller is shown the secret the directory now stores; a response carrying a
	// different one is a QR code nobody can enroll with.
	require.Len(t, userStore.UpdateUserTwoFactorSecretCalls(), 1)
	assert.Equal(t, userStore.UpdateUserTwoFactorSecretCalls()[0].Secret, result.TwoFactorSecret)
	assert.NotEmpty(t, result.TwoFactorSecret)
	// Two reads: signin's, to prove the password and generate against, and this manager's
	// afterwards for the username the QR code is labelled with. The third was the
	// directory service reading back what it had just written, which nothing here wanted.
	assert.Len(t, userStore.GetUserCalls(), 2)
	assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
}

func TestAuthManager_PasswordResetTokenRedemption_TokenNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "Abcdefghij123!@#$%^&*()"

	tokenStore := &passwordresetmock.StoreMock{
		ConsumeFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, secret string) (*passwordreset.Token, error) {
			assert.Equal(t, input.Token, secret)

			return nil, passwordreset.ErrTokenNotFound
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:                  testutils.MockDatabaseClient(),
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

	userStore := &identitymock.StoreMock{}

	tokenStore := &passwordresetmock.StoreMock{
		ConsumeFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, _ string) (*passwordreset.Token, error) {
			return nil, passwordreset.ErrTokenRedeemed
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:                  testutils.MockDatabaseClient(),
		passwordResetTokens: tokenStore,
		users:               userStore,
		directory:           directoryForTest(t, userStore),
		signIn:              signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:              loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:              tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.PasswordResetTokenRedemption(ctx, input)

	// A token spent once is refused by the store, and nothing downstream of it runs: the
	// password is never written for a link somebody else already answered.
	require.ErrorIs(t, err, passwordreset.ErrTokenRedeemed)
	assert.Empty(t, userStore.UpdateUserPasswordCalls())
}

func TestAuthManager_PasswordResetTokenRedemption_InvalidPassword(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	input := authfakes.BuildFakePasswordResetTokenRedemptionRequestInput()
	input.NewPassword = "a" // too weak for entropy 60

	tokenStore := &passwordresetmock.StoreMock{}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:                  testutils.MockDatabaseClient(),
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

	userStore := &identitymock.StoreMock{
		GetUserByEmailVerificationTokenFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, token string) (*platformidentity.User, error) {
			assert.Equal(t, input.Token, token)
			return nil, sql.ErrNoRows
		},
	}

	ctx = sessions.AttachToContext(ctx, &sessions.ContextData{})
	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.VerifyUserEmailAddress(ctx, input)

	require.Error(t, err)
	assert.Len(t, userStore.GetUserByEmailVerificationTokenCalls(), 1)
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

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, userID string) (*platformidentity.User, error) {
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
		db:            testutils.MockDatabaseClient(),
		users:         userStore,
		directory:     directoryForTest(t, userStore),
		signIn:        signInForTest(t, userStore, authenticator, &mocktotp.VerifierMock{}),
		authenticator: authenticator,
		logger:        loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:        tracing.NewTracerForTest("auth_manager"),
	}

	err := manager.UpdatePassword(ctx, password)

	require.Error(t, err)

	// Nothing was read and nothing was compared, because the password policy is checked
	// before the credential is. That is the order this adoption settled on: the entropy
	// floor is a fact about the input, it costs no I/O, and refusing on it first spares a
	// round trip and a hash for a password the caller has to retype anyway.
	assert.Empty(t, userStore.GetUserCalls())
	assert.Empty(t, authenticator.PasswordMatchesCalls())
}

func TestAuthManager_NewTOTPSecret_UserNotFound(t *testing.T) {
	t.Parallel()

	ctx := t.Context()
	userID := fake.BuildFakeID()
	input := authfakes.BuildFakeTOTPSecretRefreshInput()

	userStore := &identitymock.StoreMock{
		GetUserFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, actualUserID string) (*platformidentity.User, error) {
			assert.Equal(t, userID, actualUserID)
			return nil, sql.ErrNoRows
		},
	}

	sessionData := &sessions.ContextData{Requester: sessions.RequesterInfo{UserID: userID}}

	ctx = sessions.AttachToContext(ctx, sessionData)
	manager := &AuthManager{
		db:        testutils.MockDatabaseClient(),
		users:     userStore,
		directory: directoryForTest(t, userStore),
		signIn:    signInForTest(t, userStore, &mockauthn.AuthenticatorMock{}, &mocktotp.VerifierMock{}),
		logger:    loggingnoop.NewLogger().WithName("auth_manager"),
		tracer:    tracing.NewTracerForTest("auth_manager"),
	}

	result, err := manager.NewTOTPSecret(ctx, input)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Len(t, userStore.GetUserCalls(), 1)
}

func TestAuthManager_GetActiveSessionsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		currentSessionID := fake.BuildFakeID()

		expected := []*auth.UserSession{
			{ID: fake.BuildFakeID(), Holder: auth.SessionHolder(userID), IsCurrent: true},
		}

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			ListFunc: func(_ context.Context, holder platformsessions.Holder, actualCurrentID string) ([]*auth.UserSession, error) {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				assert.Equal(t, currentSessionID, actualCurrentID)
				return expected, nil
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, currentSessionID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, store.ListCalls(), 1)
	})

	t.Run("error from session store", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			ListFunc: func(_ context.Context, holder platformsessions.Holder, _ string) ([]*auth.UserSession, error) {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				return nil, errors.New("db error")
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		result, err := manager.GetActiveSessionsForUser(ctx, userID, "")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Len(t, store.ListCalls(), 1)
	})
}

func TestAuthManager_RevokeSession(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := fake.BuildFakeID()
		userID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeFunc: func(_ context.Context, holder platformsessions.Holder, actualSessionID string) error {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				assert.Equal(t, sessionID, actualSessionID)
				return nil
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		require.NoError(t, manager.RevokeSession(ctx, sessionID, userID))
		assert.Len(t, store.RevokeCalls(), 1)
	})

	// A session that is not the named user's is answered as absent rather than as
	// forbidden, and the manager passes that through unchanged: an error that
	// distinguished the two would confirm that somebody else's identifier names something.
	t.Run("with a session the user does not hold", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sessionID := fake.BuildFakeID()
		userID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeFunc: func(_ context.Context, holder platformsessions.Holder, actualSessionID string) error {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				assert.Equal(t, sessionID, actualSessionID)
				return platformsessions.ErrNotFound
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		err := manager.RevokeSession(ctx, sessionID, userID)

		require.ErrorIs(t, err, platformsessions.ErrNotFound)
		assert.Len(t, store.RevokeCalls(), 1)
	})
}

func TestAuthManager_RevokeAllSessionsForUserExcept(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		currentSessionID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeAllExceptFunc: func(_ context.Context, holder platformsessions.Holder, keepID string) (int, error) {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				assert.Equal(t, currentSessionID, keepID)
				return 3, nil
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		require.NoError(t, manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID))
		assert.Len(t, store.RevokeAllExceptCalls(), 1)
	})

	t.Run("error from session store", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()
		currentSessionID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeAllExceptFunc: func(context.Context, platformsessions.Holder, string) (int, error) {
				return 0, errors.New("db error")
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		require.Error(t, manager.RevokeAllSessionsForUserExcept(ctx, userID, currentSessionID))
		assert.Len(t, store.RevokeAllExceptCalls(), 1)
	})
}

func TestAuthManager_RevokeAllSessionsForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeAllFunc: func(_ context.Context, holder platformsessions.Holder) (int, error) {
				assert.Equal(t, auth.SessionHolder(userID), holder)
				return 2, nil
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		require.NoError(t, manager.RevokeAllSessionsForUser(ctx, userID))
		assert.Len(t, store.RevokeAllCalls(), 1)
	})

	t.Run("error from session store", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		userID := fake.BuildFakeID()

		store := &sessionsmock.StoreMock[auth.SessionPayload]{
			RevokeAllFunc: func(context.Context, platformsessions.Holder) (int, error) {
				return 0, errors.New("db error")
			},
		}

		manager := &AuthManager{
			db:           testutils.MockDatabaseClient(),
			sessionStore: store,
			tracer:       tracing.NewTracerForTest("auth_manager"),
		}

		require.Error(t, manager.RevokeAllSessionsForUser(ctx, userID))
		assert.Len(t, store.RevokeAllCalls(), 1)
	})
}

// directoryForTest builds the real identity service over a mocked store.
//
// The manager holds a *identity.Service rather than an interface, and that is not a seam
// this package is missing: the service is platform's, its behaviour is platform's to test,
// and what a unit test here wants to substitute is the store underneath it. So these tests
// build a real one over a mock and assert on what the store was asked for — which is a
// stronger claim than asserting on what a mocked service was told, because the service's
// own rules are in the path.
func directoryForTest(t *testing.T, store platformidentity.Store) *platformidentity.Service {
	t.Helper()

	directory, err := platformidentity.NewService(testutils.MockDatabaseClient(), store)
	require.NoError(t, err)

	return directory
}

// signInForTest builds the real sign-in service over the mocks a test supplies.
//
// The real thing rather than a mock of it, for the reason directoryForTest gives and one
// more: signin.Service is a concrete type with no interface, so the alternative was to
// invent one here. Every argument is already mocked — the identity store is its Directory,
// and the authenticator and verifier are the seams these tests drive.
//
// The token issuer is nil-free and unused: none of the credential methods this manager
// calls mints a token, and passing a mock keeps the constructor's nil checks happy without
// implying that one will be consulted.
func signInForTest(
	t *testing.T,
	store platformidentity.Store,
	authenticator authentication.Authenticator,
	verifier platformtotp.Verifier,
	generators ...random.Generator,
) *signin.Service {
	t.Helper()

	// The secret a refreshed second factor gets is signin's to mint now, so a test that
	// asserts on the value supplies the generator that produces it. The rest take the real
	// one, because a generator is not a seam any of them is about.
	generator := secretGeneratorForTest()
	if len(generators) > 0 {
		generator = generators[0]
	}

	service, err := signin.NewService(
		testutils.MockDatabaseClient(),
		store,
		authenticator,
		&mocktokens.IssuerMock{},
		signin.WithSecondFactorPolicy(signin.SecondFactorWhenEnrolled),
		signin.WithTOTPVerifier(verifier),
		signin.WithTOTPIssuer("test"),
		signin.WithSecretGenerator(generator),
		signin.WithVerifications(store),
	)
	require.NoError(t, err)

	return service
}

// secretGeneratorForTest is the real generator. It is not mocked because what these tests
// assert about a token is that one was minted and handed to the store, and a generator is
// not a seam any of them is about.
func secretGeneratorForTest() random.Generator {
	return random.NewGenerator(
		random.WithLogger(loggingnoop.NewLogger()),
		random.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)
}
