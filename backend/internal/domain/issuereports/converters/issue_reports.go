package converters

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v9/identifiers"
)

// ConvertIssueReportToIssueReportUpdateRequestInput creates an IssueReportUpdateRequestInput from an IssueReport.
func ConvertIssueReportToIssueReportUpdateRequestInput(x *types.IssueReport) *types.IssueReportUpdateRequestInput {
	out := &types.IssueReportUpdateRequestInput{
		IssueType:        &x.IssueType,
		Details:          &x.Details,
		RelevantTable:    &x.RelevantTable,
		RelevantRecordID: &x.RelevantRecordID,
	}

	return out
}

// ConvertIssueReportCreationRequestInputToIssueReportDatabaseCreationInput creates an IssueReportDatabaseCreationInput from an IssueReportCreationRequestInput.
func ConvertIssueReportCreationRequestInputToIssueReportDatabaseCreationInput(x *types.IssueReportCreationRequestInput, userID, accountID string) *types.IssueReportDatabaseCreationInput {
	out := &types.IssueReportDatabaseCreationInput{
		ID:               identifiers.New(),
		IssueType:        x.IssueType,
		Details:          x.Details,
		RelevantTable:    x.RelevantTable,
		RelevantRecordID: x.RelevantRecordID,
		CreatedByUser:    userID,
		BelongsToAccount: accountID,
	}

	return out
}

// ConvertIssueReportToIssueReportCreationRequestInput builds an IssueReportCreationRequestInput from an IssueReport.
func ConvertIssueReportToIssueReportCreationRequestInput(x *types.IssueReport) *types.IssueReportCreationRequestInput {
	return &types.IssueReportCreationRequestInput{
		IssueType:        x.IssueType,
		Details:          x.Details,
		RelevantTable:    x.RelevantTable,
		RelevantRecordID: x.RelevantRecordID,
	}
}

// ConvertIssueReportToIssueReportDatabaseCreationInput builds an IssueReportDatabaseCreationInput from an IssueReport.
func ConvertIssueReportToIssueReportDatabaseCreationInput(x *types.IssueReport) *types.IssueReportDatabaseCreationInput {
	return &types.IssueReportDatabaseCreationInput{
		ID:               x.ID,
		IssueType:        x.IssueType,
		Details:          x.Details,
		RelevantTable:    x.RelevantTable,
		RelevantRecordID: x.RelevantRecordID,
		CreatedByUser:    x.CreatedByUser,
		BelongsToAccount: x.BelongsToAccount,
	}
}
