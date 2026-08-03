package integration

import (
	"net/http"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/webhooks"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	webhookssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/webhooks"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/webhooks/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

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

		newStatus := identity.BannedUserAccountStatus.String()

		_, err = adminClient.AdminUpdateUserStatus(ctx, &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: createdUser.ID,
			NewStatus:    newStatus,
			Reason:       t.Name(),
		})
		require.NoError(t, err)

		status, err = testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		assert.NoError(t, err)
		assert.Equal(t, status.AccountStatus, newStatus)
	})

	T.Run("fails for non-admin user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		createdUser, testClient := createUserAndClientForTest(t)

		status, err := testClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)
		require.NotNil(t, status)

		_, err = testClient.AdminUpdateUserStatus(ctx, &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: createdUser.ID,
			NewStatus:    identity.BannedUserAccountStatus.String(),
			Reason:       t.Name(),
		})
		require.Error(t, err)
	})

	T.Run("nonexistent user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.AdminUpdateUserStatus(ctx, &identitysvc.AdminUpdateUserStatusRequest{
			TargetUserId: nonexistentID,
			NewStatus:    identity.BannedUserAccountStatus.String(),
			Reason:       t.Name(),
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

		t.Logf("impersonating user %s and account %s to get webhook %s", user.ID, account.Result.Id, webhook.ID)

		retrievedWebhook, err := adminClient.GetWebhook(impersonatedCtx, &webhookssvc.GetWebhookRequest{WebhookId: webhook.ID})
		assert.NoError(t, err)
		assert.NotNil(t, retrievedWebhook)
	})

	T.Run("standard user should not be able to impersonate others", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		_, testClient2 := createUserAndClientForTest(t)

		exampleWebhookInput := &webhooks.WebhookCreationRequestInput{
			ContentType: "application/json",
			Method:      http.MethodPost,
			Name:        t.Name(),
			URL:         "https://192.0.2.1/webhook",
			Events:      []string{webhooks.WebhookCreatedServiceEventType},
		}

		input := converters.ConvertWebhookCreationRequestInputToGRPCWebhookCreationRequestInput(exampleWebhookInput)

		createdWebhook, err := testClient.CreateWebhook(ctx, &webhookssvc.CreateWebhookRequest{Input: input})
		require.NoError(t, err)

		retrievedWebhook, err := testClient.GetWebhook(ctx, &webhookssvc.GetWebhookRequest{WebhookId: createdWebhook.Created.Id})
		require.NoError(t, err)
		require.NotNil(t, retrievedWebhook)

		account, err := testClient.GetActiveAccount(ctx, &authsvc.GetActiveAccountRequest{})
		require.NoError(t, err)
		require.NotNil(t, account)

		impersonatedCtx := client.ImpersonateUseAndAccountContext(ctx, user.ID, account.Result.Id)

		t.Logf("impersonating user %s and account %s to get webhook %s", user.ID, account.Result.Id, retrievedWebhook.Result.Id)

		webhook, err := testClient2.GetWebhook(impersonatedCtx, &webhookssvc.GetWebhookRequest{WebhookId: retrievedWebhook.Result.Id})
		assert.Error(t, err)
		assert.Nil(t, webhook)
	})
}
