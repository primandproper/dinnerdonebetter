package identitystore

import (
	"context"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbidentity "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// The twenty-five operations, and what each one owes the log.
//
// Three rules run through all of them.
//
// An operation that wrote several rows records several entries under one event. A
// registration writes a user, an account and a membership, and it is one thing that
// happened — so the log holds three rows and the outbox holds one user_signed_up. A
// subscriber told three times that somebody arrived would act three times.
//
// Every value comes off the arguments rather than off a read. platform hands each hook
// the rows the operation wrote and, where the operation replaced something, what was
// there before — the previous owner of a transferred account, the roles a membership
// held, whether a password change was already required. None of that is readable inside
// the transaction that overwrote it, which is why it is an argument.
//
// And the entries a chain has to be able to find go on that chain. An entry about an
// account is filed under the account; one about a person alone is filed under them. See
// record.go, and audit.ScopeFor for why the distinction is load bearing.

// AfterRegister records a person arriving: their user, their first account, and the
// membership that puts them in it.
func (h *Hooks) AfterRegister(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	registration *platformidentity.Registration,
) error {
	if registration == nil {
		return nil
	}

	return h.record(ctx, tx,
		userEntry(registration.User.ID, audit.AuditLogEventTypeCreated, ddbidentity.UserSignedUpServiceEventType),
		accountEntry(registration.Account.ID, registration.User.ID, audit.AuditLogEventTypeCreated, ""),
		membershipEntry(registration.Membership.ID, registration.User.ID, registration.Account.ID,
			audit.AuditLogEventTypeCreated, ""),
	)
}

// AfterRegisterWithInvitation records a person who arrived because somebody asked them to.
//
// The invitation is recorded as accepted rather than as a second creation: the row was
// written when it was sent, and this is the moment it was answered.
func (h *Hooks) AfterRegisterWithInvitation(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	registration *platformidentity.InvitedRegistration,
) error {
	if registration == nil {
		return nil
	}

	return h.record(ctx, tx,
		userEntry(registration.User.ID, audit.AuditLogEventTypeCreated, ddbidentity.UserSignedUpServiceEventType),
		invitationEntry(registration.Invitation.ID, registration.Membership.BelongsToAccount,
			audit.AuditLogEventTypeUpdated, ddbidentity.AccountInvitationAcceptedServiceEventType),
		membershipEntry(registration.Membership.ID, registration.User.ID,
			registration.Membership.BelongsToAccount, audit.AuditLogEventTypeCreated, ""),
	)
}

// AfterInvite records an account asking somebody to join it.
func (h *Hooks) AfterInvite(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	invitation *platformidentity.Invitation,
) error {
	if invitation == nil {
		return nil
	}

	return h.record(ctx, tx, invitationEntry(invitation.ID, invitation.BelongsToAccount,
		audit.AuditLogEventTypeCreated, ddbidentity.AccountInvitationCreatedServiceEventType).
		WithToken(invitation.Token))
}

// AfterAcceptInvitation records an invitation answered yes, and the membership it produced.
func (h *Hooks) AfterAcceptInvitation(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	acceptance *platformidentity.Acceptance,
) error {
	if acceptance == nil {
		return nil
	}

	return h.record(ctx, tx,
		invitationEntry(acceptance.Invitation.ID, acceptance.Invitation.BelongsToAccount,
			audit.AuditLogEventTypeUpdated, ddbidentity.AccountInvitationAcceptedServiceEventType),
		membershipEntry(acceptance.Membership.ID, acceptance.Membership.BelongsToUser,
			acceptance.Membership.BelongsToAccount, audit.AuditLogEventTypeCreated, ""),
	)
}

// AfterRejectInvitation records an invitation answered no.
func (h *Hooks) AfterRejectInvitation(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	invitation *platformidentity.Invitation,
) error {
	if invitation == nil {
		return nil
	}

	return h.record(ctx, tx, invitationEntry(invitation.ID, invitation.BelongsToAccount,
		audit.AuditLogEventTypeUpdated, ddbidentity.AccountInvitationRejectedServiceEventType))
}

// AfterCancelInvitation records an account withdrawing an invitation it sent.
func (h *Hooks) AfterCancelInvitation(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	invitation *platformidentity.Invitation,
) error {
	if invitation == nil {
		return nil
	}

	return h.record(ctx, tx, invitationEntry(invitation.ID, invitation.BelongsToAccount,
		audit.AuditLogEventTypeArchived, ddbidentity.AccountInvitationCanceledServiceEventType))
}

// AfterCreateAccount records somebody who was already here starting a second household.
//
// account_created rather than user_signed_up, which is the distinction platform drew this
// hook to make: a registration is a person arriving, and an audit trail that rendered the
// second as the first would say a user registered twice.
func (h *Hooks) AfterCreateAccount(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	account *platformidentity.Account,
	membership *platformidentity.Membership,
) error {
	if account == nil || membership == nil {
		return nil
	}

	return h.record(ctx, tx,
		accountEntry(account.ID, account.OwnerUserID, audit.AuditLogEventTypeCreated,
			ddbidentity.AccountCreatedServiceEventType),
		membershipEntry(membership.ID, membership.BelongsToUser, membership.BelongsToAccount,
			audit.AuditLogEventTypeCreated, ""),
	)
}

// AfterTransferAccountOwnership records an account changing hands, and who it came from.
func (h *Hooks) AfterTransferAccountOwnership(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	account *platformidentity.Account,
	previousOwnerUserID string,
) error {
	if account == nil {
		return nil
	}

	e := accountEntry(account.ID, account.OwnerUserID, audit.AuditLogEventTypeUpdated,
		ddbidentity.AccountOwnershipTransferredServiceEventType)
	// The previous owner is on the entry because the column no longer holds them: by the
	// time this runs it names the new one, so an entry that did not carry this could not
	// answer the only question a transfer raises.
	e.metadata["previousOwnerUserID"] = previousOwnerUserID

	return h.record(ctx, tx, e)
}

// AfterSetDefaultAccount records somebody moving between the households they belong to.
func (h *Hooks) AfterSetDefaultAccount(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	membership *platformidentity.Membership,
	previousAccountID string,
) error {
	if membership == nil {
		return nil
	}

	e := membershipEntry(membership.ID, membership.BelongsToUser, membership.BelongsToAccount,
		audit.AuditLogEventTypeUpdated, ddbidentity.UserChangedActiveAccountServiceEventType)
	e.metadata["previousAccountID"] = previousAccountID

	return h.record(ctx, tx, e)
}

// AfterArchiveUser records somebody leaving, and every membership that ended with them.
func (h *Hooks) AfterArchiveUser(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	endedMemberships []*platformidentity.Membership,
) error {
	if user == nil {
		return nil
	}

	entries := []*entry{userEntry(user.ID, audit.AuditLogEventTypeArchived, ddbidentity.UserArchivedServiceEventType)}
	for _, m := range endedMemberships {
		// Each account learns that it lost a member, on its own chain, because that is
		// where whoever is left will look for it.
		entries = append(entries, membershipEntry(m.ID, m.BelongsToUser, m.BelongsToAccount,
			audit.AuditLogEventTypeArchived, ddbidentity.AccountMemberRemovedServiceEventType))
	}

	return h.record(ctx, tx, entries...)
}

// AfterArchiveAccount records a household closing, and every membership that ended with it.
func (h *Hooks) AfterArchiveAccount(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	account *platformidentity.Account,
	endedMemberships []*platformidentity.Membership,
) error {
	if account == nil {
		return nil
	}

	entries := []*entry{accountEntry(account.ID, account.OwnerUserID, audit.AuditLogEventTypeArchived,
		ddbidentity.AccountArchivedServiceEventType)}
	for _, m := range endedMemberships {
		entries = append(entries, membershipEntry(m.ID, m.BelongsToUser, m.BelongsToAccount,
			audit.AuditLogEventTypeArchived, ""))
	}

	return h.record(ctx, tx, entries...)
}

// AfterUpdateUserAccountStatus records a ban, a reinstatement, or any other status move.
func (h *Hooks) AfterUpdateUserAccountStatus(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	previousStatus platformidentity.AccountStatus,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.UserStatusChangedServiceEventType)
	e.metadata["previousStatus"] = string(previousStatus)
	e.metadata["newStatus"] = string(user.AccountStatus)

	return h.record(ctx, tx, e)
}

// AfterSetUserServiceRoles records somebody becoming, or ceasing to be, an administrator.
func (h *Hooks) AfterSetUserServiceRoles(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	previousRoles []string,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.UserServiceRolesChangedServiceEventType)
	e.metadata["previousRoles"] = previousRoles
	e.metadata["newRoles"] = user.ServiceRoles

	return h.record(ctx, tx, e)
}

// AfterUpdateProfile records a change to a user's own details.
//
// One entry whatever moved, and the fields that moved on the metadata. This application
// used to publish three events here — username_changed, email_address_changed,
// user_details_changed — because it had three RPCs. platform has one operation, and
// which fields it touched is what `changed` is for.
func (h *Hooks) AfterUpdateProfile(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	changed []string,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.UserDetailsChangedEventType)
	e.metadata["changed"] = changed

	return h.record(ctx, tx, e)
}

// AfterUpdateAccount records a household's own details changing.
func (h *Hooks) AfterUpdateAccount(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	account *platformidentity.Account,
	changed []string,
) error {
	if account == nil {
		return nil
	}

	e := accountEntry(account.ID, account.OwnerUserID, audit.AuditLogEventTypeUpdated,
		ddbidentity.AccountUpdatedServiceEventType)
	e.metadata["changed"] = changed

	return h.record(ctx, tx, e)
}

// AfterRecordAgreement records somebody accepting terms.
func (h *Hooks) AfterRecordAgreement(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	agreements []platformidentity.Agreement,
) error {
	if user == nil {
		return nil
	}

	accepted := make([]string, 0, len(agreements))
	for _, a := range agreements {
		accepted = append(accepted, string(a))
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.UserAgreementRecordedServiceEventType)
	e.metadata["agreements"] = accepted

	return h.record(ctx, tx, e)
}

// AfterSetMembershipRoles records somebody's permissions in a household changing.
func (h *Hooks) AfterSetMembershipRoles(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	membership *platformidentity.Membership,
	previousRoles []string,
) error {
	if membership == nil {
		return nil
	}

	e := membershipEntry(membership.ID, membership.BelongsToUser, membership.BelongsToAccount,
		audit.AuditLogEventTypeUpdated, ddbidentity.AccountMembershipPermissionsUpdatedServiceEventType)
	e.metadata["previousRoles"] = previousRoles
	e.metadata["newRoles"] = membership.Roles

	return h.record(ctx, tx, e)
}

// AfterRemoveMembership records somebody being taken off a household's roster.
func (h *Hooks) AfterRemoveMembership(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	membership *platformidentity.Membership,
	newDefaultAccountID string,
) error {
	if membership == nil {
		return nil
	}

	e := membershipEntry(membership.ID, membership.BelongsToUser, membership.BelongsToAccount,
		audit.AuditLogEventTypeArchived, ddbidentity.AccountMemberRemovedServiceEventType)
	// Where they landed, when losing this membership moved them. A person removed from
	// the household they were working in is somewhere else afterwards, and the log
	// saying where is the difference between an entry and an explanation.
	if newDefaultAccountID != "" {
		e.metadata["newDefaultAccountID"] = newDefaultAccountID
	}

	return h.record(ctx, tx, e)
}

// AfterUpdateUserPassword records a password changing.
//
// The hash is not here and never is: platform hands the user with credentials already
// cleared, and an audit entry is the last place a stored credential belongs.
func (h *Hooks) AfterUpdateUserPassword(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	previouslyRequiredChange bool,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.PasswordChangedEventType)
	// Whether this cleared a forced reset, which is the difference between somebody
	// changing their password and somebody finally doing as they were told.
	e.metadata["satisfiedRequiredChange"] = previouslyRequiredChange

	return h.record(ctx, tx, e)
}

// AfterSetUserRequiresPasswordChange records an administrator forcing a reset.
func (h *Hooks) AfterSetUserRequiresPasswordChange(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
) error {
	if user == nil {
		return nil
	}

	return h.record(ctx, tx, userEntry(user.ID, audit.AuditLogEventTypeUpdated,
		ddbidentity.UserPasswordChangeRequiredServiceEventType))
}

// AfterUpdateUserTwoFactorSecret records a new second factor being issued.
//
// The secret is not here, for AfterUpdateUserPassword's reason.
func (h *Hooks) AfterUpdateUserTwoFactorSecret(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	previousSecretVerifiedAt *time.Time,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated, ddbidentity.TwoFactorSecretChangedServiceEventType)
	// Whether this replaced a factor they had proven. Rotating a verified secret leaves
	// somebody without a second factor until they verify the new one, which is a
	// different security event from setting one for the first time.
	e.metadata["replacedVerifiedSecret"] = previousSecretVerifiedAt != nil

	return h.record(ctx, tx, e)
}

// AfterMarkUserTwoFactorSecretVerified records somebody proving the factor they hold.
func (h *Hooks) AfterMarkUserTwoFactorSecretVerified(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
) error {
	if user == nil {
		return nil
	}

	return h.record(ctx, tx, userEntry(user.ID, audit.AuditLogEventTypeUpdated,
		ddbidentity.TwoFactorSecretVerifiedServiceEventType))
}

// AfterSetUserEmailAddressVerificationToken records that a verification mail is owed.
//
// The token is not an argument and would not be recorded if it were: platform keeps it
// off this hook deliberately, and the mail is sent from the outbox row rather than from
// here. An audit entry carrying a live verification token would be a credential in a log.
func (h *Hooks) AfterSetUserEmailAddressVerificationToken(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
	previousAddressVerifiedAt *time.Time,
) error {
	if user == nil {
		return nil
	}

	e := userEntry(user.ID, audit.AuditLogEventTypeUpdated,
		ddbidentity.UserEmailAddressVerificationEmailRequestedEventType)
	e.metadata["hadVerifiedAddress"] = previousAddressVerifiedAt != nil

	return h.record(ctx, tx, e)
}

// AfterMarkUserEmailAddressVerified records an address proven.
func (h *Hooks) AfterMarkUserEmailAddressVerified(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
) error {
	if user == nil {
		return nil
	}

	return h.record(ctx, tx, userEntry(user.ID, audit.AuditLogEventTypeUpdated,
		ddbidentity.UserEmailAddressVerifiedEventType))
}

// AfterMarkUserEmailAddressUnverified records an address that stopped being proven,
// which is what changing it does.
func (h *Hooks) AfterMarkUserEmailAddressUnverified(
	ctx context.Context,
	tx database.Tx,
	_ tenancy.Scope,
	user *platformidentity.User,
) error {
	if user == nil {
		return nil
	}

	return h.record(ctx, tx, userEntry(user.ID, audit.AuditLogEventTypeUpdated,
		ddbidentity.EmailAddressChangedEventType))
}
