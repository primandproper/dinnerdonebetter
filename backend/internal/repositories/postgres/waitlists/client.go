package waitlists

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbwaitlists "github.com/primandproper/dinnerdonebetter/backend/internal/domain/waitlists"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformwaitlists "github.com/primandproper/platform-go/v13/waitlists"
)

const (
	o11yName = "waitlists_db_client"
)

// repository is platform's waitlist store with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the reads —
// every one of them, on both tables — are the platform's own rather than
// forwarding stubs that could drift from it.
type repository struct {
	platformwaitlists.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// ProvideWaitlistsRepository provides a new waitlist store.
func ProvideWaitlistsRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (platformwaitlists.Store, error) {
	store, err := platformwaitlists.NewSQLStore(
		client,
		platformwaitlists.WithTablePrefix(ddbwaitlists.TablePrefix),
		platformwaitlists.WithStoreLogger(logger),
		platformwaitlists.WithStoreTracerProvider(tracerProvider),
		platformwaitlists.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the waitlists store")
	}

	return &repository{
		Store:             store,
		client:            client,
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
	}, nil
}

// recordAndEmit writes the audit log entry and the data change event for one write, as two
// further statements in the transaction that performed it. Together, because the failure mode of
// the two blocks it replaces was omission and omission is silent: a row nothing recorded has no
// provenance and the chain cannot notice, and a change nothing emitted leaves the search index
// stale and no webhook fired. See docs/audit.md.
//
// accountID is "" for every caller today: a waitlist entry predates any account.
func (r *repository) recordAndEmit(
	ctx context.Context,
	tx database.Tx,
	logger logging.Logger,
	entry *audit.AuditLogEntry,
	eventType, accountID string,
	metadata map[string]any,
) error {
	ctx, span := r.tracer.StartSpan(ctx)
	defer span.End()

	if err := r.auditLogEntryRepo.Record(ctx, tx, entry); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "creating audit log entry")
	}

	if err := r.events.Emit(ctx, tx, logger, eventType, accountID, metadata); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "enqueuing data change event")
	}

	return nil
}
