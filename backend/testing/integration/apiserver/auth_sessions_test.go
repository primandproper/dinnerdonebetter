package integration

import (
	"testing"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuth_ListActiveSessions(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotEmpty(t, res.Sessions)

		for _, sess := range res.Sessions {
			assert.NotEmpty(t, sess.Id)
			assert.NotEmpty(t, sess.LoginMethod)
			assert.NotNil(t, sess.CreatedAt)
			assert.NotNil(t, sess.ExpiresAt)
		}
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		assert.Error(t, err)
	})
}

func TestAuth_RevokeSession(T *testing.T) {
	T.Parallel()

	T.Run("revoke a non-current session", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		// Log in again to create a second session.
		token2 := fetchLoginTokenForUserForTest(t, user)
		client2, err := buildAuthedGRPCClient(ctx, token2)
		require.NoError(t, err)

		// List sessions from client 1 – should see at least 2.
		listRes, err := testClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(listRes.Sessions), 2)

		// Find a non-current session to revoke.
		var targetSessionID string
		for _, sess := range listRes.Sessions {
			if !sess.IsCurrent {
				targetSessionID = sess.Id
				break
			}
		}
		require.NotEmpty(t, targetSessionID, "expected at least one non-current session")

		// Revoke it.
		_, err = testClient.RevokeSession(ctx, &authsvc.RevokeSessionRequest{
			SessionId: targetSessionID,
		})
		require.NoError(t, err)

		// Verify it's gone from the list.
		listRes2, err := testClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		require.NoError(t, err)
		for _, sess := range listRes2.Sessions {
			assert.NotEqual(t, targetSessionID, sess.Id, "revoked session should not appear in active list")
		}

		// The second client should still work (we revoked its session, so it may fail
		// depending on whether the second login's session was the one we revoked).
		// Instead of asserting on client2, just ensure it was created successfully.
		_ = client2
	})

	T.Run("empty session ID returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.RevokeSession(ctx, &authsvc.RevokeSessionRequest{
			SessionId: "",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.RevokeSession(ctx, &authsvc.RevokeSessionRequest{
			SessionId: "anything",
		})
		assert.Error(t, err)
	})
}

func TestAuth_RevokeAllOtherSessions(T *testing.T) {
	T.Parallel()

	T.Run("happy path via JWT bearer", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, oauthClient := createUserAndClientForTest(t)

		// Create additional sessions by logging in more times.
		_ = fetchLoginTokenForUserForTest(t, user)
		jwt3 := fetchLoginTokenForUserForTest(t, user)

		// Use a JWT Bearer client so the interceptor knows the current session ID.
		jwtClient, err := buildAuthedGRPCClientWithBearerToken(jwt3)
		require.NoError(t, err)

		// Should have multiple sessions.
		listRes, err := oauthClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(listRes.Sessions), 3, "expected multiple sessions before revoking")

		// Revoke all other sessions via the JWT client (which knows its session).
		_, err = jwtClient.RevokeAllOtherSessions(ctx, &authsvc.RevokeAllOtherSessionsRequest{})
		require.NoError(t, err)

		// Should now have exactly one session remaining.
		listRes2, err := oauthClient.ListActiveSessions(ctx, &authsvc.ListActiveSessionsRequest{})
		require.NoError(t, err)
		require.Len(t, listRes2.Sessions, 1)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.RevokeAllOtherSessions(ctx, &authsvc.RevokeAllOtherSessionsRequest{})
		assert.Error(t, err)
	})
}

// The two properties this store exists for, end to end. Both were unenforceable before it:
// a revoked session was a flag the reads had to remember to check, and an access token
// stayed usable after the refresh that replaced it.
func TestAuth_SessionEnforcement(T *testing.T) {
	T.Parallel()

	T.Run("a revoked session's token stops authenticating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		jwt := fetchLoginTokenForUserForTest(t, user)
		jwtClient, err := buildAuthedGRPCClientWithBearerToken(jwt)
		require.NoError(t, err)

		_, err = jwtClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)

		_, err = jwtClient.RevokeCurrentSession(ctx, &authsvc.RevokeCurrentSessionRequest{})
		require.NoError(t, err)

		// The token is still perfectly valid and unexpired. The session it names is gone,
		// and that is what the interceptor refuses on.
		_, err = jwtClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.Unauthenticated, st.Code())
	})

	T.Run("a refresh retires the tokens it replaced", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		first, err := unauthedClient.LoginForToken(ctx, &authsvc.LoginForTokenRequest{
			Input: &authsvc.UserLoginInput{
				Username:  user.Username,
				Password:  user.HashedPassword,
				TotpToken: generateTOTPCodeForUserForTest(t, user),
			},
		})
		require.NoError(t, err)
		require.NotEmpty(t, first.Result.RefreshToken)

		firstClient, err := buildAuthedGRPCClientWithBearerToken(first.Result.AccessToken)
		require.NoError(t, err)

		exchanged, err := firstClient.ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{
			RefreshToken: first.Result.RefreshToken,
		})
		require.NoError(t, err)
		require.NotEqual(t, first.Result.AccessToken, exchanged.AccessToken)

		// The new access token works.
		secondClient, err := buildAuthedGRPCClientWithBearerToken(exchanged.AccessToken)
		require.NoError(t, err)
		_, err = secondClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.NoError(t, err)

		// The one it replaced does not, though the session is live and the token unexpired.
		_, err = firstClient.GetAuthStatus(ctx, &authsvc.GetAuthStatusRequest{})
		require.Error(t, err)

		// And the spent refresh token cannot be replayed for a third pair.
		_, err = secondClient.ExchangeToken(ctx, &authsvc.ExchangeTokenRequest{
			RefreshToken: first.Result.RefreshToken,
		})
		require.Error(t, err)
	})
}
