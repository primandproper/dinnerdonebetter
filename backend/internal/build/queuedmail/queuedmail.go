/*
Package queuedmail registers platform's notifications/mail QueuedMailer: the one mailer every
identity, sign-in, password reset and waitlist door here hands its mail to.

Each of those doors calls its mailer after the write it follows has committed, and the mailer
puts the mail on the outbox in a transaction of its own, under a topic only the mail Drainer
reads. The async message handler runs that Drainer, rendering each mail through
internal/services/identity/emails.

The mail topic carries live credentials — a reset's secret, an invitation's token, the waitlist's
two links — for as long as a message sits on it, so it has an outbox writer of its own: the
process's shared writer carries the search index side effect, and a side effect sees every
message it is handed, this one included. See platform's notifications/mail package
documentation for the rest of what the topic obliges.
*/
package queuedmail

import (
	queuescfg "github.com/primandproper/dinnerdonebetter/backend/internal/queues/config"

	"github.com/primandproper/platform-go/v15/notifications/mail"
	"github.com/primandproper/platform-go/v15/outbox"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// New builds the QueuedMailer over db, enqueuing on topic through an outbox writer that carries no
// side effect. It is Register's construction, for a process that assembles its services by hand.
func New(db database.Client, topic string, logger logging.Logger, tracerProvider tracing.Provider, metricsProvider metrics.Provider) (*mail.QueuedMailer, error) {
	if db == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	writer, err := outbox.NewWriter(db.Dialect(),
		outbox.WithWriterLogger(logger),
		outbox.WithWriterTracerProvider(tracerProvider),
		outbox.WithWriterMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the mail outbox writer")
	}

	return mail.NewQueuedMailer(db, writer, topic,
		mail.WithLogger(logger),
		mail.WithTracerProvider(tracerProvider),
		mail.WithMetricsProvider(metricsProvider),
	)
}

// Register registers the QueuedMailer with the injector.
func Register(i do.Injector) {
	do.Provide[*mail.QueuedMailer](i, func(i do.Injector) (*mail.QueuedMailer, error) {
		return New(
			do.MustInvoke[database.Client](i),
			topic(i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// topic is the mail topic: the configured one where this process has a queues config, and the
// default otherwise. Every process that builds the identity service owes an invitation its mail,
// and not every one of them configures a broker; the outbox holds the mail for the worker that
// does.
func topic(i do.Injector) string {
	if queues, err := do.Invoke[*queuescfg.Config](i); err == nil && queues != nil && queues.QueuedMailTopicName != "" {
		return queues.QueuedMailTopicName
	}

	return queuescfg.DefaultQueuedMailTopicName
}
