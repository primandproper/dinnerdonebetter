package webhooks

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
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
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
}

// ProvideWebhooksRepository provides a new repository.
func ProvideWebhooksRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) webhooks.Repository {
	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
