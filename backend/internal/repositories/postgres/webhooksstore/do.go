package webhooksstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWebhooksStore registers the store, with its recording hooks installed,
// and the dispatcher over it.
//
// There is no store without the hooks to register by mistake: they are built
// into the one store there is. Every endpoint write here owes an audit entry,
// and a second registration would be a way to skip that by naming the wrong
// dependency — which is how the arrangement this replaced lost them: it
// registered platform's raw store and kept the recording on three local tables
// beside it.
//
// The dispatcher is built over that store deliberately. A registration that
// reaches the database through the dispatcher is still an endpoint write, and an
// entry it did not produce would be a webhook the log does not know was created.
func RegisterWebhooksStore(i do.Injector) {
	do.Provide[platformwebhooks.Store](i, func(i do.Injector) (platformwebhooks.Store, error) {
		return ProvideStore(
			do.MustInvoke[context.Context](i),
			config(i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*platformrecording.Recorder](i),
		)
	})

	do.Provide[platformwebhooks.Dispatcher](i, func(i do.Injector) (platformwebhooks.Dispatcher, error) {
		return webhookscfg.NewDispatcher(
			do.MustInvoke[context.Context](i),
			config(i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[platformwebhooks.Store](i),
			// The catalog is generated Go rather than configuration: what an event
			// means is an application opinion, and there is no useful way to express
			// one in the environment. It is also what ListEventTypes serves.
			catalog.Catalog(),
			webhookscfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			webhookscfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			webhookscfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// config is the webhooks configuration this application runs with.
func config(i do.Injector) *webhookscfg.Config {
	if cfg, err := do.Invoke[*webhookscfg.Config](i); err == nil {
		return cfg
	}

	return &webhookscfg.Config{}
}
