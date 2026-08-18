package webhookdispatch

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v11/webhooks/config"

	"github.com/samber/do/v2"
)

// RegisterWebhookDispatch registers the webhook Store and the write-side Dispatcher.
//
// One Store serves both halves of the system: this one, and the delivery worker in the
// scheduler. They agree on which tables to use because the table prefix comes from one config
// field rather than being spelled in each process.
//
// Every process that writes rows needs this, not only the one that manages webhooks. Dispatch
// happens inside the transaction that caused the event, so any process emitting data change
// events is a process that fans out webhooks.
func RegisterWebhookDispatch(i do.Injector) {
	do.Provide[webhooks.Store](i, func(i do.Injector) (webhooks.Store, error) {
		// Resolved leniently, and defaulted when absent. The config carries a table prefix
		// and the delivery worker's knobs; a process that only writes dispatch rows needs
		// neither beyond the prefix, and the default prefix is what the migrations render.
		cfg, err := do.Invoke[*webhookscfg.Config](i)
		if err != nil || cfg == nil {
			cfg = &webhookscfg.Config{}
		}

		return webhookscfg.NewStore(
			do.MustInvoke[context.Context](i),
			cfg,
			do.MustInvoke[database.Client](i),
		)
	})

	do.Provide[*Dispatcher](i, func(i do.Injector) (*Dispatcher, error) {
		return NewDispatcher(
			do.MustInvoke[webhooks.Store](i),
			// The catalog is generated Go rather than configuration: what an event means is
			// an application opinion, and there is no useful way to express one in the
			// environment.
			catalog.Catalog(),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
