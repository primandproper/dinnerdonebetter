package sessions

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v14/callers"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// Principal adapts this application's session to the shape platform's gRPC
// servers read a caller off a context with.
//
// It exists because platform names the shape and the consumer names the type:
// callers.Principal is an interface with no default implementation, deliberately,
// because who is calling is the consumer's own answer. This is that answer, in
// one place, so that the thirteen platform surfaces this repo mounts do not each
// invent one.
type Principal struct {
	_ struct{} `json:"-"`

	data *ContextData
}

var _ callers.Principal = (*Principal)(nil)

// UserID is the calling user.
func (p *Principal) UserID() string { return p.data.GetUserID() }

// Scope is whose directory the caller is in, which here is always the global one.
//
// This deployment has a single user directory: an account is a grouping inside
// it rather than a tenant with its own users, so there is one scope and every
// caller is in it. See internal/domain/identity.
//
// It is not the account. ActiveAccountID is what carries that, and conflating
// the two would scope the user directory by account — which would make a user
// invisible to the account they have not selected.
func (*Principal) Scope() tenancy.Scope { return tenancy.Global() }

// ActiveAccountID is the account this request is against, or empty when the
// caller named none.
func (p *Principal) ActiveAccountID() string { return p.data.GetActiveAccountID() }

// PrincipalFromContext is the callers.PrincipalExtractor for this application.
//
// The false return is an unauthenticated call, which every platform surface
// answers codes.Unauthenticated. It is deliberately the same session the
// application's own handlers read, so a request cannot be one caller to an
// adopted surface and another to a local one.
func PrincipalFromContext(ctx context.Context) (callers.Principal, bool) {
	data := FromContext(ctx)
	if data == nil || data.GetUserID() == "" {
		return nil, false
	}

	return &Principal{data: data}, true
}

// GrantsFromContext is the authorization.GrantsExtractor for this application.
//
// It is the second half of what platform's surfaces read off a request, beside
// PrincipalFromContext: who is calling, and what they may do. Both are named
// here so that the surfaces this repo mounts read one answer rather than each
// assembling their own.
//
// Two sets rather than one, because this application resolves authority at two
// levels and a platform surface should see both: what the service role grants
// regardless of account, and what the active account's membership grants within
// it. authorization.NewGrants drops a set that grants nothing, so a caller with
// no membership in the active account is one set instead of two rather than a
// special case anybody has to check for.
//
// The false return is a request with no session, which every surface reads as
// granting nothing. That is the safe direction: a write gated on a grant is
// refused, and a read that would have been narrowed by one is narrowed.
func GrantsFromContext(ctx context.Context) (platformauthz.Grants, bool) {
	data := FromContext(ctx)
	if data == nil {
		return platformauthz.Grants{}, false
	}

	var sets []*platformauthz.PermissionSet

	if lister, ok := data.ServiceRolePermissionChecker().(authorization.PermissionLister); ok {
		sets = append(sets, lister.GrantedPermissions())
	}

	if lister, ok := data.AccountRolePermissionsChecker().(authorization.PermissionLister); ok {
		sets = append(sets, lister.GrantedPermissions())
	}

	return platformauthz.NewGrants(sets...), true
}

// AccountScopedPrincipal is Principal for a surface whose rows belong to an
// account rather than to the deployment.
//
// It exists because callers.Principal.Scope answers for the surface being
// called, not for the caller, and this application's domains do not agree on
// one tenancy. Comments, settings, waitlists and uploaded media are global — a
// recipe's discussion reads the same for everybody, and scoping it per account
// would make one recipe's comments depend on who was looking. Issue reports are
// their account's, and that scoping is what closed a member-visible leak.
//
// Wiring the wrong one is not a compile error and not a test failure unless the
// test spans two accounts: a global principal handed to an account-scoped
// surface files every account's rows under the global scope and serves all of
// them to everybody. So the two are named separately, and a surface's wiring
// says which it is.
type AccountScopedPrincipal struct {
	_ struct{} `json:"-"`

	data *ContextData
}

var _ callers.Principal = (*AccountScopedPrincipal)(nil)

// UserID is the calling user.
func (p *AccountScopedPrincipal) UserID() string { return p.data.GetUserID() }

// Scope is the account the request is against.
//
// tenancy.Of refuses to name nobody, so a caller with no active account yields
// the zero Scope, which every platform read refuses rather than widening to
// every tenant.
func (p *AccountScopedPrincipal) Scope() tenancy.Scope {
	return tenancy.Of(p.data.GetActiveAccountID())
}

// ActiveAccountID is the account this request is against.
func (p *AccountScopedPrincipal) ActiveAccountID() string { return p.data.GetActiveAccountID() }

// AccountScopedPrincipalFromContext is the callers.PrincipalExtractor for
// surfaces whose rows belong to an account.
func AccountScopedPrincipalFromContext(ctx context.Context) (callers.Principal, bool) {
	data := FromContext(ctx)
	if data == nil || data.GetUserID() == "" {
		return nil, false
	}

	return &AccountScopedPrincipal{data: data}, true
}
