package waitlists

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/waitlists/generated"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"
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
	events            *events.Emitter
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
}

// ProvideWaitlistsRepository provides a new repository.
// Returns concrete *Repository so the manager can wrap it; the manager is the sole provider of waitlists.Repository for services.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) *Repository {
	c := &Repository{
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

// Ensure *Repository implements the interface.
var _ waitlists.Repository = (*Repository)(nil)

// withEvent runs a write and the data change event describing it in one transaction, so the
// event cannot survive a write that rolled back — nor be lost after one that committed.
//
//nolint:unparam // accountID is "" for every caller today; see the payments repository.
func (q *Repository) withEvent(
	ctx context.Context,
	logger logging.Logger,
	eventType, accountID string,
	metadata map[string]any,
	write func(tx database.SQLQueryExecutor) error,
) error {
	return q.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		if err := write(tx); err != nil {
			return err
		}

		return q.events.Emit(ctx, tx, logger, eventType, accountID, metadata)
	})
}
