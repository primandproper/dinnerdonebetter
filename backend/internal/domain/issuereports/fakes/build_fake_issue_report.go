package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeIssueReportKind returns one of the categories this application files
// reports under.
//
// The set is closed and named here rather than randomized, because a kind is
// what a triage queue groups and routes by: a report filed under a category
// nobody watches is a report nobody reads. Platform deliberately does not
// validate it — the catalog is the consumer's — so this is the catalog.
func BuildFakeIssueReportKind() string {
	return gofakeit.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})
}

// buildFakeSubjectType returns a kind of thing a report can actually be about.
func buildFakeSubjectType() string {
	return gofakeit.RandomString([]string{"users", "accounts", "recipes", "meals"})
}

// BuildFakeIssueReport builds a faked Report, open and filed under the scope
// every report is.
//
// The status is fixed rather than randomized because a report is born open and
// the store refuses one that arrives in any other status — a randomized status
// would build a report that could never be written.
func BuildFakeIssueReport() *platformissuereports.Report {
	report := fake.BuildFakeRecord[platformissuereports.Report]()
	report.Kind = BuildFakeIssueReportKind()
	report.SubjectType = buildFakeSubjectType()
	report.Status = platformissuereports.StatusOpen
	report.Resolution = ""
	report.ClosedAt = nil
	report.Scope = issuereports.Scope()

	return report
}

// BuildFakeIssueReportList builds a faked page of Reports.
func BuildFakeIssueReportList() *filtering.QueryFilteredResult[platformissuereports.Report] {
	return fake.BuildFakePage(BuildFakeIssueReport)
}
