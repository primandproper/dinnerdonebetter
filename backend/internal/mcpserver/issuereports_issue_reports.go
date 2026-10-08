package mcpserver

import (
	"context"
	"fmt"

	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	"github.com/primandproper/dinnerdonebetter/backend/internal/mcptools"

	issuereports "github.com/primandproper/platform-go/v15/issuereports"
	"github.com/primandproper/primitives-go/v2/filtering"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var issueReportSchema = map[string]any{
	"ID":                        mcptools.StringField("The ID of the issue report"),
	"Reporter":                  mcptools.StringField("The ID of the user who filed the report"),
	"Kind":                      mcptools.StringField("The category the report was filed under"),
	"Details":                   mcptools.StringField("What the reporter actually said"),
	"SubjectType":               mcptools.StringField("The kind of thing the report is about, if any"),
	"SubjectID":                 mcptools.StringField("The ID of the thing the report is about, if any"),
	"Status":                    mcptools.StringField("Where the report stands: open, acknowledged, resolved or declined"),
	"Resolution":                mcptools.StringField("Why the report is in the terminal status it is in, if it is in one"),
	mcptools.FieldCreatedAt:     mcptools.TimestampField("When the report was filed"),
	mcptools.FieldLastUpdatedAt: mcptools.TimestampField("When the report was last updated"),
	mcptools.FieldArchivedAt:    mcptools.TimestampField("When the report was archived"),
	"ClosedAt":                  mcptools.TimestampField("When the report reached a terminal status, if it has"),
}

var getIssueReportTool = &mcp.Tool{
	Name:        "GetIssueReport",
	Description: "Get an issue report you filed, by its ID",
	InputSchema: mcptools.SchemaObject(map[string]any{
		"IssueReportID": mcptools.StringField("The ID of the issue report to get"),
	}),
	OutputSchema: mcptools.SchemaObject(issueReportSchema),
}

type GetIssueReportInvocation struct {
	IssueReportID string `jsonschema:"description=The issue report ID"`
}

// GetIssueReport reads one report the caller filed.
//
// Every report is in the one global scope, so the scope keeps nobody out and the
// reporter check is the whole boundary. Somebody else's report reads as absent,
// exactly as one that was never filed does, so the tool is no oracle for which
// reports exist. Working the queue is a service administrator's, and this server
// mounts none of it.
func (h *mcpToolManager) GetIssueReport() mcp.ToolHandlerFor[*GetIssueReportInvocation, *issuereports.Report] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportInvocation) (*mcp.CallToolResult, *issuereports.Report, error) {
		userID, err := h.reporterFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		result, err := h.issueReports.GetReport(ctx, h.reader, ddbissuereports.Scope(), x.IssueReportID)
		if err != nil {
			return nil, nil, err
		}

		if result.Reporter != userID {
			return nil, nil, issuereports.ErrReportNotFound
		}

		return nil, result, nil
	}
}

var getIssueReportsTool = &mcp.Tool{
	Name:        "GetIssueReports",
	Description: "Get the issue reports you filed, with optional filtering",
	InputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldFilter: filtering.QueryFilterSchema(),
	}),
	OutputSchema: mcptools.SchemaObject(map[string]any{
		mcptools.FieldResults: mcptools.ArrayType(mcptools.SchemaObject(issueReportSchema)),
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

// GetIssueReports pages the reports the caller filed, and nobody else's.
func (h *mcpToolManager) GetIssueReports() mcp.ToolHandlerFor[*GetIssueReportsInvocation, *GetIssueReportsResult] {
	return func(ctx context.Context, req *mcp.CallToolRequest, x *GetIssueReportsInvocation) (*mcp.CallToolResult, *GetIssueReportsResult, error) {
		userID, err := h.reporterFromRequest(req)
		if err != nil {
			return nil, nil, err
		}

		results, err := h.issueReports.ListReportsByReporter(ctx, h.reader, ddbissuereports.Scope(), userID, x.Filter)
		if err != nil {
			return nil, nil, err
		}

		return nil, &GetIssueReportsResult{Results: results.Data}, nil
	}
}

// reporterFromRequest is the user the MCP request's token was issued to.
//
// It refuses a token naming nobody rather than answering it, because a reporter
// check against the empty string would match a report filed by nobody.
func (h *mcpToolManager) reporterFromRequest(req *mcp.CallToolRequest) (string, error) {
	if req.Extra == nil || req.Extra.TokenInfo == nil || req.Extra.TokenInfo.UserID == "" {
		return "", fmt.Errorf("not authenticated")
	}

	return req.Extra.TokenInfo.UserID, nil
}
