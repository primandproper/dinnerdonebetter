package auditlog

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/filtering"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// operatorReader is the one read a ChainsResolver cannot ask for: every chain in the
// deployment.
//
// It used to do three things. Merging a caller's two chains into one page and finding an
// entry by id in either of them are both platform's now — see callerChains, and
// audit/grpc's chainsFor, which feeds the list and the get the same set so the two cannot
// disagree about which chains somebody belongs to. What is left is not a gap in that seam;
// it is the question the seam deliberately declines to answer.
//
// ChainsResolver says which chains a caller's *own* entries are in, and every scope it
// returns is bound into a query directly. There is no value in it meaning "do not narrow",
// on purpose: an unnarrowed read is not a wider list of a person's own chains, it is asking
// whether this person may audit the deployment. That is a policy decision, and a resolver
// signature that could express it by accident would be a place to make it by accident.
//
// So it is made here, in a sentence somebody wrote. A service administrator investigating
// an account they are not a member of otherwise reads an empty window, which is the one
// question an audit log exists to answer going unanswerable over the wire.
type operatorReader struct {
	platformaudit.Reader
}

// List widens a service administrator's read to every chain.
//
// The grant cannot decide this. ReadAuditLogEntriesPermission is an account member's,
// because a member reading their own account's log is the ordinary case — so it says
// whether this kind of call is allowed at all, and how wide it reaches is asked here, of
// the session, after the query is built and before any row is read. It is the division
// waitlists draws between PermissionReadSignups and AuthorizeSubjectRead.
//
// Everybody else passes through untouched: callerChains has already told the server which
// scopes to ask for, and this is called once per chain with each of them.
func (r operatorReader) List(
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

// Get reads one entry from any chain, for a service administrator.
//
// A nil scope is every tenant, which is why this is the only caller that gets one: handing
// it to anybody else is the cross-tenant disclosure the pointer exists to prevent. Everyone
// else reaches the chains callerChains named, one at a time, which is platform's search and
// not this package's business.
func (r operatorReader) Get(
	ctx context.Context,
	q database.SQLQueryExecutor,
	scope *tenancy.Scope,
	id string,
) (*platformaudit.Entry, error) {
	if !isServiceAdmin(ctx) {
		return r.Reader.Get(ctx, q, scope, id)
	}

	return r.Reader.Get(ctx, q, nil, id)
}

// isServiceAdmin reports whether the session holds the service administrator role.
func isServiceAdmin(ctx context.Context) bool {
	return sessions.FromContext(ctx).GetServicePermissions().IsServiceAdmin()
}
