package audit

import (
	"context"
	"time"

	platformaudit "github.com/primandproper/platform-go/v9/audit"
	"github.com/primandproper/platform-go/v9/database"
	"github.com/primandproper/platform-go/v9/filtering"
)

// TablePrefix namespaces the audit tables this application owns, so they read
// ddb_audit_log_entries and ddb_audit_log_chains rather than the platform's
// unprefixed defaults. Every component that touches those tables — the Recorder,
// the Reader, and the Sweeper — takes this same constant, and the migration that
// creates them renders its DDL from it.
const TablePrefix = "ddb"

// SystemActorID is the actor recorded for events with no principal behind them:
// a background job, a scheduler, a webhook from a payment processor. The platform
// requires every entry to name someone responsible, and rightly — an absent actor
// is indistinguishable from a forgotten one — so this is an explicit answer
// rather than an empty string.
const SystemActorID = "system"

type (
	// Entry is one record on its way into the audit log. It is the platform's
	// type: the Recorder assigns its ID, timestamp, and chain fields, and the
	// chain is what makes the log evidence rather than a report.
	Entry = platformaudit.Entry

	// Actor is who did the thing.
	Actor = platformaudit.Actor

	// Change is one field's before and after. Values are typed rather than
	// rendered, so a numeric field stays numeric through storage.
	Change = platformaudit.Change

	// EventType names what happened to a resource.
	EventType = platformaudit.EventType

	// Redaction declares what happens to named fields on the way into the log.
	Redaction = platformaudit.Redaction

	// VerificationResult is what a chain verification found.
	VerificationResult = platformaudit.VerificationResult
)

// The event vocabulary. These name the same events the hand-rolled log did, so
// entries written before and after the cutover read the same way, except that
// EventDeleted is now available for the hard deletions the old enum had no value
// for.
const (
	EventOther    = platformaudit.EventOther
	EventCreated  = platformaudit.EventCreated
	EventUpdated  = platformaudit.EventUpdated
	EventArchived = platformaudit.EventArchived
	EventDeleted  = platformaudit.EventDeleted
)

// Diff reports the fields that differ between two versions of a resource, in the
// shape Entry.Changes wants. Re-exported so a mutation site needs one audit
// import rather than two.
//
// Prefer it to a hand-assembled change map: the field somebody forgets to add to
// the map when they add it to the struct is exactly the field an investigation
// will want.
var Diff = platformaudit.Diff

// UserActor names a human principal, falling back to the system actor when there
// is no user behind the event.
//
// The fallback exists because many of our mutations genuinely have no requester —
// a subscription renewal arriving from Stripe, a scheduled archival — and the
// alternative to naming that is refusing to record it.
func UserActor(userID string) Actor {
	if userID == "" {
		return SystemActor()
	}

	return Actor{ID: userID, Type: platformaudit.ActorUser}
}

// SystemActor names the application acting on its own behalf.
func SystemActor() Actor {
	return Actor{ID: SystemActorID, Type: platformaudit.ActorSystem}
}

// Redactions declares what never reaches the audit tables, keyed by resource
// type. The empty key applies to every resource type.
//
// This is deliberately a reviewed Go declaration rather than a config knob. A
// secret that lands in the audit log is in the one table designed to be immutable
// and retained for years, and filtering it at query time does not un-write it, so
// "which fields must never be recorded" should show up in a diff.
//
// Dropped rather than hashed where even a digest is a liability — a password
// digest is a target, and the audit question about a password is only ever
// "did it change". Hashed where the question is "is this the same value as that
// one", which is the useful question about a rotated key or a token.
var Redactions = map[string]Redaction{
	"": {
		Drop: []string{
			"password",
			"hashed_password",
			"hashedPassword",
			"two_factor_secret",
			"twoFactorSecret",
			"token",
			"client_secret",
			"clientSecret",
			"webhook_hmac_secret",
			"webhookHmacSecret",
			"artifact_encryption_key",
		},
	},
	"users": {
		// A recovery address and a verification token are both reachable from
		// the audit log's read path, which is exposed to the account's members.
		Drop: []string{"avatar_src", "avatarSrc"},
		Hash: []string{"email_address", "emailAddress"},
	},
	"oauth2_clients": {
		Hash: []string{"client_id", "clientID"},
	},
}

type (
	// ChangeLog is one field's before and after, rendered as strings.
	//
	// It is the read and API shape rather than the stored one: the platform
	// stores typed values, and this is what the gRPC surface has always spoken.
	ChangeLog struct {
		OldValue string `json:"oldValue"`
		NewValue string `json:"newValue"`
	}

	// AuditLogEntry is a recorded entry as the API returns it.
	//
	// It is a projection of the platform's Entry, not the stored row: the chain
	// fields (Seq, PrevHash, Hash) are omitted because they mean nothing without
	// the neighbors that give them meaning, and Verify is how the chain is
	// asked a question. Scope is presented as BelongsToAccount and the actor as
	// BelongsToUser, which is what those fields have always been here.
	AuditLogEntry struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time             `json:"createdAt"`
		Changes          map[string]*ChangeLog `json:"changes"`
		BelongsToAccount *string               `json:"belongsToAccount"`
		ID               string                `json:"id"`
		ResourceType     string                `json:"resourceType"`
		RelevantID       string                `json:"relevantID"`
		EventType        string                `json:"eventType"`
		BelongsToUser    string                `json:"belongsToUser"`
	}

	// AuditLogEntryDataManager describes a structure capable of storing and
	// reading audit log entries.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, auditLogID string) (*AuditLogEntry, error)
		GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)

		// Record appends entries to the log inside the caller's transaction, so
		// an entry commits with the change it describes or not at all.
		//
		// It is variadic, and callers should use that: a transaction touching
		// three resources pays one chain-head lookup and one INSERT for three
		// entries rather than three of each, and the chain head is a row every
		// concurrent writer to the same account is queueing behind.
		Record(ctx context.Context, querier database.SQLQueryExecutor, entries ...*Entry) error

		// VerifyChain walks one account's hash chain over a time range and
		// reports the first break, or that there was none. An empty accountID is
		// the scope platform-level events belong to, and is a real chain.
		VerifyChain(ctx context.Context, accountID string, from, to time.Time) (*VerificationResult, error)
	}
)
