package authentication

import (
	"context"
	"errors"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	identitykeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/keys"

	"github.com/primandproper/platform-go/v4/authentication/totp"
	"github.com/primandproper/platform-go/v4/observability"
)

// validateLogin takes login information and returns whether the login is valid.
// In the event that there's an error, this function will return false and the error.
func (s *service) validateLogin(ctx context.Context, user *types.User, loginInput *auth.UserLoginInput) (bool, error) {
	ctx, span := s.tracer.StartSpan(ctx)
	defer span.End()

	// alias the relevant data.
	logger := s.logger.WithValue(identitykeys.UsernameKey, user.Username)

	// verify the password first; a non-match returns (false, nil) from platform's Authenticator.
	matches, err := s.authenticator.PasswordMatches(ctx, user.HashedPassword, loginInput.Password)
	if err != nil {
		return false, observability.PrepareError(err, span, "validating password")
	}
	if !matches {
		return false, authentication.ErrPasswordDoesNotMatch
	}

	// verify TOTP separately when required.
	if user.TwoFactorSecretVerifiedAt != nil {
		if verifyErr := s.totpVerifier.Verify(ctx, user.TwoFactorSecret, loginInput.TOTPToken); verifyErr != nil {
			if errors.Is(verifyErr, totp.ErrCodeRequired) || errors.Is(verifyErr, totp.ErrInvalidCode) {
				return false, verifyErr
			}
			return false, observability.PrepareError(verifyErr, span, "verifying TOTP code")
		}
	}

	logger.Debug("login validated")

	return true, nil
}
