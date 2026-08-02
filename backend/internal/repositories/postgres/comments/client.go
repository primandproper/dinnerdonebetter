package comments

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/comments/generated"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
)

const (
	o11yName = "comments_db_client"
)

// repository is the comments repository implementation.
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

// ProvideCommentsRepository provides a new repository.
func ProvideCommentsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) comments.Repository {
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
