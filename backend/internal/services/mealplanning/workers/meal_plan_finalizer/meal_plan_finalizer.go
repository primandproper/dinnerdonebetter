package mealplanfinalizer

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const (
	serviceName = "meal_plan_finalizer"
)

var _ workers.WorkerCounter = (*Worker)(nil)

type Worker struct {
	logger logging.Logger
	tracer tracing.Tracer

	dataManager             mealplanning.Repository
	finalizedRecordsCounter metrics.Int64Counter
}

// NewMealPlanFinalizer builds the worker.
//
// It takes no publisher: the finalized event is written into the outbox by the repository,
// inside the transaction that finalizes the plan.
func NewMealPlanFinalizer(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	dataManager mealplanning.Repository,
	metricsProvider metrics.Provider,
) (*Worker, error) {
	finalizedRecordsCounter, err := metricsProvider.NewInt64Counter("meal_plan_finalizer.finalized_records")
	if err != nil {
		return nil, err
	}

	return &Worker{
		dataManager:             dataManager,
		finalizedRecordsCounter: finalizedRecordsCounter,

		logger: logging.NewNamedLogger(logger, serviceName),
		tracer: tracing.NewNamedTracer(tracerProvider, serviceName),
	}, nil
}

func (w *Worker) Work(ctx context.Context) (int64, error) {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	logger := w.logger.Clone()
	logger.Info("beginning finalization of expired meal plans")

	mealPlans, err := w.dataManager.GetUnfinalizedMealPlansWithExpiredVotingPeriods(ctx)
	if err != nil {
		return -1, observability.PrepareAndLogError(err, logger, span, "fetching unfinalized and expired meal plan")
	}

	if len(mealPlans) > 0 {
		logger.WithValue("quantity", len(mealPlans)).Info("finalizing expired meal plans")
	}

	errorResult := &multierror.Error{}

	var changedCount int64
	for _, mealPlan := range mealPlans {
		var changed bool
		changed, err = w.dataManager.AttemptToFinalizeMealPlan(ctx, mealPlan.ID, mealPlan.BelongsToAccount)
		if err != nil {
			if errors.Is(err, mealplanningrepo.ErrAlreadyFinalized) {
				logger.WithValue(mealplanningkeys.MealPlanIDKey, mealPlan.ID).Info("meal plan already finalized, skipping")
				continue
			}
			// Record the failure and keep going so one bad meal plan doesn't permanently block the rest of the batch.
			errorResult = multierror.Append(errorResult, observability.PrepareError(err, span, "finalizing meal plan"))
			continue
		}

		if changed {
			changedCount++
		}
	}

	w.finalizedRecordsCounter.Add(ctx, changedCount)
	logger.WithValue("changed_count", changedCount).Info("finalized expired meal plans")

	return changedCount, errorResult.ErrorOrNil()
}
