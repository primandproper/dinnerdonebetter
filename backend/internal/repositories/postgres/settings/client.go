package settings

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/settings/generated"

	"github.com/primandproper/platform-go/v6/database"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"
)

const (
	o11yName = "settings_db_client"
)

// Repository is the settings repository implementation.
// Exported so the manager can wrap it; the manager is the sole provider of settings.Repository for services.
type Repository struct {
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
	database.Client
}

// ProvideSettingsRepository provides a new repository.
func ProvideSettingsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
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
