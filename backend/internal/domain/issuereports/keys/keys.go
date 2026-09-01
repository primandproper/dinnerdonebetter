package keys

const (
	idSuffix = ".id"

	// IssueReportIDKey is the standard key for referring to an issue report ID.
	IssueReportIDKey = "issue_report" + idSuffix

	// IssueReportStatusKey is the standard key for where a report stands.
	IssueReportStatusKey = "issue_report.status"

	// IssueReportSubjectTypeKey is the standard key for the kind of thing a
	// report is about.
	IssueReportSubjectTypeKey = "issue_report.subject_type"

	// IssueReportSubjectIDKey is the standard key for which one of them.
	IssueReportSubjectIDKey = "issue_report.subject" + idSuffix
)
