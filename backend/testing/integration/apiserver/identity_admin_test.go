package integration

import (
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v14/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAdmin_BanningUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdUser, testClient := createUserAndClientForTest(t)
		token, err := loginForConformance(ctx, createdUser, "")
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, privacyRequestsStatusFor(t, token), "the control: the HTTP routes admit this caller before the ban")

		authStatus, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		require.NotNil(t, authStatus)

		_, err = adminClient.IdentityService().UpdateUserAccountStatus(ctx, &identitypb.UpdateUserAccountStatusRequest{
			UserId:      createdUser.ID,
			Status:      identitypb.AccountStatus_ACCOUNT_STATUS_BANNED,
			Explanation: t.Name(),
		})
		require.NoError(t, err)

		// A ban takes effect on the next request, on every surface at once: the read every
		// authenticated request makes refuses a status that does not admit signing in, so
		// this session is not merely marked — it stops resolving. PermissionDenied rather
		// than Unauthenticated, because the token is genuine and refreshing it would not help.
		_, err = testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// And on the HTTP routes, which would otherwise be the way around it.
		assert.Equal(t, http.StatusForbidden, privacyRequestsStatusFor(t, token))

		banned, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: createdUser.ID})
		require.NoError(t, err)
		assert.Equal(t, identitypb.AccountStatus_ACCOUNT_STATUS_BANNED, banned.GetUser().GetAccountStatus())
	})
}

func TestAdmin_UserImpersonation(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		webhook := createWebhookForTest(t, testClient)

		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		impersonatedCtx := client.ImpersonateUseAndAccountContext(ctx, user.ID, account.Result.Id)

		t.Logf("impersonating user %s and account %s to get webhook %s", user.ID, account.Result.Id, webhook.GetId())

		retrievedWebhook, err := adminClient.WebhooksService().GetEndpoint(impersonatedCtx, &webhookspb.GetEndpointRequest{EndpointId: webhook.GetId()})
		require.NoError(t, err)
		assert.NotNil(t, retrievedWebhook)
	})

	T.Run("standard user should not be able to impersonate others", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		_, testClient2 := createUserAndClientForTest(t)

		createdWebhook, err := testClient.WebhooksService().SaveEndpoint(ctx, &webhookspb.SaveEndpointRequest{
			Endpoint: &webhookspb.WebhookEndpointInput{
				ContentType: "application/json",
				Name:        t.Name(),
				Url:         "https://192.0.2.1/webhook",
				EventTypes:  []string{webhooks.WebhookCreatedServiceEventType},
			},
			SigningKeys: signingSecretForTest(),
		})
		require.NoError(t, err)

		retrievedWebhook, err := testClient.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: createdWebhook.GetResult().GetId()})
		require.NoError(t, err)
		require.NotNil(t, retrievedWebhook)

		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		impersonatedCtx := client.ImpersonateUseAndAccountContext(ctx, user.ID, account.Result.Id)

		t.Logf("impersonating user %s and account %s to get webhook %s", user.ID, account.Result.Id, retrievedWebhook.GetResult().GetId())

		webhook, err := testClient2.WebhooksService().GetEndpoint(impersonatedCtx, &webhookspb.GetEndpointRequest{EndpointId: retrievedWebhook.GetResult().GetId()})
		require.Error(t, err)
		assert.Nil(t, webhook)
	})
}

// TestAdmin_ForcedPasswordChange pins what a user told to change their password may still
// do: learn who they are and change it, on either surface, and nothing else.
func TestAdmin_ForcedPasswordChange(T *testing.T) {
	T.Parallel()

	T.Run("the caller may read who they are, and nothing else", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		token, err := loginForConformance(ctx, user, "")
		require.NoError(t, err)

		_, err = adminClient.IdentityService().SetUserRequiresPasswordChange(ctx, &identitypb.SetUserRequiresPasswordChangeRequest{
			UserId:                 user.ID,
			RequiresPasswordChange: pointer.To(true),
		})
		require.NoError(t, err)

		principal, err := testClient.IdentityService().GetPrincipal(ctx, &identitypb.GetPrincipalRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, principal.GetPrincipal().GetUser().GetId())

		_, err = testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)

		_, err = testClient.WebhooksService().ListEndpoints(ctx, &webhookspb.ListEndpointsRequest{})
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))

		// No HTTP route is the change, so every one is refused.
		assert.Equal(t, http.StatusForbidden, privacyRequestsStatusFor(t, token))
	})
}

// privacyRequestsStatusFor is the status the privacy-request listing answers a caller holding
// token — an HTTP route every signed-in caller may otherwise use.
func privacyRequestsStatusFor(t *testing.T, token string) int {
	t.Helper()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, httpTestServerAddress+"/privacy-requests", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	require.NoError(t, res.Body.Close())

	return res.StatusCode
}
