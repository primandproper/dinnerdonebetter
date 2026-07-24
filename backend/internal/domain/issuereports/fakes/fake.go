package fakes

import (
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v6/identifiers"

	fake "github.com/brianvoe/gofakeit/v7"
)

func init() {
	if err := fake.Seed(time.Now().UnixNano()); err != nil {
		panic(err)
	}
}

// BuildFakeID builds a fake ID.
func BuildFakeID() string {
	return identifiers.New()
}

// BuildFakeTime builds a fake time.
func BuildFakeTime() time.Time {
	return fake.Date().Add(0).Truncate(time.Second).UTC()
}

// BuildFakeIssueReport builds a fake issue report.
func BuildFakeIssueReport() *types.IssueReport {
	return &types.IssueReport{
		ID:               BuildFakeID(),
		IssueType:        fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"}),
		Details:          fake.Sentence(20),
		RelevantTable:    fake.RandomString([]string{"users", "accounts"}),
		RelevantRecordID: BuildFakeID(),
		CreatedAt:        BuildFakeTime(),
		LastUpdatedAt:    nil,
		ArchivedAt:       nil,
		CreatedByUser:    BuildFakeID(),
		BelongsToAccount: BuildFakeID(),
	}
}

// BuildFakeIssueReportCreationRequestInput builds a fake IssueReportCreationRequestInput.
func BuildFakeIssueReportCreationRequestInput() *types.IssueReportCreationRequestInput {
	return &types.IssueReportCreationRequestInput{
		IssueType:        fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"}),
		Details:          fake.Sentence(20),
		RelevantTable:    fake.RandomString([]string{"users", "accounts"}),
		RelevantRecordID: BuildFakeID(),
	}
}

// BuildFakeIssueReportDatabaseCreationInput builds a fake IssueReportDatabaseCreationInput.
func BuildFakeIssueReportDatabaseCreationInput() *types.IssueReportDatabaseCreationInput {
	return &types.IssueReportDatabaseCreationInput{
		ID:               BuildFakeID(),
		IssueType:        fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"}),
		Details:          fake.Sentence(20),
		RelevantTable:    fake.RandomString([]string{"users", "accounts"}),
		RelevantRecordID: BuildFakeID(),
		CreatedByUser:    BuildFakeID(),
		BelongsToAccount: BuildFakeID(),
	}
}

// BuildFakeIssueReportUpdateRequestInput builds a fake IssueReportUpdateRequestInput.
func BuildFakeIssueReportUpdateRequestInput() *types.IssueReportUpdateRequestInput {
	issueType := fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})
	details := fake.Sentence(20)
	relevantTable := fake.RandomString([]string{"users", "accounts"})
	relevantRecordID := BuildFakeID()

	return &types.IssueReportUpdateRequestInput{
		IssueType:        &issueType,
		Details:          &details,
		RelevantTable:    &relevantTable,
		RelevantRecordID: &relevantRecordID,
	}
}
