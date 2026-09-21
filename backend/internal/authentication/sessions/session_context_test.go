package sessions

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/primandproper/primitives-go/v2/identifiers"

	"github.com/stretchr/testify/assert"
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
		// Not nil. Every caller in the tree calls IsServiceAdmin or HasPermission on
		// this immediately, so a nil interface would turn "read through a nil
		// ContextData without a preceding nil check" into a panic on exactly the
		// unauthenticated request the nil-safety is for. The account-side checker below
		// has always answered this way; the service side was the outlier.
		require.NotNil(t, x.GetServicePermissions())
		assert.False(t, x.GetServicePermissions().IsServiceAdmin())
		assert.False(t, x.GetServicePermissions().HasPermission(authorization.PermissionReadUsers))

		require.NotNil(t, x.ServiceRolePermissionChecker())
		assert.False(t, x.ServiceRolePermissionChecker().IsServiceAdmin())

		require.NotNil(t, x.AccountRolePermissionsChecker())
	})

	T.Run("a session carrying no checker answers as an anonymous one", func(t *testing.T) {
		t.Parallel()

		// The shape a hand-built ContextData has, and the one a session whose policy
		// resolution failed would have: present, but with nothing in Requester.
		x := &ContextData{Requester: RequesterInfo{UserID: "someone"}}

		require.NotNil(t, x.GetServicePermissions())
		assert.False(t, x.GetServicePermissions().IsServiceAdmin())
	})
}
