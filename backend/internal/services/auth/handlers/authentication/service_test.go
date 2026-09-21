package authentication

import (
	"context"
	"net/http"
	"testing"

	platformoauth2clients "github.com/primandproper/platform-go/v14/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v14/authentication/oauth2clients/authserver"
	oauth2clientsmock "github.com/primandproper/platform-go/v14/authentication/oauth2clients/mock"
	oauth2servercfg "github.com/primandproper/platform-go/v14/authentication/oauth2serverstore/config"
	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

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
func buildTestOAuth2Server(t *testing.T, clients platformoauth2clients.Store) *oauth2server.Server {
	t.Helper()

	cfg := &oauth2servercfg.Config{Provider: oauth2servercfg.ProviderMemory, Issuer: testIssuer}
	cfg.EnsureDefaults()

	store, err := oauth2servercfg.NewStore(t.Context(), cfg, nil)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, store.Close()) })

	registry, err := authserver.NewStore(store, clients, &databasemock.ClientMock{})
	require.NoError(t, err)

	srv, err := oauth2server.NewServer(testIssuer, registry,
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
		buildTestOAuth2Server(t, &oauth2clientsmock.StoreMock{}),
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
			buildTestOAuth2Server(t, &oauth2clientsmock.StoreMock{}),
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
			// A client rather than nil, which is what the registry decorator wants it
			// for: resolving a client_id runs on Reader(), outside any transaction,
			// because /authorize is not inside one. The store this builds is the memory
			// provider above, so nothing here reaches a database.
			&databasemock.ClientMock{},
			oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
				return &oauth2server.Subject{ID: "test_user"}, nil
			}),
			&oauth2clientsmock.StoreMock{},
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
			&databasemock.ClientMock{},
			oauth2server.SubjectAuthenticatorFunc(func(context.Context, *http.Request) (*oauth2server.Subject, error) {
				return &oauth2server.Subject{ID: "test_user"}, nil
			}),
			&oauth2clientsmock.StoreMock{},
		)

		assert.Nil(t, srv)
		assert.Error(t, err)
	})
}
