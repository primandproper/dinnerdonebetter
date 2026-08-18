package scheduler

import (
	"context"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v11/webhooks/config"

	"github.com/samber/do/v2"
)

// RegisterWebhookWorker registers the outbound webhook delivery worker with the injector.
//
// It runs here, beside the outbox relay, for the same reasons: it is a polling loop that must not
// be tied to a request, and it needs exactly what this process already has. The API service
// writes dispatch rows inside the transactions that cause them; this claims those rows, signs
// and sends them, records every attempt, and schedules retries.
//
// The same tick also reaps delivered dispatches and their attempts past the retention window, so
// there is no separate scheduled job for retention — the worker is the whole delivery side.
//
// The Store comes from the injector rather than being built here, so the claim side and the
// dispatch side of this process are reading and writing the same tables by construction.
func RegisterWebhookWorker(i do.Injector) {
	do.Provide[*webhooks.Worker](i, func(i do.Injector) (*webhooks.Worker, error) {
		return webhookscfg.NewWorker(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[*webhookscfg.Config](i),
			do.MustInvoke[webhooks.Store](i),
			webhookscfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			webhookscfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			webhookscfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
