package http

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/adapters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v15/billing"
	billingmock "github.com/primandproper/platform-go/v15/billing/mock"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	capitalismcfg "github.com/primandproper/primitives-go/v2/capitalism/config"
	caprevenuecat "github.com/primandproper/primitives-go/v2/capitalism/revenuecat"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"
	"github.com/primandproper/primitives-go/v2/webhooks/inbound"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// webhookHarness is the endpoint over the real RevenueCat processor and the real payments
// manager, with only the two stores beneath them mocked. Everything a delivery passes through
// on its way from the signed bytes to a write is the code that runs in production, so what a
// test asserts about the writes is what the endpoint does with a request rather than what a
// stub was told to say.
type webhookHarness struct {
	handler http.Handler

	// standings records every account whose billing standing a delivery wrote, in order.
	standings *[]string

	secret string
}

func buildWebhookHarness(t *testing.T, store *billingmock.StoreMock) *webhookHarness {
	t.Helper()

	logger := loggingnoop.NewLogger()
	tracerProvider := tracingnoop.NewTracerProvider()
	secret := fake.BuildFakeID()

	processor, err := adapters.NewRevenueCatPaymentProcessor(logger, tracerProvider, &caprevenuecat.Config{WebhookSecret: secret})
	require.NoError(t, err)

	standings := &[]string{}

	identityStore := &identitymock.StoreMock{
		RecordAccountSubscriptionFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, accountID string, _ platformidentity.BillingStatus, _ string) error {
			*standings = append(*standings, accountID)

			return nil
		},
		RecordAccountSubscriptionEndedFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, accountID string, _ platformidentity.BillingStatus) error {
			*standings = append(*standings, accountID)

			return nil
		},
	}

	manager, err := paymentsmanager.NewPaymentsDataManager(t.Context(), tracerProvider, logger, testutils.MockDatabaseClient(), store, identityStore)
	require.NoError(t, err)

	registry := payments.NewMapProcessorRegistry(map[string]payments.PaymentProcessor{
		capitalismcfg.RevenueCatProvider: processor,
	})

	return &webhookHarness{
		handler:   NewWebhookHandler(logger, tracerProvider, manager, registry).For(capitalismcfg.RevenueCatProvider),
		secret:    secret,
		standings: standings,
	}
}

// deliver posts body to the endpoint, at a URL carrying the given query, signed under secret.
func (h *webhookHarness) deliver(t *testing.T, query url.Values, body, secret string) *httptest.ResponseRecorder {
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

	return res
}

// initialPurchase is the RevenueCat delivery for an account buying a product.
func initialPurchase(accountID, productExternalID string) string {
	return fmt.Sprintf(`{"api_version": "1.0", "event": {
		"id": %q,
		"type": "INITIAL_PURCHASE",
		"app_user_id": %q,
		"original_transaction_id": %q,
		"product_id": %q,
		"period_type": "NORMAL"
	}}`, fake.BuildFakeID(), accountID, fake.BuildFakeID(), productExternalID)
}

// purchasableProduct is a store that sells product and opens every subscription it is asked
// to, recording the account each one was opened for.
func purchasableProduct(product *billing.Product) (store *billingmock.StoreMock, opened *[]string) {
	opened = &[]string{}

	return &billingmock.StoreMock{
		GetProductByExternalIDFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, externalID string) (*billing.Product, error) {
			if externalID != product.ExternalProductID {
				return nil, billing.ErrProductNotFound
			}

			return product, nil
		},
		GetSubscriptionByExternalIDFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*billing.Subscription, error) {
			return nil, billing.ErrSubscriptionNotFound
		},
		CreateSubscriptionFunc: func(_ context.Context, _ database.Tx, _ tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
			*opened = append(*opened, subscription.BelongsToAccount)

			return subscription, nil
		},
	}, opened
}

func TestWebhookHandler_For(T *testing.T) {
	T.Parallel()

	T.Run("the account is the signed payload's, whatever the query string names", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		store, opened := purchasableProduct(product)
		harness := buildWebhookHarness(t, store)

		signedFor := fake.BuildFakeID()
		named := fake.BuildFakeID()

		res := harness.deliver(t, url.Values{"account_id": {named}}, initialPurchase(signedFor, product.ExternalProductID), harness.secret)

		require.Equal(t, http.StatusOK, res.Code)

		// The subscription and the standing both went to the account the provider signed for,
		// and the one the query named was never written at all.
		assert.Equal(t, []string{signedFor}, *opened)
		assert.Equal(t, []string{signedFor}, *harness.standings)
		assert.NotContains(t, *opened, named)
		assert.NotContains(t, *harness.standings, named)
	})

	T.Run("a transient store failure is answered 500, so the provider redelivers", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		store, opened := purchasableProduct(product)
		store.CreateSubscriptionFunc = func(context.Context, database.Tx, tenancy.Scope, *billing.Subscription) (*billing.Subscription, error) {
			return nil, platformerrors.New("connection reset by peer")
		}
		harness := buildWebhookHarness(t, store)

		res := harness.deliver(t, nil, initialPurchase(fake.BuildFakeID(), product.ExternalProductID), harness.secret)

		assert.Equal(t, http.StatusInternalServerError, res.Code)
		assert.Empty(t, *opened)
		assert.Empty(t, *harness.standings)
	})

	T.Run("a delivery for a product nobody sells is answered 400", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		store, opened := purchasableProduct(product)
		harness := buildWebhookHarness(t, store)

		// A product this service has never heard of is not one it will have heard of on the
		// next try either, so retrying would be a redelivery loop over nothing.
		res := harness.deliver(t, nil, initialPurchase(fake.BuildFakeID(), fake.BuildFakeID()), harness.secret)

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, *opened)
	})

	T.Run("a delivery with a bad signature is answered 400 and writes nothing", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		store, opened := purchasableProduct(product)
		harness := buildWebhookHarness(t, store)

		res := harness.deliver(t, nil, initialPurchase(fake.BuildFakeID(), product.ExternalProductID), fake.BuildFakeID())

		assert.Equal(t, http.StatusBadRequest, res.Code)
		assert.Empty(t, *opened)
		assert.Empty(t, *harness.standings)
	})

	T.Run("a mounted provider with no processor is answered 500", func(t *testing.T) {
		t.Parallel()

		handler := NewWebhookHandler(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			nil,
			payments.NewMapProcessorRegistry(nil),
		).For(capitalismcfg.StripeProvider)

		req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/payments/webhooks/stripe", strings.NewReader(`{}`))
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)

		// The route exists, so this is the service's wiring rather than the delivery.
		assert.Equal(t, http.StatusInternalServerError, res.Code)
	})
}
