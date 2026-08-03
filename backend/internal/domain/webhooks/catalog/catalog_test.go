package catalog

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCatalog(T *testing.T) {
	T.Parallel()

	T.Run("is not empty", func(t *testing.T) {
		t.Parallel()

		// An empty catalog rejects every subscription and dispatches nothing — a total
		// webhook outage that presents as a series of individually plausible rejections.
		assert.NotEmpty(t, Catalog())
	})

	T.Run("partitions every published event into subscribable or excluded", func(t *testing.T) {
		t.Parallel()

		// This is what lets Dispatch skip an unknown event type instead of failing the
		// transaction that emitted it. Dispatch runs inside the caller's transaction, so an
		// error there fails the meal plan rather than the webhook — the drift has to be
		// caught here, at build time, rather than at runtime.
		subscribable := Catalog()

		for eventType := range definitions {
			if Excluded(eventType) {
				assert.NotContains(t, subscribable, eventType,
					"event type %q is excluded but still subscribable", eventType)

				continue
			}

			assert.Contains(t, subscribable, eventType,
				"event type %q is published but neither subscribable nor excluded", eventType)
		}
	})

	T.Run("excludes the events that describe account security activity", func(t *testing.T) {
		t.Parallel()

		// Named explicitly rather than derived from the exclusion list, so that deleting an
		// entry from that list fails here instead of quietly widening what a subscriber can
		// see. An endpoint URL is attacker-supplied; these would be a live feed of an
		// account's authentication activity.
		for _, eventType := range []string{
			"user_logged_in",
			"user_logged_out",
			"user_session_created",
			"user_session_revoked",
			"password_changed",
			"two_factor_secret_changed",
			"two_factor_deactivated",
			"oauth2_client_created",
		} {
			require.True(t, Published(eventType), "event type %q is no longer published; update this test", eventType)
			assert.True(t, Excluded(eventType), "event type %q must not be deliverable to a webhook", eventType)
			assert.False(t, Known(eventType), "event type %q must not be subscribable", eventType)
		}
	})

	T.Run("excludes only events the application actually publishes", func(t *testing.T) {
		t.Parallel()

		// An exclusion naming an event type nothing emits is dead weight that reads as
		// protection. This catches one left behind after its event was renamed or removed.
		for eventType := range excluded {
			assert.True(t, Published(eventType),
				"event type %q is excluded but nothing publishes it", eventType)
		}
	})

	T.Run("returns a copy", func(t *testing.T) {
		t.Parallel()

		// The caller hands this to a dispatcher that retains it. A shared map would let one
		// consumer's mutation change what every other consumer considers dispatchable.
		first := Catalog()
		require.NotEmpty(t, first)

		for eventType := range first {
			delete(first, eventType)
		}

		assert.NotEmpty(t, Catalog())
	})
}
