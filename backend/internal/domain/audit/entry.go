package audit

import (
	"context"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// Repository writes the audit log. Reading it is platform's audit.Reader, which the gRPC
// surface and the privacy collector take directly.
type Repository interface {
	// Record appends entries to the log inside the caller's transaction, so an entry commits
	// with the change it describes or not at all. Each entry's assigned ID, timestamp and
	// chain fields are written into the value it was passed.
	//
	// It is variadic because a transaction touching three resources should pay one
	// chain-head lookup and one INSERT rather than three of each. Prefer one call with three
	// entries to three calls with one.
	//
	// Every entry has to carry a scope, and NewEntry is how one gets it: an entry with the
	// zero scope is refused rather than filed somewhere nobody reads.
	Record(ctx context.Context, querier database.Tx, entries ...*platformaudit.Entry) error
}

// NewEntry builds an audit entry under this application's attribution rule.
//
// platform's Entry speaks of an Actor and a Scope, which are deliberately general: tenancy
// depth is an application's decision. Ours is two-level — an entry belongs to a user, and
// usually to an account — so a writer names those two and this decides the rest. It is the
// one place the rule is applied, which is the point: an entry that landed in the wrong chain
// would be invisible to the account read path, and a rule restated at every call site is a
// rule applied inconsistently.
//
// userID is who did it, and may be empty when the writer has no requester in hand; the entry
// is then attributed to UnattributedActorID. accountID is the account it happened in, or
// empty when it happened in none. See ScopeFor for the chain each lands in.
func NewEntry(userID, accountID, resourceType, resourceID string, eventType platformaudit.EventType) *platformaudit.Entry {
	actor := platformaudit.Actor{ID: userID, Type: platformaudit.ActorUser}
	if userID == "" {
		actor = platformaudit.Actor{ID: UnattributedActorID, Type: platformaudit.ActorSystem}
	}

	return &platformaudit.Entry{
		Actor:        actor,
		Scope:        ScopeFor(accountID, userID),
		ResourceType: resourceType,
		ResourceID:   resourceID,
		EventType:    eventType,
	}
}

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
// The global scope takes events belonging to neither, which are platform-level
// by definition and rare enough to serialize. It is tenancy.Global() rather than
// the zero Scope, and the two are not the same thing under platform-go v14: the
// zero value means nobody said, and every read refuses it. A deliberate
// platform-level event has to say so.
func ScopeFor(accountID, userID string) tenancy.Scope {
	switch {
	case accountID != "":
		return tenancy.Of(accountID)
	case userID != "":
		return tenancy.Of(userID)
	default:
		return tenancy.Global()
	}
}
