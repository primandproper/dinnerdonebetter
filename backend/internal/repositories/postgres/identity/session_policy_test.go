package identity

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	"github.com/primandproper/platform-go/v13/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestQuerier_Integration_SessionPermissionsComeFromThePolicy is the end-to-end check that
// the permissions a request carries are the ones the policy declares.
//
// It runs against the real seeded tables, through the real resolver, so it spans the whole
// path a request takes: the assignment rows a signup writes, the resolution of those role
// names, and the checker the interceptor asks. Asserting on a checker built from a slice
// would test none of that.
//
// The assertions are set equality against the declared closure, because the failure this
// adoption fixed was one of degree — a service admin holding a strict subset of what a
// service admin holds. Every containment assertion in the codebase passed throughout.
func TestQuerier_Integration_SessionPermissionsComeFromThePolicy(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	expanded, err := platformauthz.ExpandInheritance(authorization.PlatformPolicy()...)
	require.NoError(t, err)

	// A signup: service_user service-wide, account_admin in the account it creates.
	user := createUserForTest(t, ctx, nil, dbc)

	sessionCtxData, err := dbc.BuildSessionContextDataForUser(ctx, user.ID, "")
	require.NoError(t, err)
	require.NotNil(t, sessionCtxData)

	t.Run("an ordinary user holds nothing service-wide", func(t *testing.T) {
		// service_user grants no permissions, and that is the point of it: an
		// ordinary user's authority is account_member, held per account. A
		// service-wide grant would apply in every account.
		assert.False(t, sessionCtxData.ServiceRolePermissionChecker().HasPermission(authorization.ReadUserPermission))
		assert.False(t, sessionCtxData.ServiceRolePermissionChecker().HasPermission(authorization.CreateWebhooksPermission))
	})

	t.Run("and account_admin's whole closure in their own account", func(t *testing.T) {
		require.NotEmpty(t, sessionCtxData.ActiveAccountID)

		checker, ok := sessionCtxData.AccountPermissions[sessionCtxData.ActiveAccountID]
		require.True(t, ok)

		lister, ok := checker.(authorization.PermissionLister)
		require.True(t, ok)

		assert.True(t, lister.GrantedPermissions().Equal(expanded[authorization.AccountAdminRoleName]),
			"account admin resolves to %d permissions, expected %d",
			lister.GrantedPermissions().Len(), expanded[authorization.AccountAdminRoleName].Len())

		// Inherited from account_member, which is the half a flat policy lost.
		assert.True(t, checker.HasPermission(authorization.ReadWebhooksPermission))
	})

	t.Run("a service admin holds the whole closure", func(t *testing.T) {
		admin := createUserForTest(t, ctx, nil, dbc)

		_, err = dbc.writeDB.ExecContext(ctx,
			"INSERT INTO user_role_assignments (id, user_id, role_name) VALUES ($1, $2, $3)",
			identifiers.New(), admin.ID, authorization.ServiceAdminRoleName)
		require.NoError(t, err)

		adminSession, sessErr := dbc.BuildSessionContextDataForUser(ctx, admin.ID, "")
		require.NoError(t, sessErr)

		lister, ok := adminSession.ServiceRolePermissionChecker().(authorization.PermissionLister)
		require.True(t, ok)

		assert.True(t, lister.GrantedPermissions().Equal(expanded[authorization.ServiceAdminRoleName]),
			"service admin resolves to %d permissions, expected %d",
			lister.GrantedPermissions().Len(), expanded[authorization.ServiceAdminRoleName].Len())

		// The role names ride along, because IsServiceAdmin reads them rather than
		// a permission.
		assert.True(t, adminSession.ServiceRolePermissionChecker().IsServiceAdmin())
	})

	t.Run("a user with no assignment at all holds nothing", func(t *testing.T) {
		stranger := fakes.BuildFakeUser()
		created := createUserForTest(t, ctx, stranger, dbc)

		_, err = dbc.writeDB.ExecContext(ctx,
			"UPDATE user_role_assignments SET archived_at = NOW() WHERE user_id = $1", created.ID)
		require.NoError(t, err)

		strangerSession, sessErr := dbc.BuildSessionContextDataForUser(ctx, created.ID, "")
		require.NoError(t, sessErr)

		// Fail closed: an empty grant holds nothing, and resolving zero roles is
		// not an error.
		assert.False(t, strangerSession.ServiceRolePermissionChecker().HasPermission(authorization.ReadUserPermission))
		assert.False(t, strangerSession.AccountRolePermissionsChecker().HasPermission(authorization.CreateWebhooksPermission))
	})
}
