package identity

// The identity events are platform's now: identity.EventUserRegistered and the rest of
// identity.EventCatalog, signin.EventCatalog and passwordreset.EventCatalog, each
// recorded by that package's RecordingHooks on the transaction that made the rows it
// announces. What remains here is the vocabulary this application attaches to the
// notifications it builds from them.
const (
	// MobileNotificationRequestTypeHouseholdInvitationAccepted indicates a household invitation accepted notification.
	MobileNotificationRequestTypeHouseholdInvitationAccepted = "household_invitation_accepted"
	// ExcludedUserIDContextKey is the key used in MobileNotificationRequest.Context for the user to exclude.
	ExcludedUserIDContextKey = "excludedUserID"
)
