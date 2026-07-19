package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ issuereports.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// GetIssueReport is a mock function.
func (m *Repository) GetIssueReport(ctx context.Context, issueReportID string) (*issuereports.IssueReport, error) {
	args := m.Called(ctx, issueReportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*issuereports.IssueReport), args.Error(1)
}

// GetIssueReports is a mock function.
func (m *Repository) GetIssueReports(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[issuereports.IssueReport]), args.Error(1)
}

// GetIssueReportsForAccount is a mock function.
func (m *Repository) GetIssueReportsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[issuereports.IssueReport]), args.Error(1)
}

// GetIssueReportsForTable is a mock function.
func (m *Repository) GetIssueReportsForTable(ctx context.Context, tableName string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	args := m.Called(ctx, tableName, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[issuereports.IssueReport]), args.Error(1)
}

// GetIssueReportsForRecord is a mock function.
func (m *Repository) GetIssueReportsForRecord(ctx context.Context, tableName, recordID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.IssueReport], error) {
	args := m.Called(ctx, tableName, recordID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[issuereports.IssueReport]), args.Error(1)
}

// CreateIssueReport is a mock function.
func (m *Repository) CreateIssueReport(ctx context.Context, input *issuereports.IssueReportDatabaseCreationInput) (*issuereports.IssueReport, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*issuereports.IssueReport), args.Error(1)
}

// UpdateIssueReport is a mock function.
func (m *Repository) UpdateIssueReport(ctx context.Context, issueReport *issuereports.IssueReport) error {
	args := m.Called(ctx, issueReport)
	return args.Error(0)
}

// ArchiveIssueReport is a mock function.
func (m *Repository) ArchiveIssueReport(ctx context.Context, issueReportID string) error {
	args := m.Called(ctx, issueReportID)
	return args.Error(0)
}
