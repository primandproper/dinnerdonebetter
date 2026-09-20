package auditlog

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/primitives-go/v2/tenancy"
)

// callerChains answers the chains one caller's own entries are in.
//
// Two of them, because audit.ScopeFor files an entry under its account where there is one
// and under its actor otherwise — a login, a signup and a password reset happen inside no
// account — so a signed-in person legitimately belongs to their account's chain and their
// own. platform reads both in one paged call and merges them; see WithChainsResolver.
//
// Both answers come off the session, which is what the contract requires and the whole of
// the tenancy check on a multi-chain read: the scopes returned here are bound into the
// queries directly, so one derived from a request field would hand the caller a
// cross-tenant read. Neither is.
//
// A service administrator is deliberately not answered here, and gets one chain rather than
// two. Their read is wider than any list this could return — every chain in the deployment —
// and there is no way to say that in a slice of scopes. spanningReader turns that single
// read into platform's operator read instead, which is the one place the session decides
// how wide a read is rather than which rows are in it.
//
// The two are never the same scope, because an account id and a user id are separate rows
// identifiers, so nothing here dedupes: a guard against naming one chain twice would be a
// branch no input reaches, and a comment explaining a case that cannot arise is worse than
// the absence of one. A caller with no account names one chain because they have one.
//
// An empty answer means the connection's own scope, so a caller with neither a session nor
// an account reads exactly what they would have read before this existed.
func callerChains(ctx context.Context) ([]tenancy.Scope, error) {
	data := sessions.FromContext(ctx)

	if data.GetServicePermissions().IsServiceAdmin() {
		return nil, nil
	}

	chains := make([]tenancy.Scope, 0, 2)

	if accountID := data.GetActiveAccountID(); accountID != "" {
		chains = append(chains, tenancy.Of(accountID))
	}

	if userID := data.GetUserID(); userID != "" {
		chains = append(chains, tenancy.Of(userID))
	}

	return chains, nil
}
