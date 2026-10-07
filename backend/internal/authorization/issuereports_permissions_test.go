package authorization

import (
	"testing"

	issuereportsgrpc "github.com/primandproper/platform-go/v15/issuereports/grpc"
	platformauthz "github.com/primandproper/primitives-go/v2/authorization"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIssueReportPermissions pins who holds what over the issue report queue, read off the
// policy with inheritance expanded — which is what a role actually carries — rather than off the
// slices it is assembled from.
func TestIssueReportPermissions(T *testing.T) {
	T.Parallel()

	expanded, err := platformauthz.ExpandInheritance(PlatformPolicy()...)
	require.NoError(T, err)

	queue := []Permission{
		TriageIssueReportsPermission,
		TransitionIssueReportsPermission,
		UpdateIssueReportsPermission,
		ArchiveIssueReportsPermission,
	}

	T.Run("every user files and reads their own, as a person rather than a member", func(t *testing.T) {
		t.Parallel()

		user := expanded[ServiceUserRoleName]
		assert.True(t, user.Has(CreateIssueReportsPermission))
		assert.True(t, user.Has(ReadIssueReportsPermission))
	})

	T.Run("a household admin works none of the queue", func(t *testing.T) {
		t.Parallel()

		for _, role := range []string{AccountAdminRoleName, AccountMemberRoleName, ServiceUserRoleName} {
			for _, p := range queue {
				assert.False(t, expanded[role].Has(p), "%s holds %s", role, p)
			}
		}
	})

	T.Run("a service admin works all of it", func(t *testing.T) {
		t.Parallel()

		admin := expanded[ServiceAdminRoleName]
		for _, p := range queue {
			assert.True(t, admin.Has(p), "service admin lacks %s", p)
		}

		assert.True(t, admin.Has(CreateIssueReportsPermission))
		assert.True(t, admin.Has(ReadIssueReportsPermission))
	})

	T.Run("nobody holds the cross-tenant read a single queue has no use for", func(t *testing.T) {
		t.Parallel()

		for role, set := range expanded {
			assert.False(t, set.Has(issuereportsgrpc.PermissionReadAnyReports), "%s holds it", role)
		}
	})
}
