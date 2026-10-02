package auditlogentries

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
)

// toPlatformEntry renders a domain entry in the platform's vocabulary.
//
// This is where the application's two-level user/account model becomes the
// platform's actor-and-scope one, and it is the only place it does — see
// audit.ScopeFor for why the scope decision cannot be left to call sites.
func toPlatformEntry(entry *audit.AuditLogEntry) *platformaudit.Entry {
	// An entry with no actor type is a user acting, which is the overwhelming
	// majority of them. Defaulting here rather than at the call sites means the
	// ones that are not — a scheduler, another service — are the ones that have to
	// say so.
	actorType := platformaudit.ActorType(entry.ActorType)
	if actorType == "" {
		actorType = platformaudit.ActorUser
	}

	// The platform refuses an entry with no actor, and the repository methods that
	// take an ID and nothing else have none to give. Naming the absence keeps the
	// write working and keeps the gap countable; see audit.UnattributedActorID.
	actorID := entry.BelongsToUser
	if actorID == "" {
		actorID = audit.UnattributedActorID
		actorType = platformaudit.ActorSystem
	}

	return &platformaudit.Entry{
		RecordedAt: entry.CreatedAt,
		Changes:    entry.Changes,
		ID:         entry.ID,
		Actor: platformaudit.Actor{
			ID:   actorID,
			Type: actorType,
			IP:   entry.ActorIP,
		},
		Scope:        audit.ScopeFor(entry.BelongsToAccount, entry.BelongsToUser),
		ResourceType: entry.ResourceType,
		ResourceID:   entry.RelevantID,
		EventType:    platformaudit.EventType(entry.EventType),
	}
}

// applyRecorded copies back what Record assigned: identity, timestamp, the
// redacted change set, and the chain fields.
func applyRecorded(entry *audit.AuditLogEntry, recorded *platformaudit.Entry) {
	entry.ID = recorded.ID
	entry.CreatedAt = recorded.RecordedAt
	entry.Changes = recorded.Changes
	entry.Scope = recorded.Scope
	entry.Seq = recorded.Seq
	entry.PrevHash = recorded.PrevHash
	entry.Hash = recorded.Hash
}
