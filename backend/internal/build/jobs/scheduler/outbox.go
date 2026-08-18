package scheduler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/messagequeue"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/outbox"

	"github.com/samber/do/v2"
)

// RegisterOutboxRelay registers the outbox relay with the injector.
//
// The relay moves rows the API service wrote inside its own transactions onto the broker. It
// runs here rather than in the API service because it is a polling loop that must not be tied
// to a request, and because running it in one place keeps its claim contention predictable —
// though ClaimSkipLocked means several replicas would be correct too.
func RegisterOutboxRelay(i do.Injector) {
	do.Provide[*outbox.Relay](i, func(i do.Injector) (*outbox.Relay, error) {
		return outbox.NewRelay(
			do.MustInvoke[context.Context](i),
			&do.MustInvoke[*config.SchedulerConfig](i).Outbox,
			do.MustInvoke[database.Client](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			outbox.WithRelayLogger(do.MustInvoke[logging.Logger](i)),
			outbox.WithRelayTracerProvider(do.MustInvoke[tracing.Provider](i)),
			outbox.WithRelayMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}
