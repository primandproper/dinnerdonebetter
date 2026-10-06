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
	"github.com/primandproper/primitives-go/v2/tenancy"
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
// What it costs is part of the operator's console. platform-go lists reports
// across scopes for an operator (ListReportsAcrossScopes), but reads one only in
// the caller's own scope, so the admin app can list another account's report and
// not open it. See platform-go#1149.
//
// tenancy.Of maps the empty account to the zero scope rather than to the global
// one, so a report filed by a session that lost its account is refused by the
// store instead of landing somewhere every tenant can see.
func Scope(accountID string) tenancy.Scope { return tenancy.Of(accountID) }
