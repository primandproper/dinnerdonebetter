/*
Package issuereports is this application's half of platform-go's issue report
store: the namespace its table carries, the tenancy every report is filed under,
and the data change events a write emits.

The store itself is platform-go's. It owns the schema, the paging, the tenancy
column, the triage lifecycle and the erasure, because that half is the same in
every application. What is not the same is what a report is about and who may
see it, and both of those are decided here.
*/
package issuereports

import (
	"github.com/primandproper/platform-go/v13/tenancy"
)

// TablePrefix namespaces the platform-go issue reports table, rendering
// ddb_issue_reports.
//
// The platform's own default is the empty prefix, which renders
// "issue_reports" — the exact name the table this replaced carried. Its DDL says
// CREATE TABLE IF NOT EXISTS, so a deployment that kept both would get a silent
// no-op followed by a store reading columns that are not there.
const TablePrefix = "ddb"

// The data change events an issue report write emits. They are declared in the
// webhook event catalog (internal/domain/webhooks/catalog), so a subscriber is
// already able to ask for them.
const (
	// IssueReportCreatedServiceEventType indicates an issue report was created.
	IssueReportCreatedServiceEventType = "issue_report_created"
	// IssueReportUpdatedServiceEventType indicates an issue report was revised:
	// the kind, the details, or what it is about.
	IssueReportUpdatedServiceEventType = "issue_report_updated"
	// IssueReportTransitionedServiceEventType indicates an issue report moved
	// through the triage lifecycle — picked up, resolved, declined, reopened.
	//
	// It is distinct from the update event because the two answer different
	// questions and only one of them is a queue. A subscriber watching for
	// "which reports were resolved this week" cannot get that from an event that
	// also fires when somebody fixed a typo in the details.
	IssueReportTransitionedServiceEventType = "issue_report_transitioned"
	// IssueReportArchivedServiceEventType indicates an issue report was archived.
	IssueReportArchivedServiceEventType = "issue_report_archived"
)

// Scope is the tenancy an account's issue reports are filed under.
//
// The account is the tenant, which is the same reading webhooks takes of the
// same column, and it is what replaced the belongs_to_account column the local
// table carried. It is a decision rather than a default: reading an issue report
// is an account member permission, so a deployment that filed every report in
// one scope would let any member of any account read every report anybody had
// ever filed.
//
// What it costs is the operator's console. platform-go's Store deliberately
// offers no cross-scope listing, so "every tenant's open reports in one page" is
// not a call this application can make; an operator lists the accounts they
// administer and pages each. See the Store interface for why the alternative —
// a read that omits the scope — is worse.
//
// tenancy.Of maps the empty account to the zero scope rather than to the global
// one, so a report filed by a session that lost its account is refused by the
// store instead of landing somewhere every tenant can see.
func Scope(accountID string) tenancy.Scope { return tenancy.Of(accountID) }
