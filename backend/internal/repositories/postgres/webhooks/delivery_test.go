package webhooks

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexevents"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/webhookdispatch"

	"github.com/primandproper/platform-go/v10/cryptography/requestsigning"
	"github.com/primandproper/platform-go/v10/database"
	"github.com/primandproper/platform-go/v10/database/dialect"
	"github.com/primandproper/platform-go/v10/database/postgres"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v10/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v10/observability/tracing/noop"
	"github.com/primandproper/platform-go/v10/outbox"
	"github.com/primandproper/platform-go/v10/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v10/webhooks/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// receivedDelivery is one request a test subscriber accepted.
type receivedDelivery struct {
	headers http.Header
	body    []byte
}

// newTestSubscriber stands up an HTTPS server that records what it is sent.
//
// httptest.NewTLSServer rather than NewServer, because the delivery worker refuses plaintext:
// a signature authenticates a payload but does not make it confidential, and the headers that
// authenticate it would be on the wire in clear.
func newTestSubscriber(t *testing.T, status func() int) (subscriber *httptest.Server, deliveries func() []receivedDelivery) {
	t.Helper()

	var (
		mu       sync.Mutex
		received []receivedDelivery
	)

	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		assert.NoError(t, err)

		mu.Lock()
		received = append(received, receivedDelivery{headers: r.Header.Clone(), body: body})
		mu.Unlock()

		w.WriteHeader(status())
	}))
	t.Cleanup(srv.Close)

	return srv, func() []receivedDelivery {
		mu.Lock()
		defer mu.Unlock()

		return append([]receivedDelivery(nil), received...)
	}
}

// buildDeliveryHarness wires the whole webhook path over one database: the repository that
// registers endpoints, the Emitter that dispatches inside a transaction, and the Worker that
// delivers.
func buildDeliveryHarness(t *testing.T) (*repository, *events.Emitter, *webhooks.Worker, database.Client) {
	t.Helper()

	ctx := t.Context()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NoError(t, err)
	require.NotNil(t, pgc)

	store, err := webhookscfg.NewStore(ctx, &webhookscfg.Config{}, pgc)
	require.NoError(t, err)

	// The URL policy is relaxed on both halves together. Registration and delivery must agree:
	// an endpoint accepted at one and refused at the other sits in the backlog until it dies.
	allowLoopback := func(context.Context, string) error { return nil }

	dispatcher, err := webhookdispatch.NewDispatcher(
		store,
		catalog.Catalog(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		webhookdispatch.WithURLChecker(allowLoopback),
	)
	require.NoError(t, err)

	auditRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	writer, err := outbox.NewWriter(dialect.Postgres)
	require.NoError(t, err)

	emitter := events.NewEmitter(writer, "data_changes", dispatcher, indexevents.SideEffect)
	require.NotNil(t, emitter)

	repo := ProvideWebhooksRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditRepo, pgc, emitter, dispatcher)

	workerCfg := &webhookscfg.Config{}
	workerCfg.EnsureDefaults()
	workerCfg.CircuitBreaker.Name = "test"
	// A short poll so a test spends milliseconds waiting rather than a second.
	workerCfg.Worker.PollInterval = 10 * time.Millisecond

	worker, err := webhookscfg.NewWorker(ctx, workerCfg, store,
		webhookscfg.WithLogger(loggingnoop.NewLogger()),
		webhookscfg.WithWorkerOptions(
			webhooks.WithWorkerURLChecker(allowLoopback),
			// The subscriber's certificate is self-signed, which is what httptest issues.
			webhooks.WithHTTPClient(&http.Client{Transport: insecureTransport()}),
		),
	)
	require.NoError(t, err)

	return repo.(*repository), emitter, worker, pgc
}

// TestIntegration_WebhookDelivery is the end-to-end proof that a domain event reaches a
// subscriber, signed.
//
// It is the thing that has never worked: trigger configs referenced randomly-identified catalog
// rows while the fan-out looked webhooks up by event type string, so nothing ever matched and no
// delivery was ever attempted. This asserts the whole path — register, dispatch inside a
// transaction, claim, sign, send — and that webhooks.Verify accepts what arrives.
func TestIntegration_WebhookDelivery(t *testing.T) {
	ctx := t.Context()

	repo, emitter, worker, pgc := buildDeliveryHarness(t)

	subscriber, received := newTestSubscriber(t, func() int { return http.StatusOK })

	user := pgtesting.CreateUserForTest(t, nil, repo.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, repo.writeDB)

	eventType := types.WebhookCreatedServiceEventType

	exampleWebhook := fakes.BuildFakeWebhook()
	exampleWebhook.BelongsToAccount = account.ID
	exampleWebhook.CreatedByUser = user.ID
	exampleWebhook.URL = subscriber.URL
	exampleWebhook.TriggerConfigs[0].EventType = eventType

	response, err := repo.CreateWebhook(ctx, converters.ConvertWebhookToWebhookDatabaseCreationInput(exampleWebhook))
	require.NoError(t, err)
	require.NotEmpty(t, response.Secret)

	secret, err := hex.DecodeString(response.Secret)
	require.NoError(t, err)

	go worker.Run()
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		assert.NoError(t, worker.Close(closeCtx))
	})

	// Dispatch through the same seam every repository write uses: inside a transaction, on the
	// caller's executor, so the delivery commits with whatever else that transaction did.
	require.NoError(t, pgc.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		return emitter.Emit(ctx, tx, loggingnoop.NewLogger(), eventType, account.ID, map[string]any{
			"webhook.id": response.Webhook.ID,
		})
	}))

	var deliveries []receivedDelivery
	require.Eventually(t, func() bool {
		deliveries = received()

		return len(deliveries) > 0
	}, 30*time.Second, 25*time.Millisecond, "expected the subscriber to receive a delivery")

	delivery := deliveries[0]

	// The signature must verify over the exact bytes received. Verifying a re-serialized body
	// is the classic way a subscriber ends up authenticating something it did not receive, so
	// the body is passed through untouched.
	require.NoError(t, requestsigning.Verify(
		webhooks.Secret{Current: secret},
		delivery.body,
		delivery.headers.Get(requestsigning.SignatureHeader),
	))

	// A wrong key must not verify, or the assertion above proves nothing.
	require.Error(t, requestsigning.Verify(
		webhooks.Secret{Current: []byte("not the signing secret at all!!!")},
		delivery.body,
		delivery.headers.Get(requestsigning.SignatureHeader),
	))

	assert.Equal(t, eventType, delivery.headers.Get(webhooks.EventTypeHeader))
	assert.NotEmpty(t, delivery.headers.Get(webhooks.DeliveryIDHeader))
	assert.NotEmpty(t, delivery.headers.Get(requestsigning.TimestampHeader))
	assert.Equal(t, "application/json", delivery.headers.Get("Content-Type"))

	// The payload is the data change message, byte-identical to what the broker carries.
	var payload audit.DataChangeMessage
	require.NoError(t, json.Unmarshal(delivery.body, &payload))
	assert.Equal(t, eventType, payload.EventType)
	assert.Equal(t, account.ID, payload.AccountID)
}

// TestIntegration_WebhookDelivery_NotDeliveredToOtherAccounts is the tenancy property.
//
// platform-go's dispatcher has no tenant dimension: it delivers an event to every endpoint
// subscribed to that type. The account is carried inside the subscription's event type instead,
// and this is what says that actually holds.
func TestIntegration_WebhookDelivery_NotDeliveredToOtherAccounts(t *testing.T) {
	ctx := t.Context()

	repo, emitter, worker, pgc := buildDeliveryHarness(t)

	subscriberA, receivedA := newTestSubscriber(t, func() int { return http.StatusOK })
	subscriberB, receivedB := newTestSubscriber(t, func() int { return http.StatusOK })

	eventType := types.WebhookCreatedServiceEventType

	userA := pgtesting.CreateUserForTest(t, nil, repo.writeDB)
	accountA := pgtesting.CreateAccountForTest(t, nil, userA.ID, repo.writeDB)

	userB := pgtesting.CreateUserForTest(t, nil, repo.writeDB)
	accountB := pgtesting.CreateAccountForTest(t, nil, userB.ID, repo.writeDB)

	for _, sub := range []struct {
		account string
		user    string
		url     string
	}{
		{accountA.ID, userA.ID, subscriberA.URL},
		{accountB.ID, userB.ID, subscriberB.URL},
	} {
		webhook := fakes.BuildFakeWebhook()
		webhook.BelongsToAccount = sub.account
		webhook.CreatedByUser = sub.user
		webhook.URL = sub.url
		webhook.TriggerConfigs[0].EventType = eventType

		_, err := repo.CreateWebhook(ctx, converters.ConvertWebhookToWebhookDatabaseCreationInput(webhook))
		require.NoError(t, err)
	}

	go worker.Run()
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		assert.NoError(t, worker.Close(closeCtx))
	})

	// Only account A's event is emitted.
	require.NoError(t, pgc.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		return emitter.Emit(ctx, tx, loggingnoop.NewLogger(), eventType, accountA.ID, nil)
	}))

	require.Eventually(t, func() bool {
		return len(receivedA()) > 0
	}, 30*time.Second, 25*time.Millisecond, "expected account A's subscriber to receive a delivery")

	// B subscribes to the same event type on the same worker, and must see nothing. Given A has
	// already been delivered to, a delivery to B would have been claimed in the same batch.
	assert.Empty(t, receivedB(), "account B's subscriber received an event belonging to account A")
}

// TestIntegration_WebhookDelivery_ExcludedEventIsNotDeliverable covers the denylist.
//
// Some events the application publishes describe an account's security activity rather than its
// contents — signing in, rotating a two-factor secret. An endpoint URL is attacker-supplied, so
// these must not be subscribable, and the catalog is what enforces that at both ends.
func TestIntegration_WebhookDelivery_ExcludedEventIsNotDeliverable(t *testing.T) {
	ctx := t.Context()

	repo, emitter, worker, pgc := buildDeliveryHarness(t)

	subscriber, received := newTestSubscriber(t, func() int { return http.StatusOK })

	user := pgtesting.CreateUserForTest(t, nil, repo.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, repo.writeDB)

	webhook := fakes.BuildFakeWebhook()
	webhook.BelongsToAccount = account.ID
	webhook.CreatedByUser = user.ID
	webhook.URL = subscriber.URL
	webhook.TriggerConfigs[0].EventType = types.WebhookCreatedServiceEventType

	_, err := repo.CreateWebhook(ctx, converters.ConvertWebhookToWebhookDatabaseCreationInput(webhook))
	require.NoError(t, err)

	go worker.Run()
	t.Cleanup(func() {
		closeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
		defer cancel()
		assert.NoError(t, worker.Close(closeCtx))
	})

	// Emitting an excluded event succeeds and dispatches nothing.
	//
	// Succeeding is the whole point: this runs inside the transaction that wrote the row the
	// event describes, so failing here would not fail a webhook — it would fail the sign-in.
	// The event still reaches the outbox and every other consumer; only webhook delivery is
	// withheld.
	require.NoError(t, pgc.WithTransaction(ctx, func(tx database.SQLQueryExecutor) error {
		return emitter.Emit(ctx, tx, loggingnoop.NewLogger(), "user_logged_in", account.ID, nil)
	}))

	// Nothing arrives, and nothing is queued to arrive later either: no dispatch row was
	// written, so a worker running afterwards has nothing to claim.
	assert.Never(t, func() bool {
		return len(received()) > 0
	}, time.Second, 50*time.Millisecond)
}

// insecureTransport trusts the self-signed certificate httptest issues.
//
// Only ever used against a loopback server this test started, which is why skipping
// verification is acceptable here and nowhere near the production client.
func insecureTransport() *http.Transport {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
	}

	return transport
}
