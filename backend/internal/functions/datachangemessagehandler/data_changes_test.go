package datachangemessagehandler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsyncDataChangeMessageHandler_DataChangesEventHandler(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, _, _, analyticsReporter, _, _, decoder := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		// Create a test data change message
		dataChangeMessage := &audit.DataChangeMessage{
			EventType: identity.UserSignedUpServiceEventType,
			UserID:    "test-user-id",
			AccountID: "test-account-id",
			Context:   nil, // When marshaled as empty map {} and unmarshaled, becomes nil
		}

		rawMsg, err := json.Marshal(dataChangeMessage)
		assert.NoError(t, err)

		// Set up decoder mock: DataChangesEventHandler decodes rawMsg into a DataChangeMessage
		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			d := dest.(*audit.DataChangeMessage)
			*d = *dataChangeMessage
			return nil
		}

		// Set up mock expectations
		analyticsReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return nil }
		analyticsReporter.AddUserFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }
		identityRepo.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, dataChangeMessage.UserID, userID)

			return identityfakes.BuildFakeUser(), nil
		}

		assert.NoError(t, handler.DataChangesEventHandler("data_changes")(ctx, rawMsg))
	})

	t.Run("does not report an event that is not on the analytics allowlist", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, _, _, analyticsReporter, _, _, decoder := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		// Catalog table churn carries a user ID like anything else. It used to reach the
		// analytics platform for exactly that reason, which was never a reason anyone chose.
		dataChangeMessage := &audit.DataChangeMessage{
			EventType: mealplanning.ValidIngredientUpdatedServiceEventType,
			UserID:    "test-user-id",
			AccountID: "test-account-id",
		}

		rawMsg, err := json.Marshal(dataChangeMessage)
		require.NoError(t, err)

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, dest any) error {
			d := dest.(*audit.DataChangeMessage)
			*d = *dataChangeMessage
			return nil
		}

		var reported bool
		analyticsReporter.EventOccurredFunc = func(_ context.Context, _, _ string, _ map[string]any) error {
			reported = true

			return nil
		}
		identityRepo.GetUserFunc = func(_ context.Context, _ string) (*identity.User, error) {
			return identityfakes.BuildFakeUser(), nil
		}

		require.NoError(t, handler.DataChangesEventHandler("data_changes")(ctx, rawMsg))
		assert.False(t, reported, "an event off the allowlist reached the analytics platform")
	})

	t.Run("with invalid JSON", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, decoder := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()
		rawMsg := []byte("invalid json")

		decoder.DecodeBytesFunc = func(_ context.Context, _ []byte, _ any) error {
			return errors.New("invalid character 'i' looking for beginning of value")
		}

		err := handler.DataChangesEventHandler("data_changes")(ctx, rawMsg)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "decoding message body")
	})
}

func TestAsyncDataChangeMessageHandler_handleDataChangeMessage(t *testing.T) {
	t.Parallel()

	t.Run("with nil message", func(t *testing.T) {
		t.Parallel()

		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		err := handler.handleDataChangeMessage(ctx, nil, "data_changes")
		require.Error(t, err)
		assert.Equal(t, errRequiredDataIsNil, err)
	})

	t.Run("with analytics event reporting", func(t *testing.T) {
		t.Parallel()

		handler, identityRepo, _, _, analyticsEventReporter, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		dataChangeMessage := &audit.DataChangeMessage{
			EventType: identity.UserSignedUpServiceEventType,
			UserID:    "test-user-id",
			AccountID: "test-account-id",
			Context:   nil,
		}

		analyticsEventReporter.EventOccurredFunc = func(_ context.Context, _ string, _ string, _ map[string]any) error { return nil }
		analyticsEventReporter.AddUserFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }
		identityRepo.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, dataChangeMessage.UserID, userID)

			return identityfakes.BuildFakeUser(), nil
		}

		err := handler.handleDataChangeMessage(ctx, dataChangeMessage, "data_changes")
		assert.NoError(t, err)
	})
}

func TestAsyncDataChangeMessageHandler_handleOutboundNotifications(T *testing.T) {
	T.Run("with nil message", func(t *testing.T) {
		handler, _, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		err := handler.handleOutboundNotifications(ctx, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "nil data change message")
	})

	T.Run("user signed up event", func(t *testing.T) {
		// Set environment variable needed for email configuration
		t.Setenv("DINNER_DONE_BETTER_SERVICE_ENVIRONMENT", "testing")

		handler, identityRepo, _, _, analyticsEventReporter, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		user := identityfakes.BuildFakeUser()
		evf := "email-verification-token"

		dataChangeMessage := &audit.DataChangeMessage{
			EventType: identity.UserSignedUpServiceEventType,
			UserID:    user.ID,
			AccountID: "test-account-id",
			Context: map[string]any{
				identitykeys.UserEmailVerificationTokenKey: evf,
			},
		}

		identityRepo.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)

			return user, nil
		}
		analyticsEventReporter.AddUserFunc = func(_ context.Context, _ string, _ map[string]any) error { return nil }

		err := handler.handleOutboundNotifications(ctx, dataChangeMessage)
		require.NoError(t, err)

		assert.Len(t, identityRepo.GetUserCalls(), 1)
	})

	T.Run("with user fetch error", func(t *testing.T) {
		// Set environment variable needed for email configuration
		t.Setenv("DINNER_DONE_BETTER_SERVICE_ENVIRONMENT", "testing")

		handler, identityRepo, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		dataChangeMessage := &audit.DataChangeMessage{
			EventType: identity.UserSignedUpServiceEventType,
			UserID:    "test-user-id",
			AccountID: "test-account-id",
			Context:   nil,
		}

		expectedError := errors.New("user fetch error")
		identityRepo.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, "test-user-id", userID)

			return nil, expectedError
		}

		err := handler.handleOutboundNotifications(ctx, dataChangeMessage)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "getting user")

		assert.Len(t, identityRepo.GetUserCalls(), 1)
	})

	T.Run("unhandled event type", func(t *testing.T) {
		// Set environment variable needed for email configuration
		t.Setenv("DINNER_DONE_BETTER_SERVICE_ENVIRONMENT", "testing")

		handler, identityRepo, _, _, _, _, _, _ := buildTestAsyncDataChangeMessageHandler(t)

		ctx := t.Context()

		user := identityfakes.BuildFakeUser()

		dataChangeMessage := &audit.DataChangeMessage{
			EventType: "unhandled.event.type",
			UserID:    user.ID,
			AccountID: "test-account-id",
			Context:   nil,
		}

		identityRepo.GetUserFunc = func(_ context.Context, userID string) (*identity.User, error) {
			assert.Equal(t, user.ID, userID)

			return user, nil
		}

		err := handler.handleOutboundNotifications(ctx, dataChangeMessage)
		require.NoError(t, err) // Should handle gracefully with no outbound emails

		assert.Len(t, identityRepo.GetUserCalls(), 1)
	})
}
