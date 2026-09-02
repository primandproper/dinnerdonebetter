package migrations

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"
	authzdatabase "github.com/primandproper/platform-go/v13/authorization/database"
	"github.com/primandproper/platform-go/v13/database/dialect"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSeededPolicyMatchesTheDeclaredPolicy is the check that the policy in the database
// is the policy in Go.
//
// It is the load-bearing test of this adoption, because it is the only one that spans
// all three things that have to agree: the declaration in PlatformPolicy(), the rows the
// migrator seeds from it, and the recursive statement that resolves them back out. A
// unit test over ExpandInheritance proves only that the resolver agrees with itself; the
// failure this replaces — a role whose Go definition and whose database rows disagreed —
// lived precisely in the gap between them.
//
// It asserts set equality rather than containment. Containment is what the old
// assertions did, and a subset satisfies it: the flat policy this adoption corrected held
// 28 of service_admin's 240 permissions and passed every one.
func TestSeededPolicyMatchesTheDeclaredPolicy(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		resolver, err := authzdatabase.NewResolver(
			&authzdatabase.Config{Dialect: dialect.Postgres, TablePrefix: authorization.TablePrefix},
			db,
		)
		require.NoError(t, err)

		declared := authorization.PlatformPolicy()

		expanded, err := platformauthz.ExpandInheritance(declared...)
		require.NoError(t, err)

		for _, role := range declared {
			seeded, resolveErr := resolver.PermissionsForRoles(ctx, role.Name)
			require.NoError(t, resolveErr, role.Name)

			assert.True(t, seeded.Equal(expanded[role.Name]),
				"role %q resolves to %d permissions in the database, %d in the declaration",
				role.Name, seeded.Len(), expanded[role.Name].Len())
		}
	})

	// Seeding is idempotent and it revokes: a role's grants are rewritten rather than
	// added to. That is what makes the Go declaration the only one — a permission
	// deleted there is a grant that disappears on the next migration, which used to
	// need a hand-written DELETE in an unrelated migration.
	T.Run("re-seeding converges rather than accumulating", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		var before int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ddb_authz_role_permissions").Scan(&before))
		assert.Positive(t, before)

		require.NoError(t, migrator.Migrate(ctx, db))

		var after int
		require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM ddb_authz_role_permissions").Scan(&after))
		assert.Equal(t, before, after)
	})

	// An assignment naming a role nothing declares is refused at the database, rather
	// than resolving to an empty permission set that looks like a legitimate denial.
	T.Run("an assignment cannot name a role that does not exist", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		db, _ := pgtesting.BuildDatabaseContainerForTest(t)
		migrator, err := NewMigrator(loggingnoop.NewLogger())
		require.NoError(t, err)
		require.NoError(t, migrator.Migrate(ctx, db))

		// A real user, so the assignment reaches the role key rather than tripping
		// the user key first.
		_, err = db.ExecContext(ctx,
			"INSERT INTO users (id, username, email_address, hashed_password, two_factor_secret) VALUES ($1, $2, $3, '', '')",
			"user_1", "somebody", "somebody@example.com")
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			"INSERT INTO user_role_assignments (id, user_id, role_name) VALUES ($1, $2, $3)",
			"assignment_1", "user_1", "a_role_nobody_declared")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "user_role_assignments_role_fk")

		// And the same insert naming a declared role is accepted, so the failure
		// above is the key doing its job rather than the statement being wrong.
		_, err = db.ExecContext(ctx,
			"INSERT INTO user_role_assignments (id, user_id, role_name) VALUES ($1, $2, $3)",
			"assignment_1", "user_1", authorization.ServiceUserRoleName)
		require.NoError(t, err)
	})
}
