package webhooks

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	domainwebhooks "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v12/webhooks/config"

	"github.com/samber/do/v2"
)

// RegisterWebhooksRepository registers the webhooks repository with the injector, together with
// the delivery seam it is built on: the webhook Store, and the platform Dispatcher over it.
//
// The three are registered together because every process that registers one registers all of
// them. One Store serves both halves of the system — this one, and the delivery worker in the
// scheduler — and they agree on which tables to use because the table prefix comes from one
// config field rather than being spelled in each process.
//
// Every process that writes rows needs the dispatcher, not only the one that manages webhooks.
// Dispatch happens inside the transaction that caused the event, so any process emitting data
// change events is a process that fans out webhooks.
func RegisterWebhooksRepository(i do.Injector) {
	do.Provide[webhooks.Store](i, func(i do.Injector) (webhooks.Store, error) {
		return webhookscfg.NewStore(
			do.MustInvoke[context.Context](i),
			webhooksConfig(i),
			do.MustInvoke[database.Client](i),
		)
	})

	do.Provide[webhooks.Dispatcher](i, func(i do.Injector) (webhooks.Dispatcher, error) {
		return webhookscfg.NewDispatcher(
			do.MustInvoke[context.Context](i),
			webhooksConfig(i),
			do.MustInvoke[webhooks.Store](i),
			// The catalog is generated Go rather than configuration: what an event means is
			// an application opinion, and there is no useful way to express one in the
			// environment.
			catalog.Catalog(),
			webhookscfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			webhookscfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			webhookscfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[domainwebhooks.Repository](i, func(i do.Injector) (domainwebhooks.Repository, error) {
		return ProvideWebhooksRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
			do.MustInvoke[webhooks.Dispatcher](i),
			do.MustInvoke[webhooks.Store](i),
		), nil
	})
}

// webhooksConfig resolves the webhooks config leniently, defaulting when it is absent.
//
// The config carries a table prefix and the delivery worker's knobs; a process that only
// registers endpoints and writes dispatch rows needs neither beyond the prefix, and the default
// prefix is what the migrations render.
func webhooksConfig(i do.Injector) *webhookscfg.Config {
	cfg, err := do.Invoke[*webhookscfg.Config](i)
	if err != nil || cfg == nil {
		return &webhookscfg.Config{}
	}

	return cfg
}
