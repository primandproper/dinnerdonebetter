package payments

import (
	"context"
	"net/http"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"

	"github.com/primandproper/platform-go/v15/billing"
	billinghttp "github.com/primandproper/platform-go/v15/billing/http"
	"github.com/primandproper/platform-go/v15/billing/standing"
	billingsync "github.com/primandproper/platform-go/v15/billing/sync"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	capitalismcfg "github.com/primandproper/primitives-go/v2/capitalism/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// WebhookHandlers are the payment providers' webhook endpoints, one per provider this
// service takes deliveries from. Each is platform-go's billing/http handler over a
// billing/sync Syncer; what this application supplies is which provider each endpoint
// verifies against and where a subscription nobody holds yet lands — see
// payments.StripePlace and payments.RevenueCatPlace.
type WebhookHandlers struct {
	_ struct{} `json:"-"`

	// Stripe takes the web checkout's deliveries.
	Stripe http.Handler
	// RevenueCat takes the mobile stores' deliveries.
	RevenueCat http.Handler
}

// RegisterWebhookHandlers registers the payment providers' webhook endpoints with the injector.
func RegisterWebhookHandlers(i do.Injector) {
	do.Provide[*WebhookHandlers](i, func(i do.Injector) (*WebhookHandlers, error) {
		return ProvideWebhookHandlers(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*paymentscfg.Config](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[billing.Store](i),
			do.MustInvoke[platformidentity.Store](i),
		)
	})
}

// ProvideWebhookHandlers builds both endpoints.
//
// Both follow one rule: the endpoint's configured provider selects its payment manager,
// naming the noop provider gets capitalism's noop manager — which reports nothing, so every
// delivery is acknowledged and nothing is written — and an unset or unrecognized one is an
// error. Each endpoint takes only the provider it is named for: an endpoint mounted at
// /webhooks/stripe that verified against RevenueCat's secret would reject every delivery
// Stripe sent it.
//
// Every delivery is for the global scope, which is where this application keeps all of its
// billing (see internal/domain/payments), and the account's coarse standing is written on
// the delivery's transaction under standing.Strict: active is paid, trialing is a trial,
// and nothing else is. Passing it rather than having it defaulted is this deployment
// saying that is its rule — no dunning window and no grace on past_due.
func ProvideWebhookHandlers(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *paymentscfg.Config,
	client database.Client,
	store billing.Store,
	accounts platformidentity.Store,
) (*WebhookHandlers, error) {
	if cfg == nil {
		return nil, platformerrors.ErrNilInputParameter
	}

	build := func(provider string, capitalismConfig *capitalismcfg.Config, place billingsync.Place) (http.Handler, error) {
		if capitalismConfig.Provider != provider && capitalismConfig.Provider != capitalismcfg.NoopProvider {
			return nil, platformerrors.Wrapf(platformerrors.ErrUnknownProvider, "%s webhook endpoint provider %q", provider, capitalismConfig.Provider)
		}

		manager, err := capitalismcfg.NewPaymentManager(ctx, capitalismConfig,
			capitalismcfg.WithLogger(logger),
			capitalismcfg.WithTracerProvider(tracerProvider),
			capitalismcfg.WithMetricsProvider(metricsProvider),
		)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "building the %s payment manager", provider)
		}

		syncer, err := billingsync.New(store, place,
			billingsync.WithStanding(accounts, standing.Strict),
			billingsync.WithLogger(logger),
			billingsync.WithTracerProvider(tracerProvider),
			billingsync.WithMetricsProvider(metricsProvider),
		)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "building the %s billing syncer", provider)
		}

		handler, err := billinghttp.NewWebhookHandler(manager, syncer, client,
			billinghttp.WithScopeResolver(billinghttp.GlobalScope),
			billinghttp.WithLogger(logger),
			billinghttp.WithTracerProvider(tracerProvider),
		)
		if err != nil {
			return nil, platformerrors.Wrapf(err, "building the %s webhook handler", provider)
		}

		return handler, nil
	}

	stripeHandler, err := build(capitalismcfg.StripeProvider, &cfg.Capitalism, payments.StripePlace(store, accounts))
	if err != nil {
		return nil, err
	}

	// capitalism's config names one provider and this service takes deliveries from two, so
	// the mobile endpoint is the same credentials under its own selector. See
	// paymentscfg.Config.MobileProvider.
	mobile := &capitalismcfg.Config{
		Provider:   cfg.MobileProvider,
		RevenueCat: cfg.Capitalism.RevenueCat,
	}

	revenueCatHandler, err := build(capitalismcfg.RevenueCatProvider, mobile, payments.RevenueCatPlace(store, accounts))
	if err != nil {
		return nil, err
	}

	return &WebhookHandlers{
		Stripe:     stripeHandler,
		RevenueCat: revenueCatHandler,
	}, nil
}
