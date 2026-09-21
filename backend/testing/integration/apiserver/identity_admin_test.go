package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/primandproper/platform-go/v14/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v14/webhooks/webhookspb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin_BanningUsers(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdUser, testClient := createUserAndClientForTest(t)

		status, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		require.NotNil(t, status)

		_, err = adminClient.IdentityService().UpdateUserAccountStatus(ctx, &identitypb.UpdateUserAccountStatusRequest{
			UserId:      createdUser.ID,
			Status:      identitypb.AccountStatus_ACCOUNT_STATUS_BANNED,
			Explanation: t.Name(),
		})
		require.NoError(t, err)

		// A ban takes effect on the next request, on every surface at once: the read every
		// authenticated request makes refuses a status that does not admit signing in, so
		// this session is not merely marked — it stops resolving.
		_, err = testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.Error(t, err)

		banned, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: createdUser.ID})
		require.NoError(t, err)
		assert.Equal(t, identitypb.AccountStatus_ACCOUNT_STATUS_BANNED, banned.GetUser().GetAccountStatus())
	})

	T.Run("fails for non-admin user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdUser, testClient := createUserAndClientForTest(t)

		status, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		require.NotNil(t, status)

		_, err = testClient.IdentityService().UpdateUserAccountStatus(ctx, &identitypb.UpdateUserAccountStatusRequest{
			UserId:      createdUser.ID,
			Status:      identitypb.AccountStatus_ACCOUNT_STATUS_BANNED,
			Explanation: t.Name(),
		})
		require.Error(t, err)
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.IdentityService().UpdateUserAccountStatus(ctx, &identitypb.UpdateUserAccountStatusRequest{
			UserId:      nonexistentID,
			Status:      identitypb.AccountStatus_ACCOUNT_STATUS_BANNED,
			Explanation: t.Name(),
		})
		require.Error(t, err)
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
