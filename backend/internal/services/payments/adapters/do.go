package adapters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	capitalismcfg "github.com/primandproper/platform-go/v13/capitalism/config"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentProcessorRegistry registers the payment processor registry with the injector.
func RegisterPaymentProcessorRegistry(i do.Injector) {
	do.Provide[*payments.MapProcessorRegistry](i, func(i do.Injector) (*payments.MapProcessorRegistry, error) {
		return ProvidePaymentProcessorRegistry(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			*do.MustInvoke[*paymentscfg.Config](i),
		)
	})

	// Bind the interface
	do.Provide[payments.PaymentProcessorRegistry](i, func(i do.Injector) (payments.PaymentProcessorRegistry, error) {
		return do.MustInvoke[*payments.MapProcessorRegistry](i), nil
	})
}

// ProvidePaymentProcessorRegistry creates a registry with stripe and revenuecat processors.
//
// Both entries follow one rule: the endpoint's configured provider selects its processor,
// naming the noop provider gets the stub, and an unset or unrecognized one is an error. That is
// deliberately louder than the "no API key means stub" rule RevenueCat used to have — and
// Stripe before it — under which a deployment that had merely forgotten a credential looked
// exactly like one that had chosen not to bill anybody.
//
// Each endpoint takes only the provider it is named for. A provider registered under another
// one's name is an error rather than a swap, because the name in the map is the name in the
// webhook URL: a deployment that mounted the mobile adapter at /webhooks/stripe would verify
// Stripe's deliveries against RevenueCat's secret and reject every one of them.
func ProvidePaymentProcessorRegistry(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg paymentscfg.Config,
) (*payments.MapProcessorRegistry, error) {
	processors := make(map[string]payments.PaymentProcessor, 2)

	noopStubPaymentProcessor := NewStubPaymentProcessor(logger)

	// The web checkout endpoint.
	switch cfg.Capitalism.Provider {
	case capitalismcfg.StripeProvider:
		stripeProcessor, err := NewStripePaymentProcessor(logger, tracerProvider, cfg.Capitalism.Stripe)
		if err != nil {
			return nil, err
		}

		processors[capitalismcfg.StripeProvider] = stripeProcessor
	case capitalismcfg.NoopProvider:
		processors[capitalismcfg.StripeProvider] = noopStubPaymentProcessor
	default:
		return nil, platformerrors.Wrapf(platformerrors.ErrUnknownProvider, "web payments provider %q", cfg.Capitalism.Provider)
	}

	// The mobile store endpoint.
	switch cfg.MobileProvider {
	case capitalismcfg.RevenueCatProvider:
		revenueCatProcessor, err := NewRevenueCatPaymentProcessor(logger, tracerProvider, cfg.Capitalism.RevenueCat)
		if err != nil {
			return nil, err
		}

		processors[capitalismcfg.RevenueCatProvider] = revenueCatProcessor
	case capitalismcfg.NoopProvider:
		processors[capitalismcfg.RevenueCatProvider] = noopStubPaymentProcessor
	default:
		return nil, platformerrors.Wrapf(platformerrors.ErrUnknownProvider, "mobile payments provider %q", cfg.MobileProvider)
	}

	return payments.NewMapProcessorRegistry(processors), nil
}
