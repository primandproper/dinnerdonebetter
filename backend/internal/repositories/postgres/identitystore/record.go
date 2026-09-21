package identitystore

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	identitykeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/identifiers"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// entry is one thing this application wants recorded about an operation.
//
// An operation frequently wants several — registering somebody produces a user, an
// account and a membership — and they are written as a slice rather than by three calls
// so that the hook reads as a statement of what the operation did rather than as three
// statements about rows.
type entry struct {
	_ struct{} `json:"-"`

	// resourceType and relevantID name the row.
	resourceType string
	relevantID   string

	// auditEventType is created, updated or archived.
	auditEventType string

	// changeEventType is what goes on the outbox, and is empty for a row whose change
	// nothing subscribes to. An operation's several entries frequently share one event:
	// a registration is one thing that happened, not three.
	changeEventType string

	// belongsToUser and belongsToAccount file the entry on a chain.
	//
	// Which chain an entry lands on is audit.ScopeFor's decision and not this package's:
	// an entry naming an account goes on the account's chain, and one naming only a user
	// goes on theirs. What this has to get right is naming the right pair, because an
	// entry filed under the wrong account is one the account it is about cannot read.
	belongsToUser    string
	belongsToAccount string

	// metadata is what the data change event carries.
	metadata map[string]any
}

// record writes each entry's audit row and event inside the operation's transaction.
//
// One failure fails the operation. That is the property the whole package exists for:
// an identity write whose audit entry was refused is a write that happened and that the
// log does not know about, and platform calls these hooks inside the transaction
// precisely so the answer to that can be "then it did not happen".
func (h *Hooks) record(ctx context.Context, tx database.Tx, entries ...*entry) error {
	ctx, span := h.tracer.StartSpan(ctx)
	defer span.End()

	for _, e := range entries {
		if e == nil || e.relevantID == "" {
			// A hook handed a nil row. platform's service does not do this, but Hooks is
			// an interface and an entry naming nothing is worse than no entry.
			continue
		}

		logger := h.logger.WithSpan(span).WithValue(identitykeys.UserIDKey, e.belongsToUser)

		tracing.AttachToSpan(span, e.resourceType, e.relevantID)

		auditEntry := &audit.AuditLogEntry{
			ID:            identifiers.New(),
			ResourceType:  e.resourceType,
			RelevantID:    e.relevantID,
			EventType:     e.auditEventType,
			BelongsToUser: e.belongsToUser,
		}

		if e.belongsToAccount != "" {
			auditEntry.BelongsToAccount = &e.belongsToAccount
		}

		metadata := e.metadata
		if metadata == nil {
			metadata = map[string]any{}
		}

		if err := h.recorder.RecordAndEmit(ctx, tx, logger, auditEntry,
			e.changeEventType, e.belongsToAccount, metadata); err != nil {
			return err
		}
	}

	return nil
}

// userEntry names a row that is about one person and no account.
func userEntry(userID, auditEventType, changeEventType string) *entry {
	return &entry{
		resourceType:    resourceTypeUsers,
		relevantID:      userID,
		auditEventType:  auditEventType,
		changeEventType: changeEventType,
		belongsToUser:   userID,
		metadata:        map[string]any{identitykeys.UserIDKey: userID},
	}
}

// accountEntry names a row that is about an account, filed on that account's chain.
func accountEntry(accountID, ownerUserID, auditEventType, changeEventType string) *entry {
	return &entry{
		resourceType:     resourceTypeAccounts,
		relevantID:       accountID,
		auditEventType:   auditEventType,
		changeEventType:  changeEventType,
		belongsToUser:    ownerUserID,
		belongsToAccount: accountID,
		metadata: map[string]any{
			identitykeys.AccountIDKey: accountID,
			identitykeys.UserIDKey:    ownerUserID,
		},
	}
}

// membershipEntry names somebody's place in an account.
func membershipEntry(membershipID, userID, accountID, auditEventType, changeEventType string) *entry {
	return &entry{
		resourceType:     resourceTypeAccountUserMemberships,
		relevantID:       membershipID,
		auditEventType:   auditEventType,
		changeEventType:  changeEventType,
		belongsToUser:    userID,
		belongsToAccount: accountID,
		metadata: map[string]any{
			identitykeys.AccountIDKey: accountID,
			identitykeys.UserIDKey:    userID,
		},
	}
}

// invitationEntry names an invitation, filed on the inviting account's chain.
//
// The account rather than the invitee, because an invitation is the account's act and
// the invitee frequently has no user row yet — an invitation to an address nobody has
// registered is the ordinary case, and an entry filed under a user who does not exist is
// an entry nobody can read.
func invitationEntry(invitationID, accountID, auditEventType, changeEventType string) *entry {
	return &entry{
		resourceType:     resourceTypeAccountInvitations,
		relevantID:       invitationID,
		auditEventType:   auditEventType,
		changeEventType:  changeEventType,
		belongsToAccount: accountID,
		metadata: map[string]any{
			identitykeys.AccountInvitationIDKey: invitationID,
			identitykeys.AccountIDKey:           accountID,
		},
	}
}

// withToken puts an invitation's secret on the event.
//
// Only AfterInvite uses it, and only because the mail that event triggers is a link: the
// store keeps a digest, so a handler that read the invitation back would compose a link
// with an empty token in it and nothing would report the difference. The secret travels no
// further than the outbox row and the mail it is rendered into — an audit entry carries the
// invitation's id, which is what an operator reading the log is asking about.
func (e *entry) WithToken(token string) *entry {
	if token != "" {
		e.metadata[identitykeys.AccountInvitationTokenKey] = token
	}

	return e
}
