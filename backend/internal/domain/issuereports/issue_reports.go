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

// Scope is the tenancy every issue report is filed under, which is the global
// one.
//
// A report is between the person who filed it and the service's administrators —
// a bug, a complaint about another user, a creation somebody thinks is low
// quality — and it is not the business of whoever administers the reporter's
// household. Filing it under the reporter's account put it in that household's
// queue, where its admins could page, revise, resolve and archive it, a complaint
// about a fellow member included.
//
// One scope makes one queue, so what confines a caller is the policy rather than
// the tenancy: filing and reading one's own reports are every user's, the queue a
// service administrator's (see internal/authorization), and a report's reporter
// is the only other caller its reads admit (internal/build/issuereports).
func Scope() tenancy.Scope { return tenancy.Global() }
