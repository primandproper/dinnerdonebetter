package adapters

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	capitalismcfg "github.com/primandproper/platform-go/v12/capitalism/config"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRegistryForTest(t *testing.T, cfg paymentscfg.Config) (*payments.MapProcessorRegistry, error) {
	t.Helper()

	return ProvidePaymentProcessorRegistry(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), cfg)
}

func TestProvidePaymentProcessorRegistry(T *testing.T) {
	T.Parallel()

	T.Run("with nil RevenueCat config", func(t *testing.T) {
		t.Parallel()

		registry, err := buildRegistryForTest(t, paymentscfg.Config{
			Capitalism: capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
		})
		require.NoError(t, err)

		processor, ok := registry.GetProcessor("revenuecat")
		require.True(t, ok)
		assert.IsType(t, &StubPaymentProcessor{}, processor)
	})

	// The env:",init" tag on Config.RevenueCat allocates the struct before the env overlay
	// walks into it, so a deployment that configures nothing under REVENUECAT_ now arrives
	// here with an empty struct rather than a nil pointer. It must still select the stub.
	T.Run("with allocated but empty RevenueCat config", func(t *testing.T) {
		t.Parallel()

		registry, err := buildRegistryForTest(t, paymentscfg.Config{
			RevenueCat: &paymentscfg.RevenueCatConfig{},
			Capitalism: capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
		})
		require.NoError(t, err)

		processor, ok := registry.GetProcessor("revenuecat")
		require.True(t, ok)
		assert.IsType(t, &StubPaymentProcessor{}, processor)
	})

	T.Run("with configured RevenueCat config", func(t *testing.T) {
		t.Parallel()

		registry, err := buildRegistryForTest(t, paymentscfg.Config{
			RevenueCat: &paymentscfg.RevenueCatConfig{
				APIKey:            "example_api_key",
				WebhookAuthHeader: exampleRevenueCatAuthHeader,
			},
			Capitalism: capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
		})
		require.NoError(t, err)

		processor, ok := registry.GetProcessor("revenuecat")
		require.True(t, ok)
		assert.IsType(t, &RevenueCatPaymentProcessor{}, processor)
	})
}
