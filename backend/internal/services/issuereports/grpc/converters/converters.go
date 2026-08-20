package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	"github.com/primandproper/platform-go/v12/identifiers"
)

func ConvertIssueReportToGRPCIssueReport(issueReport *issuereports.IssueReport) *issuereportssvc.IssueReport {
	return &issuereportssvc.IssueReport{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(issueReport.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(issueReport.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(issueReport.LastUpdatedAt),
		Id:               issueReport.ID,
		IssueType:        issueReport.IssueType,
		Details:          issueReport.Details,
		RelevantTable:    issueReport.RelevantTable,
		RelevantRecordId: issueReport.RelevantRecordID,
		CreatedByUser:    issueReport.CreatedByUser,
		BelongsToAccount: issueReport.BelongsToAccount,
	}
}

func ConvertGRPCIssueReportToIssueReport(issueReport *issuereportssvc.IssueReport) *issuereports.IssueReport {
	return &issuereports.IssueReport{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(issueReport.CreatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(issueReport.ArchivedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(issueReport.LastUpdatedAt),
		ID:               issueReport.Id,
		IssueType:        issueReport.IssueType,
		Details:          issueReport.Details,
		RelevantTable:    issueReport.RelevantTable,
		RelevantRecordID: issueReport.RelevantRecordId,
		CreatedByUser:    issueReport.CreatedByUser,
		BelongsToAccount: issueReport.BelongsToAccount,
	}
}

func ConvertGRPCIssueReportCreationRequestInputToIssueReportDatabaseCreationInput(input *issuereportssvc.IssueReportCreationRequestInput, userID, accountID string) *issuereports.IssueReportDatabaseCreationInput {
	return &issuereports.IssueReportDatabaseCreationInput{
		ID:               identifiers.New(),
		IssueType:        input.IssueType,
		Details:          input.Details,
		RelevantTable:    input.RelevantTable,
		RelevantRecordID: input.RelevantRecordId,
		CreatedByUser:    userID,
		BelongsToAccount: accountID,
	}
}

func ConvertIssueReportCreationRequestInputToGRPCIssueReportCreationRequestInput(input *issuereports.IssueReportCreationRequestInput) *issuereportssvc.IssueReportCreationRequestInput {
	return &issuereportssvc.IssueReportCreationRequestInput{
		IssueType:        input.IssueType,
		Details:          input.Details,
		RelevantTable:    input.RelevantTable,
		RelevantRecordId: input.RelevantRecordID,
	}
}

func ConvertGRPCIssueReportUpdateRequestInputToIssueReportUpdateRequestInput(input *issuereportssvc.IssueReportUpdateRequestInput) *issuereports.IssueReportUpdateRequestInput {
	output := &issuereports.IssueReportUpdateRequestInput{}

	if input.IssueType != nil {
		output.IssueType = input.IssueType
	}

	if input.Details != nil {
		output.Details = input.Details
	}

	if input.RelevantTable != nil {
		output.RelevantTable = input.RelevantTable
	}

	if input.RelevantRecordId != nil {
		output.RelevantRecordID = input.RelevantRecordId
	}

	return output
}
