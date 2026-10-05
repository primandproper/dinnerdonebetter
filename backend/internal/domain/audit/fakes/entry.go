package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	"github.com/primandproper/primitives-go/v2/identifiers"
)

// BuildFakeEntry builds an audit entry the way a writer would, attributed to a user
// and filed under an account.
//
// Through NewEntry rather than a random record: the actor and the scope are decided by this
// application's rule, and an entry with a generated scope is one nothing reads.
func BuildFakeEntry() *platformaudit.Entry {
	return audit.NewEntry(identifiers.New(), identifiers.New(), "example", identifiers.New(), platformaudit.EventOther)
}
