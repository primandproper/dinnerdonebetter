package payments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v13/billing"
	billingcfg "github.com/primandproper/platform-go/v13/billing/config"
	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
)

const (
	o11yName = "payments_db_client"
)

// repository is platform's billing store with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the reads —
// the catalog, an account's subscriptions, purchases and ledger, and the
// current-subscription read every entitlement check makes — are the platform's
// own rather than forwarding stubs that could drift from it.
type repository struct {
	billing.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	recorder          *recording.Recorder
}

// ProvidePaymentsRepository provides a new billing store.
//
// The store is assembled through platform's own billing/config rather than by
// naming billing.NewSQLStore's options here, so the knobs are stated once
// upstream. The table prefix is the one thing this application decides, and it
// has to match the prefix the migration was rendered with — see
// internal/repositories/postgres/migrations.
func ProvidePaymentsRepository(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (billing.Store, error) {
	store, err := billingcfg.NewStore(
		ctx,
		&billingcfg.Config{TablePrefix: ddbpayments.TablePrefix},
		client,
		billingcfg.WithLogger(logger),
		billingcfg.WithTracerProvider(tracerProvider),
		billingcfg.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the billing store")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &repository{
		Store:             store,
		client:            client,
		tracer:            tracer,
		logger:            logging.NewNamedLogger(logger, o11yName),
		auditLogEntryRepo: auditLogEntryRepo,
		recorder:          recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}
