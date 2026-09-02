package identity

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"

	"github.com/stretchr/testify/assert"
)

func TestTransferAccountOwnershipInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &AccountOwnershipTransferInput{
			CurrentOwner: "123",
			NewOwner:     "321",
			Reason:       t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with account admin", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ModifyUserPermissionsInput{
			NewRole: authorization.AccountAdminRoleName,
			Reason:  t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	// The escalation this bound exists for. The role named here is written into an
	// account-scoped assignment and resolved within that account, so a service role
	// accepted here would hand a member the service-wide closure inside an account
	// somebody merely administers.
	T.Run("with a service role", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		for _, role := range []string{
			authorization.ServiceAdminRoleName,
			authorization.ServiceDataAdminRoleName,
			authorization.ServiceUserRoleName,
		} {
			x := &ModifyUserPermissionsInput{
				NewRole: role,
				Reason:  t.Name(),
			}

			assert.Error(t, x.ValidateWithContext(ctx), role)
		}
	})

	T.Run("with a role nothing declares", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ModifyUserPermissionsInput{
			NewRole: "arbitrary",
			Reason:  t.Name(),
		}

		assert.Error(t, x.ValidateWithContext(ctx))
	})
}

func TestModifyUserPermissionsInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &ModifyUserPermissionsInput{
			NewRole: authorization.AccountMemberRole.String(),
			Reason:  t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
