package authentication

import (
	"context"
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
	oauthmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/mock"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"
	oauth2servercfg "github.com/primandproper/platform-go/v12/authentication/oauth2server/config"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v12/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testIssuer is a loopback address, which is the one case the authorization server accepts an
// http issuer for — every other scheme check would refuse it, and rightly.
const testIssuer = "http://localhost:9000"

// buildTestOAuth2Server assembles a real authorization server over the memory store, with the
// client half reading from clients.
//
// A real one rather than a stub: everything worth asserting about these handlers is protocol
// behavior, and a stub that answered would be asserting the stub.
func buildTestOAuth2Server(t *testing.T, clients oauth.OAuth2ClientDataManager) *oauth2server.Server {
	t.Helper()

	cfg := &oauth2servercfg.Config{Provider: oauth2servercfg.ProviderMemory, Issuer: testIssuer}
	cfg.EnsureDefaults()

	store, err := oauth2servercfg.NewStore(t.Context(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, store.Close()) })

	srv, err := oauth2server.NewServer(testIssuer, &clientRegistryStore{Store: store, clients: clients},
		oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
			return &oauth2server.Subject{ID: "test_user"}, nil
		}),
	)
	require.NoError(t, err)

	return srv
}

func buildTestService(t *testing.T) *service {
	t.Helper()

	s, err := ProvideService(
		loggingnoop.NewLogger(),
		buildTestOAuth2Server(t, &oauthmock.RepositoryMock{}),
		tracingnoop.NewTracerProvider(),
	)
	require.NoError(t, err)

	return s.(*service)
}

func TestProvideService(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		s, err := ProvideService(
			loggingnoop.NewLogger(),
			buildTestOAuth2Server(t, &oauthmock.RepositoryMock{}),
			tracingnoop.NewTracerProvider(),
		)

		assert.NotNil(t, s)
		assert.NoError(t, err)
	})
}

func TestProvideOAuth2Server(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := &oauth2servercfg.Config{Provider: oauth2servercfg.ProviderMemory, Issuer: testIssuer}
		cfg.EnsureDefaults()

		srv, err := ProvideOAuth2Server(
			t.Context(),
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			cfg,
			nil,
			oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
				return &oauth2server.Subject{ID: "test_user"}, nil
			}),
			&oauthmock.RepositoryMock{},
		)

		require.NoError(t, err)
		assert.Equal(t, testIssuer, srv.Issuer())
	})

	T.Run("with an issuer the authorization server refuses", func(t *testing.T) {
		t.Parallel()

		cfg := &oauth2servercfg.Config{Provider: oauth2servercfg.ProviderMemory, Issuer: "http://example.com"}
		cfg.EnsureDefaults()

		srv, err := ProvideOAuth2Server(
			t.Context(),
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			metricsnoop.NewMetricsProvider(),
			cfg,
			nil,
			oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
				return &oauth2server.Subject{ID: "test_user"}, nil
			}),
			&oauthmock.RepositoryMock{},
		)

		assert.Nil(t, srv)
		assert.Error(t, err)
	})
}
