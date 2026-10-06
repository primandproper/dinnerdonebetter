/*
Package recordingspine builds platform's recording spine the way this application runs it: one
outbox writer with the search index rules registered on it, platform's webhooks Emitter fanning
out over this application's catalog, and platform's Recorder filing by subject and reading the
actor off the session.

Nothing here writes anything. Every write records and announces itself through the platform types
this package builds — recording.Recorder for an audit entry and its event, webhooks.Emitter for an
event alone, outbox.Writer's EnqueueDerived for an index event alone — and the one thing this
application adds to them, the payload its own events carry, is datachanges.Event.

Publishing an event after a repository commits is two operations against two systems that share
no commit: the row lands, the publish fails, and durable state and the event stream diverge with
nothing to detect it. Every one of those three takes the caller's database.Tx, so the event lives
or dies with the row. That guarantee is platform's, and nothing here re-states it.

There are two ways in, and they must agree: Register, for a process wired through the injector,
and New, for one that assembles its stores by hand — the local dev server, the integration
suites' fixtures, the one-shot tools, a repository test. This file is the one place the parts are
named twice.
*/
package recordingspine

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/platform-go/v15/outbox"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	recordingcfg "github.com/primandproper/platform-go/v15/recording/config"
	"github.com/primandproper/platform-go/v15/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// Spine is the three platform types a hand-built process records and publishes through.
type Spine struct {
	// Writer is the outbox writer, with the search index rules registered on it.
	Writer *outbox.Writer
	// Emitter publishes an event and fans it out to the scope's webhook subscribers.
	Emitter *webhooks.Emitter
	// Recorder writes an audit entry and the event describing the same write.
	Recorder *platformrecording.Recorder
}

// Option configures New.
type Option func(*options)

type options struct {
	pillars  *observability.Pillars
	topic    string
	webhooks webhookscfg.Config
}

// WithTopic names the topic data change events are published on. The default is
// queuescfg.DefaultDataChangesTopicName.
func WithTopic(topic string) Option {
	return func(o *options) {
		if topic != "" {
			o.topic = topic
		}
	}
}

// WithWebhooksConfig is the webhooks configuration the emitter's fan-out store is built from —
// its table prefix above all. The default is platform's defaults, which is what every process
// here runs with.
func WithWebhooksConfig(cfg *webhookscfg.Config) Option {
	return func(o *options) {
		if cfg != nil {
			o.webhooks = *cfg
		}
	}
}

// WithPillars supplies the observability pillars. Absent means noop.
func WithPillars(pillars *observability.Pillars) Option {
	return func(o *options) {
		if pillars != nil {
			o.pillars = pillars
		}
	}
}

// fileBy is where an entry is filed: by its subject where it names one — a signup's entries
// belong to the person on the list, not to the operator's request — and by the write's scope
// otherwise, which is where this application's own entries have always gone.
const fileBy = recordingcfg.FileBySubject

// NewWriter builds the outbox writer every write in a process enqueues through, with the search
// index rules registered on it.
//
// Every write to this outbox owes the search index an event, so the index event is registered
// here rather than passed per call. A repository method that changes an indexed row cannot fail
// to produce one by omitting an option it was never asked about; see internal/indexevents for
// which writes feed which index. A write that owes only the index runs the same registration
// through EnqueueDerived.
//
// One Writer serves every transaction in the process: it holds no database handle, only the
// dialect and table name, and takes the caller's executor per Enqueue.
func NewWriter(db database.Client, pillars *observability.Pillars) (*outbox.Writer, error) {
	effect, err := indexevents.NewSideEffect()
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the search index side effect")
	}

	logger, tracerProvider, metricsProvider := pillars.Deps()

	writer, err := outbox.NewWriter(db.Dialect(),
		outbox.WithWriterSideEffect(indexevents.SideEffectName, effect),
		outbox.WithWriterLogger(logger),
		outbox.WithWriterTracerProvider(tracerProvider),
		outbox.WithWriterMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the outbox writer")
	}

	return writer, nil
}

// New builds the spine over one database client, for a process that assembles its stores by
// hand rather than through the injector. Register is the same construction through platform's
// own registrations.
//
// The audit recorder is the one the process built, because the log's prefix and redactions are
// decided where the log is built. Everything else is derived.
func New(ctx context.Context, db database.Client, auditRecorder platformaudit.Recorder, opts ...Option) (*Spine, error) {
	if db == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	o := &options{topic: queuescfg.DefaultDataChangesTopicName, pillars: &observability.Pillars{}}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	writer, err := NewWriter(db, o.pillars)
	if err != nil {
		return nil, err
	}

	cfg := o.webhooks
	cfg.EmitterTopic = o.topic

	emitter, err := webhookscfg.NewEmitter(ctx, &cfg, db, writer, catalog.Catalog(), webhookscfg.WithPillars(o.pillars))
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the webhooks emitter")
	}

	recorder, err := recordingcfg.NewRecorder(ctx,
		&recordingcfg.Config{FileBy: fileBy},
		auditRecorder, emitter, sessions.PrincipalFromContext,
		recordingcfg.WithPillars(o.pillars),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the recording recorder")
	}

	return &Spine{Writer: writer, Emitter: emitter, Recorder: recorder}, nil
}

// Register registers the spine with the injector: the outbox writer, platform's webhooks Emitter
// and its recording Recorder.
//
// It needs a platform audit.Recorder registered beside it (auditlogentries.RegisterPlatformRecorder),
// which is the one piece a process builds itself: the log's prefix and redactions are decided
// where the log is built, not here.
func Register(i do.Injector) {
	do.Provide[*outbox.Writer](i, func(i do.Injector) (*outbox.Writer, error) {
		return NewWriter(do.MustInvoke[database.Client](i), &observability.Pillars{
			Logger:          do.MustInvoke[logging.Logger](i),
			TracerProvider:  do.MustInvoke[tracing.Provider](i),
			MetricsProvider: do.MustInvoke[metrics.Provider](i),
		})
	})

	// Who is writing, for every audit entry the Recorder files. It is the session's principal,
	// which is what every platform surface here is already handed positionally.
	do.ProvideValue[callers.PrincipalExtractor](i, sessions.PrincipalFromContext)

	do.ProvideValue(i, &recordingcfg.Config{FileBy: fileBy})

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
