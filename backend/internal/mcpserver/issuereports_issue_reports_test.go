package mcpserver

import (
	"context"
	"testing"

	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	issuereportfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports/fakes"

	issuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportsmock "github.com/primandproper/platform-go/v15/issuereports/mock"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/modelcontextprotocol/go-sdk/auth"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// toolRequestFrom is a tool call carrying a token issued to userID.
func toolRequestFrom(userID string) *mcp.CallToolRequest {
	return &mcp.CallToolRequest{Extra: &mcp.RequestExtra{TokenInfo: &auth.TokenInfo{UserID: userID}}}
}

func TestMCPToolManager_GetIssueReport(T *testing.T) {
	T.Parallel()

	T.Run("answers a report the caller filed, from the global scope", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		report := issuereportfakes.BuildFakeIssueReport()
		report.Reporter = userID

		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(_ context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, reportID string) (*issuereports.Report, error) {
				assert.Equal(t, ddbissuereports.Scope(), scope)
				assert.Equal(t, report.ID, reportID)

				return report, nil
			},
		}
		h := &mcpToolManager{issueReports: store}

		_, result, err := h.GetIssueReport()(t.Context(), toolRequestFrom(userID), &GetIssueReportInvocation{IssueReportID: report.ID})
		require.NoError(t, err)
		assert.Equal(t, report, result)
	})

	T.Run("answers somebody else's report as absent", func(t *testing.T) {
		t.Parallel()

		report := issuereportfakes.BuildFakeIssueReport()
		report.Reporter = fake.BuildFakeID()

		store := &issuereportsmock.StoreMock{
			GetReportFunc: func(context.Context, database.SQLQueryExecutor, tenancy.Scope, string) (*issuereports.Report, error) {
				return report, nil
			},
		}
		h := &mcpToolManager{issueReports: store}

		_, result, err := h.GetIssueReport()(t.Context(), toolRequestFrom(fake.BuildFakeID()), &GetIssueReportInvocation{IssueReportID: report.ID})
		require.ErrorIs(t, err, issuereports.ErrReportNotFound)
		assert.Nil(t, result)
	})

	T.Run("refuses a token naming nobody before reading anything", func(t *testing.T) {
		t.Parallel()

		store := &issuereportsmock.StoreMock{}
		h := &mcpToolManager{issueReports: store}

		_, result, err := h.GetIssueReport()(t.Context(), toolRequestFrom(""), &GetIssueReportInvocation{IssueReportID: fake.BuildFakeID()})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, store.GetReportCalls())
	})
}

func TestMCPToolManager_GetIssueReports(T *testing.T) {
	T.Parallel()

	T.Run("pages the caller's own reports rather than the queue", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		page := issuereportfakes.BuildFakeIssueReportList()
		filter := filtering.DefaultQueryFilter()

		store := &issuereportsmock.StoreMock{
			ListReportsByReporterFunc: func(_ context.Context, _ database.SQLQueryExecutor, scope tenancy.Scope, reporter string, f *filtering.QueryFilter) (*filtering.QueryFilteredResult[issuereports.Report], error) {
				assert.Equal(t, ddbissuereports.Scope(), scope)
				assert.Equal(t, userID, reporter)
				assert.Equal(t, filter, f)

				return page, nil
			},
		}
		h := &mcpToolManager{issueReports: store}

		_, result, err := h.GetIssueReports()(t.Context(), toolRequestFrom(userID), &GetIssueReportsInvocation{Filter: filter})
		require.NoError(t, err)
		assert.Equal(t, page.Data, result.Results)
		assert.Empty(t, store.ListReportsCalls())
	})

	T.Run("refuses a token naming nobody before reading anything", func(t *testing.T) {
		t.Parallel()

		store := &issuereportsmock.StoreMock{}
		h := &mcpToolManager{issueReports: store}

		_, result, err := h.GetIssueReports()(t.Context(), toolRequestFrom(""), &GetIssueReportsInvocation{})
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Empty(t, store.ListReportsByReporterCalls())
	})
}
