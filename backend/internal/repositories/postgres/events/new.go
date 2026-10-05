package events

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/platform-go/v15/outbox"
	recordingcfg "github.com/primandproper/platform-go/v15/recording/config"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
)

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
func WithWebhooksConfig(cfg webhookscfg.Config) Option {
	return func(o *options) { o.webhooks = cfg }
}

// WithPillars supplies the observability pillars. Absent means noop.
func WithPillars(pillars *observability.Pillars) Option {
	return func(o *options) { o.pillars = pillars }
}

// New builds the recording spine over one database client, for a process that assembles its
// stores by hand rather than through the injector: the local dev server, the integration suite's
// seeding, a repository test. RegisterOutboxEmitter is the same construction through platform's
// own registrations, and the two must agree — this is the one place the parts are named twice.
//
// The audit recorder is the one the process built, because the log's prefix and redactions are
// decided where the log is built. Everything else is derived: the outbox writer carries the
// search index rules, the webhooks Emitter fans out over this application's catalog, and the
// Recorder files by subject and attributes through the session.
func New(ctx context.Context, db database.Client, auditRecorder platformaudit.Recorder, opts ...Option) (*Emitter, error) {
	if db == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	o := &options{topic: queuescfg.DefaultDataChangesTopicName, pillars: &observability.Pillars{}}
	for _, opt := range opts {
		if opt != nil {
			opt(o)
		}
	}

	effect, err := indexevents.NewSideEffect()
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the search index side effect")
	}

	writer, err := outbox.NewWriter(db.Dialect(),
		outbox.WithWriterSideEffect(indexevents.SideEffectName, effect),
		outbox.WithWriterLogger(o.pillars.Logger),
		outbox.WithWriterTracerProvider(o.pillars.TracerProvider),
		outbox.WithWriterMetricsProvider(o.pillars.MetricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the outbox writer")
	}

	cfg := o.webhooks
	cfg.EmitterTopic = o.topic

	emitter, err := webhookscfg.NewEmitter(ctx, &cfg, db, writer, catalog.Catalog(), webhookscfg.WithPillars(o.pillars))
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the webhooks emitter")
	}

	recorder, err := recordingcfg.NewRecorder(ctx,
		&recordingcfg.Config{FileBy: recordingcfg.FileBySubject},
		auditRecorder, emitter, sessions.PrincipalFromContext,
		recordingcfg.WithPillars(o.pillars),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the recording recorder")
	}

	return NewEmitter(emitter, recorder, writer, effect)
}
