package identity

// The identity events this application publishes, and the vocabulary that travels with
// them.
//
// They are declared here rather than beside the types they describe, because the types are
// platform's now and the events are not. What an event is named is a contract with every
// webhook subscriber and every search-index rule in this repository, so the names outlived
// the structs they used to sit next to — see internal/repositories/postgres/identitystore
// for where each one is written, on the transaction that made the rows it announces.
const (
	// UserSignedUpServiceEventType indicates a user signed up.
	UserSignedUpServiceEventType = "user_signed_up"
	// UserArchivedServiceEventType indicates a user archived their account.
	UserArchivedServiceEventType = "user_archived"
	// UserStatusChangedServiceEventType indicates a user had their user status changed.
	UserStatusChangedServiceEventType = "user_status_changed"
	// UserPasswordChangeRequiredServiceEventType indicates a user was marked as requiring a password change.
	UserPasswordChangeRequiredServiceEventType = "user_password_change_required"

	// AccountCreatedServiceEventType indicates an account was created.
	AccountCreatedServiceEventType = "account_created"
	// AccountUpdatedServiceEventType indicates an account was updated.
	AccountUpdatedServiceEventType = "account_updated"
	// AccountArchivedServiceEventType indicates an account was archived.
	AccountArchivedServiceEventType = "account_archived"
	// AccountMemberRemovedServiceEventType indicates an account member was removed.
	AccountMemberRemovedServiceEventType = "account_member_removed"
	// UserServiceRolesChangedServiceEventType indicates a user's service roles were set.
	//
	// New with the platform adoption: this application had no RPC that set a service
	// role, so nothing published one. platform's SetUserServiceRoles is the operation
	// that grants and revokes them, and the hook it fires is the one place an audit
	// trail can say who became an administrator and when.
	UserServiceRolesChangedServiceEventType = "user_service_roles_changed"
	// UserAgreementRecordedServiceEventType indicates a user accepted terms.
	//
	// Also new with the adoption. platform records agreements against the user;
	// nothing here asked for them before, and the event exists so that the thing a
	// compliance question is actually about — when somebody agreed to what — is on the
	// same channel as every other user fact.
	UserAgreementRecordedServiceEventType = "user_agreement_recorded"
	// AccountMembershipPermissionsUpdatedServiceEventType indicates an account member's permissions were modified.
	AccountMembershipPermissionsUpdatedServiceEventType = "account_membership_permissions_updated"
	// AccountSetAsDefaultServiceEventType indicates an account was selected as a user's default.
	AccountSetAsDefaultServiceEventType = "account_set_as_default"
	// AccountOwnershipTransferredServiceEventType indicates an account was transferred to another owner.
	AccountOwnershipTransferredServiceEventType = "account_ownership_transferred"

	// AccountInvitationCreatedServiceEventType indicates an account invitation was created.
	AccountInvitationCreatedServiceEventType = "account_invitation_created"
	// AccountInvitationCanceledServiceEventType indicates an account invitation was canceled.
	AccountInvitationCanceledServiceEventType = "account_invitation_canceled"
	// AccountInvitationAcceptedServiceEventType indicates an account invitation was accepted.
	AccountInvitationAcceptedServiceEventType = "account_invitation_accepted"
	// AccountInvitationRejectedServiceEventType indicates an account invitation was rejected.
	AccountInvitationRejectedServiceEventType = "account_invitation_rejected"

	// MobileNotificationRequestTypeHouseholdInvitationAccepted indicates a household invitation accepted notification.
	MobileNotificationRequestTypeHouseholdInvitationAccepted = "household_invitation_accepted"
	// ExcludedUserIDContextKey is the key used in MobileNotificationRequest.Context for the user to exclude.
	ExcludedUserIDContextKey = "excludedUserID"
)
