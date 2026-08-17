package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/primandproper/platform-go/v10/encoding"
	"github.com/primandproper/platform-go/v10/healthcheck"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"
	"github.com/primandproper/platform-go/v10/routing"
	"github.com/primandproper/platform-go/v10/routing/backends/chi"
	"github.com/primandproper/platform-go/v10/version"

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
