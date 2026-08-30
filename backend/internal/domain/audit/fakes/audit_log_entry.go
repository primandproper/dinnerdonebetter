package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeAuditLogEntry builds a faked audit log entry.
func BuildFakeAuditLogEntry() *types.AuditLogEntry {
	entry := fake.BuildFakeRecord[types.AuditLogEntry]()

	// Both of these are closed vocabularies rather than free text: an event type
	// outside the five constants is one no reader knows how to render, and the
	// resource type names a table.
	entry.ResourceType = "example"
	entry.EventType = types.AuditLogEventTypeOther

	return entry
}

// BuildFakeAuditLogEntriesList builds a faked AuditLogEntryList.
func BuildFakeAuditLogEntriesList() *filtering.QueryFilteredResult[types.AuditLogEntry] {
	return fake.BuildFakePage(BuildFakeAuditLogEntry)
}
