package catalog

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth"
)

// excluded are events the application publishes but that a webhook may never subscribe to.
//
// These describe things happening *to* an account rather than within it — someone signing in,
// rotating a two-factor secret, changing a password, being removed from an account, minting an
// OAuth2 client. A subscriber receiving them learns the shape of an account's security activity,
// and an endpoint URL is attacker-supplied: whoever can create a webhook on an account they have
// any foothold in would otherwise get a live feed of that account's authentication events.
//
// This is a denylist rather than an allowlist because the default has to be safe in the other
// direction too. A new domain event that nobody remembered to classify should be deliverable —
// that is what webhooks are for — whereas a new *identity* event that nobody remembered to
// classify must not be. Keeping the exclusions here, beside the identity constants they name,
// is what makes the omission visible in review of the change that adds one.
//
// Excluded events are excluded at every layer at once: Catalog omits them, so a subscription to
// one is refused at registration, and Known reports false for them, so a dispatch of one is
// refused even if a subscription somehow existed.
var excluded = map[string]struct{}{
	identity.UserSignedUpServiceEventType:                        {},
	identity.UserArchivedServiceEventType:                        {},
	identity.TwoFactorSecretVerifiedServiceEventType:             {},
	identity.TwoFactorDeactivatedServiceEventType:                {},
	identity.TwoFactorSecretChangedServiceEventType:              {},
	identity.PasswordResetTokenCreatedEventType:                  {},
	identity.PasswordResetTokenRedeemedEventType:                 {},
	identity.PasswordChangedEventType:                            {},
	identity.EmailAddressChangedEventType:                        {},
	identity.UsernameChangedEventType:                            {},
	identity.UserDetailsChangedEventType:                         {},
	identity.UsernameReminderRequestedEventType:                  {},
	identity.UserLoggedInServiceEventType:                        {},
	identity.UserLoggedOutServiceEventType:                       {},
	identity.UserChangedActiveAccountServiceEventType:            {},
	auth.UserSessionCreatedEventType:                             {},
	auth.UserSessionRevokedEventType:                             {},
	identity.UserEmailAddressVerifiedEventType:                   {},
	identity.UserEmailAddressVerificationEmailRequestedEventType: {},
	identity.AccountInvitationAcceptedServiceEventType:           {},
	identity.AccountMemberRemovedServiceEventType:                {},
	identity.AccountMembershipPermissionsUpdatedServiceEventType: {},
	identity.AccountOwnershipTransferredServiceEventType:         {},
	oauth.OAuth2ClientCreatedServiceEventType:                    {},
	oauth.OAuth2ClientArchivedServiceEventType:                   {},
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
