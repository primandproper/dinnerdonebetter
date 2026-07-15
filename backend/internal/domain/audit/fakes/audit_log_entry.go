package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v4/filtering"
)

// BuildFakeAuditLogEntry builds a faked valid instrument.
func BuildFakeAuditLogEntry() *types.AuditLogEntry {
	return &types.AuditLogEntry{
		ID:               BuildFakeID(),
		CreatedAt:        BuildFakeTime(),
		Changes:          nil,
		BelongsToAccount: nil,
		ResourceType:     "example",
		RelevantID:       BuildFakeID(),
		EventType:        types.AuditLogEventTypeOther,
		BelongsToUser:    BuildFakeID(),
	}
}

// BuildFakeAuditLogEntriesList builds a faked AuditLogEntryList.
func BuildFakeAuditLogEntriesList() *filtering.QueryFilteredResult[types.AuditLogEntry] {
	var examples []*types.AuditLogEntry
	for range exampleQuantity {
		examples = append(examples, BuildFakeAuditLogEntry())
	}

	return &filtering.QueryFilteredResult[types.AuditLogEntry]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}
