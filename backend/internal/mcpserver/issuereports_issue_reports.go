package mcpserver

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v13/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var issueReportSchema = map[string]any{
	"ID":                  stringField("The ID of the issue report"),
	"IssueType":           stringField("The type of issue"),
	"Details":             stringField("Detailed description of the issue"),
	"RelevantTable":       stringField("The database table the issue is related to, if any"),
	"RelevantRecordID":    stringField("The ID of the record the issue is related to, if any"),
	fieldCreatedByUser:    stringField("The ID of the user who created the report"),
	fieldBelongsToAccount: stringField("The ID of the account this report belongs to"),
	fieldCreatedAt:        timestampField("When the report was created"),
	fieldLastUpdatedAt:    timestampField("When the report was last updated"),
	fieldArchivedAt:       timestampField("When the report was archived"),
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

func (h *mcpToolManager) GetIssueReport() mcp.ToolHandlerFor[*GetIssueReportInvocation, *issuereports.IssueReport] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportInvocation) (*mcp.CallToolResult, *issuereports.IssueReport, error) {
		result, err := h.issueReportsRepo.GetIssueReport(ctx, x.IssueReportID)
		if err != nil {
			return nil, nil, err
		}
		return nil, result, nil
	}
}

var getIssueReportsTool = &mcp.Tool{
	Name:        "GetIssueReports",
	Description: "Get issue reports with optional filtering",
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
		Results []*issuereports.IssueReport
	}
)

func (h *mcpToolManager) GetIssueReports() mcp.ToolHandlerFor[*GetIssueReportsInvocation, *GetIssueReportsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportsInvocation) (*mcp.CallToolResult, *GetIssueReportsResult, error) {
		results, err := h.issueReportsRepo.GetIssueReports(ctx, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetIssueReportsResult{Results: results.Data}, nil
	}
}

var getIssueReportsForAccountTool = &mcp.Tool{
	Name:        "GetIssueReportsForAccount",
	Description: "Get issue reports belonging to a specific account",
	InputSchema: schemaObject(map[string]any{
		fieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: schemaObject(map[string]any{
		fieldResults: arrayType(schemaObject(issueReportSchema)),
	}),
}

type (
	GetIssueReportsForAccountInvocation struct {
		Filter *filtering.QueryFilter
	}

	GetIssueReportsForAccountResult struct {
		Results []*issuereports.IssueReport
	}
)

func (h *mcpToolManager) GetIssueReportsForAccount() mcp.ToolHandlerFor[*GetIssueReportsForAccountInvocation, *GetIssueReportsForAccountResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportsForAccountInvocation) (*mcp.CallToolResult, *GetIssueReportsForAccountResult, error) {
		accountID, err := h.userFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		results, err := h.issueReportsRepo.GetIssueReportsForAccount(ctx, accountID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetIssueReportsForAccountResult{Results: results.Data}, nil
	}
}
