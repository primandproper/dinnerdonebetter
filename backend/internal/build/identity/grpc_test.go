package identity

import (
	"maps"
	"slices"
	"testing"

	identitygrpc "github.com/primandproper/platform-go/v15/identity/grpc"

	"github.com/stretchr/testify/assert"
)

// TestPermissionOverridesCoverEverySelfServiceMethod pins PermissionOverrides() against the list
// platform keeps of the RPCs it deliberately declines to gate.
//
// The overrides are hand-written, which means they can go stale in the one direction that fails
// quietly. platform's fragment declares a ninth self-service RPC public, so one arriving in a
// later version would be served to every signed-in caller under no grant of this application's
// own — not a compile error, not a wiring error. Upstream keeps Permissions() and
// SelfServiceMethods() exhaustive over the service with a test of its own; this is the same test
// from the consumer's side, and it is the thing that turns that version bump into a red build
// here.
//
// The reverse direction is the requirements builder's: an override of a method platform stops
// declaring is ErrOverrideUndeclared at boot, and grpcapi's tests build the real table.
func TestPermissionOverridesCoverEverySelfServiceMethod(T *testing.T) {
	T.Parallel()

	T.Run("the overrides are exactly the self-service set", func(t *testing.T) {
		t.Parallel()

		assert.ElementsMatch(t, identitygrpc.SelfServiceMethods(), slices.Collect(maps.Keys(PermissionOverrides())))
	})

	T.Run("every override names a permission", func(t *testing.T) {
		t.Parallel()

		for method, perms := range PermissionOverrides() {
			assert.NotEmpty(t, perms, "%s is overridden with no permission, which the requirements builder refuses", method)
		}
	})
}

// TestPermissions pins Permissions() as platform's fragment, unamended.
func TestPermissions(T *testing.T) {
	T.Parallel()

	T.Run("platform's gated methods keep platform's grants", func(t *testing.T) {
		t.Parallel()

		platform := identitygrpc.Permissions()
		declared := Permissions()

		for method := range platform {
			assert.Equal(t, platform[method], declared[method], "%s is platform's to gate and this package changed it", method)
		}
	})

	T.Run("the self-service methods are declared public, as platform's Require declares them", func(t *testing.T) {
		t.Parallel()

		declared := Permissions()
		platform := identitygrpc.Permissions()

		for _, method := range identitygrpc.SelfServiceMethods() {
			perms, ok := declared[method]
			assert.True(t, ok, "%s is self-service upstream and undeclared here, which the enforcer reads as denied", method)
			assert.Empty(t, perms, method)
		}

		assert.Len(t, declared, len(platform)+len(identitygrpc.SelfServiceMethods()))
	})
}
