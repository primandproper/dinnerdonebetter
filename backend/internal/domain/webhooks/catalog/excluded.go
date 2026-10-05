package catalog

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	"github.com/primandproper/platform-go/v15/authentication/oauth2clients"
	"github.com/primandproper/platform-go/v15/authentication/passkeys"
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
// Excluded events are marked Internal in Catalog, which is how platform refuses them at every
// layer at once: a subscription to one is refused at registration, a dispatch of one is refused
// even if a subscription somehow existed, and the catalog still describes them.
//
// The platform-published credential events are listed here by platform's names until platform
// marks them Internal in its own fragments (platform-go#1131); the identity ones are this
// application's until its identity hooks are platform's (platform-go#1130).
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
	identity.UserImpersonatedServiceEventType:                    {},
	identity.UserLoggedOutServiceEventType:                       {},
	identity.UserChangedActiveAccountServiceEventType:            {},
	identity.UserEmailAddressVerifiedEventType:                   {},
	identity.UserEmailAddressVerificationEmailRequestedEventType: {},
	identity.AccountInvitationAcceptedServiceEventType:           {},
	identity.AccountMemberRemovedServiceEventType:                {},
	identity.AccountMembershipPermissionsUpdatedServiceEventType: {},
	identity.AccountOwnershipTransferredServiceEventType:         {},
	identity.UserServiceRolesChangedServiceEventType:             {},
	identity.UserAgreementRecordedServiceEventType:               {},

	passkeys.EventPasskeyRegistered.String():   {},
	passkeys.EventPasskeyArchived.String():     {},
	oauth2clients.EventClientCreated.String():  {},
	oauth2clients.EventClientUpdated.String():  {},
	oauth2clients.EventClientArchived.String(): {},
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
