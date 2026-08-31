package mealplantasknotifications

import (
	"context"
	"fmt"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v13/filtering"
	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	"github.com/primandproper/platform-go/v13/observability"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/hashicorp/go-multierror"
)

const o11yName = "meal_plan_task_notification_worker"

// QueueName is the logical queue this worker claims from. One table holds every logical queue,
// partitioned by this name, so it is also what a second queue would have to differ from.
const QueueName = "meal_plan_task_notifications"

const (
	// claimBatchSize is how many tasks one Claim leases. Each task costs a few queries and one
	// push per registered device, so this is sized to finish comfortably inside claimLease.
	claimBatchSize = 100

	// claimLease is what a batch promises to finish in. Past it the tasks are handed back and
	// somebody else may do them again — which is waste rather than a second push, because the
	// stamp is read before every send.
	claimLease = 5 * time.Minute

	// releaseDelay holds a failed task back before it is offered again. It is longer than the
	// job's own timeout on purpose: a drain pass must never re-serve a task it has already
	// failed this pass, or one bad task spends the whole pass being retried.
	//
	// It does not survive the next pass, and is not meant to. Enqueue can only move an
	// outstanding item's availability *earlier* — LEAST of the delays — so the next discovery
	// offers a released task immediately. The backoff between passes is the schedule, which is
	// hourly; what stops a task that fails every time is MaxAttempts, not this.
	releaseDelay = 15 * time.Minute
)

var _ workers.WorkerCounter = (*Worker)(nil)

// Worker fills and drains the meal plan task notification queue.
//
// Both halves are one Work call rather than a discovery job and a separate claim loop. The queue
// is not here to move work between processes — it is here so that a task which cannot be
// notified stops being retried forever, and so a pass that dies does not strand what it claimed.
// Neither of those needs a second process, and a claimant that only runs on a tick is the same
// claimant.
type Worker struct {
	logger logging.Logger
	tracer tracing.Tracer

	queue        Queue
	dataManager  mealplanning.Repository
	identityRepo identity.Repository
	fanout       *push.Fanout
}

// Queue is the slice of workqueue.Queue[string] this worker drives. It is an interface so the
// tests can drive the worker without a database; the concrete type is what construction takes.
type Queue interface {
	EnqueueKeys(ctx context.Context, keys ...string) error
	Claim(ctx context.Context, limit int, lease time.Duration) ([]Item, error)
	Complete(ctx context.Context, keys ...string) error
	Release(ctx context.Context, delay time.Duration, cause error, keys ...string) error
	Reap(ctx context.Context) (int64, error)
	Stats(ctx context.Context) (Stats, error)
}

// NewWorker builds the Worker.
func NewWorker(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	queue Queue,
	dataManager mealplanning.Repository,
	identityRepo identity.Repository,
	fanout *push.Fanout,
) *Worker {
	return &Worker{
		logger:       logging.NewNamedLogger(logger, o11yName),
		tracer:       tracing.NewNamedTracer(tracerProvider, o11yName),
		queue:        queue,
		dataManager:  dataManager,
		identityRepo: identityRepo,
		fanout:       fanout,
	}
}

// Work enqueues every task that is owed a notification, drains what the queue will hand over,
// and reports how many notifications it sent.
//
// The order matters: Enqueue returns only once its keys are durably in the table, so a task
// discovered by this pass is claimable by this pass. A task already outstanding from an earlier
// pass keeps its attempt count and its delay — an enqueue can bring an item forward, never
// reset it — so re-discovering work in flight costs one upsert and changes nothing.
func (w *Worker) Work(ctx context.Context) (int64, error) {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	errorResult := &multierror.Error{}

	if err := w.discover(ctx); err != nil {
		// Recorded and carried, not returned: the queue may still hold tasks from an earlier
		// pass, and a discovery that failed is no reason to leave them unsent.
		errorResult = multierror.Append(errorResult, err)
	}

	sent, drainErr := w.drain(ctx)
	if drainErr != nil {
		errorResult = multierror.Append(errorResult, drainErr)
	}

	// Housekeeping, after the work rather than before it, so a pass that runs out of time
	// spends what it had on notifications — and skipped outright once it has, because two
	// more statements against a cancelled context report the deadline the drain already did.
	if ctx.Err() == nil {
		// Reap deletes completed items past their retention. Nothing else does, so a
		// deployment that skipped it would keep every notification it has ever sent.
		if _, err := w.queue.Reap(ctx); err != nil {
			errorResult = multierror.Append(errorResult, observability.PrepareError(err, span, "reaping meal plan task notification queue"))
		}

		// Stats records the depth and age gauges. Sampled once a pass, which is what they
		// are for: nothing in the queue fails loudly, so a backlog that has stopped moving
		// looks exactly like an idle one until somebody counts it.
		if _, err := w.queue.Stats(ctx); err != nil {
			errorResult = multierror.Append(errorResult, observability.PrepareError(err, span, "reading meal plan task notification queue stats"))
		}
	}

	return sent, errorResult.ErrorOrNil()
}

// discover enqueues every task the database says is still owed a notification.
//
// It re-derives the whole outstanding set from the tasks themselves every pass, so a key that was
// never enqueued, or one whose notification was un-stamped when its event moved, is picked up on
// the next tick rather than lost.
//
// Re-enqueueing work already in the queue is cheap and nearly inert: an outstanding item keeps
// its attempt count, and the only thing an enqueue can do to it is make it available sooner. That
// last part is why a released task is retried on the next pass rather than after its delay — the
// schedule is the backoff between passes, and MaxAttempts is what ends the retrying.
func (w *Worker) discover(ctx context.Context) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	taskIDs, err := w.dataManager.GetMealPlanTaskIDsThatNeedNotification(ctx)
	if err != nil {
		return observability.PrepareAndLogError(err, w.logger, span, "getting meal plan task IDs that need notification")
	}

	if len(taskIDs) == 0 {
		return nil
	}

	if err = w.queue.EnqueueKeys(ctx, taskIDs...); err != nil {
		return observability.PrepareAndLogError(err, w.logger, span, "enqueuing meal plan task notifications")
	}

	w.logger.WithValue("quantity", len(taskIDs)).Info("enqueued meal plan tasks needing notification")

	return nil
}

// drain claims and works batches until the queue has nothing left to hand over.
//
// Until empty rather than until short: a short batch usually does mean a nearly drained queue,
// but the batch size is capped by configuration as well as by the argument here, so treating a
// short one as the end would silently stop draining the moment somebody lowered MaxClaimBatch.
//
// It terminates without a pass counter because nothing it does re-offers work inside this pass.
// A completed task is gone, and a failed one is released with a delay far longer than the pass
// is allowed to last — discovery, which is the only thing that could hurry it, has already run.
func (w *Worker) drain(ctx context.Context) (int64, error) {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	errorResult := &multierror.Error{}

	var sent int64
	for {
		if ctx.Err() != nil {
			// The job's timeout, almost always. Whatever is still leased comes back when
			// the lease lapses, so stopping here loses a pass rather than the work.
			errorResult = multierror.Append(errorResult, ctx.Err())

			break
		}

		items, err := w.queue.Claim(ctx, claimBatchSize, claimLease)
		if err != nil {
			errorResult = multierror.Append(errorResult, observability.PrepareError(err, span, "claiming meal plan task notifications"))

			break
		}

		if len(items) == 0 {
			break
		}

		batchSent, batchErr := w.workBatch(ctx, items)
		sent += batchSent
		if batchErr != nil {
			errorResult = multierror.Append(errorResult, batchErr)
		}
	}

	return sent, errorResult.ErrorOrNil()
}

// workBatch notifies for one claimed batch, completing what it sends and releasing what it
// could not.
//
// Completion is per batch rather than per task so that a hundred tasks cost one write instead of
// a hundred. Release is likewise batched, but only for tasks that failed the same way — the
// cause is recorded on the row, and a shared cause is what makes it readable.
func (w *Worker) workBatch(ctx context.Context, items []Item) (int64, error) {
	errorResult := &multierror.Error{}

	done := make([]string, 0, len(items))
	failed := make([]string, 0)

	var lastCause error

	for idx := range items {
		item := &items[idx]

		logger := w.logger.Clone().
			WithValue(mealplanningkeys.MealPlanTaskIDKey, item.Key).
			WithValue("attempt", item.Attempts)

		if item.Reclaimed {
			// The previous holder's lease lapsed. Worth a line: a steady trickle is
			// healthy, a rate that tracks the claim rate means the lease is too short.
			logger.Info("reclaimed a lapsed meal plan task notification lease")
		}

		switch err := w.notify(ctx, logger, item.Key); {
		case err != nil:
			lastCause = err
			failed = append(failed, item.Key)
			errorResult = multierror.Append(errorResult, fmt.Errorf("notifying for meal plan task %s: %w", item.Key, err))
		default:
			done = append(done, item.Key)
		}
	}

	if len(done) > 0 {
		if err := w.queue.Complete(ctx, done...); err != nil {
			errorResult = multierror.Append(errorResult, fmt.Errorf("completing meal plan task notifications: %w", err))
		}
	}

	if len(failed) > 0 {
		// Handed back deliberately rather than left to the lease. The delay is the backoff,
		// and the cause is written to the row — which is the whole of what was missing
		// before: a failure that was logged and then forgotten.
		if err := w.queue.Release(ctx, releaseDelay, lastCause, failed...); err != nil {
			errorResult = multierror.Append(errorResult, fmt.Errorf("releasing meal plan task notifications: %w", err))
		}
	}

	return int64(len(done)), errorResult.ErrorOrNil()
}

// notify sends one task's reminder and records that it was sent.
//
// A task nobody can be reached for is completed rather than retried. There is no device to push
// to and no number of attempts that produces one, and leaving it outstanding means rediscovering
// it on every pass until its event starts — which is the loop this whole package exists to end.
// The stamp is what says the task has been dealt with, so it is written either way.
func (w *Worker) notify(ctx context.Context, logger logging.Logger, mealPlanTaskID string) error {
	ctx, span := w.tracer.StartSpan(ctx)
	defer span.End()

	// A lease can lapse while its holder is only slow, so two passes can briefly hold one
	// task. This read is what keeps that from being a second push.
	alreadySent, err := w.dataManager.MealPlanTaskNotificationHasBeenSent(ctx, mealPlanTaskID)
	if err != nil {
		return observability.PrepareError(err, span, "checking whether the notification was already sent")
	}

	if alreadySent {
		return nil
	}

	task, err := w.dataManager.GetMealPlanTask(ctx, mealPlanTaskID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching meal plan task")
	}
	if task == nil {
		return fmt.Errorf("meal plan task %s not found", mealPlanTaskID)
	}

	notificationContext, err := w.dataManager.GetMealPlanTaskNotificationContext(ctx, mealPlanTaskID)
	if err != nil {
		return observability.PrepareError(err, span, "fetching meal plan task notification context")
	}
	if notificationContext == nil {
		return fmt.Errorf("meal plan task notification context %s not found", mealPlanTaskID)
	}

	recipients, err := w.recipients(ctx, task)
	if err != nil {
		return observability.PrepareError(err, span, "resolving notification recipients")
	}

	title, body := content(notificationContext)

	result, err := w.fanout.Send(ctx, RequestType, recipients, platformnotifications.PushMessage{Title: title, Body: body})
	if err != nil {
		return observability.PrepareError(err, span, "sending meal plan task notification")
	}

	if !result.Reached() && !result.Unreachable() {
		// There were devices and every one of them refused, which is usually the push
		// provider rather than the task. Left unstamped so the queue offers it again after
		// the release delay.
		//
		// Unreachable is the other case and is deliberately not this one: recipients with no
		// registered device are stamped and completed below, because no number of attempts
		// conjures a phone.
		return errNoDeviceAccepted
	}

	if err = w.dataManager.MarkMealPlanTaskNotificationSent(ctx, mealPlanTaskID); err != nil {
		return observability.PrepareError(err, span, "marking meal plan task notification sent")
	}

	logger.Info("sent meal plan task notification")

	return nil
}

// recipients resolves who a task's reminder goes to: whoever it is assigned to, or everybody in
// the account when it is assigned to nobody.
func (w *Worker) recipients(ctx context.Context, task *mealplanning.MealPlanTask) ([]string, error) {
	if task.AssignedToUser != nil && *task.AssignedToUser != "" {
		return []string{*task.AssignedToUser}, nil
	}

	accountID, err := w.dataManager.GetMealPlanTaskAccountID(ctx, task.ID)
	if err != nil {
		return nil, fmt.Errorf("getting account ID for task: %w", err)
	}
	if accountID == "" {
		return nil, errTaskHasNoAccount
	}

	users, err := w.identityRepo.GetUsersForAccount(ctx, accountID, filtering.DefaultQueryFilter())
	if err != nil {
		return nil, fmt.Errorf("getting users for account: %w", err)
	}

	userIDs := make([]string, 0, len(users.Data))
	for _, user := range users.Data {
		if user != nil && user.ID != "" {
			userIDs = append(userIDs, user.ID)
		}
	}

	return userIDs, nil
}
