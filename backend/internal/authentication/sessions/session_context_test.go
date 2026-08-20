package sessions

import (
	"context"
	"testing"

	"github.com/primandproper/platform-go/v12/identifiers"

	"github.com/stretchr/testify/require"
)

func TestFromContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := &ContextData{ActiveAccountID: identifiers.New()}
		ctx := context.WithValue(t.Context(), SessionContextDataKey, expected)

		require.Same(t, expected, FromContext(ctx))
	})

	T.Run("missing data", func(t *testing.T) {
		t.Parallel()

		require.Nil(t, FromContext(t.Context()))
	})

	T.Run("wrong type under the key", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(t.Context(), SessionContextDataKey, "not session context data")

		require.Nil(t, FromContext(ctx))
	})
}

func TestRequireFromContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		expected := &ContextData{ActiveAccountID: identifiers.New()}
		ctx := context.WithValue(t.Context(), SessionContextDataKey, expected)

		actual, err := RequireFromContext(ctx)
		require.NoError(t, err)
		require.Same(t, expected, actual)
	})

	T.Run("missing data", func(t *testing.T) {
		t.Parallel()

		actual, err := RequireFromContext(t.Context())
		require.ErrorIs(t, err, ErrAuthenticationNotFound)
		require.Nil(t, actual)
	})

	T.Run("explicit nil under the key", func(t *testing.T) {
		t.Parallel()

		ctx := context.WithValue(t.Context(), SessionContextDataKey, (*ContextData)(nil))

		actual, err := RequireFromContext(ctx)
		require.ErrorIs(t, err, ErrAuthenticationNotFound)
		require.Nil(t, actual)
	})
}

func TestContextData_gettersAreNilSafe(T *testing.T) {
	T.Parallel()

	T.Run("populated", func(t *testing.T) {
		t.Parallel()

		userID, accountID, sessionID := identifiers.New(), identifiers.New(), identifiers.New()
		x := &ContextData{
			ActiveAccountID: accountID,
			SessionID:       sessionID,
			Requester: RequesterInfo{
				UserID:       userID,
				EmailAddress: "requester@example.com",
				Username:     "requester",
			},
		}

		require.Equal(t, userID, x.GetUserID())
		require.Equal(t, accountID, x.GetActiveAccountID())
		require.Equal(t, sessionID, x.GetSessionID())
		require.Equal(t, "requester@example.com", x.GetEmailAddress())
		require.Equal(t, "requester", x.GetUsername())
	})

	T.Run("nil receiver yields zero values", func(t *testing.T) {
		t.Parallel()

		var x *ContextData

		require.Empty(t, x.GetUserID())
		require.Empty(t, x.GetActiveAccountID())
		require.Empty(t, x.GetSessionID())
		require.Empty(t, x.GetEmailAddress())
		require.Empty(t, x.GetUsername())
		require.Nil(t, x.GetServicePermissions())
		require.Nil(t, x.ServiceRolePermissionChecker())
		require.NotNil(t, x.AccountRolePermissionsChecker())
	})
}
