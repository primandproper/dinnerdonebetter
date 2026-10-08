package indexing

import (
	"github.com/primandproper/platform-go/v15/identity"
	"github.com/primandproper/platform-go/v15/searchsync"
)

// userIDKey is the field of platform's identity.UserEvent that names the user. It is a JSON
// field name rather than a context key: a payload that does not answer for itself is read by
// its field names (see searchsync's doc).
const userIDKey = "userID"

// IndexRules is identity's entry in the table internal/indexevents registers on the outbox
// writer: which of platform's user events feed the users index.
//
// Both handle doors on sign-in write through the profile operation, so a username or email
// change is a profile update here.
func IndexRules() []searchsync.Rule {
	return []searchsync.Rule{
		{EventType: identity.EventUserRegistered.String(), Topic: IndexTypeUsers, IDKey: userIDKey, Op: searchsync.OpUpsert},
		{EventType: identity.EventUserProfileUpdated.String(), Topic: IndexTypeUsers, IDKey: userIDKey, Op: searchsync.OpUpsert},
		{EventType: identity.EventUserArchived.String(), Topic: IndexTypeUsers, IDKey: userIDKey, Op: searchsync.OpDelete},
	}
}
