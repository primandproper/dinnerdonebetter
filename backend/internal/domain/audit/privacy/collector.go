/*
Package privacy is the audit log's contribution to a subject access request.

The audit log is the section of an export most likely to be misread as an
oversight, so it is worth saying why it is here. An audit entry about a person is
personal data — it says what they did and when — and a subject access request
covers it like anything else. What a subject may not do is have it erased, which
is the asymmetry platform-go models by registering collectors and erasers
separately: this domain exports in full and erases only whole chains it can
remove without making the rest of the log unverifiable. See
internal/domain/audit/privacy/eraser.go for that half.
*/
package privacy

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// NewCollector builds the audit log collector: every entry recorded about the
// subject, paged to the end and encoded, or nothing if the log holds none.
// The request scope is not consulted. An entry about the subject is filed under
// whichever account it happened in, and this read filters on the actor rather
// than the chain, so confining it to one scope would drop entries that are still
// about the subject. See internal/domain/audit.ScopeFor.
func NewCollector(repo audit.Repository) platformdataprivacy.Collector {
	return platformdataprivacy.CollectorFor(func(ctx context.Context, _ tenancy.Scope, subject platformdataprivacy.Subject, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[audit.AuditLogEntry], error) {
		return repo.GetAuditLogEntriesForUser(ctx, subject.ID, filter)
	})
}
