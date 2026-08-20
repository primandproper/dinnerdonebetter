package adapters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	capitalismcfg "github.com/primandproper/platform-go/v12/capitalism/config"
	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

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
// The Stripe entry follows the capitalism provider: naming the noop provider gets the stub, and
// an unset or unrecognized one is an error. That is deliberately louder than the old
// "no API key means stub" rule, under which a deployment that had merely forgotten to set its
// Stripe secret looked exactly like one that had chosen not to bill anybody.
//
// RevenueCat has no such selector, and still falls back to the stub when unconfigured.
func ProvidePaymentProcessorRegistry(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	cfg paymentscfg.Config,
) (*payments.MapProcessorRegistry, error) {
	processors := make(map[string]payments.PaymentProcessor)

	noopStubPaymentProcessor := NewStubPaymentProcessor(logger)

	switch cfg.Capitalism.Provider {
	case capitalismcfg.StripeProvider:
		stripeProcessor, err := NewStripePaymentProcessor(logger, tracerProvider, cfg.Capitalism.Stripe)
		if err != nil {
			return nil, err
		}

		processors["stripe"] = stripeProcessor
	case capitalismcfg.NoopProvider:
		processors["stripe"] = noopStubPaymentProcessor
	default:
		return nil, platformerrors.Wrapf(platformerrors.ErrUnknownProvider, "payments provider %q", cfg.Capitalism.Provider)
	}

	// RevenueCat: use real adapter when configured, else stub
	if cfg.RevenueCat != nil && cfg.RevenueCat.APIKey != "" {
		processors["revenuecat"] = NewRevenueCatPaymentProcessor(logger, tracerProvider, &RevenueCatConfig{
			APIKey:            cfg.RevenueCat.APIKey,
			WebhookAuthHeader: cfg.RevenueCat.WebhookAuthHeader,
		})
	} else {
		processors["revenuecat"] = noopStubPaymentProcessor
	}

	return payments.NewMapProcessorRegistry(processors), nil
}
