package notifications

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v13/database"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "notifications_db_client"
)

// Repository is the notifications repository implementation.
// Exported so the manager can wrap it; the manager is the sole provider of notifications.Repository for services.
type Repository struct {
	database.Client
	tracer           tracing.Tracer
	logger           logging.Logger
	generatedQuerier generated.Querier
	events           *events.Emitter
	recorder         *recording.Recorder
	readDB           database.SQLQueryExecutor
	writeDB          database.SQLQueryExecutor
}

// ProvideNotificationsRepository provides a new repository.
func ProvideNotificationsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	cfg *databasecfg.Config,
	client database.Client,
	eventEmitter *events.Emitter,
) *Repository {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	c := &Repository{
		Client:           client,
		readDB:           client.Reader(),
		writeDB:          client.Writer(),
		tracer:           tracer,
		generatedQuerier: generated.New(),
		events:           eventEmitter,
		recorder:         recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		logger:           logging.NewNamedLogger(logger, o11yName),
	}

	return c
}
