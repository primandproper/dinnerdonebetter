package webhooks

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/clock"
	"github.com/primandproper/platform-go/v13/database"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/webhooks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The retention window and the reap cadence these tests run the worker with.
//
// They are the configuration's own floors rather than something smaller, because the floors are
// validated: a retention under a minute and a reap interval under a second are rejected outright,
// on the grounds that a deployment configured that way is deleting its delivery log faster than
// anybody could read it. A test cannot wait a minute either, which is what advanceableClock is
// for.
const (
	testRetention    = time.Minute
	testReapInterval = time.Second
)

// advanceableClock is the wall clock with an offset a test can move forward.
//
// The reap window is "delivered more than Retention ago", measured against the worker's own
// clock, and Retention cannot be configured below a minute — a deployment sweeping its delivery
// log faster than that is deleting evidence nobody could have read. Waiting out a minute per
// test is not an option either, and skewing the clock from the start buys nothing: the worker
// stamps delivered_at from the same clock it measures the window with, so a constant offset
// moves both ends together.
//
// Moving it after the delivery lands is what actually reaches the window, and it reaches it
// through the same arithmetic the worker does in production.
//
// Only Now and Since are offset. Sleep and NewTicker stay on the wall clock, because the poll
// and reap loops have to run at real speed for a test to observe them at all — a fake ticker
// would mean driving the loop rather than watching it.
type advanceableClock struct {
	wall   clock.WallClock
	offset atomic.Int64
}

// newAdvanceableClock returns a clock that starts in step with the wall clock.
func newAdvanceableClock() *advanceableClock {
	return &advanceableClock{wall: clock.NewClock()}
}

// advance moves the clock forward. It is safe to call while the worker is running, which is the
// only way it is ever called.
func (c *advanceableClock) advance(d time.Duration) { c.offset.Add(int64(d)) }

func (c *advanceableClock) Now() time.Time {
	return c.wall.Now().Add(time.Duration(c.offset.Load()))
}

func (c *advanceableClock) Since(t time.Time) time.Duration { return c.Now().Sub(t) }

func (c *advanceableClock) Sleep(ctx context.Context, d time.Duration) error {
	return c.wall.Sleep(ctx, d)
}

func (c *advanceableClock) NewTicker(d time.Duration) clock.Ticker { return c.wall.NewTicker(d) }

// dispatchState is what the delivery worker has written about one dispatch row.
//
// It is read with SQL rather than through the Store because the Store deliberately offers no
// way to read one back: Claim, MarkDelivered and RecordFailure are the whole of its dispatch
// surface, because nothing in production needs to ask what state a dispatch is in. A test
// asserting that a dispatch was marked dead rather than lost does, and the row is the fact.
type dispatchState struct {
	lastError string
	attempts  int
	dead      bool
	delivered bool
}

// onlyDispatch reads the single dispatch row in this test's database.
//
// Single because every test here stands up an isolated database and emits one event to one
// subscriber, so a second row would mean the fan-out did something nobody asked for — which is
// worth failing on rather than filtering out.
func onlyDispatch(t *testing.T, ctx context.Context, pgc database.Client) dispatchState {
	t.Helper()

	rows, err := pgc.Reader().QueryContext(ctx,
		"SELECT attempts, dead, delivered_at IS NOT NULL, COALESCE(last_error, '') FROM webhooks_dispatches")
	require.NoError(t, err)

	defer func() { assert.NoError(t, rows.Close()) }()

	var states []dispatchState
	for rows.Next() {
		var state dispatchState
		require.NoError(t, rows.Scan(&state.attempts, &state.dead, &state.delivered, &state.lastError))
		states = append(states, state)
	}
	require.NoError(t, rows.Err())

	require.Len(t, states, 1, "expected exactly one dispatch row")

	return states[0]
}

// countRows counts a table, for the reaper's assertions.
func countRows(t *testing.T, ctx context.Context, pgc database.Client, table string) int {
	t.Helper()

	var count int
	require.NoError(t, pgc.Reader().QueryRowContext(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count))

	return count
}

// subscribeAndEmit registers a subscriber for eventType and emits one event to it, returning
// nothing: every assertion below is about what the worker did afterwards, not about the write.
//
// The emit goes through the Emitter inside a transaction, which is the only way this
// application dispatches — a delivery row and the state change that caused it commit together.
func subscribeAndEmit(t *testing.T, ctx context.Context, repo *repository, emitter *events.Emitter, pgc database.Client, subscriberURL, eventType string) {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, repo.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, repo.writeDB)

	webhook := fakes.BuildFakeWebhook()
	webhook.BelongsToAccount = account.ID
	webhook.CreatedByUser = user.ID
	webhook.URL = subscriberURL
	webhook.TriggerConfigs[0].EventType = eventType

	_, err := repo.CreateWebhook(ctx, converters.ConvertWebhookToWebhookDatabaseCreationInput(webhook))
	require.NoError(t, err)

	require.NoError(t, pgc.WithTransaction(ctx, func(tx database.Tx) error {
		return emitter.Emit(ctx, tx, loggingnoop.NewLogger(), eventType, account.ID, nil)
	}))
}

// runWorker starts the delivery loop and stops it when the test ends.
func runWorker(t *testing.T, worker *webhooks.Worker) {
	t.Helper()

	go worker.Run()

	t.Cleanup(func() {
		// A background context: t.Context() is already cancelled by the time cleanups run,
		// and a Close given an expired deadline reports a failure that is only about the
		// deadline.
		closeCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		assert.NoError(t, worker.Close(closeCtx))
	})
}

// TestIntegration_WebhookDelivery_RetriesThenDies is the failure half of delivery.
//
// A subscriber that is down is the ordinary case, and the property that matters is that the
// dispatch is retried a bounded number of times and then marked dead — not retried forever, and
// not quietly forgotten. Dead is a row an operator can find and replay; lost is an event a
// subscriber will never see and nobody can name.
//
// Two attempts rather than production's ten, and a backoff of milliseconds rather than an hour,
// because the point is to reach the ceiling. Both are plain config fields, so nothing here
// reaches past the worker's constructor to get there.
func TestIntegration_WebhookDelivery_RetriesThenDies(t *testing.T) {
	ctx := t.Context()

	repo, emitter, worker, pgc := buildDeliveryHarness(t, func(tuning *deliveryTuning) {
		tuning.Config.Worker.Backoff.MaxAttempts = 2
		tuning.Config.Worker.Backoff.InitialDelay = 25 * time.Millisecond
		tuning.Config.Worker.Backoff.MaxDelay = 25 * time.Millisecond
		// Jitter would make the second attempt land anywhere in a window, which is right in
		// production and only makes this test's budget wider.
		tuning.Config.Worker.Backoff.UseJitter = false
	})

	// 500 rather than a 4xx: a subscriber that understood and refused goes straight to dead
	// without a retry, which would pass this test while proving nothing about the schedule.
	subscriber, received := newTestSubscriber(t, func() int { return http.StatusInternalServerError })

	subscribeAndEmit(t, ctx, repo, emitter, pgc, subscriber.URL, types.WebhookCreatedServiceEventType)

	runWorker(t, worker)

	// Dead is the terminal state, and reaching it is what bounds the retries.
	var final dispatchState
	require.Eventually(t, func() bool {
		final = onlyDispatch(t, ctx, pgc)

		return final.dead
	}, 30*time.Second, 25*time.Millisecond, "expected the dispatch to be marked dead once it ran out of attempts")

	assert.Equal(t, 2, final.attempts, "a dispatch must die at its attempt ceiling, not before or after it")
	assert.False(t, final.delivered, "a dead dispatch was never delivered")
	assert.NotEmpty(t, final.lastError, "the row has to say why it died, or dead is indistinguishable from lost")

	// Every attempt reached the subscriber, numbered, so a receiver can tell a retry from a
	// second event.
	deliveries := received()
	require.Len(t, deliveries, 2, "expected exactly one request per attempt")
	for i, delivery := range deliveries {
		assert.Equal(t, strconv.Itoa(i+1), delivery.headers.Get(webhooks.AttemptHeader))
	}

	// And nothing further. A dead dispatch is excluded from the claim, so a worker that kept
	// running has nothing to re-serve.
	assert.Never(t, func() bool {
		return len(received()) > 2
	}, time.Second, 50*time.Millisecond, "a dead dispatch must stop being retried")
}

// TestIntegration_WebhookDelivery_ReapsDeliveredDispatches covers the other end of the
// lifecycle.
//
// The delivery log is the record of what a subscriber was sent, and it grows with every event
// this deployment publishes. Nothing else trims it: retention has no scheduled job of its own,
// because the worker's own tick reaps — so a worker that delivers and does not reap is a table
// that only grows, and the absence would be invisible until the disk was.
//
// The reap interval is a plain config field and is turned right down. The retention window
// cannot be — its floor is a minute — so the clock is moved past it once the delivery has
// landed, which is the same window the worker computes in production.
func TestIntegration_WebhookDelivery_ReapsDeliveredDispatches(t *testing.T) {
	ctx := t.Context()

	workerClock := newAdvanceableClock()

	repo, emitter, worker, pgc := buildDeliveryHarness(t, func(tuning *deliveryTuning) {
		tuning.Config.Worker.ReapInterval = testReapInterval
		tuning.Config.Worker.Retention = testRetention
		tuning.Clock = workerClock
	})

	subscriber, received := newTestSubscriber(t, func() int { return http.StatusOK })

	subscribeAndEmit(t, ctx, repo, emitter, pgc, subscriber.URL, types.WebhookCreatedServiceEventType)

	runWorker(t, worker)

	require.Eventually(t, func() bool {
		return len(received()) > 0
	}, 30*time.Second, 25*time.Millisecond, "expected the subscriber to receive the delivery")

	// Delivered, and now older than the retention window. Advanced only after the delivery has
	// landed: the worker stamps delivered_at from this same clock, so moving it beforehand
	// would carry the row's age forward with it and never reach the window.
	require.Eventually(t, func() bool {
		return onlyDispatch(t, ctx, pgc).delivered
	}, 30*time.Second, 25*time.Millisecond, "expected the dispatch to be marked delivered")

	workerClock.advance(2 * testRetention)

	// The dispatch, the delivery it belonged to, and the attempts recorded against it all go.
	// Reaping the dispatch alone would leave the payload behind forever, which is the row that
	// actually holds the event's contents.
	require.Eventually(t, func() bool {
		return countRows(t, ctx, pgc, "webhooks_dispatches") == 0 &&
			countRows(t, ctx, pgc, "webhooks_deliveries") == 0 &&
			countRows(t, ctx, pgc, "webhooks_attempts") == 0
	}, 30*time.Second, 25*time.Millisecond, "expected the delivered dispatch, its delivery, and its attempts to be reaped")

	// The endpoint outlives its deliveries: reaping the log must not unsubscribe anybody.
	assert.Equal(t, 1, countRows(t, ctx, pgc, "webhooks_endpoints"))
	assert.Positive(t, countRows(t, ctx, pgc, "webhooks_subscriptions"))
}

// TestIntegration_WebhookDelivery_UndeliveredIsNotReaped pins the half of retention that would
// be silent if it broke.
//
// The reaper deletes delivered dispatches. One still owed to a subscriber that has been down for
// longer than the retention window must survive it — reaping by age alone would delete exactly
// the events a subscriber has not seen, which is the opposite of what retention is for.
func TestIntegration_WebhookDelivery_UndeliveredIsNotReaped(t *testing.T) {
	ctx := t.Context()

	workerClock := newAdvanceableClock()

	repo, emitter, worker, pgc := buildDeliveryHarness(t, func(tuning *deliveryTuning) {
		tuning.Config.Worker.ReapInterval = testReapInterval
		tuning.Config.Worker.Retention = testRetention
		tuning.Clock = workerClock
		// Long enough that the dispatch stays outstanding for the whole test rather than
		// dying and being reaped for some other reason. The clock below moves past the
		// retention window but not past this, so the row is still owed a retry throughout.
		tuning.Config.Worker.Backoff.MaxAttempts = 100
		tuning.Config.Worker.Backoff.InitialDelay = time.Hour
		tuning.Config.Worker.Backoff.MaxDelay = time.Hour
		tuning.Config.Worker.Backoff.UseJitter = false
	})

	subscriber, received := newTestSubscriber(t, func() int { return http.StatusInternalServerError })

	subscribeAndEmit(t, ctx, repo, emitter, pgc, subscriber.URL, types.WebhookCreatedServiceEventType)

	runWorker(t, worker)

	require.Eventually(t, func() bool {
		return len(received()) > 0
	}, 30*time.Second, 25*time.Millisecond, "expected the subscriber to be attempted at least once")

	// Older than the retention window, and still undelivered.
	workerClock.advance(2 * testRetention)

	// Several reap intervals, all of them past the retention window.
	assert.Never(t, func() bool {
		return countRows(t, ctx, pgc, "webhooks_dispatches") == 0
	}, 4*testReapInterval, 50*time.Millisecond, "an undelivered dispatch must survive retention")

	state := onlyDispatch(t, ctx, pgc)
	assert.False(t, state.delivered)
	assert.False(t, state.dead)
	assert.Positive(t, state.attempts)
}
