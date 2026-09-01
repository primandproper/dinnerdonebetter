package adapters

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v13/capitalism"
	caprevenuecat "github.com/primandproper/platform-go/v13/capitalism/revenuecat"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/webhooks/inbound"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// exampleRevenueCatWebhookSecret is the signing secret from a RevenueCat webhook integration,
// which is what capitalism verifies deliveries against. The Authorization header the dashboard
// offers beside it proves only that the sender knew a secret, and says nothing about the body
// it arrived with; capitalism implements the signed mode alone, and so this is the only
// credential the adapter takes.
const exampleRevenueCatWebhookSecret = "rcsec_example_secret"

func buildRevenueCatProcessorForTest(t *testing.T) *RevenueCatPaymentProcessor {
	t.Helper()

	processor, err := NewRevenueCatPaymentProcessor(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&caprevenuecat.Config{WebhookSecret: exampleRevenueCatWebhookSecret},
	)
	require.NoError(t, err)

	return processor
}

// buildSignedRevenueCatRequest builds the request RevenueCat would send for a given event body,
// signed under the example secret in the t=…,v1= scheme Stripe published and RevenueCat adopted.
func buildSignedRevenueCatRequest(t *testing.T, body string) *http.Request {
	t.Helper()

	return buildRevenueCatRequest(t, body, revenueCatSignature(t, body, exampleRevenueCatWebhookSecret))
}

func buildRevenueCatRequest(t *testing.T, body, signature string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/payments/webhooks/revenuecat", strings.NewReader(body))
	if signature != "" {
		req.Header.Set(inbound.RevenueCatSignatureHeader, signature)
	}

	return req
}

func revenueCatSignature(t *testing.T, body, secret string) string {
	t.Helper()

	seconds := fmt.Sprintf("%d", time.Now().Unix())

	mac := hmac.New(sha256.New, []byte(secret))
	_, err := mac.Write([]byte(seconds + "." + body))
	require.NoError(t, err)

	return fmt.Sprintf("t=%s,v1=%s", seconds, hex.EncodeToString(mac.Sum(nil)))
}

// revenueCatDelivery renders the envelope every RevenueCat delivery arrives in, around a given
// event object.
func revenueCatDelivery(event string) string {
	return fmt.Sprintf(`{"api_version": "1.0", "event": %s}`, event)
}

func TestNewRevenueCatPaymentProcessor(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		assert.NotNil(t, buildRevenueCatProcessorForTest(t))
	})

	// Refused at construction rather than at the first delivery: RevenueCat is inbound-only,
	// so a manager without a signing secret could do nothing at all.
	T.Run("with no webhook secret", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRevenueCatPaymentProcessor(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&caprevenuecat.Config{},
		)
		require.Error(t, err)
		assert.Nil(t, processor)
	})

	T.Run("with nil config", func(t *testing.T) {
		t.Parallel()

		processor, err := NewRevenueCatPaymentProcessor(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			nil,
		)
		require.Error(t, err)
		assert.Nil(t, processor)
	})
}

func TestRevenueCatPaymentProcessor_HandleWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{
			"id": "evt_example",
			"type": "INITIAL_PURCHASE",
			"app_user_id": "account_example",
			"transaction_id": "txn_example",
			"original_transaction_id": "original_txn_example",
			"product_id": "product_example",
			"period_type": "NORMAL"
		}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, caprevenuecat.EventTypeInitialPurchase, actual.EventType)
		assert.Equal(t, "account_example", actual.AccountID)
		assert.Equal(t, "product_example", actual.ProductID)
		assert.Equal(t, payments.SubscriptionStatusActive, actual.Status)

		// The original transaction, not the current one: RevenueCat mints a fresh
		// transaction_id on every renewal and holds the original fixed, so the original is the
		// handle a subscription still answers to a year later.
		assert.Equal(t, "original_txn_example", actual.SubscriptionID)
	})

	// A purchase with no original transaction — a non-renewing one is its own first
	// transaction — falls back to the transaction it arrived with.
	T.Run("with no original transaction", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "NON_RENEWING_PURCHASE", "app_user_id": "account_example", "transaction_id": "txn_example"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "txn_example", actual.SubscriptionID)
	})

	T.Run("with expiration event", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "EXPIRATION", "app_user_id": "account_example", "transaction_id": "txn_example"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusCancelled, actual.Status)
	})

	// A cancellation is auto-renew being switched off, and the subscriber keeps the period
	// they paid for until EXPIRATION arrives. The manager marks the subscription cancelled on
	// the event type either way; what must not happen is the status reading as ended.
	T.Run("with cancellation event", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "CANCELLATION", "app_user_id": "account_example", "transaction_id": "txn_example", "cancel_reason": "UNSUBSCRIBE"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusActive, actual.Status)
	})

	// A refund does end it now, and the cancel reason is the only thing that says so.
	T.Run("with cancellation that ends access", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "CANCELLATION", "app_user_id": "account_example", "transaction_id": "txn_example", "cancel_reason": "CUSTOMER_SUPPORT"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusCancelled, actual.Status)
	})

	T.Run("with trial purchase", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "INITIAL_PURCHASE", "app_user_id": "account_example", "transaction_id": "txn_example", "period_type": "TRIAL"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusTrialing, actual.Status)
	})

	T.Run("with billing issue", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "BILLING_ISSUE", "app_user_id": "account_example", "transaction_id": "txn_example"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusPastDue, actual.Status)
	})

	// An event RevenueCat documents as carrying no subscription standing arrives with a nil
	// Subscription, and must not be rendered as a subscription whose status we could not read.
	T.Run("with event carrying no subscription", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "TEST", "app_user_id": "account_example"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, caprevenuecat.EventTypeTest, actual.EventType)
		assert.Empty(t, actual.AccountID)
		assert.Empty(t, actual.SubscriptionID)
		assert.Empty(t, actual.Status)
	})

	// An event type nobody has seen is a genuine delivery whose standing we cannot place. It
	// comes back with the type intact and no status, rather than guessed onto a neighbor.
	T.Run("with unrecognized event type", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "SOMETHING_NEW", "app_user_id": "account_example", "transaction_id": "txn_example"}`)

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "SOMETHING_NEW", actual.EventType)
		assert.Equal(t, "account_example", actual.AccountID)
		assert.Empty(t, actual.Status)
	})

	T.Run("with wrong signature", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "INITIAL_PURCHASE", "app_user_id": "account_example"}`)

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, revenueCatSignature(t, body, "rcsec_wrong_secret")))
		require.ErrorIs(t, err, inbound.ErrInvalidSignature)
		assert.Nil(t, actual)
	})

	// The header the old hand-rolled adapter verified. It is not the scheme capitalism
	// implements, and a delivery carrying only it is unverified.
	T.Run("with authorization header instead of a signature", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "INITIAL_PURCHASE", "app_user_id": "account_example"}`)

		req := buildRevenueCatRequest(t, body, "")
		req.Header.Set("Authorization", "Bearer "+exampleRevenueCatWebhookSecret)

		actual, err := processor.HandleWebhook(req)
		require.ErrorIs(t, err, inbound.ErrInvalidSignature)
		assert.Nil(t, actual)
	})

	T.Run("with missing signature", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := revenueCatDelivery(`{"type": "INITIAL_PURCHASE", "app_user_id": "account_example"}`)

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, ""))
		require.ErrorIs(t, err, inbound.ErrInvalidSignature)
		assert.Nil(t, actual)
	})

	T.Run("with malformed payload", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := `{"api_version":`

		actual, err := processor.HandleWebhook(buildSignedRevenueCatRequest(t, body))
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestRevenueCatSubscriptionStatus(T *testing.T) {
	T.Parallel()

	T.Run("maps every status capitalism defines", func(t *testing.T) {
		t.Parallel()

		expected := map[capitalism.SubscriptionStatus]string{
			capitalism.SubscriptionStatusActive:            payments.SubscriptionStatusActive,
			capitalism.SubscriptionStatusTrialing:          payments.SubscriptionStatusTrialing,
			capitalism.SubscriptionStatusPastDue:           payments.SubscriptionStatusPastDue,
			capitalism.SubscriptionStatusCanceled:          payments.SubscriptionStatusCancelled,
			capitalism.SubscriptionStatusIncomplete:        payments.SubscriptionStatusIncomplete,
			capitalism.SubscriptionStatusIncompleteExpired: payments.SubscriptionStatusIncomplete,
			capitalism.SubscriptionStatusUnpaid:            payments.SubscriptionStatusPastDue,
			capitalism.SubscriptionStatusPaused:            payments.SubscriptionStatusCancelled,
			capitalism.SubscriptionStatusUnknown:           "",
		}

		for status, want := range expected {
			assert.Equal(t, want, revenueCatSubscriptionStatus(status), "mapping %s", status.String())
		}
	})
}
