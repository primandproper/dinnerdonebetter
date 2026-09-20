package auditlog

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// spanningReader covers the two reads platform's chains resolver does not.
//
// It used to merge the paged list as well. platform does that now — see callerChains and
// audit/grpc.WithChainsResolver — and it does it in the one place the set of chains is
// derivable, with one cursor across all of them. Two things are left over.
//
// The first is Get. It still takes the single scope the connection resolved, so a caller
// who finds an entry of their own in a list and then asks for it by id would be told it
// does not exist. The list and the get disagreeing about which chains somebody belongs to
// is the gap this closes.
//
// The second is the operator. platform's reader has the read — a nil query scope narrows
// nothing — and its surface declines to use it, because that surface "has no operator".
// This deployment has one, and a service administrator investigating an account they are
// not a member of otherwise reads an empty window. A ChainsResolver cannot express it:
// every chain in the deployment is not a slice anybody can enumerate.
type spanningReader struct {
	platformaudit.Reader
}

// List widens a service administrator's read to every chain.
//
// One read rather than a merge, because nil is already every chain. Everybody else is
// untouched: callerChains has already told the server which scopes to ask for, and this is
// called once per chain with each of them.
//
// The grant cannot decide this. ReadAuditLogEntriesPermission is an account member's,
// because a member reading their own account's log is the ordinary case — so it says
// whether this kind of call is allowed at all, and how wide it reaches is asked here, of
// the session, after the query is built and before any row is read. It is the division
// waitlists draws between PermissionReadSignups and AuthorizeSubjectRead.
func (r spanningReader) List(
	ctx context.Context,
	q database.SQLQueryExecutor,
	query *platformaudit.Query,
	filter *filtering.QueryFilter,
) (*filtering.QueryFilteredResult[platformaudit.Entry], error) {
	if query == nil || query.Scope == nil || !isServiceAdmin(ctx) {
		return r.Reader.List(ctx, q, query, filter)
	}

	// A copy, because the query belongs to the handler that built it and a scope written
	// through it would outlive this call.
	across := *query
	across.Scope = nil

	return r.Reader.List(ctx, q, &across, filter)
}

// Get reads one entry from either of the caller's chains, or from any chain for an operator.
//
// Two reads rather than one unscoped read, for everybody but the operator. A nil scope is
// every tenant here, and handing that to a caller's read is the cross-tenant disclosure the
// pointer exists to prevent — so each chain is asked for the entry by name, and an entry in
// neither is not found.
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

	own, ok := ownChainScope(ctx)
	if !ok || (scope != nil && *scope == own) {
		return nil, err
	}

	// The first error is dropped rather than joined: both are "not in this chain", and a
	// caller told an entry is in neither wants one answer, not the same one twice.
	return r.Reader.Get(ctx, q, &own, id)
}

// ownChainScope is the chain this caller's account-less events are filed under: their own.
func ownChainScope(ctx context.Context) (tenancy.Scope, bool) {
	userID := sessions.FromContext(ctx).GetUserID()
	if userID == "" {
		return tenancy.Scope{}, false
	}

	return tenancy.Of(userID), true
}

// isServiceAdmin reports whether the session holds the service administrator role.
func isServiceAdmin(ctx context.Context) bool {
	return sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin()
}
