package mealplantasknotifications

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	domainnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	notificationsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications/push"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	identitymock "github.com/primandproper/platform-go/v14/identity/mock"
	"github.com/primandproper/primitives-go/v2/database"
	databasemock "github.com/primandproper/primitives-go/v2/database/mock"
	"github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	platformnotifications "github.com/primandproper/primitives-go/v2/notifications/mobile"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// queueSpy is an in-memory stand-in for the leased queue. It records what the worker did rather
// than reimplementing leasing: the lease semantics belong to platform-go and are tested there,
// and what matters here is which keys the worker completed and which it handed back.
type queueSpy struct {
	cause    error
	claimErr error

	// ready is handed out one batch at a time, so a worker draining until empty terminates.
	ready [][]Item

	enqueued  []string
	completed []string
	released  []string

	mu sync.Mutex

	reaps     int
	statReads int
}

func (q *queueSpy) EnqueueKeys(_ context.Context, keys ...string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.enqueued = append(q.enqueued, keys...)

	return nil
}

func (q *queueSpy) Claim(_ context.Context, _ int, _ time.Duration) ([]Item, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.claimErr != nil {
		return nil, q.claimErr
	}

	if len(q.ready) == 0 {
		return nil, nil
	}

	batch := q.ready[0]
	q.ready = q.ready[1:]

	return batch, nil
}

// Complete and Release take the claimed items rather than their keys as of platform-go v14:
// a lease is released by the claim that holds it, not by the key it happens to name, so a
// stale claimant cannot release work somebody else has since taken. The spy still records the
// keys, because the keys are what the assertions are about.
func (q *queueSpy) Complete(_ context.Context, items ...Item) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range items {
		q.completed = append(q.completed, item.Key)
	}

	return nil
}

func (q *queueSpy) Release(_ context.Context, _ time.Duration, cause error, items ...Item) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	for _, item := range items {
		q.released = append(q.released, item.Key)
	}
	q.cause = cause

	return nil
}

func (q *queueSpy) Reap(context.Context) (int64, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.reaps++

	return 0, nil
}

func (q *queueSpy) Stats(context.Context) (Stats, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.statReads++

	return Stats{}, nil
}

func readyWith(keys ...string) [][]Item {
	batch := make([]Item, 0, len(keys))
	for _, key := range keys {
		batch = append(batch, Item{Key: key, Attempts: 1})
	}

	return [][]Item{batch}
}

// stubSender stands in for APNs. The default accepts everything; a test that wants a refusal
// swaps in its own.
type stubSender struct {
	send func(ctx context.Context, platform, token string, msg platformnotifications.PushMessage) error
}

func (s *stubSender) SendPush(ctx context.Context, platform, token string, msg platformnotifications.PushMessage) error {
	if s.send == nil {
		return nil
	}

	return s.send(ctx, platform, token, msg)
}

type testWorker struct {
	worker  *Worker
	queue   *queueSpy
	repo    *mealplanningmock.RepositoryMock
	users   *identitymock.StoreMock
	devices *notificationsmock.RepositoryMock
	sender  *stubSender
}

func buildTestWorker(t *testing.T) *testWorker {
	t.Helper()

	logger := loggingnoop.NewLogger()
	devices := &notificationsmock.RepositoryMock{}
	sender := &stubSender{}

	fanout, err := push.NewFanout(logger, devices, sender, metricsnoop.NewMetricsProvider())
	require.NoError(t, err)

	queue := &queueSpy{}
	repo := &mealplanningmock.RepositoryMock{}
	users := &identitymock.StoreMock{}

	// The roster read runs on the client's reader. The store is a mock and never touches
	// what it is handed, so a nil executor is the honest value: anything else would be a
	// second thing the test is pretending about.
	db := &databasemock.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }}

	return &testWorker{
		worker:  NewWorker(logger, tracingnoop.NewTracerProvider(), queue, repo, users, db, fanout),
		queue:   queue,
		repo:    repo,
		users:   users,
		devices: devices,
		sender:  sender,
	}
}

// notifiable wires the mocks so that taskID is a task assigned to one user with one live device.
func (w *testWorker) notifiable(t *testing.T, taskID string) (assignedUser string) {
	t.Helper()

	assignedUser = fake.BuildFakeID()

	task := mealplanningfakes.BuildFakeMealPlanTask()
	task.ID = taskID
	task.AssignedToUser = &assignedUser

	w.repo.MealPlanTaskNotificationHasBeenSentFunc = func(context.Context, string) (bool, error) { return false, nil }
	w.repo.GetMealPlanTaskFunc = func(context.Context, string) (*mealplanning.MealPlanTask, error) { return task, nil }
	w.repo.GetMealPlanTaskNotificationContextFunc = func(context.Context, string) (*mealplanning.MealPlanTaskNotificationContext, error) {
		return &mealplanning.MealPlanTaskNotificationContext{
			PrepTaskName: "Chop onions",
			MealName:     mealplanning.DinnerMealName,
			StartsAt:     time.Date(2025, 3, 3, 18, 0, 0, 0, time.UTC), // a Monday
		}, nil
	}
	w.repo.MarkMealPlanTaskNotificationSentFunc = func(context.Context, string) error { return nil }
	w.devices.GetUserDeviceTokensFunc = func(_ context.Context, userID string, _ *filtering.QueryFilter, _ *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
		return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{
			Data: []*domainnotifications.UserDeviceToken{
				{
					ID:            fake.BuildFakeID(),
					DeviceToken:   fake.BuildFakeID(),
					Platform:      domainnotifications.UserDeviceTokenPlatformIOS,
					BelongsToUser: userID,
				},
			},
		}, nil
	}

	return assignedUser
}

func TestWorker_Work(t *testing.T) {
	t.Parallel()

	t.Run("enqueues what needs notifying, sends it, and completes it", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) {
			return []string{taskID}, nil
		}
		w.queue.ready = readyWith(taskID)
		w.notifiable(t, taskID)

		sent, err := w.worker.Work(t.Context())

		require.NoError(t, err)
		assert.Equal(t, int64(1), sent)
		assert.Equal(t, []string{taskID}, w.queue.enqueued)
		assert.Equal(t, []string{taskID}, w.queue.completed)
		assert.Empty(t, w.queue.released)
		assert.Len(t, w.repo.MarkMealPlanTaskNotificationSentCalls(), 1)
		assert.Equal(t, 1, w.queue.reaps)
		assert.Equal(t, 1, w.queue.statReads)
	})

	// The stamp and the queue's completion have to mean the same thing, which is the whole
	// reason the send happens under the lease. A task somebody else already stamped is finished,
	// not re-sent.
	t.Run("completes an already stamped task without pushing again", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.ready = readyWith(taskID)
		w.repo.MealPlanTaskNotificationHasBeenSentFunc = func(context.Context, string) (bool, error) { return true, nil }

		sent, err := w.worker.Work(t.Context())

		require.NoError(t, err)
		assert.Equal(t, int64(1), sent)
		assert.Equal(t, []string{taskID}, w.queue.completed)
		assert.Empty(t, w.repo.GetMealPlanTaskCalls())
		assert.Empty(t, w.repo.MarkMealPlanTaskNotificationSentCalls())
	})

	// This is the failure the old job could not express: a task it could not build a
	// notification for was logged and rediscovered on every tick forever. Released with a
	// cause, it now carries its own backoff and counts against the attempt ceiling.
	t.Run("releases a task whose notification cannot be built", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()
		expected := errors.New("no such task")

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.ready = readyWith(taskID)
		w.repo.MealPlanTaskNotificationHasBeenSentFunc = func(context.Context, string) (bool, error) { return false, nil }
		w.repo.GetMealPlanTaskFunc = func(context.Context, string) (*mealplanning.MealPlanTask, error) {
			return nil, expected
		}

		sent, err := w.worker.Work(t.Context())

		require.Error(t, err)
		assert.Equal(t, int64(0), sent)
		assert.Equal(t, []string{taskID}, w.queue.released)
		require.ErrorIs(t, w.queue.cause, expected)
		assert.Empty(t, w.queue.completed)
	})

	// A push that nobody's device accepted is a transient failure of the provider far more
	// often than of the task, so the stamp is withheld and the queue offers it again.
	t.Run("releases a task no device accepted", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.ready = readyWith(taskID)
		w.notifiable(t, taskID)
		w.sender.send = func(context.Context, string, string, platformnotifications.PushMessage) error {
			return errors.New("APNs is having a moment")
		}

		_, err := w.worker.Work(t.Context())

		require.Error(t, err)
		assert.Equal(t, []string{taskID}, w.queue.released)
		require.ErrorIs(t, w.queue.cause, errNoDeviceAccepted)
		assert.Empty(t, w.repo.MarkMealPlanTaskNotificationSentCalls())
	})

	// Nobody to reach is not a failure and never becomes one, so the task is stamped and
	// completed rather than left to be rediscovered until its event starts.
	t.Run("completes a task whose recipients have no devices", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.ready = readyWith(taskID)
		w.notifiable(t, taskID)
		w.devices.GetUserDeviceTokensFunc = func(context.Context, string, *filtering.QueryFilter, *string) (*filtering.QueryFilteredResult[domainnotifications.UserDeviceToken], error) {
			return &filtering.QueryFilteredResult[domainnotifications.UserDeviceToken]{}, nil
		}

		sent, err := w.worker.Work(t.Context())

		require.NoError(t, err)
		assert.Equal(t, int64(1), sent)
		assert.Equal(t, []string{taskID}, w.queue.completed)
		assert.Len(t, w.repo.MarkMealPlanTaskNotificationSentCalls(), 1)
	})

	t.Run("notifies every member of the account when a task is unassigned", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()
		accountID := fake.BuildFakeID()
		memberA, memberB := fake.BuildFakeID(), fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.ready = readyWith(taskID)
		w.notifiable(t, taskID)

		task := mealplanningfakes.BuildFakeMealPlanTask()
		task.ID = taskID
		task.AssignedToUser = nil
		w.repo.GetMealPlanTaskFunc = func(context.Context, string) (*mealplanning.MealPlanTask, error) { return task, nil }
		w.repo.GetMealPlanTaskAccountIDFunc = func(context.Context, string) (string, error) { return accountID, nil }
		w.users.ListAccountMembersFunc = func(
			_ context.Context,
			_ database.SQLQueryExecutor,
			_ tenancy.Scope,
			id string,
			_ *filtering.QueryFilter,
		) (*filtering.QueryFilteredResult[platformidentity.MembershipWithUser], error) {
			assert.Equal(t, accountID, id)

			// One short page, which is what ends the walk.
			return &filtering.QueryFilteredResult[platformidentity.MembershipWithUser]{
				Data: []*platformidentity.MembershipWithUser{
					{User: &platformidentity.User{ID: memberA}},
					{User: &platformidentity.User{ID: memberB}},
				},
			}, nil
		}

		sent, err := w.worker.Work(t.Context())

		require.NoError(t, err)
		assert.Equal(t, int64(1), sent)
		assert.Len(t, w.devices.GetUserDeviceTokensCalls(), 2)
	})

	// A discovery that fails must not cost the queue its backlog: work enqueued by an earlier
	// pass is still owed, and the pass that cannot find new work can still finish old work.
	t.Run("drains the queue even when discovery fails", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)
		taskID := fake.BuildFakeID()

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) {
			return nil, errors.New("the database is having a moment")
		}
		w.queue.ready = readyWith(taskID)
		w.notifiable(t, taskID)

		sent, err := w.worker.Work(t.Context())

		require.Error(t, err)
		assert.Equal(t, int64(1), sent)
		assert.Equal(t, []string{taskID}, w.queue.completed)
	})

	t.Run("stops draining when a claim fails", func(t *testing.T) {
		t.Parallel()

		w := buildTestWorker(t)

		w.repo.GetMealPlanTaskIDsThatNeedNotificationFunc = func(context.Context) ([]string, error) { return nil, nil }
		w.queue.claimErr = errors.New("the queue is having a moment")

		sent, err := w.worker.Work(t.Context())

		require.Error(t, err)
		assert.Equal(t, int64(0), sent)
		assert.Empty(t, w.queue.completed)
	})
}
