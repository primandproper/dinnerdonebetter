package events

import (
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"

	"github.com/primandproper/platform-go/v10/database/dialect"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	"github.com/primandproper/platform-go/v10/outbox"

	"github.com/samber/do/v2"
)

// RegisterOutboxEmitter registers the outbox writer and the data-changes Emitter built on it.
//
// One Writer serves every transaction in the process: it holds no database handle, only the
// dialect and table name, and takes the caller's executor per Enqueue.
func RegisterOutboxEmitter(i do.Injector) {
	do.Provide[*outbox.Writer](i, func(i do.Injector) (*outbox.Writer, error) {
		return outbox.NewWriter(
			dialect.Postgres,
			outbox.WithWriterLogger(do.MustInvoke[logging.Logger](i)),
			outbox.WithWriterTracerProvider(do.MustInvoke[tracing.Provider](i)),
			outbox.WithWriterMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	do.Provide[*Emitter](i, func(i do.Injector) (*Emitter, error) {
		// A process with no queues config has no topic to emit to — the MCP server, and the
		// one-shot CLI tools, register repositories but publish nothing. NewEmitter returns
		// nil for an empty topic and a nil Emitter emits nothing, so those keep working
		// rather than failing to construct a repository they only read through.
		topic := ""
		if queues, err := do.Invoke[*queuescfg.Config](i); err == nil {
			topic = queues.DataChangesTopicName
		}

		// The dispatcher is resolved leniently for the same reason the topic is: the MCP
		// server and the one-shot CLI tools register repositories they only read through,
		// and have no webhook tables to dispatch into. A nil dispatcher dispatches nothing.
		dispatcher, err := do.Invoke[*webhookdispatch.Dispatcher](i)
		if err != nil {
			dispatcher = nil
		}

		// NewEmitter returns nil for an empty topic, and a nil Emitter emits nothing.
		return NewEmitter(do.MustInvoke[*outbox.Writer](i), topic, dispatcher), nil
	})
}
