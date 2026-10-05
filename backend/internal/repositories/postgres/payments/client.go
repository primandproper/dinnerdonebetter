package payments

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v15/billing"
	billingcfg "github.com/primandproper/platform-go/v15/billing/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "payments_db_client"
)

// ProvidePaymentsRepository provides platform's billing store, with this
// application's recording hung off its writes.
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
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return newStore(ctx, logger, tracerProvider, metricsProvider, client, &hooks{
		logger:            logging.NewNamedLogger(logger, o11yName),
		auditLogEntryRepo: auditLogEntryRepo,
		recorder:          recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	})
}

// newStore builds the billing store with h installed as its hooks. It is
// ProvidePaymentsRepository with the recording already assembled, which is what
// lets a test install recording that fails.
func newStore(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	h *hooks,
) (billing.Store, error) {
	store, err := billingcfg.NewStore(
		ctx,
		&billingcfg.Config{TablePrefix: branding.TablePrefix},
		client,
		billingcfg.WithLogger(logger),
		billingcfg.WithTracerProvider(tracerProvider),
		billingcfg.WithMetricsProvider(metricsProvider),
		billingcfg.WithStoreOptions(billing.WithHooks(h)),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the billing store")
	}

	return store, nil
}
