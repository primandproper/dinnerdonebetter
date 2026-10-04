/*
Package webhooksstore is platform-go's webhook store with this application's
recording around it.

It replaces internal/repositories/postgres/webhooks, which owned three tables of
its own — webhooks, webhook_trigger_configs and webhook_trigger_events — and ran
them in parallel with platform's endpoints, joined by a shared identifier. The
endpoints were the delivery machinery and the local rows were the model; now
there is one model, and it is platform's.

One field did not survive, and it was already inert. The local webhooks row
carried a method column, validated on every write and echoed back on every read,
while delivery went through platform's dispatcher — which posts. A client that
asked for PUT was told PUT and sent POST. Dropping the column removes the lie
rather than a capability.

What is not platform's is the audit entry and the data change event every write
owes, which is what lives here. Both are written from platform's webhooks.Hooks,
which the store calls on the caller's transaction once each endpoint or
subscription write has landed. A hook's error fails the write, so the row, the
entry and the event commit together or not at all.
*/
package webhooksstore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformwebhooks "github.com/primandproper/platform-go/v14/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v14/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const o11yName = "webhooks_db_client"

// ProvideStore builds platform's webhook store, with this application's
// recording hung off its writes.
func ProvideStore(
	ctx context.Context,
	cfg *webhookscfg.Config,
	client database.Client,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	eventEmitter *events.Emitter,
) (platformwebhooks.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return webhookscfg.NewStore(ctx, cfg, client, &hooks{
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	})
}
