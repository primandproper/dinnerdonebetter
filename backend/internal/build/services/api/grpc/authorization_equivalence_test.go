package grpcapi

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	analyticsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/analytics/grpc"
	auditgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/audit/grpc"
	authgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"
	commentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/comments/grpc"
	dataprivacygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/dataprivacy/grpc"
	identitygrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/grpc"
	internalopsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/internalops/grpc"
	issuereportsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/issuereports/grpc"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc"
	notificationsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/notifications/grpc"
	oauthgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/oauth/grpc"
	paymentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc"
	settingsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/grpc"
	uploadedmediagrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc"
	waitlistsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/waitlists/grpc"
	webhooksgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// realMethodPermissions is the same table the running server enforces, assembled the same way.
func realMethodPermissions() interceptors.MethodPermissionsMap {
	return AggregateMethodPermissions(
		analyticsgrpc.ProvideMethodPermissions(),
		auditgrpc.ProvideMethodPermissions(),
		authgrpc.ProvideMethodPermissions(),
		commentsgrpc.ProvideMethodPermissions(),
		dataprivacygrpc.ProvideMethodPermissions(),
		identitygrpc.ProvideMethodPermissions(),
		internalopsgrpc.ProvideMethodPermissions(),
		issuereportsgrpc.ProvideMethodPermissions(),
		mealplanninggrpc.ProvideMethodPermissions(),
		notificationsgrpc.ProvideMethodPermissions(),
		oauthgrpc.ProvideMethodPermissions(),
		paymentsgrpc.ProvideMethodPermissions(),
		settingsgrpc.ProvideMethodPermissions(),
		uploadedmediagrpc.ProvideMethodPermissions(),
		waitlistsgrpc.ProvideMethodPermissions(),
		webhooksgrpc.ProvideMethodPermissions(),
	)
}

// interceptorWouldAllow reimplements the hand-rolled check exactly as AuthInterceptor performs
// it: every required permission must be held by the service-wide or the per-account checker, and
// a method with no entry in the table is refused.
func interceptorWouldAllow(
	perms interceptors.MethodPermissionsMap,
	method string,
	service authorization.ServiceRolePermissionChecker,
	account authorization.AccountRolePermissionsChecker,
) bool {
	required, declared := perms[method]
	if !declared {
		return false
	}

	for _, p := range required {
		if !service.HasPermission(p) && !account.HasPermission(p) {
			return false
		}
	}

	return true
}

// TestAuthorizationEnforcerMatchesTheHandRolledCheck is what makes enforcement safe to turn on.
//
// The argument for audit-only mode is that flipping enforcement across a large hand-written
// permission table is a coin flip on whether the two tables agree. This test replaces that coin
// flip with a proof: for every method the server actually declares and every role a principal
// can actually hold, the platform enforcer reaches the same verdict as the check it replaces.
//
// If this test ever fails, the enforcer and the interceptor have diverged, and the deployed
// service is refusing or admitting something the other would not.
func TestAuthorizationEnforcerMatchesTheHandRolledCheck(t *testing.T) {
	t.Parallel()

	perms := realMethodPermissions()
	require.NotEmpty(t, perms)

	authInterceptor := interceptors.ProvideAuthInterceptor(nil, loggingnoop.NewLogger(), nil, nil, nil, "", nil, perms)

	// Built enforcing, not audit-only: an audit-only enforcer allows everything, so comparing
	// one against the real check would prove nothing.
	enforcer, err := ProvideAuthorizationEnforcer(perms, authInterceptor, loggingnoop.NewLogger(), metricsnoop.NewMetricsProvider(), false)
	require.NoError(t, err)

	interceptor := enforcer.UnaryServerInterceptor()

	public := map[string]struct{}{}
	for _, m := range authInterceptor.UnauthenticatedRoutes() {
		public[m] = struct{}{}
	}

	// Every role a principal can hold, expanded through the policy's own inheritance
	// rather than written out by hand.
	//
	// Hand-written rows are how this test used to describe three principals that
	// cannot exist: a service admin holding 28 permissions, an account admin holding
	// 43, and a combination of the two. The database has granted those roles their
	// inherited permissions since #1215, so the sets this drove were subsets of what
	// any real caller carries — and a subset is exactly the shape that makes an
	// equivalence proof pass without proving anything about real traffic.
	expanded, err := platformauthz.ExpandInheritance(authorization.PlatformPolicy()...)
	require.NoError(t, err)

	held := func(role string) []authorization.Permission {
		set, ok := expanded[role]
		require.True(t, ok, "policy declares no role %q", role)

		out := make([]authorization.Permission, 0, set.Len())
		for _, p := range set.Slice() {
			out = append(out, authorization.Permission(p))
		}

		return out
	}

	roles := map[string]struct {
		service []authorization.Permission
		account []authorization.Permission
	}{
		"service admin":      {service: held(authorization.ServiceAdminRoleName)},
		"service data admin": {service: held(authorization.ServiceDataAdminRoleName)},
		"service user":       {service: held(authorization.ServiceUserRoleName)},
		"account admin":      {account: held(authorization.AccountAdminRoleName)},
		"account member":     {account: held(authorization.AccountMemberRoleName)},
		// What an ordinary signed-in user actually is: service_user service-wide,
		// account_member in the account they are acting on.
		"an ordinary user in their own account": {
			service: held(authorization.ServiceUserRoleName),
			account: held(authorization.AccountMemberRoleName),
		},
		"admin of an account they also administer": {
			service: held(authorization.ServiceAdminRoleName),
			account: held(authorization.AccountAdminRoleName),
		},
		"nothing at all": {},
	}

	handler := func(context.Context, any) (any, error) { return struct{}{}, nil }

	compared := 0
	for roleName, role := range roles {
		service := authorization.NewServiceRolePermissionChecker(nil, role.service)
		account := authorization.NewAccountRolePermissionChecker(role.account)

		ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
			Requester:          sessions.RequesterInfo{UserID: "user_1", ServicePermissions: service},
			AccountPermissions: map[string]authorization.AccountRolePermissionsChecker{"account_1": account},
			ActiveAccountID:    "account_1",
		})

		for method := range perms {
			// Public methods never reach the permission check in either implementation.
			if _, ok := public[method]; ok {
				continue
			}

			_, enforcerErr := interceptor(ctx, struct{}{}, &grpc.UnaryServerInfo{FullMethod: method}, handler)

			assert.Equal(t, interceptorWouldAllow(perms, method, service, account), enforcerErr == nil,
				"enforcer and interceptor disagree on %q for a %s", method, roleName)
			compared++
		}
	}

	t.Logf("compared %d method/role decisions", compared)
	assert.Positive(t, compared)
}
