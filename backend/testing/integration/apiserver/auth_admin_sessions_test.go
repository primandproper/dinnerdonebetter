package integration

import (
	"testing"

	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuth_AdminListSessionsForUser(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		res, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
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

	T.Run("empty user ID returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: "",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("non-admin user is rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, testClient := createUserAndClientForTest(t)

		_, err := testClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: "anything",
		})
		assert.Error(t, err)
	})
}

func TestAuth_AdminRevokeUserSession(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		// Log in again to create a second session.
		_ = fetchLoginTokenForUserForTest(t, user)

		// Admin lists sessions for this user.
		listRes, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(listRes.Sessions), 2)

		// Pick one session to revoke.
		targetSessionID := listRes.Sessions[0].Id

		// Revoke it.
		_, err = adminClient.AdminRevokeUserSession(ctx, &authsvc.AdminRevokeUserSessionRequest{
			UserId:    user.ID,
			SessionId: targetSessionID,
		})
		require.NoError(t, err)

		// Verify it's gone.
		listRes2, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		for _, sess := range listRes2.Sessions {
			assert.NotEqual(t, targetSessionID, sess.Id, "revoked session should not appear in active list")
		}
	})

	T.Run("empty user ID returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.AdminRevokeUserSession(ctx, &authsvc.AdminRevokeUserSessionRequest{
			UserId:    "",
			SessionId: "anything",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("empty session ID returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.AdminRevokeUserSession(ctx, &authsvc.AdminRevokeUserSessionRequest{
			UserId:    nonexistentID,
			SessionId: "",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("non-admin user is rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.AdminRevokeUserSession(ctx, &authsvc.AdminRevokeUserSessionRequest{
			UserId:    nonexistentID,
			SessionId: "anything",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.AdminRevokeUserSession(ctx, &authsvc.AdminRevokeUserSessionRequest{
			UserId:    "anything",
			SessionId: "anything",
		})
		assert.Error(t, err)
	})
}

func TestAuth_AdminRevokeAllUserSessions(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, _ := createUserAndClientForTest(t)

		// Create additional sessions.
		_ = fetchLoginTokenForUserForTest(t, user)
		_ = fetchLoginTokenForUserForTest(t, user)

		// Should have multiple sessions.
		listRes, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(listRes.Sessions), 3, "expected multiple sessions before revoking")

		// Revoke all sessions.
		_, err = adminClient.AdminRevokeAllUserSessions(ctx, &authsvc.AdminRevokeAllUserSessionsRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)

		// Should now have zero sessions.
		listRes2, err := adminClient.AdminListSessionsForUser(ctx, &authsvc.AdminListSessionsForUserRequest{
			UserId: user.ID,
		})
		require.NoError(t, err)
		require.Empty(t, listRes2.Sessions, "expected all sessions to be revoked")
	})

	T.Run("empty user ID returns error", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.AdminRevokeAllUserSessions(ctx, &authsvc.AdminRevokeAllUserSessionsRequest{
			UserId: "",
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("non-admin user is rejected", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		_, err := testClient.AdminRevokeAllUserSessions(ctx, &authsvc.AdminRevokeAllUserSessionsRequest{
			UserId: nonexistentID,
		})
		require.Error(t, err)
		st, ok := status.FromError(err)
		require.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		unauthedClient := buildUnauthenticatedGRPCClientForTest(t)
		_, err := unauthedClient.AdminRevokeAllUserSessions(ctx, &authsvc.AdminRevokeAllUserSessionsRequest{
			UserId: "anything",
		})
		assert.Error(t, err)
	})
}
