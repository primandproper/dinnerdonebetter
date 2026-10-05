package events

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	recordingcfg "github.com/primandproper/platform-go/v15/recording/config"
	"github.com/primandproper/platform-go/v15/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterOutboxEmitter registers the recording spine: the outbox writer with the search index
// rules on it, platform's webhooks Emitter and recording Recorder, and the Emitter here that
// builds this application's events for them.
//
// It needs a platform audit.Recorder registered beside it (auditlogentries.RegisterPlatformRecorder),
// which is the one piece a process builds itself: the log's prefix and redactions are decided
// where the log is built, not here.
//
// One Writer serves every transaction in the process: it holds no database handle, only the
// dialect and table name, and takes the caller's executor per Enqueue. The same is true of the
// Emitter and the Recorder, which is why one of each is enough.
func RegisterOutboxEmitter(i do.Injector) {
	// The one side effect, built once and shared: the Writer runs it inside every Enqueue, and
	// EmitIndex runs it over a message it never enqueues. Two instances would be two copies of
	// one table, which is the drift registering it on the writer exists to remove.
	do.Provide[outbox.SideEffect](i, func(do.Injector) (outbox.SideEffect, error) {
		return indexevents.NewSideEffect()
	})

	do.Provide[*outbox.Writer](i, func(i do.Injector) (*outbox.Writer, error) {
		return outbox.NewWriter(
			do.MustInvoke[database.Client](i).Dialect(),
			// Every write to this outbox owes the search index an event, so the index event is
			// registered here rather than passed per call. A repository method that changes an
			// indexed row cannot fail to produce one by omitting an option it was never asked
			// about; see internal/indexevents for which writes feed which index.
			outbox.WithWriterSideEffect(indexevents.SideEffectName, do.MustInvoke[outbox.SideEffect](i)),
			outbox.WithWriterLogger(do.MustInvoke[logging.Logger](i)),
			outbox.WithWriterTracerProvider(do.MustInvoke[tracing.Provider](i)),
			outbox.WithWriterMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})

	// Who is writing, for every audit entry the Recorder files. It is the session's principal,
	// which is what every platform surface here is already handed positionally.
	do.ProvideValue[callers.PrincipalExtractor](i, sessions.PrincipalFromContext)

	// Where an entry is filed. By its subject where it names one — a signup's entries belong
	// to the person on the list, not to the operator's request — and by the write's scope
	// otherwise, which is where this application's own entries have always gone.
	do.ProvideValue(i, &recordingcfg.Config{FileBy: recordingcfg.FileBySubject})

	do.Provide[*webhooks.Emitter](i, func(i do.Injector) (*webhooks.Emitter, error) {
		pillars, err := observability.InvokePillars(i)
		if err != nil {
			return nil, err
		}

		// The emitter builds its own fan-out store and dispatcher over the same tables the
		// webhooks store writes, which is what keeps the store's hooks — which record through
		// the Recorder, which emits through this — from needing the dispatcher that reads them.
		// See platform's webhooks/recordinghooks for the cycle that avoids.
		cfg := webhooksConfig(i)
		cfg.EmitterTopic = dataChangesTopic(i)

		return webhookscfg.NewEmitter(
			do.MustInvoke[context.Context](i),
			&cfg,
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*outbox.Writer](i),
			// The catalog is generated Go rather than configuration: what an event means is an
			// application opinion, and there is no useful way to express one in the environment.
			catalog.Catalog(),
			webhookscfg.WithPillars(pillars),
		)
	})

	// platform's own registration of the Recorder, over the three it needs: the audit recorder,
	// the Emitter above and the principal extractor.
	recordingcfg.Register(i)

	do.Provide[*Emitter](i, func(i do.Injector) (*Emitter, error) {
		return NewEmitter(
			do.MustInvoke[*webhooks.Emitter](i),
			do.MustInvoke[*platformrecording.Recorder](i),
			do.MustInvoke[*outbox.Writer](i),
			do.MustInvoke[outbox.SideEffect](i),
		)
	})
}

// webhooksConfig is a copy of the webhooks configuration this process runs with, or the zero
// value — which platform fills with defaults — for a process that has none. The MCP server and
// the async handler build repositories and have no webhook worker to configure.
func webhooksConfig(i do.Injector) webhookscfg.Config {
	if cfg, err := do.Invoke[*webhookscfg.Config](i); err == nil && cfg != nil {
		return *cfg
	}

	return webhookscfg.Config{}
}

// dataChangesTopic is the topic every data change event is published on: the configured one
// where this process has a queues config, and the default otherwise. A process with no broker
// still writes its events to the outbox under that topic, for the worker that has one.
func dataChangesTopic(i do.Injector) string {
	if queues, err := do.Invoke[*queuescfg.Config](i); err == nil && queues != nil && queues.DataChangesTopicName != "" {
		return queues.DataChangesTopicName
	}

	return queuescfg.DefaultDataChangesTopicName
}
