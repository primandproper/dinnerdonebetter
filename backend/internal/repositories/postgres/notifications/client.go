package notifications

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"

	"github.com/primandproper/platform-go/v7/database"
	databasecfg "github.com/primandproper/platform-go/v7/database/config"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/tracing"
)

const (
	o11yName = "notifications_db_client"
)

// Repository is the notifications repository implementation.
// Exported so the manager can wrap it; the manager is the sole provider of notifications.Repository for services.
type Repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
}

// ProvideNotificationsRepository provides a new repository.
func ProvideNotificationsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	cfg *databasecfg.Config,
	client database.Client,
) *Repository {
	c := &Repository{
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
