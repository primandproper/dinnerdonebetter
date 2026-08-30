package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPasswordResetTokenCreationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &PasswordResetTokenCreationRequestInput{
			EmailAddress: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestPasswordResetTokenRedemptionRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &PasswordResetTokenRedemptionRequestInput{
			Token:       t.Name(),
			NewPassword: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestUsernameReminderRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UsernameReminderRequestInput{
			EmailAddress: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
