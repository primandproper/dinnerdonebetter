package audit

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v13/fake"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"

	"github.com/stretchr/testify/assert"
)

func Test_buildDataChangeMessageFromContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		sessionContextData := &sessions.ContextData{
			Requester:       sessions.RequesterInfo{UserID: fake.BuildFakeID()},
			ActiveAccountID: fake.BuildFakeID(),
		}
		ctx = sessions.AttachToContext(ctx, sessionContextData)

		expected := &DataChangeMessage{
			EventType: mealplanning.MealCreatedServiceEventType,
			Context: map[string]any{
				"things": "stuff",
			},
			UserID:    sessionContextData.Requester.UserID,
			AccountID: sessionContextData.ActiveAccountID,
		}

		actual := BuildDataChangeMessageFromContext(ctx, loggingnoop.NewLogger(), expected.EventType, expected.Context)

		assert.Equal(t, expected, actual)
	})
}
