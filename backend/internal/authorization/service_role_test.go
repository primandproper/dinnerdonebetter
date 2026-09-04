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
		assert.True(t, r.CanManageUserSessions())
	})

	T.Run("both", func(t *testing.T) {
		t.Parallel()

		r := NewServiceRolePermissionChecker([]string{ServiceUserRole.String(), ServiceAdminRoleName}, nil)

		assert.True(t, r.IsServiceAdmin())
	})
}
