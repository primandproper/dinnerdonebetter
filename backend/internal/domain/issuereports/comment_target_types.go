package issuereports

import (
	comments "github.com/primandproper/platform-go/v15/comments"
)

// CommentTargetTypeIssueReports is the kind of thing in this domain that a
// comment may be about. See mealplanning's for why it is a declared type.
const CommentTargetTypeIssueReports comments.TargetType = "issue_reports"

// CommentTargets is this domain's entry in the comment catalog.
//
// It carries no existence check. Reports were unchecked while they were filed per
// account and comments globally, and that reason is gone — every report is filed
// under the global scope now (see Scope), so the hook's scope is the report's and a
// check is possible. It has not been added yet.
func CommentTargets() comments.Targets {
	return comments.Targets{
		CommentTargetTypeIssueReports: {Description: "An issue report."},
	}
}
