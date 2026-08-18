package mealplanning

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"

	"github.com/primandproper/platform-go/v11/database"
	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

const (
	o11yName = "meal_planning_db_client"
)

// repository is the meal planning repository implementation.
type repository struct {
	database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	identityRepo      identity.Repository
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
}

// ProvideMealPlanningRepository provides a new repository.
//
// eventEmitter may be nil, in which case no data change events are written to the outbox and
// the caller is expected to publish them itself — which is what every unconverted method here
// still does.
func ProvideMealPlanningRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	auditLogEntryRepo audit.Repository,
	identityRepo identity.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) mealplanning.Repository {
	c := &repository{
		Client:            client,
		readDB:            client.Reader(),
		writeDB:           client.Writer(),
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		generatedQuerier:  generated.New(),
		auditLogEntryRepo: auditLogEntryRepo,
		identityRepo:      identityRepo,
		events:            eventEmitter,
		logger:            logging.NewNamedLogger(logger, o11yName),
	}

	return c
}

// withEvent runs a write and the data change event describing it in one transaction, so the
// event cannot survive a write that rolled back — nor be lost after one that committed.
//
// accountID is passed explicitly wherever the repository knows it; see events.Emitter.Emit.
// The enumeration tables (valid ingredients, vessels, preparations, and friends) are global
// catalog data owned by no account, so they pass "".
//
// It takes no EmitOptions. It used to, so that a write to an indexed table could pass the search
// index event as one — which made a thing every such write owes into a thing a call site could
// forget. That obligation is registered on the outbox writer now; see internal/indexevents.
func (q *repository) withEvent(
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
