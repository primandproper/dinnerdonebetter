package datachanges

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
)

func TestMessageFromContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: fake.BuildFakeID()},
			ActiveAccountID: fake.BuildFakeID(),
		}
		ctx = sessions.AttachToContext(ctx, sessionContextData)

		expected := &Message{
			EventType: "meal_created",
			Context: map[string]any{
				"things": "stuff",
			},
			UserID:    sessionContextData.Requester.UserID,
			AccountID: sessionContextData.ActiveAccountID,
		}

		actual := MessageFromContext(ctx, loggingnoop.NewLogger(), expected.EventType, expected.Context)

		assert.Equal(t, expected, actual)
	})

	T.Run("without a session", func(t *testing.T) {
		t.Parallel()

		actual := MessageFromContext(t.Context(), loggingnoop.NewLogger(), "meal_created", nil)

		assert.Equal(t, "meal_created", actual.EventType)
		assert.Empty(t, actual.UserID)
		assert.Empty(t, actual.AccountID)
	})
}

func TestEvent(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: fake.BuildFakeID()},
			ActiveAccountID: fake.BuildFakeID(),
		}
		ctx := sessions.AttachToContext(t.Context(), sessionContextData)

		eventType := fake.BuildFakeString()
		metadata := map[string]any{fake.BuildFakeString(): fake.BuildFakeString()}

		event, msg := Event(ctx, loggingnoop.NewLogger(), eventType, "", metadata)

		assert.Equal(t, webhooks.EventType(eventType), event.EventType)
		// The payload platform publishes is the message handed back, so what a caller reads
		// off the second return is what the consumer will decode.
		assert.Same(t, msg, event.Payload)
		assert.Equal(t, eventType, msg.EventType)
		assert.Equal(t, metadata, msg.Context)
		assert.Equal(t, sessionContextData.Requester.UserID, msg.UserID)
		assert.Equal(t, sessionContextData.ActiveAccountID, msg.AccountID)
	})

	T.Run("an account the writer names overrides the session's", func(t *testing.T) {
		t.Parallel()

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: fake.BuildFakeID()},
			ActiveAccountID: fake.BuildFakeID(),
		})
		accountID := fake.BuildFakeID()

		_, msg := Event(ctx, loggingnoop.NewLogger(), fake.BuildFakeString(), accountID, nil)

		assert.Equal(t, accountID, msg.AccountID)
	})

	T.Run("with no session and no logger", func(t *testing.T) {
		t.Parallel()

		// A background job: nobody to attribute it to, and nothing to log through.
		event, msg := Event(t.Context(), nil, fake.BuildFakeString(), "", nil)

		assert.NotNil(t, event)
		assert.Empty(t, msg.UserID)
		assert.Empty(t, msg.AccountID)
	})
}

func TestScope(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()

		assert.Equal(t, tenancy.Of(accountID), Scope(accountID))
	})

	T.Run("with no account", func(t *testing.T) {
		t.Parallel()

		assert.True(t, Scope("").IsGlobal())
	})
}
