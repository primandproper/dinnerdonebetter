package waitlists

import (
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/waitlists/generated"

	"github.com/primandproper/platform-go/v4/database"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

const (
	o11yName = "waitlists_db_client"
)

// Repository is the waitlists repository implementation.
type Repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	readDB            *sql.DB
	writeDB           *sql.DB
}

// ProvideWaitlistsRepository provides a new repository.
// Returns concrete *Repository so the manager can wrap it; the manager is the sole provider of waitlists.Repository for services.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
) *Repository {
	c := &Repository{
		Client:            client,
		readDB:            client.ReadDB(),
		writeDB:           client.WriteDB(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}

// Ensure *Repository implements the interface.
var _ waitlists.Repository = (*Repository)(nil)
