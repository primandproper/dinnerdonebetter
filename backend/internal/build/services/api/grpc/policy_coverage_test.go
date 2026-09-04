package grpcapi

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	platformauthz "github.com/primandproper/platform-go/v13/authorization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMethodTableIsCoveredByThePolicy is the check that a permission a method requires is
// a permission some role can hold.
//
// It replaces a test that read the migration text looking for permission names, which was
// the only thing keeping the Go slices and the hand-written seed in agreement. That seed is
// gone — the policy is seeded from PlatformPolicy(), so a permission declared in Go and
// missing from the database is no longer expressible.
//
// The failure it was catching is not gone, only relocated. A method mapped to a permission
// no role grants is a method nobody can call, and nothing says so: the request is refused
// exactly as it would be for a caller who genuinely lacked it. That is not hypothetical —
// the three data privacy permissions were declared, mapped to the data privacy service's
// methods, and held by no role, so no user could export their own data, read it back, or
// ask to be erased.
func TestMethodTableIsCoveredByThePolicy(T *testing.T) {
	T.Parallel()

	T.Run("every required permission is held by some role", func(t *testing.T) {
		t.Parallel()

		expanded, err := platformauthz.ExpandInheritance(authorization.PlatformPolicy()...)
		require.NoError(t, err)

		grantable := map[authorization.Permission]struct{}{}
		for _, set := range expanded {
			for _, p := range set.Slice() {
				grantable[authorization.Permission(p)] = struct{}{}
			}
		}

		perms := realMethodPermissions()
		require.NotEmpty(t, perms)

		for method, required := range perms {
			for _, p := range required {
				assert.Contains(t, grantable, p,
					"method %q requires %q, which no role grants", method, p)
			}
		}
	})
}
