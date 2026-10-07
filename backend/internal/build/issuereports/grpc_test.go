package issuereports

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v15/callers"
	platformissuereports "github.com/primandproper/platform-go/v15/issuereports"
	issuereportsgrpc "github.com/primandproper/platform-go/v15/issuereports/grpc"
	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signedIn is a request from userID holding perms service-wide.
func signedIn(t *testing.T, userID string, perms ...authorization.Permission) (context.Context, callers.Principal) {
	t.Helper()

	ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             userID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, perms),
		},
		ActiveAccountID: fake.BuildFakeID(),
	})

	principal, ok := sessions.PrincipalFromContext(ctx)
	require.True(t, ok)

	return ctx, principal
}

func TestOwnReportOrAdmin_AuthorizeReport(T *testing.T) {
	T.Parallel()

	authorizer := ownReportOrAdmin{grants: sessions.GrantsFromContext}

	T.Run("permits the reporter", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		ctx, principal := signedIn(t, userID)

		assert.NoError(t, authorizer.AuthorizeReport(ctx, principal, &platformissuereports.Report{Reporter: userID}))
	})

	T.Run("refuses somebody else's report to a household admin", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.AccountAdminPermissions...)

		err := authorizer.AuthorizeReport(ctx, principal, &platformissuereports.Report{Reporter: fake.BuildFakeID()})
		assert.ErrorIs(t, err, callers.ErrTargetNotPermitted)
	})

	T.Run("permits somebody else's report to a holder of the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.TriageIssueReportsPermission)

		assert.NoError(t, authorizer.AuthorizeReport(ctx, principal, &platformissuereports.Report{Reporter: fake.BuildFakeID()}))
	})

	T.Run("a service admin holds the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ServiceAdminPermissions...)

		assert.NoError(t, authorizer.AuthorizeReport(ctx, principal, &platformissuereports.Report{Reporter: fake.BuildFakeID()}))
	})

	T.Run("refuses a request with no session", func(t *testing.T) {
		t.Parallel()

		err := authorizer.AuthorizeReport(t.Context(), nil, &platformissuereports.Report{Reporter: fake.BuildFakeID()})
		assert.ErrorIs(t, err, callers.ErrTargetNotPermitted)
	})
}

func TestOwnReportOrAdmin_AuthorizeReporter(T *testing.T) {
	T.Parallel()

	authorizer := ownReportOrAdmin{grants: sessions.GrantsFromContext}

	T.Run("permits a caller naming themselves", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		ctx, principal := signedIn(t, userID)

		assert.NoError(t, authorizer.AuthorizeReporter(ctx, principal, userID))
	})

	T.Run("refuses somebody else to a household admin", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.AccountAdminPermissions...)

		assert.ErrorIs(t, authorizer.AuthorizeReporter(ctx, principal, fake.BuildFakeID()), callers.ErrTargetNotPermitted)
	})

	T.Run("permits somebody else to a holder of the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.TriageIssueReportsPermission)

		assert.NoError(t, authorizer.AuthorizeReporter(ctx, principal, fake.BuildFakeID()))
	})

	T.Run("a service admin holds the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), authorization.ServiceAdminPermissions...)

		assert.NoError(t, authorizer.AuthorizeReporter(ctx, principal, fake.BuildFakeID()))
	})
}

func TestPermissionOverrides(T *testing.T) {
	T.Parallel()

	T.Run("puts the cross-scope reads behind the grant a service admin works the queue with", func(t *testing.T) {
		t.Parallel()

		overrides := PermissionOverrides()
		require.Len(t, overrides, 2)

		for method, required := range overrides {
			assert.Equal(t, []authorization.Permission{authorization.TriageIssueReportsPermission}, required, method)
			assert.NotContains(t, required, issuereportsgrpc.PermissionReadAnyReports, method)
		}
	})
}
