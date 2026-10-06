package integration

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	internalopssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"
	"github.com/primandproper/dinnerdonebetter/backend/internal/localdev"

	"github.com/primandproper/platform-go/v15/authentication/signin"
	"github.com/primandproper/platform-go/v15/authentication/signin/signinpb"
	"github.com/primandproper/platform-go/v15/identity/identitypb"
	webhookspb "github.com/primandproper/platform-go/v15/webhooks/webhookspb"
	"github.com/primandproper/primitives-go/v2/pointer"

	"github.com/pquerna/otp/totp"
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

		authStatus, err := testClient.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
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
		_, err = testClient.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// And on the HTTP routes, which would otherwise be the way around it.
		assert.Equal(t, http.StatusForbidden, privacyRequestsStatusFor(t, token))

		banned, err := adminClient.IdentityService().GetUser(ctx, &identitypb.GetUserRequest{UserId: createdUser.ID})
		require.NoError(t, err)
		assert.Equal(t, identitypb.AccountStatus_ACCOUNT_STATUS_BANNED, banned.GetUser().GetAccountStatus())
	})
}

// TestAdmin_UserImpersonation pins this application's half of impersonation: who may ask for a
// token (an operator holding imitate.user, and nobody acting through one already), that the token
// acts as the subject in the account it names, and that what it does is recorded as the
// operator's. The token itself — its claims, its lifetime, its login — is platform's.
// TestAdmin_OperatorGrantsRideOnlyOnTheAdministrativeDoor pins the claim that decides whether a
// token carries an operator's grants: the administrative door sets it, and every other door does
// not, whatever service roles the person holds.
func TestAdmin_OperatorGrantsRideOnlyOnTheAdministrativeDoor(T *testing.T) {
	T.Parallel()

	T.Run("an operator signed in through the ordinary door is a person, not an operator", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		code, err := totp.GenerateCode(strings.ToUpper(premadeAdminUser.TwoFactorSecret), time.Now().UTC())
		require.NoError(t, err)

		token, err := localdev.FetchLoginTokenForUser(ctx, fmt.Sprintf(":%d", apiServiceConfig.GRPCServer.Port), &signinpb.Credentials{
			Username: premadeAdminUser.Username,
			Password: adminUserPassword,
			TotpCode: code,
		})
		require.NoError(t, err)

		ordinary, err := buildAuthedGRPCClientWithBearerToken(token)
		require.NoError(t, err)
		t.Cleanup(func() { require.NoError(t, ordinary.Close()) })

		user, _ := createUserAndClientForTest(t)

		_, err = ordinary.ImpersonateUser(ctx, &internalopssvc.ImpersonateUserRequest{SubjectId: user.ID})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

		// The control: the same person through the administrative door.
		res, err := adminClient.ImpersonateUser(ctx, &internalopssvc.ImpersonateUserRequest{SubjectId: user.ID})
		require.NoError(t, err)
		assert.NotEmpty(t, res.GetToken())
	})
}

func TestAdmin_UserImpersonation(T *testing.T) {
	T.Parallel()

	T.Run("an operator acts as a user, in the user's account", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		webhook := createWebhookForTest(t, testClient)
		accountID := getAccountIDForTest(t, testClient)

		impersonated := impersonationClientForTest(t, adminClient, user.ID, accountID)

		retrievedWebhook, err := impersonated.WebhooksService().GetEndpoint(ctx, &webhookspb.GetEndpointRequest{EndpointId: webhook.GetId()})
		require.NoError(t, err)
		assert.Equal(t, webhook.GetId(), retrievedWebhook.GetResult().GetId())

		authStatus, err := impersonated.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, user.ID, authStatus.GetStatus().GetUser().GetId())
		assert.Equal(t, accountID, authStatus.GetStatus().GetActiveAccountId())
	})

	T.Run("the impersonation is recorded as the subject's authentication, naming the operator", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		impersonationClientForTest(t, adminClient, user.ID, getAccountIDForTest(t, testClient))

		// platform records an impersonation the way it records every proven sign-in — as the
		// subject's signin.user.authenticated, through RecordAs, with the operator on the
		// entry as its Impersonator and on the event as its actor. The subject also signed in
		// for themselves above, which is an authentication naming no actor, so the payload is
		// found by the operator's ID as well as the subject's and the event's name.
		recorded, err := outboxPayloads(ctx,
			`convert_from(payload, 'UTF8') LIKE '%' || $1 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $2 || '%' AND convert_from(payload, 'UTF8') LIKE '%' || $3 || '%'`,
			user.ID, signin.EventUserAuthenticated.String(), premadeAdminUser.ID)
		require.NoError(t, err)
		require.Len(t, recorded, 1)
		assert.Equal(t, premadeAdminUser.ID, findStringKey(recorded[0], "actorID"))
		assert.Equal(t, user.ID, findStringKey(recorded[0], "userID"))
	})

	T.Run("an ordinary user may not impersonate anybody", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		_, err := otherClient.ImpersonateUser(ctx, &internalopssvc.ImpersonateUserRequest{
			SubjectId: user.ID,
			AccountId: getAccountIDForTest(t, testClient),
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("an impersonation carries the subject's grants and none of the operator's", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		other, _ := createUserAndClientForTest(t)

		impersonated := impersonationClientForTest(t, adminClient, user.ID, getAccountIDForTest(t, testClient))

		// Banning somebody is an operator's, and the operator behind this token holds it.
		_, err := impersonated.IdentityService().UpdateUserAccountStatus(ctx, &identitypb.UpdateUserAccountStatusRequest{
			UserId:      other.ID,
			Status:      identitypb.AccountStatus_ACCOUNT_STATUS_BANNED,
			Explanation: t.Name(),
		})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("an impersonation cannot start another", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)
		other, _ := createUserAndClientForTest(t)

		impersonated := impersonationClientForTest(t, adminClient, user.ID, getAccountIDForTest(t, testClient))

		_, err := impersonated.ImpersonateUser(ctx, &internalopssvc.ImpersonateUserRequest{SubjectId: other.ID})
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
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

		_, err = testClient.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
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
