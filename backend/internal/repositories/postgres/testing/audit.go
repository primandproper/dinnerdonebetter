package testing

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertAuditLogContains fetches audit log entries for the account and asserts that entries exist
// matching each expected (eventType, resourceType, relevantID).
func AssertAuditLogContains(t *testing.T, ctx context.Context, auditRepo audit.Repository, accountID string, expected []*audit.AuditLogEntry) {
	t.Helper()
	filter := filtering.DefaultQueryFilter()
	entries, err := auditRepo.GetAuditLogEntriesForAccount(ctx, accountID, filter)
	require.NoError(t, err)
	require.NotNil(t, entries)
	for _, exp := range expected {
		var found bool
		for _, e := range entries.Data {
			if e.EventType == exp.EventType && e.ResourceType == exp.ResourceType && e.RelevantID == exp.RelevantID {
				found = true
				break
			}
		}
		assert.True(t, found, "expected audit log entry with EventType=%q ResourceType=%q RelevantID=%q", exp.EventType, exp.ResourceType, exp.RelevantID)
	}
}

// AssertAuditLogContainsForUser fetches audit log entries for the user and asserts that entries exist
// matching each expected (eventType, resourceType, relevantID). Use for user-scoped resources.
func AssertAuditLogContainsForUser(t *testing.T, ctx context.Context, auditRepo audit.Repository, userID string, expected []*audit.AuditLogEntry) {
	t.Helper()
	filter := filtering.DefaultQueryFilter()
	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, userID, filter)
	require.NoError(t, err)
	require.NotNil(t, entries)
	for _, exp := range expected {
		var found bool
		for _, e := range entries.Data {
			if e.EventType == exp.EventType && e.ResourceType == exp.ResourceType && e.RelevantID == exp.RelevantID {
				found = true
				break
			}
		}
		assert.True(t, found, "expected audit log entry with EventType=%q ResourceType=%q RelevantID=%q", exp.EventType, exp.ResourceType, exp.RelevantID)
	}
}
