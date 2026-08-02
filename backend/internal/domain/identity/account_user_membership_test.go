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
