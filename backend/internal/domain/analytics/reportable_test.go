package analytics

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/stretchr/testify/assert"
)

func TestReportable(T *testing.T) {
	T.Parallel()

	T.Run("reports the events on the allowlist", func(t *testing.T) {
		t.Parallel()

		assert.True(t, Reportable(identity.UserSignedUpServiceEventType))
		assert.True(t, Reportable(mealplanning.MealPlanFinalizedServiceEventType))
	})

	T.Run("does not report events off it", func(t *testing.T) {
		t.Parallel()

		// This is the whole behavior change: catalog table churn used to reach the analytics
		// platform because it carried a user ID, which was never a reason for anyone to want
		// it there.
		assert.False(t, Reportable(mealplanning.ValidIngredientUpdatedServiceEventType))
		assert.False(t, Reportable(mealplanning.RecipeStepUpdatedServiceEventType))
	})

	T.Run("does not report an unknown event type", func(t *testing.T) {
		t.Parallel()

		// An event added to a domain is unreported until someone puts it on the list, which
		// is what makes the list the whole of the decision rather than half of it.
		assert.False(t, Reportable("some_event_nobody_classified"))
		assert.False(t, Reportable(""))
	})

	T.Run("names only events the application publishes", func(t *testing.T) {
		t.Parallel()

		// The constants make a deleted event a compile error here, but they do not stop an
		// event from being listed that no code ever emits — an allowlist entry for an event
		// that never fires is a metric that reads as a flat zero rather than as a mistake.
		// catalog.Published is generated from the same constants, so this asserts the entry
		// corresponds to a real declared event rather than a value that drifted.
		for eventType := range reportable {
			assert.True(t, catalog.Published(eventType),
				"event type %q is reported to analytics but is not one the application publishes", eventType)
		}
	})

	T.Run("is not empty", func(t *testing.T) {
		t.Parallel()

		// An empty allowlist silently turns off product analytics entirely, and every
		// dashboard it feeds reads as zero traffic rather than as a broken pipeline.
		assert.NotEmpty(t, reportable)
	})
}
