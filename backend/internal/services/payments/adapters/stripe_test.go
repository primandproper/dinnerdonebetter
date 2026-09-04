package adapters

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/primandproper/platform-go/v13/capitalism"
	capstripe "github.com/primandproper/platform-go/v13/capitalism/stripe"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stripe/stripe-go/v81/webhook"
)

const (
	exampleWebhookSecret = "whsec_example_secret"

	// exampleAPIVersion is the Stripe API version stripe-go v81 expects. Its own copy of this
	// string is unexported, and webhook.ConstructEvent — which is what capitalism verifies
	// with — refuses an event stamped with any other one. A Stripe webhook endpoint therefore
	// has to be configured at this version, and bumping stripe-go means bumping it there too;
	// this constant is where that coupling shows up when the SDK moves.
	exampleAPIVersion = "2025-02-24.acacia"
)

// buildStripeProcessorForTest returns a processor whose webhook secret is known, so that
// tests can sign payloads the way Stripe would.
func buildStripeProcessorForTest(t *testing.T) *StripePaymentProcessor {
	t.Helper()

	processor, err := NewStripePaymentProcessor(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		&capstripe.Config{WebhookSecret: exampleWebhookSecret},
	)
	require.NoError(t, err)

	return processor
}

// buildSignedStripeRequest builds the request Stripe would send for a given event body.
func buildSignedStripeRequest(t *testing.T, body string) *http.Request {
	t.Helper()

	now := time.Now()
	signature := webhook.ComputeSignature(now, []byte(body), exampleWebhookSecret)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/api/payments/webhooks/stripe", strings.NewReader(body))
	req.Header.Set("Stripe-Signature", fmt.Sprintf("t=%d,v1=%x", now.Unix(), signature))

	return req
}

func TestStripePaymentProcessor_HandleWebhook(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		processor := buildStripeProcessorForTest(t)

		body := fmt.Sprintf(`{
			"id": "evt_example",
			"api_version": %q,
			"type": "customer.subscription.updated",
			"data": {
				"object": {
					"id": "sub_example",
					"status": "active",
					"customer": "cus_example",
					"items": {"data": [{"price": {"id": "price_example"}}]}
				}
			}
		}`, exampleAPIVersion)

		actual, err := processor.HandleWebhook(buildSignedStripeRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "customer.subscription.updated", actual.EventType)
		assert.Equal(t, "sub_example", actual.SubscriptionID)
		assert.Equal(t, capitalism.SubscriptionStatusActive, actual.Status)
		assert.Equal(t, "cus_example", actual.AccountID)
		assert.Equal(t, "price_example", actual.ProductID)
	})

	T.Run("with unhandled event type", func(t *testing.T) {
		t.Parallel()

		processor := buildStripeProcessorForTest(t)

		body := fmt.Sprintf(`{"id": "evt_example", "api_version": %q, "type": "invoice.paid", "data": {"object": {"id": "in_example"}}}`, exampleAPIVersion)

		actual, err := processor.HandleWebhook(buildSignedStripeRequest(t, body))
		require.NoError(t, err)
		require.NotNil(t, actual)

		// The event verified, so it is reported; nothing about it names a subscription, so the
		// manager's switch will no-op on it.
		assert.Equal(t, "invoice.paid", actual.EventType)
		assert.Empty(t, actual.SubscriptionID)
	})

	T.Run("with invalid signature", func(t *testing.T) {
		t.Parallel()

		processor := buildStripeProcessorForTest(t)

		body := fmt.Sprintf(`{"id": "evt_example", "api_version": %q, "type": "customer.subscription.updated", "data": {"object": {}}}`, exampleAPIVersion)

		req := buildSignedStripeRequest(t, body)
		req.Header.Set("Stripe-Signature", "t=1,v1=deadbeef")

		actual, err := processor.HandleWebhook(req)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with missing signature", func(t *testing.T) {
		t.Parallel()

		processor := buildStripeProcessorForTest(t)

		body := fmt.Sprintf(`{"id": "evt_example", "api_version": %q, "type": "customer.subscription.updated", "data": {"object": {}}}`, exampleAPIVersion)

		req := buildSignedStripeRequest(t, body)
		req.Header.Del("Stripe-Signature")

		actual, err := processor.HandleWebhook(req)
		require.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with concurrent requests", func(t *testing.T) {
		t.Parallel()

		// One processor serves every request, and the capitalism event handler it registers is
		// shared across all of them. This asserts each caller gets its own event back.
		processor := buildStripeProcessorForTest(t)

		for i := range 16 {
			t.Run(fmt.Sprintf("request_%d", i), func(t *testing.T) {
				t.Parallel()

				subscriptionID := fmt.Sprintf("sub_example_%d", i)
				body := fmt.Sprintf(
					`{"id": "evt_example_%d", "api_version": %q, "type": "customer.subscription.updated", "data": {"object": {"id": %q, "status": "active"}}}`,
					i, exampleAPIVersion, subscriptionID,
				)

				actual, err := processor.HandleWebhook(buildSignedStripeRequest(t, body))
				require.NoError(t, err)
				require.NotNil(t, actual)

				assert.Equal(t, subscriptionID, actual.SubscriptionID)
			})
		}
	})
}

func TestParseStripeEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		actual, err := parseStripeEvent(&capitalism.Event{
			ID:      "evt_example",
			Type:    "customer.subscription.deleted",
			Payload: []byte(`{"id": "sub_example", "status": "canceled", "customer": "cus_example"}`),
		})
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "customer.subscription.deleted", actual.EventType)
		assert.Equal(t, "sub_example", actual.SubscriptionID)
		assert.Equal(t, capitalism.SubscriptionStatusCanceled, actual.Status)
		assert.Equal(t, "cus_example", actual.AccountID)
		assert.Empty(t, actual.ProductID)
	})

	T.Run("with expanded customer object", func(t *testing.T) {
		t.Parallel()

		// Stripe renders a relation as a bare ID or as the whole object depending on what the
		// endpoint expanded; stripe-go's decoding is what makes both land in the same field.
		actual, err := parseStripeEvent(&capitalism.Event{
			ID:      "evt_example",
			Type:    "customer.subscription.created",
			Payload: []byte(`{"id": "sub_example", "status": "active", "customer": {"id": "cus_example", "email": "user@example.com"}}`),
		})
		require.NoError(t, err)
		require.NotNil(t, actual)

		assert.Equal(t, "cus_example", actual.AccountID)
	})

	T.Run("with malformed payload", func(t *testing.T) {
		t.Parallel()

		actual, err := parseStripeEvent(&capitalism.Event{
			ID:      "evt_example",
			Type:    "customer.subscription.updated",
			Payload: []byte(`{"id":`),
		})
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}
