package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/primandproper/platform-go/v7/email"

	"github.com/stretchr/testify/assert"
)

func TestAsyncDataChangeMessageHandler_OutboundEmailsEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return nil }
		analyticsEventReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return nil }

		err = handler.OutboundEmailsEventHandler("outbound_emails")(ctx, rawMsg)
		assert.NoError(t, err)
	})

	t.Run("with invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		rawMsg := []byte("invalid json")

		err := handler.OutboundEmailsEventHandler("outbound_emails")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decoding JSON body")
	})

	t.Run("with email sending error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		expectedError := errors.New("email sending error")
		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return expectedError }

		err = handler.OutboundEmailsEventHandler("outbound_emails")(ctx, rawMsg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sending email")
	})

	t.Run("with analytics error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		rawMsg, err := json.Marshal(emailMessage)
		assert.NoError(t, err)

		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return nil }
		expectedError := errors.New("analytics error")
		analyticsEventReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return expectedError }

		err = handler.OutboundEmailsEventHandler("outbound_emails")(ctx, rawMsg)
		assert.NoError(t, err) // Should not return error, just log it
	})
}

func TestAsyncDataChangeMessageHandler_handleEmailRequest(t *testing.T) {
	t.Parallel()

	t.Run("with nil email message", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		err := handler.handleEmailRequest(ctx, nil)
		assert.Error(t, err)
		assert.Equal(t, errRequiredDataIsNil, err)
	})

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return nil }
		analyticsEventReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return nil }

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.NoError(t, err)
	})

	t.Run("with email sending error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		expectedError := errors.New("email sending error")
		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return expectedError }

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "sending email")
	})

	t.Run("with analytics error", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, analyticsEventReporter, emailer, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		emailMessage := &email.OutboundEmailMessage{
			ToAddress:   "test@example.com",
			ToName:      "Test User",
			FromAddress: "noreply@example.com",
			FromName:    "Test App",
			Subject:     "Test Subject",
			HTMLContent: "<p>Test content</p>",
			UserID:      "test-user-id",
		}

		emailer.SendEmailFunc = func(_ context.Context, _ *email.OutboundEmailMessage) error { return nil }
		expectedError := errors.New("analytics error")
		analyticsEventReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return expectedError }

		err := handler.handleEmailRequest(ctx, emailMessage)
		assert.NoError(t, err) // Should not return error, just log it
	})
}
