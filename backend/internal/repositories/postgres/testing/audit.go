package testing

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExpectedAuditEntry is what an assertion below looks for: an entry of this event type, about
// this resource. Everything else on an entry — the actor, the chain, the hash — is the
// recorder's to decide and not what these tests pin.
type ExpectedAuditEntry struct {
	EventType    string
	ResourceType string
	ResourceID   string
}

// auditReader reads the log at this application's prefix, which is the one the recorder writes.
func auditReader(t *testing.T, db database.Client) platformaudit.Reader {
	t.Helper()

	reader, err := platformaudit.NewReader(db.Dialect(), platformaudit.WithReaderTablePrefix(branding.TablePrefix))
	require.NoError(t, err)

	return reader
}

// AuditEntriesForActor reads every entry the actor recorded, in every chain.
func AuditEntriesForActor(t *testing.T, ctx context.Context, db database.Client, actorID string) []*platformaudit.Entry {
	t.Helper()

	page, err := auditReader(t, db).ListAcrossScopes(ctx, db.Reader(), &platformaudit.Query{ActorID: actorID}, filtering.DefaultQueryFilter())
	require.NoError(t, err)

	return page.Data
}

// AuditEntriesForAccount reads the first page of one account's chain.
func AuditEntriesForAccount(t *testing.T, ctx context.Context, db database.Client, accountID string) []*platformaudit.Entry {
	t.Helper()

	page, err := auditReader(t, db).List(ctx, db.Reader(), tenancy.Of(accountID), nil, filtering.DefaultQueryFilter())
	require.NoError(t, err)

	return page.Data
}

// AssertAuditLogContains asserts that the account's chain holds an entry matching each expected one.
func AssertAuditLogContains(t *testing.T, ctx context.Context, db database.Client, accountID string, expected []ExpectedAuditEntry) {
	t.Helper()

	assertContains(t, AuditEntriesForAccount(t, ctx, db, accountID), expected)
}

// AssertAuditLogContainsForUser asserts that the user recorded an entry matching each expected
// one, in whichever chain it was filed.
func AssertAuditLogContainsForUser(t *testing.T, ctx context.Context, db database.Client, userID string, expected []ExpectedAuditEntry) {
	t.Helper()

	assertContains(t, AuditEntriesForActor(t, ctx, db, userID), expected)
}

func assertContains(t *testing.T, entries []*platformaudit.Entry, expected []ExpectedAuditEntry) {
	t.Helper()

	for i := range expected {
		exp := &expected[i]

		var found bool
		for _, e := range entries {
			if string(e.EventType) == exp.EventType && e.ResourceType == exp.ResourceType && e.ResourceID == exp.ResourceID {
				found = true
				break
			}
		}

		assert.True(t, found, "expected audit log entry with EventType=%q ResourceType=%q ResourceID=%q", exp.EventType, exp.ResourceType, exp.ResourceID)
	}
}
