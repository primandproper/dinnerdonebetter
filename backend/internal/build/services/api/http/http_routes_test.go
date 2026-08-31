package api

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentswebhook "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/http"

	"github.com/primandproper/platform-go/v13/encoding"
	"github.com/primandproper/platform-go/v13/healthcheck"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/routing"
	"github.com/primandproper/platform-go/v13/routing/backends/chi"
	routingcfg "github.com/primandproper/platform-go/v13/routing/config"
	"github.com/primandproper/platform-go/v13/version"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// stubRegistry answers CheckAll with whatever the test wants the probe to see.
type stubRegistry struct {
	result *healthcheck.Result
}

func (r *stubRegistry) Register(healthcheck.Checker) {}

func (r *stubRegistry) CheckAll(context.Context) *healthcheck.Result { return r.result }

func buildOpsHandler(t *testing.T, registry healthcheck.Registry) http.Handler {
	t.Helper()

	logger := loggingnoop.NewLogger()

	backend := chi.NewBackend(
		&chi.Config{ServiceName: t.Name(), SilenceRouteLogging: true},
		chi.WithLogger(logger),
		chi.WithTracerProvider(tracingnoop.NewTracerProvider()),
	)

	router := routing.New(backend, encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger)))

	registerOpsRoutes(router, registry)
	require.NoError(t, router.Err())

	return router.Handler()
}

func getOps(t *testing.T, handler http.Handler, path string) *httptest.ResponseRecorder {
	t.Helper()

	res := httptest.NewRecorder()
	handler.ServeHTTP(res, httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, http.NoBody))

	return res
}

func Test_registerOpsRoutes(T *testing.T) {
	T.Parallel()

	T.Run("live answers 200 with no body", func(t *testing.T) {
		t.Parallel()

		// Liveness checks nothing but that the process is serving, so an unhealthy registry
		// must not change what it says — a dependency being down restarts no pods.
		res := getOps(t, buildOpsHandler(t, &stubRegistry{
			result: &healthcheck.Result{Status: healthcheck.StatusDown},
		}), "/_ops_/live")

		assert.Equal(t, http.StatusOK, res.Code)
		assert.Empty(t, res.Body.String())
	})

	T.Run("ready answers 200 when every component is up", func(t *testing.T) {
		t.Parallel()

		result := &healthcheck.Result{
			Status:     healthcheck.StatusUp,
			Components: map[string]healthcheck.ComponentResult{"database": {Status: healthcheck.StatusUp}},
		}

		res := getOps(t, buildOpsHandler(t, &stubRegistry{result: result}), "/_ops_/ready")

		assert.Equal(t, http.StatusOK, res.Code)

		// Unenveloped: the body is the Result itself, as it was before these became typed
		// routes.
		var body healthcheck.Result
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		assert.Equal(t, *result, body)
	})

	T.Run("ready answers 503 when a component is down", func(t *testing.T) {
		t.Parallel()

		result := &healthcheck.Result{
			Status:     healthcheck.StatusDown,
			Components: map[string]healthcheck.ComponentResult{"database": {Status: healthcheck.StatusDown, Message: "nope"}},
		}

		res := getOps(t, buildOpsHandler(t, &stubRegistry{result: result}), "/_ops_/ready")

		// The status rides the return rather than the handler failing, so the body is the
		// same shape it is on the way up and nothing logs a fault.
		assert.Equal(t, http.StatusServiceUnavailable, res.Code)

		var body healthcheck.Result
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		assert.Equal(t, *result, body)
	})

	T.Run("version answers the build info", func(t *testing.T) {
		t.Parallel()

		res := getOps(t, buildOpsHandler(t, &stubRegistry{
			result: &healthcheck.Result{Status: healthcheck.StatusUp},
		}), "/_ops_/version")

		assert.Equal(t, http.StatusOK, res.Code)

		var body version.Info
		require.NoError(t, json.Unmarshal(res.Body.Bytes(), &body))
		assert.Equal(t, version.Get(), body)
	})
}

// stubAuthService records whether the OAuth2 handlers the router mounts were reached, which is
// what a body refused before the handler runs has to be checked against.
type stubAuthService struct {
	reached bool
}

func (s *stubAuthService) AuthorizeHandler(res http.ResponseWriter, req *http.Request) {
	s.serve(res, req)
}

func (s *stubAuthService) TokenHandler(res http.ResponseWriter, req *http.Request) {
	s.serve(res, req)
}

func (s *stubAuthService) RevokeHandler(res http.ResponseWriter, req *http.Request) {
	s.serve(res, req)
}

func (s *stubAuthService) AuthorizationServerMetadataHandler(res http.ResponseWriter, req *http.Request) {
	s.serve(res, req)
}

// serve reads the body the way the real OAuth2 handlers do, so a bound that only takes effect on
// the read is still visible to a test.
func (s *stubAuthService) serve(res http.ResponseWriter, req *http.Request) {
	s.reached = true

	if _, err := io.ReadAll(req.Body); err != nil {
		res.WriteHeader(http.StatusBadRequest)
		return
	}

	res.WriteHeader(http.StatusOK)
}

// stubProcessorRegistry records whether the payments webhook handler got as far as resolving a
// processor. It answers no to every provider: reaching the lookup at all is the signal.
type stubProcessorRegistry struct {
	consulted bool
}

func (r *stubProcessorRegistry) GetProcessor(string) (payments.PaymentProcessor, bool) {
	r.consulted = true

	return nil, false
}

func buildAPIRouter(t *testing.T, authService *stubAuthService, registry *stubProcessorRegistry) http.Handler {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()

	// The webhook handler is built with no payments manager: every request this test sends is
	// either refused before it runs or turned away at the processor lookup, both of which come
	// before the manager is touched.
	router, err := ProvideAPIRouter(
		t.Context(),
		routingcfg.Config{
			Provider: routingcfg.ProviderChi,
			Chi:      &chi.Config{ServiceName: t.Name(), SilenceRouteLogging: true},
		},
		logger,
		tracerProvider,
		metricsnoop.NewMetricsProvider(),
		authService,
		paymentswebhook.NewWebhookHandler(logger, tracerProvider, nil, registry),
		&stubRegistry{result: &healthcheck.Result{Status: healthcheck.StatusUp}},
	)
	require.NoError(t, err)

	return router.Handler()
}

func postTo(t *testing.T, handler http.Handler, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, path, strings.NewReader(body))
	res := httptest.NewRecorder()
	handler.ServeHTTP(res, req)

	return res
}

func Test_ProvideAPIRouter_requestBodyBound(T *testing.T) {
	T.Parallel()

	T.Run("an OAuth2 request within the bound reaches its handler", func(t *testing.T) {
		t.Parallel()

		authService := &stubAuthService{}

		res := postTo(t, buildAPIRouter(t, authService, &stubProcessorRegistry{}), "/token", "grant_type=authorization_code")

		assert.Equal(t, http.StatusOK, res.Code)
		assert.True(t, authService.reached)
	})

	T.Run("an oversized OAuth2 request is refused before its handler runs", func(t *testing.T) {
		t.Parallel()

		authService := &stubAuthService{}

		res := postTo(t, buildAPIRouter(t, authService, &stubProcessorRegistry{}), "/token", strings.Repeat("x", maxRequestBodyBytes+1))

		// 413 rather than 400: a client told its request was malformed sends the same one again.
		assert.Equal(t, http.StatusRequestEntityTooLarge, res.Code)
		assert.False(t, authService.reached)
	})

	T.Run("an OAuth2 request declaring no length is cut off mid-read", func(t *testing.T) {
		t.Parallel()

		authService := &stubAuthService{}

		// Chunked, or lying about its Content-Length: there is no length to check up front, so
		// the handler runs and the read is what fails. Without the bound it would read as much
		// as arrives.
		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/token", strings.NewReader(strings.Repeat("x", maxRequestBodyBytes+1)))
		req.ContentLength = -1
		res := httptest.NewRecorder()
		buildAPIRouter(t, authService, &stubProcessorRegistry{}).ServeHTTP(res, req)

		assert.True(t, authService.reached)
		assert.Equal(t, http.StatusBadRequest, res.Code)
	})

	T.Run("a payments webhook within the bound reaches its handler", func(t *testing.T) {
		t.Parallel()

		registry := &stubProcessorRegistry{}

		res := postTo(t, buildAPIRouter(t, &stubAuthService{}, registry), "/api/payments/webhooks/stripe", `{"type":"test"}`)

		// The stub registry knows no providers, so the handler's own answer is a 400 — which is
		// only reachable by running.
		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.True(t, registry.consulted)
	})

	T.Run("an oversized payments webhook is refused before its handler runs", func(t *testing.T) {
		t.Parallel()

		registry := &stubProcessorRegistry{}

		res := postTo(t, buildAPIRouter(t, &stubAuthService{}, registry), "/api/payments/webhooks/stripe", strings.Repeat("x", maxRequestBodyBytes+1))

		// Signature verification reads the body, so a body nobody has to hold is one nobody has
		// to verify.
		assert.Equal(t, http.StatusRequestEntityTooLarge, res.Code)
		assert.False(t, registry.consulted)
	})
}
