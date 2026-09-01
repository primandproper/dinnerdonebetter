package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/issuereports"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	platformissuereports "github.com/primandproper/platform-go/v13/issuereports"

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

// BuildFakeIssueReport builds a faked Report, open and belonging to a faked
// account.
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
	report.Scope = issuereports.Scope(fake.BuildFakeID())

	return report
}

// BuildFakeIssueReportForScope builds a faked Report filed under the given
// account.
func BuildFakeIssueReportForScope(accountID string) *platformissuereports.Report {
	report := BuildFakeIssueReport()
	report.Scope = issuereports.Scope(accountID)

	return report
}

// BuildFakeIssueReportList builds a faked page of Reports in one scope.
//
// Every element carries the scope, because that is what the read path filtered
// on: a page of one account's reports is the only page the read path returns.
func BuildFakeIssueReportList(scope string) *filtering.QueryFilteredResult[platformissuereports.Report] {
	return fake.BuildFakePage(func() *platformissuereports.Report {
		return BuildFakeIssueReportForScope(scope)
	})
}
