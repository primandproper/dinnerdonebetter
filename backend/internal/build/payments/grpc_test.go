package payments

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/platform-go/v15/callers"
	"github.com/primandproper/primitives-go/v2/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signedIn is a request from userID with accountID active, holding perms service-wide.
func signedIn(t *testing.T, userID, accountID string, perms ...authorization.Permission) (context.Context, callers.Principal) {
	t.Helper()

	ctx := sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{
			UserID:             userID,
			ServicePermissions: authorization.NewServiceRolePermissionChecker(nil, perms),
		},
		ActiveAccountID: accountID,
	})

	principal, ok := sessions.PrincipalFromContext(ctx)
	require.True(t, ok)

	return ctx, principal
}

func TestOwnAccountOrAdmin(T *testing.T) {
	T.Parallel()

	authorizer := ownAccountOrAdmin(sessions.GrantsFromContext)

	T.Run("permits the caller's active account", func(t *testing.T) {
		t.Parallel()

		accountID := fake.BuildFakeID()
		ctx, principal := signedIn(t, fake.BuildFakeID(), accountID)

		assert.NoError(t, authorizer.AuthorizeAccount(ctx, principal, accountID))
	})

	T.Run("refuses another account to a caller without the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), fake.BuildFakeID(), authorization.AccountAdminPermissions...)

		assert.ErrorIs(t, authorizer.AuthorizeAccount(ctx, principal, fake.BuildFakeID()), callers.ErrTargetNotPermitted)
	})

	T.Run("permits another account to a holder of the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), fake.BuildFakeID(), authorization.ReadAnyBillingAccountPermission)

		assert.NoError(t, authorizer.AuthorizeAccount(ctx, principal, fake.BuildFakeID()))
	})

	T.Run("a service admin holds the grant", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), fake.BuildFakeID(), authorization.ServiceAdminPermissions...)

		assert.NoError(t, authorizer.AuthorizeAccount(ctx, principal, fake.BuildFakeID()))
	})

	T.Run("refuses an empty account id, even one matching nothing active", func(t *testing.T) {
		t.Parallel()

		ctx, principal := signedIn(t, fake.BuildFakeID(), "")

		assert.ErrorIs(t, authorizer.AuthorizeAccount(ctx, principal, ""), callers.ErrTargetNotPermitted)
	})

	T.Run("refuses a request with no session", func(t *testing.T) {
		t.Parallel()

		assert.ErrorIs(t, authorizer.AuthorizeAccount(t.Context(), nil, fake.BuildFakeID()), callers.ErrTargetNotPermitted)
	})
}
