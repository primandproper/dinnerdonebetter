package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// buildFakeIssueType returns one of the types an issue report may have.
//
// Both this and buildFakeRelevantTable name closed sets: the type is one of five the
// triage path branches on, and the table is one a report can actually point at.
func buildFakeIssueType() string {
	return gofakeit.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})
}

func buildFakeRelevantTable() string {
	return gofakeit.RandomString([]string{"users", "accounts"})
}

// BuildFakeIssueReport builds a fake issue report.
func BuildFakeIssueReport() *types.IssueReport {
	report := fake.BuildFakeRecord[types.IssueReport]()
	report.IssueType = buildFakeIssueType()
	report.RelevantTable = buildFakeRelevantTable()

	return report
}

// BuildFakeIssueReportCreationRequestInput builds a fake IssueReportCreationRequestInput.
func BuildFakeIssueReportCreationRequestInput() *types.IssueReportCreationRequestInput {
	input := fake.BuildFakeRecord[types.IssueReportCreationRequestInput]()
	input.IssueType = buildFakeIssueType()
	input.RelevantTable = buildFakeRelevantTable()

	return input
}

// BuildFakeIssueReportDatabaseCreationInput builds a fake IssueReportDatabaseCreationInput.
func BuildFakeIssueReportDatabaseCreationInput() *types.IssueReportDatabaseCreationInput {
	input := fake.BuildFakeRecord[types.IssueReportDatabaseCreationInput]()
	input.IssueType = buildFakeIssueType()
	input.RelevantTable = buildFakeRelevantTable()

	return input
}

// BuildFakeIssueReportUpdateRequestInput builds a fake IssueReportUpdateRequestInput.
//
// Every field on an update input is optional, and BuildFakeRecord leaves an optional
// field absent — which for this type would be an update that updates nothing. So the
// fields are filled here, from the value builders above.
func BuildFakeIssueReportUpdateRequestInput() *types.IssueReportUpdateRequestInput {
	return &types.IssueReportUpdateRequestInput{
		IssueType:        pointer.To(buildFakeIssueType()),
		Details:          pointer.To(fake.BuildFakeString()),
		RelevantTable:    pointer.To(buildFakeRelevantTable()),
		RelevantRecordID: pointer.To(fake.BuildFakeID()),
	}
}
