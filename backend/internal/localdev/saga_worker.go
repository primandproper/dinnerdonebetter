package localdev

import (
	"context"
	"fmt"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	identityrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/identity"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/database/dialect"
	pglock "github.com/primandproper/platform-go/v9/distributedlock/postgres"
	"github.com/primandproper/platform-go/v9/observability/logging"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/outbox"
	"github.com/primandproper/platform-go/v9/saga"
)

// StartSagaWorker builds a saga worker over the given database and runs it until the returned
// stop function is called.
//
// The API service starts sagas and does not advance them — advancing is the scheduler's job, and
// tying a durable process to the lifetime of a request is the thing a saga exists to avoid. An
// in-process harness has no scheduler, so without this the sagas the API starts would sit at step
// zero and nothing downstream of finalization would ever happen.
//
// The dependencies are constructed here rather than taken from the API server's container
// because that container deliberately does not expose itself. They are cheap — a repository over
// the caller's existing pool, and two stateless generators — and building them twice costs
// nothing a test would notice.
func StartSagaWorker(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	databaseClient database.Client,
) (stop func(context.Context) error, err error) {
	outboxWriter, err := outbox.NewWriter(dialect.Postgres, outbox.WithWriterLogger(logger))
	if err != nil {
		return nil, fmt.Errorf("building outbox writer: %w", err)
	}

	auditRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, metricsnoop.NewMetricsProvider(), databaseClient)
	if err != nil {
		return nil, fmt.Errorf("building audit log repository: %w", err)
	}

	identityRepo := identityrepo.ProvideIdentityRepository(logger, tracerProvider, auditRepo, databaseClient, nil)

	registry := saga.NewRegistry()
	if err = mealplanfinalization.Register(
		registry,
		mealplanningrepo.ProvideMealPlanningRepository(logger, tracerProvider, auditRepo, identityRepo, databaseClient, nil),
		recipeanalysis.NewRecipeAnalyzer(logger, tracerProvider),
		grocerylistpreparation.NewGroceryListCreator(logger, tracerProvider),
		logger,
	); err != nil {
		return nil, fmt.Errorf("registering meal plan finalization saga: %w", err)
	}

	store, err := saga.NewSQLStore(databaseClient, saga.WithStoreLogger(logger), saga.WithStoreTracerProvider(tracerProvider))
	if err != nil {
		return nil, fmt.Errorf("building saga store: %w", err)
	}

	publisher, err := saga.NewOutboxPublisher(outboxWriter)
	if err != nil {
		return nil, fmt.Errorf("building saga event publisher: %w", err)
	}

	// The transaction-scoped locker, which needs only the safe database.Client surface. The
	// worker takes a per-instance lock so two of them cannot step through the same saga; there
	// is only one here, but the worker requires one and a noop would remove the guarantee this
	// harness is meant to be exercising.
	locker, err := pglock.NewPostgresScopedLocker(&pglock.Config{}, databaseClient, nil,
		pglock.WithLogger(logger),
		pglock.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("building saga locker: %w", err)
	}

	cfg := &saga.WorkerConfig{}
	cfg.EnsureDefaults()

	worker, err := saga.NewWorker(ctx, cfg, store, registry, locker,
		saga.WithWorkerEventPublisher(publisher),
		saga.WithWorkerLogger(logger),
		saga.WithWorkerTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, fmt.Errorf("building saga worker: %w", err)
	}

	go worker.Run()

	return worker.Close, nil
}
