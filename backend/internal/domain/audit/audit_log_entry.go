package audit

import (
	"context"
	"time"

	platformaudit "github.com/primandproper/platform-go/v10/audit"
	"github.com/primandproper/platform-go/v10/database"
	"github.com/primandproper/platform-go/v10/filtering"
)

const (
	AuditLogEventTypeOther    = string(platformaudit.EventOther)
	AuditLogEventTypeCreated  = string(platformaudit.EventCreated)
	AuditLogEventTypeUpdated  = string(platformaudit.EventUpdated)
	AuditLogEventTypeArchived = string(platformaudit.EventArchived)
	AuditLogEventTypeDeleted  = string(platformaudit.EventDeleted)
)

type (
	// Change is one field's before and after. It is the platform's type rather than
	// a local one so that a value survives storage as the type it was written as: a
	// numeric field stays numeric, and a caller reading the log back is not left
	// parsing rendered strings.
	Change = platformaudit.Change

	// AuditLogEntry is one record in the audit log, in this application's
	// vocabulary rather than the platform's.
	//
	// The platform speaks of an Actor and a Scope, which are deliberately general:
	// tenancy depth is an application's decision. Ours is two-level and always has
	// been — an entry belongs to a user, and usually to an account — so this type
	// keeps those names and Recorder translates. That translation is the only place
	// Scope is decided, which is the point: an entry that landed in the wrong chain
	// would be invisible to the account read path, and a rule applied at eighty call
	// sites is a rule applied inconsistently.
	//
	// The chain fields below are assigned by Recorder.Record and populated by the
	// Reader. Nothing else should set them.
	AuditLogEntry struct {
		_ struct{} `json:"-"`

		// CreatedAt is when the event happened. Record stamps it when zero.
		CreatedAt time.Time `json:"createdAt"`

		// Changes is the per-field before/after of the event. Diff builds it from a
		// before/after pair; the recorder's redactions filter it on the way in.
		Changes map[string]Change `json:"changes,omitempty"`

		// BelongsToAccount is the account the event happened within, where there was
		// one. Account-scoped events chain per account; see Scope.
		BelongsToAccount *string `json:"belongsToAccount"`

		// ID identifies the entry. Record assigns one when empty.
		ID string `json:"id"`

		// ResourceType names the kind of thing acted on. Required.
		ResourceType string `json:"resourceType"`

		// RelevantID identifies the instance acted on.
		RelevantID string `json:"relevantID"`

		// EventType names what happened. Required — use AuditLogEventTypeOther
		// rather than the empty string for events outside the vocabulary above.
		EventType string `json:"eventType"`

		// BelongsToUser is who did it. Required: a background job that belongs to no
		// user is the job's name with ActorType system, not an absence.
		BelongsToUser string `json:"belongsToUser"`

		// ActorType says what kind of principal acted. Empty means a user.
		ActorType string `json:"actorType"`

		// ActorIP is the address the action arrived from, where there was one. It is
		// recorded rather than derived at read time because the association between
		// a principal and an address is what an investigation needs and is not
		// recoverable afterwards.
		ActorIP string `json:"actorIP"`

		// Scope is the hash chain this entry belongs to. Assigned by Record from
		// BelongsToAccount and BelongsToUser; see ScopeFor.
		Scope string `json:"scope"`

		// PrevHash is the hash of the preceding entry in this scope, or empty for the
		// first. Assigned by Record.
		PrevHash string `json:"prevHash"`

		// Hash is this entry's digest over PrevHash and its own canonical image.
		// Assigned by Record. Publishing it somewhere this database's owner does not
		// control is what raises tamper evidence to tamper proof.
		Hash string `json:"hash"`

		// Seq is the entry's position in its scope's chain, starting at zero.
		// Assigned by Record, and unique per scope in the database — so the chain
		// cannot fork, rather than forking detectably.
		Seq int64 `json:"seq"`
	}

	// AuditLogEntryDataManager describes a structure capable of storing and
	// retrieving audit log entries.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, auditLogID string) (*AuditLogEntry, error)
		GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)

		// Record appends entries to the log inside the caller's transaction, so an
		// entry commits with the change it describes or not at all. It writes each
		// entry's assigned ID, timestamp, and chain fields back into the value it was
		// passed.
		//
		// It is variadic because a transaction touching three resources should pay
		// one chain-head lookup and one INSERT rather than three of each. Prefer one
		// call with three entries to three calls with one.
		Record(ctx context.Context, querier database.SQLQueryExecutor, entries ...*AuditLogEntry) error

		// VerifyChain walks one scope's hash chain over a time range and reports the
		// first break, or that there was none. See ScopeFor for what a scope is here.
		VerifyChain(ctx context.Context, scope string, from, to time.Time) (*VerificationResult, error)
	}
)

// VerificationResult is what a VerifyChain found. It is the platform's type: a
// break is described by position and reason, neither of which this application
// has anything to add to.
type VerificationResult = platformaudit.VerificationResult

// UnattributedActorID is the actor recorded when a write reaches the log with no
// requester in scope.
//
// The platform requires an actor on every entry, and it is right to: an event
// with nobody responsible for it is half a record. Many of this application's
// repository methods genuinely do not have one — ArchiveServiceSetting takes an
// ID and nothing else — and the old schema let belongs_to_user be NULL, so the
// gap predates this package and is not created by it.
//
// Recording it under a name rather than leaving it blank makes the gap a thing
// you can find: "SELECT ... WHERE actor_id = 'unattributed'" is the list of call
// sites still owed a requester. A blank would just look like a bug in the reader.
// Threading the requester through those methods is the fix, and it is a bigger
// change than adopting the log — see docs/audit.md.
const UnattributedActorID = "unattributed"

// TablePrefix namespaces the audit tables, rendering ddb_audit_log_entries and
// ddb_audit_log_chains.
//
// A prefix rather than the platform's empty default, which would render
// audit_log_entries — the name the hand-rolled log this replaces already holds.
// That collision is not hypothetical: the platform's DDL says CREATE TABLE IF NOT
// EXISTS, so against a database that ever applied the old migration the new
// schema is a silent no-op and the audit code then runs against the wrong
// columns. goose records the old migration as applied, so deleting its file does
// not undo it. A prefix makes the two unable to collide, which is worth more than
// a shorter table name and a paragraph in the release notes.
//
// It is declared here, and referenced by the migration that creates the tables,
// the Recorder that writes them, the Reader that queries them, and the Sweeper
// that prunes them. All four have to agree, and a prefix that differs between the
// writer and the reader is the one misconfiguration that stays invisible until
// somebody asks the log a question and gets an empty answer.
const TablePrefix = "ddb"

// ScopeFor resolves the hash chain an entry belongs to.
//
// The chain is partitioned so that unrelated writers do not serialize against
// each other — Record holds a scope's chain row for the length of the caller's
// transaction, so everything sharing a scope shares that lock. An account is the
// natural partition and covers most events.
//
// The events that have no account are the ones that happen before or outside
// one: signup, login, password reset. Filing those under the empty scope would
// be faithful to the platform's model and would also put every login in the
// application behind a single row lock, so they chain per user instead. Reads
// are unaffected, because the account read path filters on scope while the user
// read path filters on the actor.
//
// The empty scope remains for events belonging to neither, which are
// platform-level by definition and rare enough to serialize.
func ScopeFor(belongsToAccount *string, belongsToUser string) string {
	switch {
	case belongsToAccount != nil && *belongsToAccount != "":
		return *belongsToAccount
	default:
		return belongsToUser
	}
}

// Diff reports the fields that differ between two versions of a resource, in the
// shape AuditLogEntry.Changes wants.
//
// Prefer it to a hand-assembled map. Hand assembly is tedious where it is right
// and silently incomplete where it is wrong, and the field somebody forgot to add
// to the map when they added it to the struct is exactly the field an
// investigation will want. Either side may be nil, for a creation or a deletion.
//
// Field names come from the json tag where there is one, so the audit log and the
// API speak the same vocabulary. A field tagged json:"-" or audit:"-" is skipped.
func Diff(before, after any) (map[string]Change, error) {
	return platformaudit.Diff(before, after)
}
