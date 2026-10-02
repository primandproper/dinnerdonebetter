package auditlogentries

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	auditkeys "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit/keys"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

var (
	_ audit.Repository = (*repository)(nil)
)

// attachImpersonator names the operator on an entry the request's subject is recorded as acting
// in, when the request came through an impersonation token.
//
// The entry stays the subject's — it is their data, and a query for what happened to their
// account has to find it — and Impersonator is the second slot that stops it saying they did it.
// An entry filed under somebody else, or under nobody, is left alone: the operator was not
// acting as them.
func attachImpersonator(ctx context.Context, entry *platformaudit.Entry) {
	session := sessions.FromContext(ctx)
	if session == nil || session.ImpersonatorID == "" {
		return
	}

	if entry.Actor.Type == platformaudit.ActorUser && entry.Actor.ID == session.GetUserID() {
		entry.Actor.Impersonator = session.ImpersonatorID
	}
}

// Record appends audit log entries inside the caller's transaction.
func (q *repository) Record(ctx context.Context, querier database.Tx, entries ...*platformaudit.Entry) error {
	ctx, span := q.tracer.StartSpan(ctx)
	defer span.End()

	logger := q.logger.Clone()

	if len(entries) == 0 {
		return nil
	}

	// The platform Recorder appends to one chain per call, and a chain is
	// identified by scope. A batch reaching this method may span scopes — an
	// account-scoped change and the user-scoped login that authorized it are
	// legitimately recorded together — so the batch is split by scope and each
	// group is appended to its own chain. Order within a scope is preserved,
	// which is the only order the chain defines.
	order := make([]tenancy.Scope, 0, len(entries))
	groups := map[tenancy.Scope][]*platformaudit.Entry{}
	for _, entry := range entries {
		if entry == nil {
			return observability.PrepareAndLogError(platformerrors.ErrNilInputParameter, logger, span, "recording audit log entries")
		}

		// The zero scope means nobody decided which chain this belongs to, which is
		// what audit.NewEntry is for. Every read refuses it, so recording one would
		// write an entry nothing can find.
		if entry.Scope == (tenancy.Scope{}) {
			return observability.PrepareAndLogError(errUnscopedEntry, logger, span, "recording audit log entries")
		}

		attachImpersonator(ctx, entry)

		if _, ok := groups[entry.Scope]; !ok {
			order = append(order, entry.Scope)
		}
		groups[entry.Scope] = append(groups[entry.Scope], entry)
	}

	// Record assigns the ID, timestamp and chain fields, and applies redaction to
	// the changes, in the entries it is handed — so a caller that logs or returns
	// the entry it just wrote describes the row that actually landed.
	for _, scope := range order {
		if err := q.recorder.Record(ctx, querier, scope, groups[scope]...); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "recording audit log entries")
		}
	}

	tracing.AttachToSpan(span, auditkeys.AuditLogEntryIDKey, entries[0].ID)

	return nil
}

// errUnscopedEntry is the refusal of an entry that names no chain.
var errUnscopedEntry = platformerrors.New("audit entry has no scope; build it with audit.NewEntry")
