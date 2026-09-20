package identity

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/tenancy"
)

// SessionBuilder turns a user and an account into the session context every authenticated
// request is served with.
//
// It replaces identity.BuildSessionContextDataForUser, which read memberships, resolved a
// default account, checked the named one, read role assignments out of a table of this
// application's own and then read the user — five queries assembled by hand, with the
// membership check as one of the five. platform's Store.GetPrincipal is all of that in one
// method, and its documentation is blunt about why it is a method rather than a recipe:
// the membership check "is the one every hand-built session context eventually forgets,
// and forgetting it hands one account's data to another account's member".
//
// It also refuses a user whose status does not admit signing in, before the membership
// reads. This application checked that separately, in the interceptor, which meant a ban
// took effect wherever somebody had remembered to look; it now takes effect on the read
// every authenticated request already makes.
//
// What is left here is the part that was never platform's: turning role names into
// permissions. That is this application's policy, keyed by role name and cached for every
// principal holding the role, and platform has no opinion about it — Principal carries the
// names, and what they grant is resolved below.
type SessionBuilder struct {
	_ struct{} `json:"-"`

	client database.Client
	store  platformidentity.Store
	policy platformauthz.PolicyResolver
	tracer tracing.Tracer
}

// NewSessionBuilder builds a SessionBuilder.
func NewSessionBuilder(
	client database.Client,
	store platformidentity.Store,
	policy platformauthz.PolicyResolver,
	tracerProvider tracing.Provider,
) (*SessionBuilder, error) {
	if client == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil database client")
	}

	if store == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil identity store")
	}

	if policy == nil {
		return nil, platformerrors.Wrap(platformerrors.ErrNilInputParameter, "nil policy resolver")
	}

	return &SessionBuilder{
		client: client,
		store:  store,
		policy: policy,
		tracer: tracing.NewNamedTracer(tracerProvider, "identity_session_builder"),
	}, nil
}

// BuildSessionContextDataForUser resolves the principal and renders it as session context.
//
// An empty activeAccountID means "wherever they land", which platform answers with the
// user's default account — or with no account at all for somebody nobody has put in one
// yet, which is a state a sign-in has to survive. A named account must be one they are a
// live member of, and is refused otherwise.
func (b *SessionBuilder) BuildSessionContextDataForUser(
	ctx context.Context,
	userID, activeAccountID string,
) (*sessions.ContextData, error) {
	ctx, span := b.tracer.StartSpan(ctx)
	defer span.End()

	if userID == "" {
		return nil, platformerrors.ErrInvalidIDProvided
	}

	// The reader rather than a transaction: this is a read on the request path, and the
	// three statements behind it do not need to agree with each other more than the
	// clock does.
	principal, err := b.store.GetPrincipal(ctx, b.client.Reader(), tenancy.Global(), userID, activeAccountID)
	if err != nil {
		return nil, err
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
