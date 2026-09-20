package auditlog

import (
	"context"
	"sort"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// spanningReader reads a caller's two chains as one.
//
// This application files an entry under its account when it has one and under its actor
// otherwise — see audit.ScopeFor, and the row lock that reasoning is about. A signed-in
// caller therefore belongs to two chains at once: their account's, and their own, which is
// where a login, a signup and a password reset land because none of those happens inside an
// account.
//
// platform's read takes one scope. audit.Query.Scope is a *tenancy.Scope with three
// readings — nil narrows nothing, a scope narrows to it, and the zero value is refused —
// and there is no fourth for "these two". So a caller reading over the wire saw their
// account's entries and never their own, which is the half a person asking "what happened
// to my account" most wants.
//
// This reads both and merges. It is not the right place for this to live: the fix upstream
// is one predicate, `scope = ANY($1)` in place of `scope = $1`, and it needs no new cursor
// semantics because the paged list already orders and pages by id rather than by seq. Filed;
// this deletes when it lands.
type spanningReader struct {
	platformaudit.Reader
}

// List reads the resolved scope and the caller's own chain, and merges them.
//
// The merge is exact rather than approximate, and it is the id cursor that makes it so.
// Both sub-reads answer with "the n rows past this cursor, in this scope, in id order", so
// the n smallest of the union are among the 2n returned — taking n after a merge-sort is
// the same page the union would have produced. The cursor to follow it is the last row's
// id, which is what the merged result carries.
//
// What it cannot reproduce is the counts. Each sub-read counts its own chain, and adding
// them over-counts nothing but describes two pages rather than one; the result is built
// without counts instead, so a client reads the cursor rather than a total that would be
// a guess. See filtering.NewQueryFilteredResultWithoutCounts.
func (r spanningReader) List(
	ctx context.Context,
	q database.SQLQueryExecutor,
	query *platformaudit.Query,
	filter *filtering.QueryFilter,
) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
	if across, ok := everyChain(ctx, query); ok {
		return r.Reader.List(ctx, q, across, filter)
	}

	own, ok := r.ownChain(ctx, query)
	if !ok {
		return r.Reader.List(ctx, q, query, filter)
	}

	resolved, err := r.Reader.List(ctx, q, query, filter)
	if err != nil {
		return nil, err
	}

	// A copy, because the query belongs to the handler that built it and a scope written
	// through it would outlive this call.
	mine := *query
	mine.Scope = &own

	actor, err := r.Reader.List(ctx, q, &mine, filter)
	if err != nil {
		return nil, err
	}

	return mergePages(resolved, actor, filter), nil
}

// Get reads one entry from either chain.
//
// Two reads rather than one unscoped read. A nil scope is "every tenant" here, and handing
// that to a caller's read is the cross-tenant disclosure the pointer exists to prevent —
// so each chain is asked for the entry by name, and an entry in neither is not found.
func (r spanningReader) Get(
	ctx context.Context,
	q database.SQLQueryExecutor,
	scope *tenancy.Scope,
	id string,
) (*platformaudit.Entry, error) {
	if isServiceAdmin(ctx) {
		return r.Reader.Get(ctx, q, nil, id)
	}

	entry, err := r.Reader.Get(ctx, q, scope, id)
	if err == nil {
		return entry, nil
	}

	own, ok := r.ownChainScope(ctx)
	if !ok || (scope != nil && *scope == own) {
		return nil, err
	}

	// The first error is dropped rather than joined: both are "not in this chain", and a
	// caller told an entry is in neither wants one answer, not the same one twice.
	return r.Reader.Get(ctx, q, &own, id)
}

// everyChain answers with the operator's read — every chain in the deployment — when the
// caller is a service administrator.
//
// platform's reader already has this: Query.Scope is a *tenancy.Scope in which nil "narrows
// nothing", and its own gRPC surface documents that it never passes nil because it "has no
// operator". This deployment has one, so the decision is made here, where the session is.
//
// It is the same division the two surfaces adopted alongside this one draw. A grant on the
// method says whether this kind of call is allowed at all — ReadAuditLogEntriesPermission is
// an account member's, because a member reading their own account's log is the ordinary
// case — and it cannot say which chains. Which chains is asked here, of the session, after
// the query is built and before any row is read.
//
// Without it a service administrator investigating somebody else's account reads an empty
// window, since ScopeFor files an entry under its account and an admin is not a member of
// the account they are investigating. That is not a narrower answer than the log can give;
// it is the one question an audit log exists to answer going unanswerable over the wire,
// with the privacy export — a subject's own read of their own data — as the only remaining
// path to it.
//
// The query is copied rather than written through, for List's reason: it belongs to the
// handler that built it.
func everyChain(ctx context.Context, query *platformaudit.Query) (*platformaudit.Query, bool) {
	if query == nil || query.Scope == nil || !isServiceAdmin(ctx) {
		return nil, false
	}

	across := *query
	across.Scope = nil

	return &across, true
}

// isServiceAdmin reports whether the session holds the service administrator role.
func isServiceAdmin(ctx context.Context) bool {
	return sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin()
}

// ownChain answers with the caller's own chain when it is worth a second read.
func (r spanningReader) ownChain(ctx context.Context, query *platformaudit.Query) (tenancy.Scope, bool) {
	own, ok := r.ownChainScope(ctx)
	switch {
	case !ok:
		return tenancy.Scope{}, false
	// Nil already spans every chain, so there is nothing to add and adding it would narrow.
	case query == nil || query.Scope == nil:
		return tenancy.Scope{}, false
	// The resolved scope is already the caller's own, which is what an account-less session
	// reads. A second identical read would double every row.
	case *query.Scope == own:
		return tenancy.Scope{}, false
	default:
		return own, true
	}
}

// ownChainScope is the chain this caller's account-less events are filed under: their own.
func (r spanningReader) ownChainScope(ctx context.Context) (tenancy.Scope, bool) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return tenancy.Scope{}, false
	}

	return tenancy.Of(userID), true
}

// mergePages interleaves two id-ordered pages and keeps the first page's worth.
func mergePages(
	first, second *filtering.QueryFilteredResult[platformaudit.Entry],
	filter *filtering.QueryFilter,
) *filtering.QueryFilteredResult[platformaudit.Entry] {
	merged := make([]*platformaudit.Entry, 0, len(first.Data)+len(second.Data))
	merged = append(merged, first.Data...)
	merged = append(merged, second.Data...)

	descending := filter.SortsDescending()
	sort.SliceStable(merged, func(i, j int) bool {
		if descending {
			return merged[i].ID > merged[j].ID
		}

		return merged[i].ID < merged[j].ID
	})

	// The page the caller asked for, which is the filter's own ceiling rather than either
	// sub-read's length: two full pages merged are twice as many rows as anybody wanted.
	if filter != nil && filter.MaxResponseSize != nil {
		if limit := int(*filter.MaxResponseSize); limit > 0 && len(merged) > limit {
			merged = merged[:limit]
		}
	}

	return filtering.NewQueryFilteredResultWithoutCounts(
		merged,
		func(e *platformaudit.Entry) string { return e.ID },
		filter,
	)
}
