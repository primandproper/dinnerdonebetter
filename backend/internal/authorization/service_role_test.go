package authorization

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServiceRoles(T *testing.T) {
	T.Parallel()

	T.Run("service user", func(t *testing.T) {
		t.Parallel()

		r := NewServiceRolePermissionChecker([]string{ServiceUserRole.String()}, nil)

		assert.False(t, r.IsServiceAdmin())
	})

	T.Run("service admin", func(t *testing.T) {
		t.Parallel()

		allPerms := slices.Concat(ServiceAdminPermissions, ServiceDataAdminPermissions, AccountAdminPermissions, AccountMemberPermissions)
		r := NewServiceRolePermissionChecker([]string{ServiceAdminRoleName}, allPerms)

		assert.True(t, r.IsServiceAdmin())
		assert.True(t, r.CanUpdateUserAccountStatuses())
		assert.True(t, r.CanImpersonateUsers())
	})

	T.Run("both", func(t *testing.T) {
		t.Parallel()

		r := NewServiceRolePermissionChecker([]string{ServiceUserRole.String(), ServiceAdminRoleName}, nil)

		assert.True(t, r.IsServiceAdmin())
	})
}

func TestOrdinaryServiceRoles(T *testing.T) {
	T.Parallel()

	T.Run("keeps what a person holds about themselves", func(t *testing.T) {
		t.Parallel()

		assert.Equal(t, []string{ServiceUserRoleName}, OrdinaryServiceRoles(t.Context(), []string{ServiceUserRoleName}))
	})

	T.Run("drops an operator's roles, leaving the person", func(t *testing.T) {
		t.Parallel()

		for _, role := range AdministrativeServiceRoleNames() {
			kept := OrdinaryServiceRoles(t.Context(), []string{role})

			assert.Equal(t, []string{ServiceUserRoleName}, kept, role)
		}
	})

	T.Run("gives nobody a role they did not hold", func(t *testing.T) {
		t.Parallel()

		assert.Empty(t, OrdinaryServiceRoles(t.Context(), nil))
	})
}
