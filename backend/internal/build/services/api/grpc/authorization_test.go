package grpcapi

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/auth/grpc/interceptors"

	platformauthz "github.com/primandproper/platform-go/v9/authorization"
	"github.com/primandproper/platform-go/v9/authorization/static"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v9/observability/metrics/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func TestPlatformPolicy(T *testing.T) {
	T.Parallel()

	T.Run("is a valid platform policy", func(t *testing.T) {
		t.Parallel()

		// ValidateRoles runs inside NewResolver, so this rejects a policy the database
		// resolver would also reject — which is what keeps the two interchangeable.
		resolver, err := static.NewResolver(authorization.PlatformPolicy())
		require.NoError(t, err)
		require.NotNil(t, resolver)
	})

	T.Run("resolves each role to the permissions this package already grants", func(t *testing.T) {
		t.Parallel()

		resolver, err := static.NewResolver(authorization.PlatformPolicy())
		require.NoError(t, err)

		for _, tc := range []struct {
			role  string
			perms []authorization.Permission
		}{
			{authorization.ServiceAdminRoleName, authorization.ServiceAdminPermissions},
			{authorization.ServiceDataAdminRoleName, authorization.ServiceDataAdminPermissions},
			{authorization.AccountAdminRoleName, authorization.AccountAdminPermissions},
			{authorization.AccountMemberRoleName, authorization.AccountMemberPermissions},
		} {
			set, resolveErr := resolver.PermissionsForRoles(t.Context(), tc.role)
			require.NoError(t, resolveErr, tc.role)

			for _, p := range tc.perms {
				assert.True(t, set.Has(platformauthz.Permission(p)),
					"role %q lost permission %q in the platform policy", tc.role, p)
			}
		}
	})
}

func TestPlatformGrants(T *testing.T) {
	T.Parallel()

	T.Run("unions service and account authority", func(t *testing.T) {
		t.Parallel()

		service := authorization.NewServiceRolePermissionChecker(
			[]string{authorization.ServiceAdminRoleName}, []authorization.Permission{authorization.ReadUserPermission})
		account := authorization.NewAccountRolePermissionChecker(
			[]authorization.Permission{authorization.CreateWebhooksPermission})

		grants := authorization.PlatformGrants(service, account)

		assert.True(t, grants.Has(platformauthz.Permission(authorization.ReadUserPermission)))
		assert.True(t, grants.Has(platformauthz.Permission(authorization.CreateWebhooksPermission)))
		assert.False(t, grants.Has(platformauthz.Permission(authorization.ArchiveUserPermission)))
	})

	T.Run("with no checkers at all", func(t *testing.T) {
		t.Parallel()

		// Fail closed: an empty grant holds nothing.
		grants := authorization.PlatformGrants(nil, nil)
		assert.True(t, grants.IsEmpty())
	})
}

func buildTestEnforcer(t *testing.T) *grpc.UnaryServerInterceptor {
	t.Helper()

	perms := interceptors.MethodPermissionsMap{
		paymentssvc.PaymentsService_CreateSubscription_FullMethodName: {authorization.CreateSubscriptionsPermission},
	}

	enforcer, err := ProvideAuthorizationEnforcer(
		perms,
		interceptors.ProvideAuthInterceptor(nil, loggingnoop.NewLogger(), nil, nil, nil, nil, perms),
		loggingnoop.NewLogger(),
		metricsnoop.NewMetricsProvider(),
		true,
	)
	require.NoError(t, err)

	i := enforcer.UnaryServerInterceptor()

	return &i
}

func TestProvideAuthorizationEnforcer(T *testing.T) {
	T.Parallel()

	T.Run("denies nothing, because it is audit-only", func(t *testing.T) {
		t.Parallel()

		interceptor := *buildTestEnforcer(t)

		called := 0
		handler := func(context.Context, any) (any, error) {
			called++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}

		// A caller with no session at all would be refused under enforcement. Audit mode
		// records that and runs the handler anyway — which is the whole point, and the
		// property that makes deploying this safe.
		_, err := interceptor(t.Context(), &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)
		assert.Equal(t, 1, called)
	})

	T.Run("with an authorized caller", func(t *testing.T) {
		t.Parallel()

		interceptor := *buildTestEnforcer(t)

		called := 0
		handler := func(context.Context, any) (any, error) {
			called++

			return &paymentssvc.CreateSubscriptionResponse{}, nil
		}

		ctx := context.WithValue(t.Context(), sessions.SessionContextDataKey, &sessions.ContextData{
			Requester: sessions.RequesterInfo{
				UserID: "user_1",
				ServicePermissions: authorization.NewServiceRolePermissionChecker(
					[]string{authorization.ServiceAdminRoleName},
					[]authorization.Permission{authorization.CreateSubscriptionsPermission},
				),
			},
		})

		info := &grpc.UnaryServerInfo{FullMethod: paymentssvc.PaymentsService_CreateSubscription_FullMethodName}

		_, err := interceptor(ctx, &paymentssvc.CreateSubscriptionRequest{}, info, handler)
		require.NoError(t, err)
		assert.Equal(t, 1, called)
	})
}
