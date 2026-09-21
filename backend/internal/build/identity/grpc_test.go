package identity

import (
	"testing"

	identitygrpc "github.com/primandproper/platform-go/v14/identity/grpc"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPermissionsCoversEverySelfServiceMethod pins Permissions() against the list platform
// keeps of the RPCs it deliberately declines to gate.
//
// The amendment in Permissions() is hand-written, which means it can go stale in the one
// direction that fails quietly. platform's enforcer treats an undeclared method as denied,
// so a ninth self-service RPC arriving in a later version would not be a compile error or a
// wiring error — it would be a method every caller is refused, discovered by somebody
// finding a screen broken. Upstream keeps Permissions() and SelfServiceMethods() exhaustive
// over the service with a test of its own; this is the same test from the consumer's side,
// and it is the thing that turns that version bump into a red build here.
//
// The reverse direction matters too. A method platform starts gating that this file also
// declares would have the application's grant silently replace the platform's.
func TestPermissionsCoversEverySelfServiceMethod(T *testing.T) {
	T.Parallel()

	T.Run("every self-service method is declared here", func(t *testing.T) {
		t.Parallel()

		declared := Permissions()

		for _, method := range identitygrpc.SelfServiceMethods() {
			perms, ok := declared[method]
			assert.True(t, ok, "%s is self-service upstream and undeclared here, which the enforcer reads as denied", method)
			assert.NotEmpty(t, perms, "%s is declared with no permission, which the requirements builder refuses", method)
		}
	})

	T.Run("nothing is declared twice", func(t *testing.T) {
		t.Parallel()

		platform := identitygrpc.Permissions()
		declared := Permissions()

		require.Greater(t, len(declared), len(platform))

		for method := range platform {
			assert.Equal(t, platform[method], declared[method],
				"%s is platform's to gate and this package overrode it", method)
		}
	})

	T.Run("the amendment is exactly the self-service set", func(t *testing.T) {
		t.Parallel()

		platform := identitygrpc.Permissions()

		var added []string
		for method := range Permissions() {
			if _, ok := platform[method]; !ok {
				added = append(added, method)
			}
		}

		assert.ElementsMatch(t, identitygrpc.SelfServiceMethods(), added)
	})
}
