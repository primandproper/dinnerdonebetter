package datachangemessagehandler

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	webhooksfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"

	"github.com/primandproper/platform-go/v6/identifiers"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAsyncDataChangeMessageHandler_handleWebhookExecutionRequest(t *testing.T) {
	t.Parallel()

	t.Run("with nil request", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		err := handler.handleWebhookExecutionRequest(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, errRequiredDataIsNil, err)
	})

	t.Run("with account fetch error", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		accountID := identifiers.New()
		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: identifiers.New(),
			AccountID: accountID,
			RequestID: identifiers.New(),
			Payload:   &audit.DataChangeMessage{},
		}

		expectedError := errors.New("account fetch error")
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, accountID).Return((*identity.Account)(nil), expectedError)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "getting account")

		mock.AssertExpectationsForObjects(t, identityRepo)
	})

	t.Run("with webhook fetch error", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		account := identityfakes.BuildFakeAccount()
		webhookID := identifiers.New()

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhookID,
			AccountID: account.ID,
			RequestID: identifiers.New(),
			Payload:   &audit.DataChangeMessage{},
		}

		expectedError := errors.New("webhook fetch error")
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhookID, account.ID).Return((*webhooks.Webhook)(nil), expectedError)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.NoError(t, err) // Should not return error, just log it

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})

	t.Run("with invalid webhook encryption key", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "invalid-hex-key" // Invalid hex

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/json"

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: identifiers.New(),
			Payload:   &audit.DataChangeMessage{},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding webhook encryption key")

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})

	t.Run("with successful JSON webhook execution", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer webhookServer.Close()

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "deadbeefdeadbeefdeadbeefdeadbeef" // Valid 32-char hex key

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/json"
		webhook.Method = http.MethodPost
		webhook.URL = webhookServer.URL

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: identifiers.New(),
			Payload: &audit.DataChangeMessage{
				EventType: identity.UserSignedUpServiceEventType,
				UserID:    identifiers.New(),
				AccountID: account.ID,
			},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})

	t.Run("with non-2xx webhook response", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		webhookServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer webhookServer.Close()

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "deadbeefdeadbeefdeadbeefdeadbeef" // Valid 32-char hex key

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/json"
		webhook.Method = http.MethodPost
		webhook.URL = webhookServer.URL

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: identifiers.New(),
			Payload: &audit.DataChangeMessage{
				EventType: identity.UserSignedUpServiceEventType,
				UserID:    identifiers.New(),
				AccountID: account.ID,
			},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected status code")

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})

	t.Run("with XML webhook payload marshaling error", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "deadbeefdeadbeefdeadbeefdeadbeef" // Valid 32-char hex key

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/xml"
		webhook.Method = http.MethodPost

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: identifiers.New(),
			Payload: &audit.DataChangeMessage{
				EventType: identity.UserSignedUpServiceEventType,
				UserID:    identifiers.New(),
				AccountID: account.ID,
				Context:   nil, // explicit nil to avoid map[string]interface{} marshaling issues
			},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		// XML marshaling of the payload's map fields is not supported, so the request fails before any HTTP call.
		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "marshaling webhook payload")

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})
}
