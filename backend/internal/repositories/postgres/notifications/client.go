package notifications

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/notifications/generated"

	"github.com/primandproper/platform-go/v13/database"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/observability"
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
	tracer            tracing.Tracer
	logger            logging.Logger
	generatedQuerier  generated.Querier
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
	readDB            database.SQLQueryExecutor
	writeDB           database.SQLQueryExecutor
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

// recordAndEmit writes the audit log entry and the data change event for one write, as two
// further statements in the transaction that performed it. Together, because the failure mode of
// the two blocks it replaces was omission and omission is silent: a row nothing recorded has no
// provenance and the chain cannot notice, and a change nothing emitted leaves the search index
// stale and no webhook fired. See docs/audit.md.
//
//nolint:unparam // accountID is "" for every caller today; notifications are addressed to a user
func (q *Repository) recordAndEmit(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *audit.AuditLogEntry,
	eventType, accountID string,
	metadata map[string]any,
) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	if err := q.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
	}

	if err := q.events.Emit(ctx, tx, logger, eventType, accountID, metadata); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "enqueuing data change event")
	}

	return nil
}
