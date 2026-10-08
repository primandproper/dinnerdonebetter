package mcpbuild

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInjector(T *testing.T) {
	T.Parallel()

	// Registration is where a duplicate declaration panics — a component registered by the
	// build and again by a domain's registration — so building the container at all, with
	// nothing resolved, is the test that catches it without a database.
	T.Run("registers every provider once", func(t *testing.T) {
		t.Parallel()

		configs := (&config.EnvironmentConfigSet{RootConfig: environments.BuildIntegrationTestsConfig()}).Derive()

		var i *do.RootScope

		require.NotPanics(t, func() {
			i = BuildInjector(t.Context(), configs.MCPService)
		})

		services := i.ListProvidedServices()
		assert.NotEmpty(t, services, "expected providers to be registered")
		assert.Greater(t, len(services), 10, "expected many providers to be registered")
	})
}
