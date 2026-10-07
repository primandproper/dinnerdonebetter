package grpcapi

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	authzgrpc "github.com/primandproper/primitives-go/v2/authorization/grpc"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
)

// auditOnlyAuthorization decides whether the platform enforcer records its verdict or acts on it.
//
// It is false: the enforcer denies, and it is the only permission check on either chain.
// TestAuthorizationEnforcerMatchesTheMethodPermissionTable drives every method the server
// declares against every role a principal can hold and asserts the enforcer reaches the verdict
// the merged table — MethodPermissions — gives. Both are built from the same fragments and
// overrides: the table from them merged, the enforcer by declaring one and overriding with the
// other.
//
// That test is load-bearing. It is what would catch a policy change here drifting from the
// permission slices in internal/authorization, and it caught one real divergence when the
// enforcer was introduced beside AuthInterceptor's own check, since deleted: 39 methods mapped to
// an empty permission slice, which that check admitted and the platform would have refused as
// undeclared.
const auditOnlyAuthorization = false

// ProvideAuthorizationEnforcer builds primitives' authorization enforcer over this deployment's
// method permission table. It is the server's only permission check, on the unary and the stream
// chain alike.
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
// The table is each surface's fragment, declared as it ships, with this deployment's amendments
// applied through RequirementsBuilder.Override rather than written over a copy first. Override is
// checked against what was declared, so an amendment to a method no fragment serves fails the
// build here instead of being declared quietly. The public methods come from the interceptor's
// own allow-list.
func ProvideAuthorizationEnforcer(
	fragments interceptors.MethodPermissionsMap,
	overrides map[string][]authorization.Permission,
	authInterceptor *interceptors.AuthInterceptor,
	logger logging.Logger,
	metricsProvider metrics.Provider,
	auditOnly bool,
) (*authzgrpc.Enforcer, error) {
	builder := authzgrpc.NewRequirements()

	declared := map[string]struct{}{}

	for method, perms := range fragments {
		declared[method] = struct{}{}

		// A method mapped to an empty permission slice means "any authenticated caller" —
		// AuthInterceptor demands a session for it, and nothing more is asked. The platform
		// refuses to express that as a requirement, for
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

	for method, perms := range overrides {
		builder.Override(method, authorization.ToPlatformPermissions(perms)...)
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

	return authzgrpc.NewEnforcer(reqs, sessions.GrantsFromContext, opts...)
}
