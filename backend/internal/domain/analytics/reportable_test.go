package analytics

import (
	"encoding/json"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/datachanges"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks/catalog"

	"github.com/primandproper/platform-go/v15/billing"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// envelopeFor wraps payload the way platform's Emitter does for the broker.
func envelopeFor(t *testing.T, eventType webhooks.EventType, payload any) *webhooks.Envelope {
	t.Helper()

	raw, err := json.Marshal(payload)
	require.NoError(t, err)

	return &webhooks.Envelope{EventType: eventType, ID: identifiers.New(), Payload: raw}
}

func TestReportable(T *testing.T) {
	T.Parallel()

	T.Run("reports the events on the allowlist", func(t *testing.T) {
		t.Parallel()

		assert.True(t, Reportable(identity.EventUserRegistered.String()))
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
		// The catalog is composed from the same constants, this application's and
		// platform's, so this asserts the entry corresponds to a real declared event.
		for eventType := range reportable {
			assert.True(t, catalog.Published(eventType.String()),
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

func TestRead(T *testing.T) {
	T.Parallel()

	T.Run("reads this application's own event off its message", func(t *testing.T) {
		t.Parallel()

		userID, accountID := identifiers.New(), identifiers.New()
		event := envelopeFor(t, webhooks.EventType(mealplanning.RecipeCreatedServiceEventType), &datachanges.Message{
			EventType: mealplanning.RecipeCreatedServiceEventType,
			UserID:    userID,
			AccountID: accountID,
			Context:   map[string]any{"recipe.id": "r1"},
		})

		report, ok, err := Read(event)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, accountID, report.AccountID)
		assert.Equal(t, "r1", report.Properties["recipe.id"])
	})

	T.Run("reads a registration without its verification link", func(t *testing.T) {
		t.Parallel()

		// The link is a bearer secret. It is on the event because the mail needs it; the
		// analytics vendor does not, and the old arrangement — forward the whole context —
		// sent it anyway.
		userID, accountID := identifiers.New(), identifiers.New()
		event := envelopeFor(t, identity.EventUserRegistered, &identity.UserEvent{
			UserID:                        userID,
			AccountID:                     accountID,
			EmailAddressVerificationToken: "secret",
		})

		report, ok, err := Read(event)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, userID, report.UserID)
		assert.Equal(t, accountID, report.Properties["accountID"])
		assert.NotContains(t, report.Properties, "emailAddressVerificationToken")
		for _, v := range report.Properties {
			assert.NotEqual(t, "secret", v)
		}
	})

	T.Run("attributes an account opened to its owner", func(t *testing.T) {
		t.Parallel()

		owner := identifiers.New()
		event := envelopeFor(t, identity.EventAccountCreated, &identity.AccountEvent{AccountID: identifiers.New(), OwnerUserID: owner})

		report, ok, err := Read(event)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, owner, report.UserID)
	})

	T.Run("attributes an invitation to its sender and an acceptance to its acceptor", func(t *testing.T) {
		t.Parallel()

		from, to := identifiers.New(), identifiers.New()

		sent, ok, err := Read(envelopeFor(t, identity.EventInvitationCreated, &identity.InvitationEvent{
			InvitationID: identifiers.New(), AccountID: identifiers.New(), FromUser: from, Token: "secret",
		}))
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, from, sent.UserID)
		for _, v := range sent.Properties {
			assert.NotEqual(t, "secret", v)
		}

		accepted, ok, err := Read(envelopeFor(t, identity.EventInvitationAccepted, &identity.InvitationEvent{
			InvitationID: identifiers.New(), AccountID: identifiers.New(), FromUser: from, ToUser: &to,
		}))
		require.NoError(t, err)
		require.True(t, ok)
		assert.Equal(t, to, accepted.UserID)
	})

	T.Run("names the account and nobody for a subscription", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()
		event := envelopeFor(t, billing.EventSubscriptionCreated, &billing.SubscriptionEvent{
			SubscriptionID: identifiers.New(), AccountID: accountID, ProductID: "pro",
		})

		report, ok, err := Read(event)
		require.NoError(t, err)
		require.True(t, ok)
		assert.Empty(t, report.UserID)
		assert.Equal(t, accountID, report.AccountID)
		assert.Equal(t, "pro", report.Properties["productID"])
	})

	T.Run("reports false for an event off the allowlist", func(t *testing.T) {
		t.Parallel()

		_, ok, err := Read(envelopeFor(t, identity.EventUserProfileUpdated, &identity.UserEvent{UserID: identifiers.New()}))
		require.NoError(t, err)
		assert.False(t, ok)
	})

	T.Run("refuses a payload that is not what its event promises", func(t *testing.T) {
		t.Parallel()

		event := &webhooks.Envelope{EventType: identity.EventUserRegistered, Payload: json.RawMessage(`"not an object"`)}

		_, _, err := Read(event)
		require.Error(t, err)
	})

	T.Run("refuses nil", func(t *testing.T) {
		t.Parallel()

		_, _, err := Read(nil)
		require.Error(t, err)
	})
}
