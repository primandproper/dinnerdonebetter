package authentication

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	mockauthn "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/mock"

	"github.com/primandproper/platform-go/v9/authentication/totp"
	mocktotp "github.com/primandproper/platform-go/v9/authentication/totp/mock"

	"github.com/stretchr/testify/assert"
)

func TestAuthenticationService_validateLogin(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, helper.exampleUser.HashedPassword, hash)
				assert.Equal(t, helper.exampleLoginInput.Password, password)
				return true, nil
			},
		}
		helper.service.authenticator = authenticator
		helper.service.totpVerifier = &mocktotp.VerifierMock{
			VerifyFunc: func(_ context.Context, _, _ string) error { return nil },
		}

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.True(t, actual)
		assert.NoError(t, err)

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
	})

	T.Run("with invalid two factor code", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, helper.exampleUser.HashedPassword, hash)
				assert.Equal(t, helper.exampleLoginInput.Password, password)
				return true, nil
			},
		}
		helper.service.authenticator = authenticator

		// Force the TOTP path: user has a verified 2FA secret and the verifier returns ErrInvalidCode.
		now := time.Now()
		helper.exampleUser.TwoFactorSecretVerifiedAt = &now
		helper.service.totpVerifier = &mocktotp.VerifierMock{
			VerifyFunc: func(_ context.Context, _, _ string) error { return totp.ErrInvalidCode },
		}

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)
		assert.ErrorIs(t, err, totp.ErrInvalidCode)

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
	})

	T.Run("with error returned from validator", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		expectedErr := errors.New("arbitrary")

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, helper.exampleUser.HashedPassword, hash)
				assert.Equal(t, helper.exampleLoginInput.Password, password)
				return false, expectedErr
			},
		}
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
	})

	T.Run("with invalid login", func(t *testing.T) {
		t.Parallel()

		helper := buildTestHelper(t)

		authenticator := &mockauthn.AuthenticatorMock{
			PasswordMatchesFunc: func(_ context.Context, hash, password string) (bool, error) {
				assert.Equal(t, helper.exampleUser.HashedPassword, hash)
				assert.Equal(t, helper.exampleLoginInput.Password, password)
				return false, nil
			},
		}
		helper.service.authenticator = authenticator

		actual, err := helper.service.validateLogin(helper.ctx, helper.exampleUser, helper.exampleLoginInput)
		assert.False(t, actual)
		assert.Error(t, err)
		assert.ErrorIs(t, err, authentication.ErrPasswordDoesNotMatch)

		assert.Len(t, authenticator.PasswordMatchesCalls(), 1)
	})
}
