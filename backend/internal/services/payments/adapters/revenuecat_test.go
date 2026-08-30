package adapters

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleRevenueCatAuthHeader = "Bearer example_token"

func buildRevenueCatProcessorForTest(t *testing.T) *RevenueCatPaymentProcessor {
	t.Helper()

	return NewRevenueCatPaymentProcessor(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&RevenueCatConfig{WebhookAuthHeader: exampleRevenueCatAuthHeader},
	)
}

func buildRevenueCatRequest(t *testing.T, body, authHeader string) *http.Request {
	t.Helper()

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/payments/webhooks/revenuecat", strings.NewReader(body))
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	return req
}

func TestRevenueCatPaymentProcessor_HandleWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := `{"type": "INITIAL_PURCHASE", "app_user_id": "account_example", "transaction_id": "txn_example", "product_id": "product_example"}`

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, exampleRevenueCatAuthHeader))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "INITIAL_PURCHASE", actual.EventType)
		assert.Equal(t, "account_example", actual.AccountID)
		assert.Equal(t, "txn_example", actual.SubscriptionID)
		assert.Equal(t, "product_example", actual.ProductID)
		assert.Equal(t, payments.SubscriptionStatusActive, actual.Status)
	})

	T.Run("with expiration event", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := `{"type": "EXPIRATION", "app_user_id": "account_example", "transaction_id": "txn_example"}`

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, exampleRevenueCatAuthHeader))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, payments.SubscriptionStatusCancelled, actual.Status)
	})

	T.Run("with wrong auth header", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := `{"type": "INITIAL_PURCHASE", "app_user_id": "account_example"}`

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, "Bearer wrong_token"))
		require.ErrorIs(t, err, ErrInvalidWebhookSignature)
		assert.Nil(t, actual)
	})

	T.Run("with missing auth header", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		body := `{"type": "INITIAL_PURCHASE", "app_user_id": "account_example"}`

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, ""))
		require.ErrorIs(t, err, ErrInvalidWebhookSignature)
		assert.Nil(t, actual)
	})

	T.Run("with no configured auth header", func(t *testing.T) {
		t.Parallel()

		// An unconfigured header accepts everything, which is what makes the adapter usable
		// against RevenueCat's webhook tester in local development.
		processor := NewRevenueCatPaymentProcessor(
			loggingnoop.NewLogger(),
			tracingnoop.NewTracerProvider(),
			&RevenueCatConfig{},
		)

		body := `{"type": "RENEWAL", "app_user_id": "account_example", "transaction_id": "txn_example"}`

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, body, ""))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "RENEWAL", actual.EventType)
	})

	T.Run("with malformed payload", func(t *testing.T) {
		t.Parallel()

		processor := buildRevenueCatProcessorForTest(t)

		actual, err := processor.HandleWebhook(buildRevenueCatRequest(t, `{"type":`, exampleRevenueCatAuthHeader))
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
