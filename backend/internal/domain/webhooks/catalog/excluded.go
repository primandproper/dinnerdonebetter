package catalog

import (
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
	"github.com/primandproper/platform-go/v15/authentication/passwordreset"
	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/webhooks"
)

// excluded are events the application publishes but that a webhook may never subscribe to.
//
// These describe things happening *to* an account rather than within it — someone signing in,
// rotating a two-factor secret, changing a password, being invited, joining or being removed,
// registering a passkey, minting an OAuth2 client. A subscriber receiving them learns the shape
// of an account's security activity, and an endpoint URL is attacker-supplied: whoever can
// create a webhook on an account they have any foothold in would otherwise get a live feed of
// that account's authentication and membership events.
//
// The set is whole fragments rather than a list of names, and that is the policy rather than a
// shortcut. A new domain event that nobody remembered to classify should be deliverable — that
// is what webhooks are for — whereas a new *identity* event that nobody remembered to classify
// must not be. Platform already marks its credential events Internal in its own fragments; this
// agrees with it and extends it to the rest of each package, so nothing about a person's
// account, memberships or sign-ins reaches a third party from here. The three events this
// application queues for its own mail — a reset link, a verification link, a handle reminder —
// carry the secret the mail is rendered from, and are excluded for that reason.
//
// Excluded events are marked Internal in Catalog, which is how platform refuses them at every
// layer at once: a subscription to one is refused at registration, a dispatch of one is refused
// even if a subscription somehow existed, and the catalog still describes them.
var excluded = excludedEventTypes()

func excludedEventTypes() map[string]struct{} {
	out := map[string]struct{}{
		ddbidentity.PasswordResetTokenCreatedEventType:                  {},
		ddbidentity.UsernameReminderRequestedEventType:                  {},
		ddbidentity.UserEmailAddressVerificationEmailRequestedEventType: {},
		ddbidentity.AccountInvitationMailRequestedEventType:             {},
	}

	for _, fragment := range []webhooks.Catalog{
		identity.EventCatalog(),
		signin.EventCatalog(),
		passwordreset.EventCatalog(),
		passkeys.EventCatalog(),
		oauth2clients.EventCatalog(),
	} {
		for eventType := range fragment {
			out[eventType.String()] = struct{}{}
		}
	}

	return out
}

// Excluded reports whether eventType is published but deliberately not deliverable.
//
// It is exported so a test can assert the two sets are disjoint from the catalog's side, and so
// an operator debugging "why does my webhook never fire" has something to point at other than an
// absence.
func Excluded(eventType string) bool {
	_, found := excluded[eventType]

	return found
}
