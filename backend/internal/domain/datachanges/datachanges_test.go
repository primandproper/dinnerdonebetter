package datachanges

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/primitives-go/v2/fake"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"

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
