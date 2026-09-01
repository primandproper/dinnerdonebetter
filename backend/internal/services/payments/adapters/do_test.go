package adapters

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	capitalismcfg "github.com/primandproper/platform-go/v13/capitalism/config"
	caprevenuecat "github.com/primandproper/platform-go/v13/capitalism/revenuecat"
	capstripe "github.com/primandproper/platform-go/v13/capitalism/stripe"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRegistryForTest(t *testing.T, cfg paymentscfg.Config) (*payments.MapProcessorRegistry, error) {
	t.Helper()

	return ProvidePaymentProcessorRegistry(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), cfg)
}

// noopPaymentsConfigForTest is the configuration of a deployment that has deliberately chosen
// to bill nobody, on either endpoint.
func noopPaymentsConfigForTest() paymentscfg.Config {
	return paymentscfg.Config{
		Capitalism:     capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
		MobileProvider: capitalismcfg.NoopProvider,
	}
}

func TestProvidePaymentProcessorRegistry(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.Capitalism.Stripe = &capstripe.Config{WebhookSecret: exampleWebhookSecret}
		cfg.Capitalism.Provider = capitalismcfg.StripeProvider
		cfg.Capitalism.RevenueCat = &caprevenuecat.Config{WebhookSecret: exampleRevenueCatWebhookSecret}
		cfg.MobileProvider = capitalismcfg.RevenueCatProvider

		registry, err := buildRegistryForTest(t, cfg)
		require.NoError(t, err)

		stripeProcessor, ok := registry.GetProcessor(capitalismcfg.StripeProvider)
		require.True(t, ok)
		assert.IsType(t, &StripePaymentProcessor{}, stripeProcessor)

		revenueCatProcessor, ok := registry.GetProcessor(capitalismcfg.RevenueCatProvider)
		require.True(t, ok)
		assert.IsType(t, &RevenueCatPaymentProcessor{}, revenueCatProcessor)
	})

	T.Run("with both providers noop", func(t *testing.T) {
		t.Parallel()

		registry, err := buildRegistryForTest(t, noopPaymentsConfigForTest())
		require.NoError(t, err)

		for _, provider := range []string{capitalismcfg.StripeProvider, capitalismcfg.RevenueCatProvider} {
			processor, ok := registry.GetProcessor(provider)
			require.True(t, ok, provider)
			assert.IsType(t, &StubPaymentProcessor{}, processor, provider)
		}
	})

	// The rule both endpoints now share. A deployment that merely forgot to name its provider
	// must not come up looking like one that chose not to bill anybody.
	T.Run("with unset web provider", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.Capitalism.Provider = ""

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, registry)
	})

	T.Run("with unset mobile provider", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.MobileProvider = ""

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, registry)
	})

	// Each endpoint takes only the provider it is named for: the name in the map is the name in
	// the webhook URL, so a swap would verify one provider's deliveries against the other's
	// secret.
	T.Run("with the mobile provider named on the web endpoint", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.Capitalism.Provider = capitalismcfg.RevenueCatProvider

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, registry)
	})

	T.Run("with the web provider named on the mobile endpoint", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.MobileProvider = capitalismcfg.StripeProvider

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, registry)
	})

	// RevenueCat is inbound-only, so a manager without a signing secret could do nothing at
	// all. Selecting it without one fails here rather than at the first delivery.
	T.Run("with revenuecat selected but no webhook secret", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.MobileProvider = capitalismcfg.RevenueCatProvider
		cfg.Capitalism.RevenueCat = &caprevenuecat.Config{}

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, caprevenuecat.ErrWebhookSecretNotConfigured)
		assert.Nil(t, registry)
	})

	T.Run("with revenuecat selected but no config at all", func(t *testing.T) {
		t.Parallel()

		cfg := noopPaymentsConfigForTest()
		cfg.MobileProvider = capitalismcfg.RevenueCatProvider

		registry, err := buildRegistryForTest(t, cfg)
		require.ErrorIs(t, err, caprevenuecat.ErrNilConfig)
		assert.Nil(t, registry)
	})
}
