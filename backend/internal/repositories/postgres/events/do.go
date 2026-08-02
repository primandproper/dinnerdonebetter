package events

import (
	queuescfg "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v9/database/dialect"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/outbox"

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
			outbox.WithWriterTracerProvider(do.MustInvoke[tracing.TracerProvider](i)),
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

		// NewEmitter returns nil for an empty topic, and a nil Emitter emits nothing.
		return NewEmitter(do.MustInvoke[*outbox.Writer](i), topic), nil
	})
}
