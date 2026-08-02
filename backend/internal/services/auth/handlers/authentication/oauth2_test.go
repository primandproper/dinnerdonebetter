package authentication

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	mockauthn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitymanagermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"

	"github.com/primandproper/platform-go/v9/authentication/tokens/paseto"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	"github.com/primandproper/platform-go/v9/random"

	oauth2errors "github.com/go-oauth2/oauth2/v4/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvideOAuth2ClientManager(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		cfg := &OAuth2Config{
			Domain: "example.com",
		}
		dataManager := &oauthmock.RepositoryMock{}

		manager := ProvideOAuth2ClientManager(logger, tracerProvider, cfg, dataManager)

		assert.NotNil(t, manager)
	})
}

func TestProvideOAuth2ServerImplementation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		identityDataManager := &identitymanagermock.IdentityDataManagerMock{}
		authenticator := &mockauthn.AuthenticatorMock{}

		ctx := t.Context()
		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewSigner("dinner-done-better", t.Name(), signingKey, paseto.WithLogger(logger), paseto.WithTracerProvider(tracerProvider))
		require.NoError(t, err)

		cfg := &OAuth2Config{
			Domain: "example.com",
		}
		dataManager := &oauthmock.RepositoryMock{}
		manager := ProvideOAuth2ClientManager(logger, tracerProvider, cfg, dataManager)

		server := ProvideOAuth2ServerImplementation(logger, tracerProvider, identityDataManager, authenticator, tokenIssuer, manager)

		assert.NotNil(t, server)
	})
}

func TestBuildOAuth2ErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		handler := buildOAuth2ErrorHandler(logger)

		assert.NotNil(t, handler)

		// Test that the handler doesn't panic
		assert.NotPanics(t, func() {
			handler(&oauth2errors.Response{
				Error: errors.New("test error"),
			})
		})
	})
}

func TestBuildInternalErrorHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		handler := buildInternalErrorHandler(logger)

		assert.NotNil(t, handler)

		testErr := errors.New("test error")
		result := handler(testErr)

		assert.NotNil(t, result)
		assert.Equal(t, testErr, result.Error)
		assert.Equal(t, -1, result.ErrorCode)
		assert.Equal(t, testErr.Error(), result.Description)
		assert.Equal(t, http.StatusInternalServerError, result.StatusCode)
	})
}

func TestBuildClientInfoHandler(T *testing.T) {
	T.Parallel()

	T.Run("with form data", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Form: url.Values{
				"client_id":     []string{"test-client-id"},
				"client_secret": []string{"test-client-secret"},
			},
		}

		clientID, clientSecret, err := handler(req)

		assert.NoError(t, err)
		assert.Equal(t, "test-client-id", clientID)
		assert.Equal(t, "test-client-secret", clientSecret)
	})

	T.Run("with basic auth", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Basic dGVzdC1jbGllbnQtaWQ6dGVzdC1jbGllbnQtc2VjcmV0"},
			},
			Form: url.Values{},
		}

		clientID, clientSecret, err := handler(req)

		assert.NoError(t, err)
		assert.Equal(t, "test-client-id", clientID)
		assert.Equal(t, "test-client-secret", clientSecret)
	})

	T.Run("with no auth", func(t *testing.T) {
		t.Parallel()

		handler := buildClientInfoHandler()
		assert.NotNil(t, handler)

		req := &http.Request{
			Form: url.Values{},
		}

		clientID, clientSecret, err := handler(req)

		assert.Error(t, err)
		assert.Equal(t, oauth2errors.ErrInvalidClient, err)
		assert.Empty(t, clientID)
		assert.Empty(t, clientSecret)
	})
}

func TestBuildPasswordAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, user.HashedPassword, hash)
				assert.Equal(t, "password", password)
				return true, nil
			},
		}

		dataManager := &identitymanagermock.IdentityDataManagerMock{
			GetUserByUsernameFunc: func(_ context.Context, username string) (*identity.User, error) {
				assert.Equal(t, user.Username, username)
				return user, nil
			},
		}

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "password")

		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, dataManager.GetUserByUsernameCalls(), 1)
	})

	T.Run("with invalid username", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		authenticator := &mockauthn.AuthenticatorMock{}
		dataManager := &identitymanagermock.IdentityDataManagerMock{
			GetUserByUsernameFunc: func(_ context.Context, username string) (*identity.User, error) {
				assert.Equal(t, "unknown-user", username)
				return nil, errors.New("user not found")
			},
		}

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", "unknown-user", "password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		assert.Empty(t, authenticator.PasswordMatchesCalls())
		assert.Len(t, dataManager.GetUserByUsernameCalls(), 1)
	})

	T.Run("with invalid credentials", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, user.HashedPassword, hash)
				assert.Equal(t, "wrong-password", password)
				return false, nil
			},
		}

		dataManager := &identitymanagermock.IdentityDataManagerMock{
			GetUserByUsernameFunc: func(_ context.Context, username string) (*identity.User, error) {
				assert.Equal(t, user.Username, username)
				return user, nil
			},
		}

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "wrong-password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, dataManager.GetUserByUsernameCalls(), 1)
	})

	T.Run("with validation error", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		user := fakes.BuildFakeUser()

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, user.HashedPassword, hash)
				assert.Equal(t, "password", password)
				return false, errors.New("validation error")
			},
		}

		dataManager := &identitymanagermock.IdentityDataManagerMock{
			GetUserByUsernameFunc: func(_ context.Context, username string) (*identity.User, error) {
				assert.Equal(t, user.Username, username)
				return user, nil
			},
		}

		handler := buildPasswordAuthorizationHandler(logger, authenticator, dataManager)
		assert.NotNil(t, handler)

		userID, err := handler(ctx, "client-id", user.Username, "password")

		assert.Error(t, err)
		assert.Empty(t, userID)
		assert.Contains(t, err.Error(), "invalid username or password")

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
		assert.Len(t, dataManager.GetUserByUsernameCalls(), 1)
	})
}

func TestBuildUserAuthorizationHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewSigner("dinner-done-better", t.Name(), signingKey, paseto.WithLogger(logger), paseto.WithTracerProvider(tracingnoop.NewTracerProvider()))
		require.NoError(t, err)

		user := fakes.BuildFakeUser()
		token, _, err := tokenIssuer.IssueToken(ctx, user.ID, time.Hour, nil)
		require.NoError(t, err)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Bearer " + token},
			},
		}
		req = req.WithContext(ctx)

		handler := buildUserAuthorizationHandler(tracer, logger, tokenIssuer)
		assert.NotNil(t, handler)

		userID, err := handler(nil, req)

		assert.NoError(t, err)
		assert.Equal(t, user.ID, userID)
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()
		tracer := tracing.NewTracerForTest("test")

		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewSigner("dinner-done-better", t.Name(), signingKey, paseto.WithLogger(logger), paseto.WithTracerProvider(tracingnoop.NewTracerProvider()))
		require.NoError(t, err)

		req := &http.Request{
			Header: http.Header{
				"Authorization": []string{"Bearer invalid-token"},
			},
		}
		req = req.WithContext(ctx)

		handler := buildUserAuthorizationHandler(tracer, logger, tokenIssuer)
		assert.NotNil(t, handler)

		userID, err := handler(nil, req)

		assert.Error(t, err)
		assert.Equal(t, oauth2errors.ErrAccessDenied, err)
		assert.Empty(t, userID)
	})
}

func TestAuthorizeScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		handler := AuthorizeScopeHandler(logger)
		assert.NotNil(t, handler)

		req := &http.Request{
			URL: &url.URL{
				RawQuery: "scope=read%20write",
			},
		}

		scope, err := handler(nil, req)

		assert.NoError(t, err)
		assert.Equal(t, "read write", scope)
	})
}

func TestAccessTokenExpHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		handler := AccessTokenExpHandler(logger)
		assert.NotNil(t, handler)

		duration, err := handler(nil, nil)

		assert.NoError(t, err)
		assert.Equal(t, 24*time.Hour, duration)
	})
}

func TestClientScopeHandler(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		handler := ClientScopeHandler(logger)
		assert.NotNil(t, handler)

		allowed, err := handler(nil)

		assert.NoError(t, err)
		assert.True(t, allowed)
	})
}

func TestValidateRedirectURI(T *testing.T) {
	T.Parallel()

	T.Run("with exact host match", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, validateRedirectURI("https://dinnerdonebetter.com", "https://dinnerdonebetter.com/callback"))
	})

	T.Run("with subdomain", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, validateRedirectURI("https://dinnerdonebetter.com", "https://api.dinnerdonebetter.com/callback"))
	})

	T.Run("with differing ports on the same host", func(t *testing.T) {
		t.Parallel()

		assert.NoError(t, validateRedirectURI("http://localhost:9000", "http://localhost:55123/callback"))
	})

	T.Run("with foreign host", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, validateRedirectURI("https://dinnerdonebetter.com", "https://evil.example.com/callback"))
	})

	T.Run("with suffix-spoofed host", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, validateRedirectURI("https://dinnerdonebetter.com", "https://evildinnerdonebetter.com/callback"))
	})

	T.Run("with empty redirect host", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, validateRedirectURI("https://dinnerdonebetter.com", "/relative/path"))
	})

	T.Run("with unparseable URI", func(t *testing.T) {
		t.Parallel()

		assert.Error(t, validateRedirectURI("https://dinnerdonebetter.com", "://not-a-uri"))
	})
}
