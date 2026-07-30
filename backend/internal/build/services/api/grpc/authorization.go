package grpcapi

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authorization"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	platformauthz "github.com/primandproper/platform-go/v8/authorization"
	authzgrpc "github.com/primandproper/platform-go/v8/authorization/grpc"
	"github.com/primandproper/platform-go/v8/authorization/static"
	platformerrors "github.com/primandproper/platform-go/v8/errors"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/metrics"
)

// ProvideAuthorizationEnforcer builds platform-go's authorization enforcer over the same policy
// and the same method table the hand-rolled AuthInterceptor already uses.
//
// It runs in audit-only mode: it evaluates every call and records the decision, and denies
// nothing. The existing interceptor remains the thing that actually refuses requests.
//
// That is deliberate, and it is the package's own documented migration path. Turning enforcement
// on across a service that already has a large hand-written permission table is a coin flip on
// whether the two tables agree; audit mode turns that coin flip into a measurement. Watch
// authorization_denied for methods this would have refused but the current interceptor allows —
// each one is either a policy bug here or a gap there — and flip enforcement on only once it
// stays at zero under real traffic.
//
// Nothing is derived twice: the policy comes from the same permission slices the checkers are
// built from, the required permissions come from the same aggregated map, and the public methods
// come from the interceptor's own allow-list.
func ProvideAuthorizationEnforcer(
	methodPermissions interceptors.MethodPermissionsMap,
	authInterceptor *interceptors.AuthInterceptor,
	logger logging.Logger,
	metricsProvider metrics.Provider,
) (*authzgrpc.Enforcer, error) {
	// The static resolver is the right backend while the policy is compiled in. Moving to the
	// database resolver later is a configuration change: it takes the same []Role.
	if _, err := static.NewResolver(authorization.PlatformPolicy()); err != nil {
		return nil, platformerrors.Wrap(err, "validating authorization policy")
	}

	builder := authzgrpc.NewRequirements()

	for method, perms := range methodPermissions {
		// A method that declares zero permissions is an authorization hole wearing the
		// costume of a requirement, so the platform rejects it. The current interceptor
		// denies such a method outright; leaving it undeclared here does the same.
		if len(perms) == 0 {
			continue
		}

		builder.Require(method, authorization.ToPlatformPermissions(perms)...)
	}

	for _, method := range authInterceptor.UnauthenticatedRoutes() {
		// A method may be both public and permissioned in the current setup; public wins
		// there, so it must win here, and declaring both would be a duplicate.
		if _, permissioned := methodPermissions[method]; permissioned {
			continue
		}

		builder.Public(method)
	}

	reqs, err := builder.Build()
	if err != nil {
		return nil, platformerrors.Wrap(err, "building authorization requirements")
	}

	return authzgrpc.NewEnforcer(
		reqs,
		grantsFromSession,
		authzgrpc.WithAuditOnly(),
		authzgrpc.WithLogger(logger),
		authzgrpc.WithMetricsProvider(metricsProvider),
	)
}

// grantsFromSession bridges this service's session type onto the platform's Grants.
//
// Service-wide and per-account authority are handed over as separate sets and unioned, which is
// how the existing check already treats them: a permission held in either place is held.
func grantsFromSession(ctx context.Context) (platformauthz.Grants, bool) {
	sessionCtxData, err := sessions.FetchContextDataFromContext(ctx)
	if err != nil || sessionCtxData == nil {
		return platformauthz.Grants{}, false
	}

	return authorization.PlatformGrants(
		sessionCtxData.ServiceRolePermissionChecker(),
		sessionCtxData.AccountRolePermissionsChecker(),
	), true
}
