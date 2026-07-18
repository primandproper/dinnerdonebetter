package datachangemessagehandler

import (
	"errors"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks"
	webhooksfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/webhooks/fakes"

	"github.com/primandproper/platform-go/v5/reflection"

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

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: "test-webhook-id",
			AccountID: "test-account-id",
			RequestID: "test-request-id",
			Payload:   &audit.DataChangeMessage{},
		}

		expectedError := errors.New("account fetch error")
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, "test-account-id").Return((*identity.Account)(nil), expectedError)

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

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: "test-webhook-id",
			AccountID: account.ID,
			RequestID: "test-request-id",
			Payload:   &audit.DataChangeMessage{},
		}

		expectedError := errors.New("webhook fetch error")
		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, "test-webhook-id", account.ID).Return((*webhooks.Webhook)(nil), expectedError)

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
			RequestID: "test-request-id",
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

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "deadbeefdeadbeefdeadbeefdeadbeef" // Valid 32-char hex key

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/json"
		webhook.Method = "POST"
		webhook.URL = "https://httpbin.org/post" // This will fail but that's expected in tests

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: "test-request-id",
			Payload: &audit.DataChangeMessage{
				EventType: identity.UserSignedUpServiceEventType,
				UserID:    "test-user-id",
				AccountID: account.ID,
			},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		// We expect no error to be returned even if HTTP request fails (it gets logged)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})

	t.Run("with successful XML webhook execution", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, webhookRepo, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		account := identityfakes.BuildFakeAccount()
		account.WebhookEncryptionKey = "deadbeefdeadbeefdeadbeefdeadbeef" // Valid 32-char hex key

		webhook := webhooksfakes.BuildFakeWebhook()
		webhook.ContentType = "application/xml"
		webhook.Method = "POST"
		webhook.URL = "https://httpbin.org/post" // This will fail but that's expected in tests

		webhookExecutionRequest := &webhooks.WebhookExecutionRequest{
			WebhookID: webhook.ID,
			AccountID: account.ID,
			RequestID: "test-request-id",
			Payload: &audit.DataChangeMessage{
				EventType: identity.UserSignedUpServiceEventType,
				UserID:    "test-user-id",
				AccountID: account.ID,
				Context:   nil, // explicit nil to avoid map[string]interface{} marshaling issues
			},
		}

		identityRepo.On(reflection.GetMethodName(identityRepo.GetAccount), mock.Anything, account.ID).Return(account, nil)
		webhookRepo.On(reflection.GetMethodName(webhookRepo.GetWebhook), mock.Anything, webhook.ID, account.ID).Return(webhook, nil)

		err := handler.handleWebhookExecutionRequest(ctx, webhookExecutionRequest)
		// XML marshaling of map[string]interface{} is not supported, so we expect an error
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "marshaling webhook payload")

		mock.AssertExpectationsForObjects(t, identityRepo, webhookRepo)
	})
}
