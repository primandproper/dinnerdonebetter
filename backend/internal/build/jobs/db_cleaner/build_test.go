package dbcleaner

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config/environments"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildInjector(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		configs := (&config.EnvironmentConfigSet{RootConfig: environments.BuildIntegrationTestsConfig()}).Derive()

		i, err := BuildInjector(t.Context(), configs.DBCleaner)
		require.NoError(t, err)

		services := i.ListProvidedServices()
		assert.NotEmpty(t, services, "expected providers to be registered")
		assert.Greater(t, len(services), 5, "expected many providers to be registered")
	})

	T.Run("refuses a config with no database", func(t *testing.T) {
		t.Parallel()

		configs := (&config.EnvironmentConfigSet{RootConfig: environments.BuildIntegrationTestsConfig()}).Derive()
		configs.DBCleaner.Service.Database = nil

		_, err := BuildInjector(t.Context(), configs.DBCleaner)
		assert.Error(t, err)
	})
}
