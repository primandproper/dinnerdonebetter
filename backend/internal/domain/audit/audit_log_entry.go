package audit

import (
	"context"
	"time"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/filtering"
)

const (
	AuditLogEventTypeOther    = "other"
	AuditLogEventTypeCreated  = "created"
	AuditLogEventTypeUpdated  = "updated"
	AuditLogEventTypeArchived = "archived"
	AuditLogEventTypeDeleted  = "deleted"
)

type (
	ChangeLog struct {
		OldValue string `json:"oldValue"`
		NewValue string `json:"newValue"`
	}

	AuditLogEntry struct {
		_ struct{} `json:"-"`

		CreatedAt        time.Time             `json:"createdAt"`
		Changes          map[string]*ChangeLog `json:"changes"`
		BelongsToAccount *string               `json:"belongsToAccount"`
		ID               string                `json:"id"`
		ResourceType     string                `json:"resourceType"`
		RelevantID       string                `json:"relevantID"`
		EventType        string                `json:"eventType"`
		BelongsToUser    string                `json:"belongsToUser"`
	}

	AuditLogEntryDatabaseCreationInput struct {
		_ struct{} `json:"-"`

		Changes          map[string]*ChangeLog `json:"-"`
		BelongsToAccount *string               `json:"-"`
		ID               string                `json:"-"`
		ResourceType     string                `json:"-"`
		RelevantID       string                `json:"-"`
		EventType        string                `json:"-"`
		BelongsToUser    string                `json:"-"`
	}

	// AuditLogEntryDataManager describes a structure capable of storing audit log entries.
	AuditLogEntryDataManager interface {
		GetAuditLogEntry(ctx context.Context, auditLogID string) (*AuditLogEntry, error)
		GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[AuditLogEntry], error)
		CreateAuditLogEntry(ctx context.Context, querier database.SQLQueryExecutor, input *AuditLogEntryDatabaseCreationInput) (*AuditLogEntry, error)
	}
)
