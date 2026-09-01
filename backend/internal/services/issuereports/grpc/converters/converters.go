package converters

import (
	ddbissuereports "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	issuereportssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/issue_reports"

	issuereports "github.com/primandproper/platform-go/v13/issuereports"
)

// ConvertIssueReportToGRPCIssueReport converts a stored report to proto.
//
// The scope is deliberately not on the wire. It is the account the report was
// filed under, every read is already restricted to the caller's account, and a
// client that could see the column would be a client that could tell one account
// from another by looking at a report.
func ConvertIssueReportToGRPCIssueReport(input *issuereports.Report) *issuereportssvc.IssueReport {
	if input == nil {
		return nil
	}

	return &issuereportssvc.IssueReport{
		Id:            input.ID,
		Reporter:      input.Reporter,
		Kind:          input.Kind,
		Details:       input.Details,
		SubjectType:   input.SubjectType,
		SubjectId:     input.SubjectID,
		Status:        input.Status.String(),
		Resolution:    input.Resolution,
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		ClosedAt:      grpcconverters.ConvertTimePointerToPBTimestamp(input.ClosedAt),
	}
}

// ConvertGRPCIssueReportToIssueReport converts a proto report back to the
// platform's, for a client asserting against what it was handed.
func ConvertGRPCIssueReportToIssueReport(input *issuereportssvc.IssueReport) *issuereports.Report {
	if input == nil {
		return nil
	}

	return &issuereports.Report{
		ID:            input.GetId(),
		Reporter:      input.GetReporter(),
		Kind:          input.GetKind(),
		Details:       input.GetDetails(),
		SubjectType:   input.GetSubjectType(),
		SubjectID:     input.GetSubjectId(),
		Status:        issuereports.Status(input.GetStatus()),
		Resolution:    input.GetResolution(),
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.GetCreatedAt()),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.GetLastUpdatedAt()),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.GetArchivedAt()),
		ClosedAt:      grpcconverters.ConvertPBTimestampToTimePointer(input.GetClosedAt()),
	}
}

// ConvertGRPCIssueReportCreationRequestInputToIssueReport builds the report the
// store writes from what the client sent.
//
// The reporter and the scope are parameters rather than fields on the input,
// because both come from the authenticated session: a report that could name its
// own reporter is a report anybody could file as anybody, and one that could name
// its own scope is one anybody could file into another account's queue. The
// status is left unset, because a report is born open and the store refuses one
// that arrives in any other status.
func ConvertGRPCIssueReportCreationRequestInputToIssueReport(input *issuereportssvc.IssueReportCreationRequestInput, reporter, accountID string) *issuereports.Report {
	if input == nil {
		return nil
	}

	return &issuereports.Report{
		Reporter:    reporter,
		Kind:        input.GetKind(),
		Details:     input.GetDetails(),
		SubjectType: input.GetSubjectType(),
		SubjectID:   input.GetSubjectId(),
		Scope:       ddbissuereports.Scope(accountID),
	}
}

// ConvertIssueReportToGRPCIssueReportCreationRequestInput builds the creation
// input a client would have sent to produce this report. It is what the
// integration suite files its fakes with.
func ConvertIssueReportToGRPCIssueReportCreationRequestInput(input *issuereports.Report) *issuereportssvc.IssueReportCreationRequestInput {
	if input == nil {
		return nil
	}

	return &issuereportssvc.IssueReportCreationRequestInput{
		Kind:        input.Kind,
		Details:     input.Details,
		SubjectType: input.SubjectType,
		SubjectId:   input.SubjectID,
	}
}

// ApplyGRPCIssueReportUpdateRequestInput merges an update input into a report
// the caller has already read.
//
// It is a mutation of the read row rather than a conversion into a fresh one,
// because platform's UpdateReport takes a whole Report: a value built from the
// request alone would write empty strings over every field the client did not
// send.
func ApplyGRPCIssueReportUpdateRequestInput(report *issuereports.Report, input *issuereportssvc.IssueReportUpdateRequestInput) {
	if report == nil || input == nil {
		return
	}

	if input.Kind != nil {
		report.Kind = input.GetKind()
	}

	if input.Details != nil {
		report.Details = input.GetDetails()
	}

	if input.SubjectType != nil {
		report.SubjectType = input.GetSubjectType()
	}

	if input.SubjectId != nil {
		report.SubjectID = input.GetSubjectId()
	}
}
