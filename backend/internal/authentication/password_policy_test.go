package authentication

import (
	"context"
	"errors"
	"testing"

	"github.com/primandproper/platform-go/v15/authentication/signin"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// changeTo is a change to candidate on an account whose current password is current.
func changeTo(current, candidate string) *signin.PasswordChange {
	return &signin.PasswordChange{
		NewPassword: candidate,
		MatchesCurrent: func(_ context.Context, password string) (bool, error) {
			return password == current, nil
		},
	}
}

func TestAccountPasswordPolicy(T *testing.T) {
	T.Parallel()

	T.Run("admits a new password", func(t *testing.T) {
		t.Parallel()

		current := gofakeit.Password(true, true, true, true, false, 32)

		assert.NoError(t, AccountPasswordPolicy(t.Context(), changeTo(current, current+gofakeit.Password(true, true, true, true, false, 8))))
	})

	T.Run("refuses the password the account already holds", func(t *testing.T) {
		t.Parallel()

		current := gofakeit.Password(true, true, true, true, false, 32)

		assert.ErrorIs(t, AccountPasswordPolicy(t.Context(), changeTo(current, current)), ErrPasswordUnchanged)
	})

	T.Run("refuses when the comparison cannot be made", func(t *testing.T) {
		t.Parallel()

		expected := errors.New(gofakeit.Sentence())
		change := &signin.PasswordChange{
			NewPassword:    gofakeit.Password(true, true, true, true, false, 32),
			MatchesCurrent: func(context.Context, string) (bool, error) { return false, expected },
		}

		err := AccountPasswordPolicy(t.Context(), change)
		require.Error(t, err)
		assert.ErrorIs(t, err, expected)
	})
}
