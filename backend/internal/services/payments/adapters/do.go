package adapters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/config"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentProcessorRegistry registers the payment processor registry with the injector.
func RegisterPaymentProcessorRegistry(i do.Injector) {
	do.Provide[*payments.MapProcessorRegistry](i, func(i do.Injector) (*payments.MapProcessorRegistry, error) {
		return ProvidePaymentProcessorRegistry(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			*do.MustInvoke[*paymentscfg.Config](i),
		), nil
	})

	// Bind the interface
	do.Provide[payments.PaymentProcessorRegistry](i, func(i do.Injector) (payments.PaymentProcessorRegistry, error) {
		return do.MustInvoke[*payments.MapProcessorRegistry](i), nil
	})
}

// ProvidePaymentProcessorRegistry creates a registry with stripe and revenuecat processors.
// Uses stub when a provider is not configured.
func ProvidePaymentProcessorRegistry(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	cfg paymentscfg.Config,
) *payments.MapProcessorRegistry {
	processors := make(map[string]payments.PaymentProcessor)

	noopStubPaymentProcessor := NewStubPaymentProcessor(logger)

	// Stripe: use real adapter when configured, else stub
	if cfg.Stripe != nil && cfg.Stripe.APIKey != "" {
		processors["stripe"] = NewStripePaymentProcessor(logger, tracerProvider, &StripeConfig{
			APIKey:        cfg.Stripe.APIKey,
			WebhookSecret: cfg.Stripe.WebhookSecret,
		})
	} else {
		processors["stripe"] = noopStubPaymentProcessor
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

	return payments.NewMapProcessorRegistry(processors)
}
