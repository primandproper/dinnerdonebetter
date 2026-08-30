package integration

import (
	"context"
	"testing"

	auditgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/audit"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v13/filtering/filteringpb"

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
func entryMatches(actual *auditgrpc.AuditLogEntry, exp *ExpectedAuditEntry) bool {
	if exp == nil {
		return false
	}
	if exp.EventType != "" && actual.GetEventType() != exp.EventType {
		return false
	}
	if exp.ResourceType != "" && actual.GetResourceType() != exp.ResourceType {
		return false
	}
	if exp.RelevantID != "" && actual.GetRelevantId() != exp.RelevantID {
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
		if c.GetNewValue() != want {
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
	resp, err := c.GetAuditLogEntriesForAccount(ctx, &auditgrpc.GetAuditLogEntriesForAccountRequest{
		AccountId: accountID,
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
	resp, err := c.GetAuditLogEntriesForUser(ctx, &auditgrpc.GetAuditLogEntriesForUserRequest{
		UserId: userID,
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
