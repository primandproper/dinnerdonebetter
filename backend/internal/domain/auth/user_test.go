package auth

import (
	"errors"
	"testing"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/stretchr/testify/assert"
)

func TestTOTPSecretVerificationInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &TOTPSecretVerificationInput{
			UserID:    "123",
			TOTPToken: "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestUserLoginInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserLoginInput{
			Username:  t.Name(),
			Password:  t.Name(),
			TOTPToken: "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("without token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserLoginInput{
			Username: t.Name(),
			Password: t.Name(),
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})

	T.Run("with invalid token", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserLoginInput{
			Username:  t.Name(),
			Password:  t.Name(),
			TOTPToken: "not_real",
		}

		err := x.ValidateWithContext(ctx)
		var validationErr validation.Errors
		if errors.As(err, &validationErr) {
			assert.Error(t, validationErr["totpToken"])
		}

		assert.Error(t, err)
	})

	T.Run("with desired account ID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserLoginInput{
			Username:         t.Name(),
			Password:         t.Name(),
			TOTPToken:        "123456",
			DesiredAccountID: "550e8400-e29b-41d4-a716-446655440000",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestPasswordUpdateInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &PasswordUpdateInput{
			NewPassword:     "new_password",
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx, 1))
	})

	T.Run("with identical passwords", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &PasswordUpdateInput{
			NewPassword:     t.Name(),
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.Error(t, x.ValidateWithContext(ctx, 1))
	})
}

func TestTOTPSecretRefreshInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &TOTPSecretRefreshInput{
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestEmailAddressVerificationRequestInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &EmailAddressVerificationRequestInput{
			Token: "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestUsernameUpdateInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UsernameUpdateInput{
			NewUsername:     t.Name(),
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}

func TestUserEmailAddressUpdateInput_ValidateWithContext(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		x := &UserEmailAddressUpdateInput{
			NewEmailAddress: t.Name(),
			CurrentPassword: t.Name(),
			TOTPToken:       "123456",
		}

		assert.NoError(t, x.ValidateWithContext(ctx))
	})
}
