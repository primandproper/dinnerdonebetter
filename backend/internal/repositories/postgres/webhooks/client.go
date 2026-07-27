package webhooks

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/webhooks/generated"

	"github.com/primandproper/platform-go/v7/database"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
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
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
}

// ProvideWebhooksRepository provides a new repository.
func ProvideWebhooksRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) webhooks.Repository {
	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
