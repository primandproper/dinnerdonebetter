package sessions

import (
	"context"

	"github.com/primandproper/platform-go/v14/callers"
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
