package integration

import (
	"context"
	"errors"
	"sync"
	"testing"

	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/mealplanning/generated"
	mealplantasknotifications "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_notifications"

	platformnotifications "github.com/primandproper/platform-go/v13/notifications/mobile"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/workqueue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recordingPushSender stands in for APNs and counts what reached each device token.
//
// It counts per token rather than in total because this suite runs in parallel and the worker's
// discovery query is database-wide: a pass sweeps up every other test's outstanding tasks too.
// Those recipients have registered no device, so they cost nothing and reach nobody — but the
// only count that means anything here is the one scoped to this test's own token.
type recordingPushSender struct {
	counts map[string]int
	// err, when set, is what every device refuses with.
	err error
	mu  sync.Mutex
}

func newRecordingPushSender() *recordingPushSender {
	return &recordingPushSender{counts: map[string]int{}}
}

func (s *recordingPushSender) SendPush(_ context.Context, _, token string, _ platformnotifications.PushMessage) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.counts[token]++

	return s.err
}

func (s *recordingPushSender) countFor(token string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.counts[token]
}

// notificationSweepMu orders the tests that run a notification worker against each other.
//
// The worker's discovery query is database-wide: it asks for every meal plan task in the database
// still owed a reminder, not for this test's. Partitioning the queue keeps two workers' *claims*
// apart, but nothing can keep their *discoveries* apart, because the tasks are the shared thing —
// so a worker with a succeeding sender would sweep up and stamp the tasks a worker with a failing
// sender is asserting on, and both tests would be measuring the other's.
//
// Every other test in this suite is unaffected and stays parallel: their tasks get swept and
// stamped, and nothing asserts on that. Only tests that assert on a sweep have to take turns.
var notificationSweepMu sync.Mutex

// serializeNotificationSweep holds that order for the whole of a test, from before it creates the
// tasks it cares about until after its last assertion about them.
func serializeNotificationSweep(t *testing.T) {
	t.Helper()

	notificationSweepMu.Lock()
	t.Cleanup(notificationSweepMu.Unlock)
}

// buildNotificationWorkerForTest stands the real worker up over the real queue, in a partition
// named for the test, pushing through sender.
func buildNotificationWorkerForTest(t *testing.T, sender platformnotifications.PushNotificationSender, queueOpts ...func(*workqueue.Config)) *mealplantasknotifications.Worker {
	t.Helper()

	worker, closeQueue, err := localdev.NewMealPlanTaskNotificationWorker(
		t.Context(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		databaseClient,
		sender,
		t.Name(),
		queueOpts...,
	)
	require.NoError(t, err)

	// Closed with a background context: the queue owns a goroutine, and t.Context() is already
	// cancelled by the time cleanups run.
	t.Cleanup(func() { require.NoError(t, closeQueue(context.Background())) })

	return worker
}

// notificationHasBeenSent reads the stamp through the generated query the worker itself checks
// first, rather than through a SELECT written here.
//
// That query is `SELECT notification_sent_at IS NOT NULL`, whose column has no name and whose
// generated signature is therefore `interface{}`; the repository type-asserts it to bool. Nothing
// but a real driver can say whether that assertion holds, and if it stopped holding, every
// notification would fail on its first step with a type error.
func notificationHasBeenSent(t *testing.T, ctx context.Context, mealPlanTaskID string) bool {
	t.Helper()

	result, err := generated.New().MealPlanTaskNotificationHasBeenSent(ctx, databaseClient.Reader(), mealPlanTaskID)
	require.NoError(t, err)

	sent, ok := result.(bool)
	require.True(t, ok, "the driver must scan `notification_sent_at IS NOT NULL` as a bool, got %T", result)

	return sent
}

// TestMealPlanTaskNotifications_Worker is the assertion the work queue was adopted for.
//
// Before it, the scheduler published one message per unnotified task on every tick and the async
// consumer stamped notification_sent_at some time later, in another process. Between those two
// moments every tick republished the same task, and a task whose notification could not be built
// was rediscovered and re-failed until its event started. The queue's completion and the stamp
// are now the same fact, written by the same pass — and the only way to show that is to run two
// passes against a real database and count what reached the device.
//
// Each subtest claims from a partition of its own. Two claimants on one logical queue is exactly
// what the lease permits, so siblings sharing production's queue name could legitimately claim
// each other's tasks and push them through the wrong sender — a per-sender count would then be
// non-deterministic without anything being wrong. The name is a partition key, so a test that
// wants isolation takes its own; TestMealPlanTaskNotifications_QueueName is what pins the one
// production uses.
func TestMealPlanTaskNotifications_Worker(T *testing.T) {
	T.Parallel()

	T.Run("notifies once and does not notify again", func(t *testing.T) {
		t.Parallel()
		serializeNotificationSweep(t)

		ctx := t.Context()

		mealPlanID, userClient := createFinalizedMealPlanWithTasks(t)
		tasks := awaitMealPlanTasks(t, ctx, userClient, mealPlanID)
		require.NotEmpty(t, tasks)

		status, err := userClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
		require.NoError(t, err)
		require.NotEmpty(t, status.UserId)

		// The tasks the finalization saga wrote are unassigned, so their recipients are
		// everybody in the account — which includes this user, and now this device.
		deviceToken := createUserDeviceTokenForTest(t, status.UserId)

		sender := newRecordingPushSender()
		worker := buildNotificationWorkerForTest(t, sender)

		// One pass: discover, enqueue, claim under a lease, send, stamp, complete.
		sent, err := worker.Work(ctx)
		require.NoError(t, err)
		require.NotZero(t, sent)

		firstPass := sender.countFor(deviceToken.DeviceToken)
		require.NotZero(t, firstPass, "the first pass should have pushed to this device")

		// The stamp and the queue's completion have to agree, or the next discovery finds the
		// task unsent, re-enqueues it, and restarts the item it just completed.
		for _, task := range tasks {
			assert.True(t, notificationHasBeenSent(t, ctx, task.Id),
				"task %s should be stamped by the pass that notified it", task.Id)
		}

		// The second pass is the whole test. Discovery no longer returns these tasks, so
		// nothing re-enqueues them, so the completed items stay completed.
		_, err = worker.Work(ctx)
		require.NoError(t, err)

		assert.Equal(t, firstPass, sender.countFor(deviceToken.DeviceToken),
			"a second pass must not re-notify a task the first one completed")
	})
}

// TestMealPlanTaskNotifications_PoisonTask covers the other half of the lifecycle, and is the
// failure the old design could not express at all.
//
// A push every device refused was logged and forgotten. The task stayed unstamped, so the next
// tick rediscovered it, republished it, and failed again — for as long as its event was in the
// future, with no attempt count, no ceiling, and no record of why. What ends that now is
// MaxAttempts: the task is claimed, failed and released with its cause until it runs out of
// attempts, and then it is stalled rather than retried forever.
//
// Two attempts here rather than production's five, because the point is to reach the ceiling.
func TestMealPlanTaskNotifications_PoisonTask(T *testing.T) {
	T.Parallel()

	T.Run("stops being retried once it runs out of attempts", func(t *testing.T) {
		t.Parallel()
		serializeNotificationSweep(t)

		ctx := t.Context()

		mealPlanID, userClient := createFinalizedMealPlanWithTasks(t)
		tasks := awaitMealPlanTasks(t, ctx, userClient, mealPlanID)
		require.NotEmpty(t, tasks)

		status, err := userClient.GetAuthStatus(ctx, &authgrpc.GetAuthStatusRequest{})
		require.NoError(t, err)

		deviceToken := createUserDeviceTokenForTest(t, status.UserId)

		sender := newRecordingPushSender()
		sender.err = errors.New("APNs is having a moment")
		worker := buildNotificationWorkerForTest(t, sender, func(cfg *workqueue.Config) {
			cfg.MaxAttempts = 2
		})

		// Each pass claims every outstanding task exactly once: the release delay is longer
		// than the pass, so the drain loop cannot re-serve what it has just failed.
		//
		// The pass reports the failure rather than swallowing it, so a job that cannot notify
		// anybody is a failed run rather than a quiet one. Its count of completed tasks is
		// not asserted on and cannot be: discovery is database-wide, so a pass also finishes
		// every device-less task the rest of the suite has left lying around — correctly,
		// as unreachable. Only what reached this test's own device is this test's.
		_, err = worker.Work(ctx)
		require.Error(t, err)

		afterFirst := sender.countFor(deviceToken.DeviceToken)
		require.NotZero(t, afterFirst, "the send should have been attempted")

		// A failed send must never stamp the task, or the reminder is silently lost.
		for _, task := range tasks {
			assert.False(t, notificationHasBeenSent(t, ctx, task.Id),
				"task %s must stay unstamped when nothing accepted its push", task.Id)
		}

		// The second pass is the last one allowed. Discovery re-offers the released task —
		// an enqueue can only make an outstanding item available sooner — but the claim
		// still counts against the ceiling.
		_, err = worker.Work(ctx)
		require.Error(t, err)

		afterSecond := sender.countFor(deviceToken.DeviceToken)
		require.Greater(t, afterSecond, afterFirst, "the second attempt should have been made")

		// The third finds this test's tasks unclaimable. They are stalled: excluded from
		// every claim, counted by Stats, and still in the table for an operator to read.
		_, err = worker.Work(ctx)
		require.NoError(t, err, "a pass whose only failing tasks are stalled is not a failed pass")

		assert.Equal(t, afterSecond, sender.countFor(deviceToken.DeviceToken),
			"a task out of attempts must stop being retried")
	})
}

// TestMealPlanTaskNotifications_QueueName pins the logical queue's name.
//
// One table holds every logical queue, partitioned by this string, and the operations tier is a
// second queue in it. A rename would fail nothing loudly — it would start claiming from an empty
// partition and leave the old one's rows outstanding forever.
func TestMealPlanTaskNotifications_QueueName(T *testing.T) {
	T.Parallel()

	T.Run("is stable", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, "meal_plan_task_notifications", mealplantasknotifications.QueueName)
	})
}
