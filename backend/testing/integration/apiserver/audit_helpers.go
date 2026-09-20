package integration

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"
	auditgrpc "github.com/primandproper/platform-go/v14/audit/auditpb"

	"github.com/primandproper/primitives-go/v2/filtering/filteringpb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ExpectedAuditEntry describes fuzzy match criteria for an audit log entry.
type ExpectedAuditEntry struct {
	ChangesEquals  map[string]string
	EventType      string
	ResourceType   string
	RelevantID     string
	ChangesHasKeys []string
}

// entryMatches returns true if the actual proto entry matches all non-empty expected criteria.
func entryMatches(actual *auditgrpc.Entry, exp *ExpectedAuditEntry) bool {
	if exp == nil {
		return false
	}
	if exp.EventType != "" && actual.GetEventType() != exp.EventType {
		return false
	}
	if exp.ResourceType != "" && actual.GetResourceType() != exp.ResourceType {
		return false
	}
	if exp.RelevantID != "" && actual.GetResourceId() != exp.RelevantID {
		return false
	}
	for _, k := range exp.ChangesHasKeys {
		c, ok := actual.GetChanges()[k]
		if !ok || c == nil {
			return false
		}
	}
	for k, want := range exp.ChangesEquals {
		c, ok := actual.GetChanges()[k]
		if !ok || c == nil {
			return false
		}
		// platform types a change's values as structpb.Value rather than string,
		// which is what lets a numeric change read as a number rather than as its
		// rendering. Every expectation here is a string, so this reads the string
		// arm; a non-string expectation would want its own.
		if c.GetNewValue().GetStringValue() != want {
			return false
		}
	}
	return true
}

// AssertAuditLogContainsFuzzy fetches up to limit audit log entries for the account via the gRPC API
// and asserts that each expected entry has at least one matching actual entry in that window.
func AssertAuditLogContainsFuzzy(t *testing.T, ctx context.Context, c client.Client, accountID string, limit int, expected []*ExpectedAuditEntry) {
	t.Helper()

	limit32 := uint32(limit)
	// The account is the scope the connection resolves, so it is not a request
	// field any more — platform binds the query's scope off the principal and the
	// proto argues at length why there can be no scope field. accountID is kept
	// in the signature because every caller has it and the assertion message
	// names it.
	_ = accountID

	resp, err := c.ListEntries(ctx, &auditgrpc.ListEntriesRequest{
		Filter: &filteringpb.QueryFilter{
			MaxResponseSize: &limit32,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	entries := resp.GetResults()
	for _, exp := range expected {
		var found bool
		for _, e := range entries {
			if entryMatches(e, exp) {
				found = true
				break
			}
		}
		assert.True(t, found,
			"expected audit log entry with EventType=%q ResourceType=%q RelevantID=%q within %d entries",
			exp.EventType, exp.ResourceType, exp.RelevantID, limit)
	}
}

// AssertAuditLogContainsFuzzyForUser fetches up to limit audit log entries for the user via the gRPC API
// and asserts that each expected entry has at least one matching actual entry in that window.
func AssertAuditLogContainsFuzzyForUser(t *testing.T, ctx context.Context, c client.Client, userID string, limit int, expected []*ExpectedAuditEntry) {
	t.Helper()

	limit32 := uint32(limit)
	// By actor, within the scope the connection resolved. That is narrower than
	// the RPC this replaced, which filtered on actor across every chain the user
	// appeared in; platform's surface reads one scope per request. See
	// internal/build/auditlog for why, and why the cross-chain view is the
	// privacy export rather than an API read.
	resp, err := c.ListEntries(ctx, &auditgrpc.ListEntriesRequest{
		Query: &auditgrpc.EntryQuery{ActorId: userID},
		Filter: &filteringpb.QueryFilter{
			MaxResponseSize: &limit32,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, resp)

	entries := resp.GetResults()
	for _, exp := range expected {
		var found bool
		for _, e := range entries {
			if entryMatches(e, exp) {
				found = true
				break
			}
		}
		assert.True(t, found,
			"expected audit log entry with EventType=%q ResourceType=%q RelevantID=%q within %d entries",
			exp.EventType, exp.ResourceType, exp.RelevantID, limit)
	}
}
