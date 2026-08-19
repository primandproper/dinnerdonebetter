package authentication

import (
	"errors"
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"

	"github.com/primandproper/platform-go/v11/authentication/tokens/paseto"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"
	"github.com/primandproper/platform-go/v11/random"

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

		ctx := t.Context()
		signingKey := random.MustGenerateRawBytes(ctx, 32)
		tokenIssuer, err := paseto.NewSigner("dinner-done-better", t.Name(), signingKey, paseto.WithLogger(logger), paseto.WithTracerProvider(tracerProvider))
		require.NoError(t, err)

		cfg := &OAuth2Config{
			Domain: "example.com",
		}
		dataManager := &oauthmock.RepositoryMock{}
		manager := ProvideOAuth2ClientManager(logger, tracerProvider, cfg, dataManager)

		server := ProvideOAuth2ServerImplementation(logger, tracerProvider, tokenIssuer, manager)

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

		require.NoError(t, err)
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

		require.NoError(t, err)
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

		require.Error(t, err)
		assert.Equal(t, oauth2errors.ErrInvalidClient, err)
		assert.Empty(t, clientID)
		assert.Empty(t, clientSecret)
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

		require.NoError(t, err)
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

		require.Error(t, err)
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

		require.NoError(t, err)
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

		require.NoError(t, err)
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

		require.NoError(t, err)
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
