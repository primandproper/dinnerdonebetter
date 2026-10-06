package datachangemessagehandler

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

		i, err := BuildInjector(t.Context(), configs.AsyncMessageHandler)
		require.NoError(t, err)

		services := i.ListProvidedServices()
		assert.NotEmpty(t, services, "expected providers to be registered")
		assert.Greater(t, len(services), 10, "expected many providers to be registered")
	})

	T.Run("refuses a config with no broker", func(t *testing.T) {
		t.Parallel()

		configs := (&config.EnvironmentConfigSet{RootConfig: environments.BuildIntegrationTestsConfig()}).Derive()
		configs.AsyncMessageHandler.Service.MessageQueue = nil

		_, err := BuildInjector(t.Context(), configs.AsyncMessageHandler)
		assert.Error(t, err)
	})
}
