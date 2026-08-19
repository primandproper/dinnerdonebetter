package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMCPServer_OpsRoutes covers the paths a deployment's probes hit rather than a
// client. They are mounted outside the bearer middleware, and a probe that got a 401
// instead of a 200 would be a replica that never goes ready — with the server behind it
// working the whole time.
func TestMCPServer_OpsRoutes(T *testing.T) {
	T.Parallel()

	T.Run("answers its health probes unauthenticated", func(t *testing.T) {
		t.Parallel()

		for _, path := range []string{"/_ops_/live", "/_ops_/ready"} {
			res := primary.get(t, primary.address+path)
			assert.Equal(t, http.StatusOK, res.StatusCode, path)
			require.NoError(t, res.Body.Close())
		}
	})

	T.Run("reports its version unauthenticated", func(t *testing.T) {
		t.Parallel()

		version := map[string]any{}
		getJSON(t, primary.address+"/_ops_/version", &version)

		assert.NotEmpty(t, version)
	})
}
