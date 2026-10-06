package grpcapi

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInjector(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		i, err := BuildInjector(t.Context(), environments.BuildIntegrationTestsConfig())
		require.NoError(t, err)

		services := i.ListProvidedServices()
		assert.NotEmpty(t, services, "expected providers to be registered")
		assert.Greater(t, len(services), 10, "expected many providers to be registered")
	})

	T.Run("refuses a config with no gRPC server", func(t *testing.T) {
		t.Parallel()

		cfg := environments.BuildIntegrationTestsConfig()
		cfg.Service.GRPCServer = nil

		_, err := BuildInjector(t.Context(), cfg)
		assert.Error(t, err)
	})
}
