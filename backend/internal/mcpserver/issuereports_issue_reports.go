package mcpserver

import (
	"context"

	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v13/filtering"
	issuereports "github.com/primandproper/platform-go/v13/issuereports"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var issueReportSchema = map[string]any{
	"ID":               stringField("The ID of the issue report"),
	"Reporter":         stringField("The ID of the user who filed the report"),
	"Kind":             stringField("The category the report was filed under"),
	"Details":          stringField("What the reporter actually said"),
	"SubjectType":      stringField("The kind of thing the report is about, if any"),
	"SubjectID":        stringField("The ID of the thing the report is about, if any"),
	"Status":           stringField("Where the report stands: open, acknowledged, resolved or declined"),
	"Resolution":       stringField("Why the report is in the terminal status it is in, if it is in one"),
	fieldCreatedAt:     timestampField("When the report was filed"),
	fieldLastUpdatedAt: timestampField("When the report was last updated"),
	fieldArchivedAt:    timestampField("When the report was archived"),
	"ClosedAt":         timestampField("When the report reached a terminal status, if it has"),
}

var getIssueReportTool = &mcp.Tool{
	Name:        "GetIssueReport",
	Description: "Get an issue report by its ID",
	InputSchema: schemaObject(map[string]any{
		"IssueReportID": stringField("The ID of the issue report to get"),
	}),
	OutputSchema: schemaObject(issueReportSchema),
}

type GetIssueReportInvocation struct {
	IssueReportID string `jsonschema:"description=The issue report ID"`
}

// GetIssueReport reads one of the caller's account's reports.
//
// The account comes off the token rather than from the invocation, which is what
// keeps a model from asking for a report in somebody else's account: a report
// outside the scope reads as absent.
func (h *mcpToolManager) GetIssueReport() mcp.ToolHandlerFor[*GetIssueReportInvocation, *issuereports.Report] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportInvocation) (*mcp.CallToolResult, *issuereports.Report, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		result, err := h.issueReports.GetReport(ctx, ddbissuereports.Scope(accountID), x.IssueReportID)
		if err != nil {
			return nil, nil, err
		}

		return nil, result, nil
	}
}

var getIssueReportsTool = &mcp.Tool{
	Name:        "GetIssueReports",
	Description: "Get the active account's issue reports, with optional filtering",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(issueReportSchema)),
	}),
}

type (
	GetIssueReportsInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetIssueReportsResult struct {
		Results []*issuereports.Report
	}
)

func (h *mcpToolManager) GetIssueReports() mcp.ToolHandlerFor[*GetIssueReportsInvocation, *GetIssueReportsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportsInvocation) (*mcp.CallToolResult, *GetIssueReportsResult, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		results, err := h.issueReports.ListReports(ctx, ddbissuereports.Scope(accountID), x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetIssueReportsResult{Results: results.Data}, nil
	}
}

var getIssueReportsByStatusTool = &mcp.Tool{
	Name:        "GetIssueReportsByStatus",
	Description: "Get the active account's issue reports in one triage status: open, acknowledged, resolved or declined",
	InputSchema: schemaObject(map[string]any{
		"Status":    stringField("One of open, acknowledged, resolved or declined"),
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(issueReportSchema)),
	}),
}

type (
	GetIssueReportsByStatusInvocation struct {
		Filter *filtering.QueryFilter
		Status string `jsonschema:"description=One of open, acknowledged, resolved or declined"`
	}

	GetIssueReportsByStatusResult struct {
		Results []*issuereports.Report
	}
)

// GetIssueReportsByStatus is the triage queue.
//
// A status this application does not serve is refused rather than answered with
// an empty page, because an empty page is indistinguishable from a queue nobody
// has filed into — and a model that asked for "closed" would report that
// everything had been dealt with.
func (h *mcpToolManager) GetIssueReportsByStatus() mcp.ToolHandlerFor[*GetIssueReportsByStatusInvocation, *GetIssueReportsByStatusResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportsByStatusInvocation) (*mcp.CallToolResult, *GetIssueReportsByStatusResult, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		status, ok := issuereports.ParseStatus(x.Status)
		if !ok {
			return nil, nil, issuereports.ErrUnknownStatus
		}

		results, err := h.issueReports.ListReportsByStatus(ctx, ddbissuereports.Scope(accountID), status, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetIssueReportsByStatusResult{Results: results.Data}, nil
	}
}
