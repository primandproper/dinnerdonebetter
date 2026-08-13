package webhooks

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v10/database"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

const (
	o11yName = "webhook_db_client"
)

// repository is the webhook repository client.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
	// dispatcher owns the delivery side of a webhook: the endpoint, its signing secret, and
	// its subscriptions. It is required rather than optional, because a webhook that is stored
	// and not registered is one the account was told exists and that will never fire.
	dispatcher *webhookdispatch.Dispatcher
	readDB     database.SQLQueryExecutor
	writeDB    database.SQLQueryExecutor
}

// ProvideWebhooksRepository provides a new repository.
func ProvideWebhooksRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
	dispatcher *webhookdispatch.Dispatcher,
) webhooks.Repository {
	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
		dispatcher:        dispatcher,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
