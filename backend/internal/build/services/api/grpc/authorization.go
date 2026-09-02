package grpcapi

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	authzgrpc "github.com/primandproper/platform-go/v13/authorization/grpc"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
)

// auditOnlyAuthorization decides whether the platform enforcer records its verdict or acts on it.
//
// It is false: the enforcer denies. That is safe because it is not a guess.
// TestAuthorizationEnforcerMatchesTheHandRolledCheck drives every method the server declares
// against every role a principal can hold — 2,920 decisions — and asserts the enforcer reaches
// the same verdict as the hand-rolled check it replaces. Audit-only mode exists for services
// that cannot make that comparison ahead of time; this one can, because both sides read the same
// method table and the same permission checkers.
//
// That test is load-bearing. It is what would catch a policy change here drifting from the
// permission slices in internal/authorization, and it already caught one real divergence: 39
// methods mapped to an empty permission slice, which the interceptor admits and the platform
// would have refused as undeclared.
const auditOnlyAuthorization = false

// ProvideAuthorizationEnforcer builds platform-go's authorization enforcer over the method table
// the hand-rolled AuthInterceptor already uses.
//
// It takes no resolver. Enforcement reads a caller's Grants, which the session already carries,
// and resolution — turning the roles a principal holds into the permissions they carry — happens
// once when that session is built, in the identity repository, against the policy tables. That
// split is the platform package's whole shape: resolve may do I/O and happens per session, checks
// never fail and happen per call.
//
// This used to construct a static resolver over PlatformPolicy() and throw it away, purely to run
// the policy through ValidateRoles. That validation now runs where it can act on the answer — in
// Seed, at migration time, where a policy with an unknown parent or an inheritance cycle fails the
// deploy rather than being noticed at boot and discarded.
//
// Nothing is derived twice: the required permissions come from the same aggregated map the
// interceptor reads, and the public methods come from the interceptor's own allow-list.
func ProvideAuthorizationEnforcer(
	methodPermissions interceptors.MethodPermissionsMap,
	authInterceptor *interceptors.AuthInterceptor,
	logger logging.Logger,
	metricsProvider metrics.Provider,
	auditOnly bool,
) (*authzgrpc.Enforcer, error) {
	builder := authzgrpc.NewRequirements()

	declared := map[string]struct{}{}

	for method, perms := range methodPermissions {
		declared[method] = struct{}{}

		// A method mapped to an empty permission slice means "any authenticated caller" —
		// AuthInterceptor still demands a session, then loops over zero permissions and
		// admits the request. The platform refuses to express that as a requirement, for
		// good reason: zero required permissions reads as a check while behaving as an
		// allow. Public is how it says the same thing honestly. Authentication is still
		// enforced by the interceptor ahead of this one, so Public here scopes to
		// authorization only.
		//
		// This is not cosmetic. Treating these as undeclared instead would deny 39 methods
		// the current service admits — including UpdateUserDetails.
		if len(perms) == 0 {
			builder.Public(method)

			continue
		}

		builder.Require(method, authorization.ToPlatformPermissions(perms)...)
	}

	for _, method := range authInterceptor.UnauthenticatedRoutes() {
		// A method may be both unauthenticated and permissioned in the current setup;
		// skipping authentication wins there, so it must win here, and declaring a method
		// twice is an error.
		if _, ok := declared[method]; ok {
			continue
		}

		builder.Public(method)
	}

	reqs, err := builder.Build()
	if err != nil {
		return nil, platformerrors.Wrap(err, "building authorization requirements")
	}

	opts := []authzgrpc.Option{
		authzgrpc.WithLogger(logger),
		authzgrpc.WithMetricsProvider(metricsProvider),
	}
	if auditOnly {
		opts = append(opts, authzgrpc.WithAuditOnly())
	}

	return authzgrpc.NewEnforcer(reqs, grantsFromSession, opts...)
}

// grantsFromSession bridges this service's session type onto the platform's Grants.
//
// Service-wide and per-account authority are handed over as separate sets and unioned, which is
// how the existing check already treats them: a permission held in either place is held.
func grantsFromSession(ctx context.Context) (platformauthz.Grants, bool) {
	sessionCtxData, err := sessions.RequireFromContext(ctx)
	if err != nil || sessionCtxData == nil {
		return platformauthz.Grants{}, false
	}

	return authorization.PlatformGrants(
		sessionCtxData.ServiceRolePermissionChecker(),
		sessionCtxData.AccountRolePermissionsChecker(),
	), true
}
