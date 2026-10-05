package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformidentity "github.com/primandproper/platform-go/v15/identity"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"
	authzmock "github.com/primandproper/primitives-go/v2/authorization/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resolverGranting(perms ...authorization.Permission) *authzmock.PolicyResolverMock {
	return &authzmock.PolicyResolverMock{
		PermissionsForRolesFunc: func(context.Context, ...string) (*platformauthz.PermissionSet, error) {
			return platformauthz.NewPermissionSet(authorization.ToPlatformPermissions(perms)...), nil
		},
	}
}

func TestNewImpersonationPolicy(T *testing.T) {
	T.Parallel()

	T.Run("admits an operator whose roles grant imitate.user", func(t *testing.T) {
		t.Parallel()

		policy := NewImpersonationPolicy(resolverGranting(authorization.ImpersonateUserPermission))

		err := policy(t.Context(), &platformidentity.User{ServiceRoles: []string{authorization.ServiceAdminRoleName}}, &platformidentity.User{})
		assert.NoError(t, err)
	})

	T.Run("refuses an operator whose roles do not", func(t *testing.T) {
		t.Parallel()

		policy := NewImpersonationPolicy(resolverGranting(authorization.PermissionReadUsers))

		err := policy(t.Context(), &platformidentity.User{ServiceRoles: []string{authorization.ServiceUserRoleName}}, &platformidentity.User{})
		assert.ErrorIs(t, err, ErrImpersonationNotPermitted)
	})

	T.Run("refuses nobody as the operator", func(t *testing.T) {
		t.Parallel()

		policy := NewImpersonationPolicy(resolverGranting(authorization.ImpersonateUserPermission))

		assert.ErrorIs(t, policy(t.Context(), nil, &platformidentity.User{}), ErrImpersonationNotPermitted)
	})

	T.Run("refuses when the roles cannot be resolved", func(t *testing.T) {
		t.Parallel()

		resolveErr := errors.New("blah")
		policy := NewImpersonationPolicy(&authzmock.PolicyResolverMock{
			PermissionsForRolesFunc: func(context.Context, ...string) (*platformauthz.PermissionSet, error) {
				return nil, resolveErr
			},
		})

		err := policy(t.Context(), &platformidentity.User{ServiceRoles: []string{authorization.ServiceAdminRoleName}}, &platformidentity.User{})
		require.ErrorIs(t, err, resolveErr)
	})
}
