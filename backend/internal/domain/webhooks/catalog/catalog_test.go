package catalog

import (
	"testing"

	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/waitlists"
	"github.com/primandproper/platform-go/v15/webhooks"

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

	T.Run("composes without a collision", func(t *testing.T) {
		t.Parallel()

		// Catalog panics on one, which is what a process should do; this is where it is
		// caught first. Two packages naming one event is a programming error, and platform's
		// fragments are prefixed with their package for exactly this reason.
		assert.NotPanics(t, func() { Catalog() })
	})

	T.Run("carries every published event, this application's and platform's", func(t *testing.T) {
		t.Parallel()

		catalog := Catalog()
		for eventType := range definitions {
			assert.Contains(t, catalog, eventType, "generated event type %q is missing from the catalog", eventType)
		}

		for eventType := range waitlists.EventCatalog() {
			assert.Contains(t, catalog, eventType, "platform event type %q is missing from the catalog", eventType)
		}

		for eventType := range comments.EventCatalog() {
			assert.Contains(t, catalog, eventType, "platform event type %q is missing from the catalog", eventType)
		}

		for eventType := range identity.EventCatalog() {
			assert.Contains(t, catalog, eventType, "platform event type %q is missing from the catalog", eventType)
		}
	})

	T.Run("marks every excluded event internal, and no other", func(t *testing.T) {
		t.Parallel()

		// This is what lets a dispatch of an excluded event be refused rather than fail the
		// transaction that emitted it, and what keeps the emitter's unsubscribable counter
		// counting only constants that fell out of the catalog.
		catalog := Catalog()
		for eventType, definition := range catalog {
			assert.Equal(t, Excluded(eventType.String()), definition.Internal,
				"event type %q: excluded=%t but internal=%t", eventType, Excluded(eventType.String()), definition.Internal)
			assert.Equal(t, !definition.Internal, catalog.Subscribable(eventType))
		}
	})

	T.Run("excludes the events that describe account security activity", func(t *testing.T) {
		t.Parallel()

		// Named explicitly rather than derived from the exclusion list, so that dropping a
		// fragment from that list fails here instead of quietly widening what a subscriber
		// can see. An endpoint URL is attacker-supplied; these would be a live feed of an
		// account's authentication activity, and two of them carry a bearer secret.
		for _, eventType := range []string{
			signin.EventUserAuthenticated.String(),
			signin.EventSignInsRevoked.String(),
			signin.EventPasswordUpdated.String(),
			signin.EventTOTPSecretRefreshed.String(),
			identity.EventUserRegistered.String(),
			identity.EventUserPasswordChanged.String(),
			identity.EventInvitationCreated.String(),
			identity.EventMembershipRemoved.String(),
			passwordreset.EventTokenIssued.String(),
			ddbidentity.PasswordResetTokenCreatedEventType,
			ddbidentity.UserEmailAddressVerificationEmailRequestedEventType,
			passkeys.EventPasskeyRegistered.String(),
			passkeys.EventPasskeyArchived.String(),
			oauth2clients.EventClientCreated.String(),
		} {
			require.True(t, Published(eventType), "event type %q is no longer published; update this test", eventType)
			assert.True(t, Excluded(eventType), "event type %q must not be deliverable to a webhook", eventType)
			assert.False(t, Known(eventType), "event type %q must not be subscribable", eventType)
		}
	})

	T.Run("excludes only events something actually publishes", func(t *testing.T) {
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

	T.Run("a subscribable event is one a dispatcher accepts", func(t *testing.T) {
		t.Parallel()

		var known webhooks.EventType
		for eventType := range definitions {
			if !Excluded(eventType.String()) {
				known = eventType
				break
			}
		}
		require.NotEmpty(t, known)
		assert.True(t, Known(known.String()))
	})
}
