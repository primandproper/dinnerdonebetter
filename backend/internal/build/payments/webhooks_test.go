package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	paymentsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/payments"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"
	paymentscfg "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v15/billing"
	billingmock "github.com/primandproper/platform-go/v15/billing/mock"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/capitalism"
	capitalismcfg "github.com/primandproper/primitives-go/v2/capitalism/config"
	caprevenuecat "github.com/primandproper/primitives-go/v2/capitalism/revenuecat"
	capstripe "github.com/primandproper/primitives-go/v2/capitalism/stripe"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/webhooks/inbound"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// provideForTest builds the endpoints from cfg over stores nothing will reach: every case
// that uses it is decided before a delivery gets as far as the database.
func provideForTest(t *testing.T, cfg *paymentscfg.Config) (*WebhookHandlers, error) {
	t.Helper()

	return ProvideWebhookHandlers(
		t.Context(),
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		cfg,
		testutils.MockDatabaseClient(),
		&billingmock.StoreMock{},
		&identitymock.StoreMock{},
	)
}

func TestProvideWebhookHandlers(T *testing.T) {
	T.Parallel()

	T.Run("each endpoint takes the provider it is named for", func(t *testing.T) {
		t.Parallel()

		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism: capitalismcfg.Config{
				Provider:   capitalismcfg.StripeProvider,
				Stripe:     &capstripe.Config{WebhookSecret: fake.BuildFakeID()},
				RevenueCat: &caprevenuecat.Config{WebhookSecret: fake.BuildFakeID()},
			},
			MobileProvider: capitalismcfg.RevenueCatProvider,
		})
		require.NoError(t, err)

		assert.NotNil(t, handlers.Stripe)
		assert.NotNil(t, handlers.RevenueCat)
	})

	T.Run("a noop endpoint acknowledges every delivery and writes nothing", func(t *testing.T) {
		t.Parallel()

		// The stores are mocks with no functions set, so a delivery that reached either
		// would panic rather than pass.
		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism:     capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
			MobileProvider: capitalismcfg.NoopProvider,
		})
		require.NoError(t, err)

		for _, handler := range []http.Handler{handlers.Stripe, handlers.RevenueCat} {
			res := httptest.NewRecorder()
			handler.ServeHTTP(res, httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{}`)))

			assert.Equal(t, http.StatusOK, res.Code)
		}
	})

	T.Run("the web endpoint refuses the mobile provider", func(t *testing.T) {
		t.Parallel()

		// It would verify Stripe's deliveries against RevenueCat's secret and reject every one.
		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism: capitalismcfg.Config{
				Provider:   capitalismcfg.RevenueCatProvider,
				RevenueCat: &caprevenuecat.Config{WebhookSecret: fake.BuildFakeID()},
			},
			MobileProvider: capitalismcfg.NoopProvider,
		})

		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, handlers)
	})

	T.Run("the mobile endpoint refuses the web provider", func(t *testing.T) {
		t.Parallel()

		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism:     capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
			MobileProvider: capitalismcfg.StripeProvider,
		})

		require.ErrorIs(t, err, platformerrors.ErrUnknownProvider)
		assert.Nil(t, handlers)
	})

	T.Run("an unset provider is an error rather than billing nobody", func(t *testing.T) {
		t.Parallel()

		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism: capitalismcfg.Config{Provider: capitalismcfg.NoopProvider},
		})

		require.Error(t, err)
		assert.Nil(t, handlers)
	})

	T.Run("RevenueCat without a signing secret is refused at startup", func(t *testing.T) {
		t.Parallel()

		handlers, err := provideForTest(t, &paymentscfg.Config{
			Capitalism: capitalismcfg.Config{
				Provider:   capitalismcfg.NoopProvider,
				RevenueCat: &caprevenuecat.Config{},
			},
			MobileProvider: capitalismcfg.RevenueCatProvider,
		})

		require.Error(t, err)
		assert.Nil(t, handlers)
	})
}

// webhookHarness is the RevenueCat endpoint exactly as ProvideWebhookHandlers builds it for
// production, over a real database: the recording billing store, platform's identity store,
// capitalism's verifier, billing/sync and billing/http. Nothing between the signed bytes and
// the rows is a stand-in, so what a test asserts about the rows is what a delivery does.
type webhookHarness struct {
	handler  http.Handler
	store    billing.Store
	accounts platformidentity.Store
	db       database.Client
	secret   string
}

func buildWebhookHarness(t *testing.T) *webhookHarness {
	t.Helper()

	ctx := t.Context()
	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	metricsProvider := metricsnoop.NewMetricsProvider()

	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	db, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(logger), postgres.WithTracerProvider(tracerProvider))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, db.Close()) })

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(logger, tracerProvider, metricsProvider, db)
	require.NoError(t, err)

	auditRecorder, ok := auditlogentries.RecorderFrom(auditLogEntryRepo)
	require.True(t, ok)

	store, err := paymentsrepo.ProvidePaymentsRepository(ctx, logger, tracerProvider, metricsProvider,
		pgtesting.NewRecorderForTest(t, ctx, db, auditRecorder), db)
	require.NoError(t, err)

	accounts, err := platformidentity.NewSQLStore(db, platformidentity.WithTablePrefix(branding.TablePrefix))
	require.NoError(t, err)

	secret := fake.BuildFakeID()

	handlers, err := ProvideWebhookHandlers(ctx, logger, tracerProvider, metricsProvider, &paymentscfg.Config{
		Capitalism: capitalismcfg.Config{
			Provider:   capitalismcfg.NoopProvider,
			RevenueCat: &caprevenuecat.Config{WebhookSecret: secret},
		},
		MobileProvider: capitalismcfg.RevenueCatProvider,
	}, db, store, accounts)
	require.NoError(t, err)

	return &webhookHarness{
		handler:  handlers.RevenueCat,
		store:    store,
		accounts: accounts,
		db:       db,
		secret:   secret,
	}
}

// account creates an account, owned by a user of its own.
func (h *webhookHarness) account(t *testing.T) *platformidentity.Account {
	t.Helper()

	user := pgtesting.CreateUserForTest(t, nil, h.db.Writer())

	return pgtesting.CreateAccountForTest(t, nil, user.ID, h.db.Writer())
}

// product adds one recurring product to the catalog.
func (h *webhookHarness) product(t *testing.T) *billing.Product {
	t.Helper()

	var product *billing.Product

	require.NoError(t, h.db.WithTransaction(t.Context(), func(tx database.Tx) error {
		var err error
		product, err = h.store.CreateProduct(t.Context(), tx, tenancy.Global(), fakes.BuildFakeProduct())

		return err
	}))

	return product
}

// reread is the account as the database now holds it.
func (h *webhookHarness) reread(t *testing.T, accountID string) *platformidentity.Account {
	t.Helper()

	account, err := h.accounts.GetAccount(t.Context(), h.db.Reader(), tenancy.Global(), accountID)
	require.NoError(t, err)

	return account
}

// subscription is the agreement stored under the provider's id, or nil if there is none.
func (h *webhookHarness) subscription(t *testing.T, externalID string) *billing.Subscription {
	t.Helper()

	subscription, err := h.store.GetSubscriptionByExternalID(t.Context(), h.db.Reader(), tenancy.Global(), externalID)
	if errors.Is(err, billing.ErrSubscriptionNotFound) {
		return nil
	}
	require.NoError(t, err)

	return subscription
}

// deliver posts body to the endpoint, at a URL carrying the given query, signed under secret.
func (h *webhookHarness) deliver(t *testing.T, query url.Values, body, secret string) int {
	t.Helper()

	seconds := fmt.Sprintf("%d", time.Now().Unix())

	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(seconds + "." + body))
	require.NoError(t, err)

	target := "/api/payments/webhooks/revenuecat"
	if len(query) > 0 {
		target += "?" + query.Encode()
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, target, strings.NewReader(body))
	req.Header.Set(inbound.RevenueCatSignatureHeader, fmt.Sprintf("t=%s,v1=%s", seconds, hex.EncodeToString(mac.Sum(nil))))

	res := httptest.NewRecorder()
	h.handler.ServeHTTP(res, req)

	return res.Code
}

// revenueCatEvent is one RevenueCat delivery about a subscription.
type revenueCatEvent struct {
	start, end    time.Time
	eventType     string
	accountID     string
	transactionID string
	productID     string
}

// purchase is an account buying a product, paid for a month from now.
func purchase(accountID, productID string) *revenueCatEvent {
	start := time.Now().UTC().Truncate(time.Millisecond)

	return &revenueCatEvent{
		eventType:     caprevenuecat.EventTypeInitialPurchase,
		accountID:     accountID,
		transactionID: fake.BuildFakeID(),
		productID:     productID,
		start:         start,
		end:           start.AddDate(0, 1, 0),
	}
}

func (e *revenueCatEvent) body() string {
	return fmt.Sprintf(`{"api_version": "1.0", "event": {
		"id": %q,
		"type": %q,
		"app_user_id": %q,
		"original_transaction_id": %q,
		"product_id": %q,
		"period_type": "NORMAL",
		"purchased_at_ms": %d,
		"expiration_at_ms": %d
	}}`, fake.BuildFakeID(), e.eventType, e.accountID, e.transactionID, e.productID, e.start.UnixMilli(), e.end.UnixMilli())
}

func TestWebhookHandlers_Integration_RevenueCat(T *testing.T) {
	T.Parallel()

	T.Run("a purchase opens the subscription and makes the account paid", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		event := purchase(account.ID, product.ExternalProductID)
		require.Equal(t, http.StatusOK, h.deliver(t, nil, event.body(), h.secret))

		subscription := h.subscription(t, event.transactionID)
		require.NotNil(t, subscription)
		assert.Equal(t, account.ID, subscription.BelongsToAccount)
		assert.Equal(t, product.ID, subscription.ProductID)
		assert.Equal(t, capitalism.SubscriptionStatusActive, subscription.Status)

		stored := h.reread(t, account.ID)
		assert.Equal(t, platformidentity.BillingPaid, stored.BillingStatus)
		require.NotNil(t, stored.SubscriptionPlanID)
		assert.Equal(t, product.ID, *stored.SubscriptionPlanID)
	})

	T.Run("the account is the signed payload's, whatever the query string names", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		signedFor := h.account(t)
		named := h.account(t)
		product := h.product(t)

		event := purchase(signedFor.ID, product.ExternalProductID)
		require.Equal(t, http.StatusOK, h.deliver(t, url.Values{"account_id": {named.ID}}, event.body(), h.secret))

		// The subscription and the standing both went to the account the provider signed for,
		// and the one the query named was never written at all.
		subscription := h.subscription(t, event.transactionID)
		require.NotNil(t, subscription)
		assert.Equal(t, signedFor.ID, subscription.BelongsToAccount)

		assert.Equal(t, platformidentity.BillingPaid, h.reread(t, signedFor.ID).BillingStatus)
		assert.Equal(t, named.BillingStatus, h.reread(t, named.ID).BillingStatus)
		assert.Nil(t, h.reread(t, named.ID).SubscriptionPlanID)
	})

	T.Run("a delivery a redelivery could fix is answered 500, and the redelivery lands", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		externalProductID := fake.BuildFakeID()

		// A product the operator has not added to the catalog yet. This used to be answered
		// 400, which told the provider to drop a delivery that would succeed tomorrow.
		event := purchase(account.ID, externalProductID)
		body := event.body()

		assert.Equal(t, http.StatusInternalServerError, h.deliver(t, nil, body, h.secret))
		assert.Nil(t, h.subscription(t, event.transactionID))
		assert.Equal(t, account.BillingStatus, h.reread(t, account.ID).BillingStatus)

		product := fakes.BuildFakeProduct()
		product.ExternalProductID = externalProductID
		require.NoError(t, h.db.WithTransaction(t.Context(), func(tx database.Tx) error {
			_, err := h.store.CreateProduct(t.Context(), tx, tenancy.Global(), product)

			return err
		}))

		// The provider sends the same bytes again.
		require.Equal(t, http.StatusOK, h.deliver(t, nil, body, h.secret))
		require.NotNil(t, h.subscription(t, event.transactionID))
		assert.Equal(t, platformidentity.BillingPaid, h.reread(t, account.ID).BillingStatus)
	})

	T.Run("a delivery with a bad signature is answered 400 and writes nothing", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		event := purchase(account.ID, product.ExternalProductID)

		assert.Equal(t, http.StatusBadRequest, h.deliver(t, nil, event.body(), fake.BuildFakeID()))
		assert.Nil(t, h.subscription(t, event.transactionID))
		assert.Equal(t, account.BillingStatus, h.reread(t, account.ID).BillingStatus)
	})

	T.Run("the paid period is the one the provider reported", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		// An annual plan. The local handler wrote now + one month whatever the provider said.
		event := purchase(account.ID, product.ExternalProductID)
		event.end = event.start.AddDate(1, 0, 0)
		require.Equal(t, http.StatusOK, h.deliver(t, nil, event.body(), h.secret))

		subscription := h.subscription(t, event.transactionID)
		require.NotNil(t, subscription)
		assert.True(t, event.start.Equal(subscription.CurrentPeriodStart), "start: %s != %s", event.start, subscription.CurrentPeriodStart)
		assert.True(t, event.end.Equal(subscription.CurrentPeriodEnd), "end: %s != %s", event.end, subscription.CurrentPeriodEnd)
	})

	T.Run("a status nobody can place changes nothing", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		// An event type RevenueCat adds next year. The local handler read a delivery with no
		// placeable status as active, which is the reading that keeps a lapsed account paid.
		event := purchase(account.ID, product.ExternalProductID)
		event.eventType = strings.ToUpper(fake.BuildFakeID())
		require.Equal(t, http.StatusOK, h.deliver(t, nil, event.body(), h.secret))

		assert.Nil(t, h.subscription(t, event.transactionID))
		assert.Equal(t, account.BillingStatus, h.reread(t, account.ID).BillingStatus)
	})

	T.Run("an expiration ends the subscription and the account's paid standing", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		event := purchase(account.ID, product.ExternalProductID)
		require.Equal(t, http.StatusOK, h.deliver(t, nil, event.body(), h.secret))
		require.Equal(t, platformidentity.BillingPaid, h.reread(t, account.ID).BillingStatus)

		event.eventType = caprevenuecat.EventTypeExpiration
		require.Equal(t, http.StatusOK, h.deliver(t, nil, event.body(), h.secret))

		subscription := h.subscription(t, event.transactionID)
		require.NotNil(t, subscription)
		assert.Equal(t, capitalism.SubscriptionStatusCanceled, subscription.Status)
		assert.Equal(t, platformidentity.BillingUnpaid, h.reread(t, account.ID).BillingStatus)
	})

	T.Run("a redelivered event is acknowledged and opens nothing new", func(t *testing.T) {
		t.Parallel()

		h := buildWebhookHarness(t)
		account := h.account(t)
		product := h.product(t)

		body := purchase(account.ID, product.ExternalProductID).body()
		require.Equal(t, http.StatusOK, h.deliver(t, nil, body, h.secret))
		require.Equal(t, http.StatusOK, h.deliver(t, nil, body, h.secret))

		subscriptions, err := h.store.ListSubscriptionsForAccount(t.Context(), h.db.Reader(), tenancy.Global(), account.ID, nil)
		require.NoError(t, err)
		assert.Len(t, subscriptions.Data, 1)
	})
}
