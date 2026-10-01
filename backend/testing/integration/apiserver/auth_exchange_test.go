package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v14/authentication/signin/signinpb"
	"github.com/primandproper/platform-go/v14/identity/identitypb"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestAuth_ExchangeToken pins AuthService.ExchangeToken, the door somebody in two households moves
// between them through until SignInService can (platform-go#1069). Its authority is the refresh
// token SignInService issued, so it is called with nobody signed in.
func TestAuth_ExchangeToken(T *testing.T) {
	T.Parallel()

	T.Run("naming another of the user's accounts lands the login there, and retires the token it spent", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		signIn := buildSignInClientForTest(t)
		token, _ := signInForTest(t, signIn)

		signedIn, err := buildAuthedGRPCClientWithBearerToken(token.GetToken())
		require.NoError(t, err)

		created, err := signedIn.IdentityService().CreateAccount(ctx, &identitypb.CreateAccountRequest{
			Name:       t.Name(),
			OwnerRoles: []string{authorization.AccountAdminRoleName},
		})
		require.NoError(t, err)
		otherAccount := created.GetAccount().GetId()
		require.NotEqual(t, token.GetActiveAccountId(), otherAccount)

		anonymous := buildUnauthenticatedGRPCClientForTest(t)

		exchanged, err := anonymous.ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{
			RefreshToken:     token.GetRefreshToken(),
			DesiredAccountId: otherAccount,
		})
		require.NoError(t, err)
		assert.Equal(t, otherAccount, exchanged.GetAccountId())

		switched, err := buildAuthedGRPCClientWithBearerToken(exchanged.GetAccessToken())
		require.NoError(t, err)

		authStatus, err := switched.GetAuthStatus(ctx, &signinpb.GetAuthStatusRequest{})
		require.NoError(t, err)
		assert.Equal(t, otherAccount, authStatus.GetStatus().GetActiveAccountId())

		// The new pair refreshes like any other.
		_, err = signIn.ExchangeRefreshToken(ctx, &signinpb.ExchangeRefreshTokenRequest{RefreshToken: exchanged.GetRefreshToken()})
		require.NoError(t, err)

		// And the token the switch spent is spent.
		_, err = anonymous.ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{RefreshToken: token.GetRefreshToken()})
		assert.Equal(t, codes.Unauthenticated, status.Code(err))
	})

	T.Run("naming no account is a refresh", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		token, _ := signInForTest(t, buildSignInClientForTest(t))

		exchanged, err := buildUnauthenticatedGRPCClientForTest(t).ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{
			RefreshToken: token.GetRefreshToken(),
		})
		require.NoError(t, err)
		assert.Equal(t, token.GetActiveAccountId(), exchanged.GetAccountId())
		assert.NotEqual(t, token.GetRefreshToken(), exchanged.GetRefreshToken())
	})

	T.Run("naming an account the user is not in is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		token, _ := signInForTest(t, buildSignInClientForTest(t))
		_, neighbor := createUserAndClientForTest(t)

		_, err := buildUnauthenticatedGRPCClientForTest(t).ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{
			RefreshToken:     token.GetRefreshToken(),
			DesiredAccountId: getAccountIDForTest(t, neighbor),
		})
		require.Error(t, err)
		assert.NotEqual(t, codes.Internal, status.Code(err))
	})
}
