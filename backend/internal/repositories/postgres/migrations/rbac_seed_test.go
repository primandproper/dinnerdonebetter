package migrations

import (
	"io/fs"
	"strings"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRBACSeed_CoversEveryPermissionARoleHolds is the check that a permission a role holds in Go
// is a permission that exists in the database.
//
// Authorization is decided against rows: a role's permissions come from user_role_permissions,
// seeded here, while the slices in internal/authorization are what the method table and the
// platform policy are written from. A permission declared there and never seeded is a method no
// principal can call, and nothing says so — the request is simply denied, exactly as it would be
// for a caller who genuinely lacked it.
//
// That is not hypothetical either. The three data privacy permissions were declared, mapped to
// the data privacy service's methods, and listed under AccountMemberPermissions, and none of them
// existed in the database — so no user could request an export of their own data, ask for it
// back, or ask to be erased, and the only symptom was a permission denial that looked correct.
//
// It reads the migration text rather than a migrated database on purpose: the seed is what is
// under test, and a test that needed a container would not run in the ordinary unit pass, which
// is where a missing permission should be caught.
func TestRBACSeed_CoversEveryPermissionARoleHolds(T *testing.T) {
	T.Parallel()

	T.Run("every declared permission is seeded", func(t *testing.T) {
		t.Parallel()

		seeded := seededMigrationText(t)

		for _, set := range []struct {
			name  string
			perms []authorization.Permission
		}{
			{"service_admin", authorization.ServiceAdminPermissions},
			{"service_data_admin", authorization.ServiceDataAdminPermissions},
			{"account_admin", authorization.AccountAdminPermissions},
			{"account_member", authorization.AccountMemberPermissions},
		} {
			for _, permission := range set.perms {
				assert.Contains(t, seeded, "'"+string(permission)+"'",
					"%s holds %q, which no migration seeds", set.name, permission)
			}
		}
	})
}

// seededMigrationText is every migration file's text, concatenated.
//
// All of them rather than the RBAC file alone: permissions are seeded in more than one place —
// the core set with the RBAC tables, the meal planning set with the meal planning schema — and a
// test that knew which file a permission ought to live in would fail the next time somebody
// chose differently, for no reason a reader could act on.
func seededMigrationText(t *testing.T) string {
	t.Helper()

	files, err := fs.Sub(rawMigrations, "migration_files")
	require.NoError(t, err)

	entries, err := fs.ReadDir(files, ".")
	require.NoError(t, err)
	require.NotEmpty(t, entries)

	var sql strings.Builder
	for _, entry := range entries {
		contents, readErr := fs.ReadFile(files, entry.Name())
		require.NoError(t, readErr)

		sql.Write(contents)
	}

	return sql.String()
}
