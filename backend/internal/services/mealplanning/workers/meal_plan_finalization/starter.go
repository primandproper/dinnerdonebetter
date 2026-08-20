/*
Package mealplanfinalization carries a meal plan from "voting is over" to "there are prep tasks
and a grocery list", as a platform-go saga.

It used to be three interval-polled jobs — a finalizer, a task creator, and a grocery list
initializer — coordinated by two boolean columns, each rediscovering its own work with a query
for "finalized but not yet X". That cost a plan the sum of the three intervals before its
grocery list existed, had no compensation when a stage failed for good, and left no record of
which stage a plan was in or how many times it had been tried.

What is here instead is a Starter, which claims plans and writes one saga instance each, and the
saga's Steps, which the shared saga.Worker advances. A pass runs as many steps as its budget
allows, so a plan whose deadline passes is finalized, tasked, and shopped for in one go.

The two flag columns stayed. They stopped being the coordinator and became each step's
idempotency guard, which is what they were always better at: each is written in the same
transaction as the work it describes, and no idempotency key that commits separately can promise
as much.
*/
package mealplanfinalization

import (
	"context"
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	mealplanningrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v12/database"
	"github.com/primandproper/platform-go/v12/observability"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/metrics"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/saga"

	"github.com/hashicorp/go-multierror"
)

const serviceName = "meal_plan_finalization_starter"

// candidateBatchSize bounds how many sagas one tick starts.
//
// It is a bound on the first tick after a deploy, not on throughput: a plan is claimed once and
// stays claimed, so the backlog drains at this many per tick and steady state is however many
// plans came due since the last one. Unbounded, the tick that first sees a backlog would write
// every instance in it and then hand the saga worker more in-flight sagas than its batch size
// and concurrency were sized for.
const candidateBatchSize = 100

var _ workers.WorkerCounter = (*Starter)(nil)

// Starter is the scheduled half of meal plan finalization: it finds plans the pipeline owes
// something to and writes one saga instance for each.
//
// It does no pipeline work itself. Advancing is the saga worker's job, and a Starter that ran
// the steps inline would be the three jobs again with extra rows.
type Starter struct {
	logger logging.Logger
	tracer tracing.Tracer

	dataManager    mealplanning.Repository
	runner         saga.Runner[mealplanning.MealPlanFinalizationState]
	startedCounter metrics.Int64Counter
}

// NewStarter builds the Starter.
func NewStarter(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	dataManager mealplanning.Repository,
	runner saga.Runner[mealplanning.MealPlanFinalizationState],
	metricsProvider metrics.Provider,
) (*Starter, error) {
	startedCounter, err := metricsProvider.NewInt64Counter("meal_plan_finalization.sagas_started")
	if err != nil {
		return nil, err
	}

	return &Starter{
		dataManager:    dataManager,
		runner:         runner,
		startedCounter: startedCounter,
		logger:         logging.NewNamedLogger(logger, serviceName),
		tracer:         tracing.NewNamedTracer(tracerProvider, serviceName),
	}, nil
}

// Work starts a finalization saga for every unclaimed plan that needs one, and reports how many
// it started.
func (w *Starter) Work(ctx context.Context) (int64, error) {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	logger := w.logger.Clone()

	candidates, err := w.dataManager.GetMealPlansAwaitingFinalizationSaga(ctx, candidateBatchSize)
	if err != nil {
		return -1, observability.PrepareAndLogError(err, logger, span, "fetching meal plans awaiting finalization")
	}

	if len(candidates) > 0 {
		logger.WithValue("quantity", len(candidates)).Info("starting meal plan finalization sagas")
	}

	errorResult := &multierror.Error{}

	var started int64
	for _, candidate := range candidates {
		l := logger.Clone().WithValue(mealplanningkeys.MealPlanIDKey, candidate.MealPlanID)

		sagaID, startErr := w.start(ctx, candidate)
		switch {
		case errors.Is(startErr, mealplanningrepo.ErrFinalizationSagaAlreadyAttached):
			// Another replica read the same page and got there first. Not an error, and not
			// worth a log line per tick per plan.
			continue
		case startErr != nil:
			// Recorded and skipped, so one plan that cannot be claimed does not cost the rest
			// of the batch a tick.
			errorResult = multierror.Append(errorResult, observability.PrepareError(startErr, span, "starting meal plan finalization saga"))
			continue
		}

		started++
		l.WithValue("saga_instance_id", sagaID).Info("started meal plan finalization saga")
	}

	w.startedCounter.Add(ctx, started)

	return started, errorResult.ErrorOrNil()
}

// EnsureStarted puts one meal plan into the finalization pipeline if nothing has yet, and does
// nothing for one that is already in it.
//
// It is for the caller that has just finalized a plan on a user's request and wants its tasks and
// grocery list built. That used to mean running the two downstream workers inline, in the
// request, over every finalized plan in the database; this writes one row and returns, and the
// saga worker does the rest of it durably.
func (w *Starter) EnsureStarted(ctx context.Context, mealPlanID, accountID string) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	_, err := w.start(ctx, &mealplanning.MealPlanFinalizationCandidate{MealPlanID: mealPlanID, AccountID: accountID})
	if err != nil && !errors.Is(err, mealplanningrepo.ErrFinalizationSagaAlreadyAttached) {
		return observability.PrepareError(err, span, "starting meal plan finalization saga")
	}

	return nil
}

// start writes one instance and claims its plan, in one transaction.
func (w *Starter) start(ctx context.Context, candidate *mealplanning.MealPlanFinalizationCandidate) (string, error) {
	return w.dataManager.AttachMealPlanFinalizationSaga(ctx, candidate.MealPlanID,
		func(ctx context.Context, q database.SQLQueryExecutor) (string, error) {
			instance, err := w.runner.StartInTransaction(ctx, q, mealplanning.MealPlanFinalizationSagaName,
				mealplanning.MealPlanFinalizationState{
					MealPlanID: candidate.MealPlanID,
					AccountID:  candidate.AccountID,
				},
			)
			if err != nil {
				return "", err
			}

			return instance.ID, nil
		},
	)
}
