package identity

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// SessionBuilder turns a resolved principal into the session context every authenticated
// request is served with.
//
// The principal is platform's: Store.GetPrincipal reads the user, refuses one whose status does
// not admit signing in, and checks the active account is one they are a live member of — the
// check its documentation calls "the one every hand-built session context eventually forgets".
// The sign-in extractor makes that read on every request, so nothing here reads the directory.
//
// What is left is the part that was never platform's: turning role names into permissions. That
// is this application's policy, keyed by role name and cached for every principal holding the
// role, and platform has no opinion about it — Principal carries the names, and what they grant
// is resolved below.
type SessionBuilder struct {
	_ struct{} `json:"-"`

	policy platformauthz.PolicyResolver
	tracer tracing.Tracer
}

// NewSessionBuilder builds a SessionBuilder.
func NewSessionBuilder(
	policy platformauthz.PolicyResolver,
	tracerProvider tracing.Provider,
) (*SessionBuilder, error) {
	if policy == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil policy resolver")
	}

	return &SessionBuilder{
		policy: policy,
		tracer: tracing.NewNamedTracer(tracerProvider, "identity_session_builder"),
	}, nil
}

// SessionForPrincipal renders a principal platform's sign-in extractor already resolved as
// session context.
//
// The extractor makes the read — Store.GetPrincipal, with its status refusal and its membership
// check — and narrows the service roles of a caller who came through an ordinary door, so this is
// only the half that was never platform's: what the roles grant. Reading the principal again here
// would be a second query for an answer the request already holds, and a second place for the
// service roles to come back un-narrowed.
func (b *SessionBuilder) SessionForPrincipal(
	ctx context.Context,
	principal *platformidentity.Principal,
) (*sessions.ContextData, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if principal == nil || principal.User == nil {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	return b.render(ctx, principal)
}

// render turns a principal into session context, resolving what its roles grant.
func (b *SessionBuilder) render(ctx context.Context, principal *platformidentity.Principal) (*sessions.ContextData, error) {
	servicePerms, err := b.policy.PermissionsForRoles(ctx, principal.User.ServiceRoles...)
	if err != nil {
		return nil, err
	}

	// Every account the user belongs to, not just the active one. The checker map is what
	// answers "may they do this in that account", and a map holding only the active
	// account would refuse every cross-account read the API deliberately allows.
	accountPermissions := make(map[string]authorization.AccountRolePermissionsChecker, len(principal.Memberships))
	for _, membership := range principal.Memberships {
		perms, permErr := b.policy.PermissionsForRoles(ctx, membership.Roles...)
		if permErr != nil {
			return nil, permErr
		}

		accountPermissions[membership.BelongsToAccount] =
			authorization.NewAccountRolePermissionCheckerFromSet(membership.Roles, perms)
	}

	return &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:                   principal.User.ID,
			Username:                 principal.User.Username,
			EmailAddress:             principal.User.EmailAddress,
			AccountStatus:            string(principal.User.AccountStatus),
			AccountStatusExplanation: principal.User.AccountStatusExplanation,
			ServicePermissions: authorization.NewServiceRolePermissionCheckerFromSet(
				principal.User.ServiceRoles, servicePerms),
		},
		AccountPermissions: accountPermissions,
		ActiveAccountID:    principal.ActiveAccountID,
	}, nil
}
