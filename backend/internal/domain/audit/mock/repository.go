package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ audit.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// GetAuditLogEntry is a mock function.
func (m *Repository) GetAuditLogEntry(ctx context.Context, auditLogID string) (*audit.AuditLogEntry, error) {
	args := m.Called(ctx, auditLogID)
	return args.Get(0).(*audit.AuditLogEntry), args.Error(1)
}

// GetAuditLogEntriesForUser is a mock function.
func (m *Repository) GetAuditLogEntriesForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[audit.AuditLogEntry]), args.Error(1)
}

// GetAuditLogEntriesForUserAndResourceTypes is a mock function.
func (m *Repository) GetAuditLogEntriesForUserAndResourceTypes(ctx context.Context, userID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	args := m.Called(ctx, userID, resourceTypes, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[audit.AuditLogEntry]), args.Error(1)
}

// GetAuditLogEntriesForAccount is a mock function.
func (m *Repository) GetAuditLogEntriesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	args := m.Called(ctx, accountID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[audit.AuditLogEntry]), args.Error(1)
}

// GetAuditLogEntriesForAccountAndResourceTypes is a mock function.
func (m *Repository) GetAuditLogEntriesForAccountAndResourceTypes(ctx context.Context, accountID string, resourceTypes []string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
	args := m.Called(ctx, accountID, resourceTypes, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[audit.AuditLogEntry]), args.Error(1)
}

// CreateAuditLogEntry is a mock function.
func (m *Repository) CreateAuditLogEntry(ctx context.Context, querier database.SQLQueryExecutor, input *audit.AuditLogEntryDatabaseCreationInput) (*audit.AuditLogEntry, error) {
	args := m.Called(ctx, querier, input)
	return args.Get(0).(*audit.AuditLogEntry), args.Error(1)
}
