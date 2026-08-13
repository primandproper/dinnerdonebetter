package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"

	httperrors "github.com/primandproper/platform-go/v10/errors/http"
)

func init() {
	httperrors.RegisterHTTPErrorMapper(authSessionIdentityHTTPMapper{})
}

type authSessionIdentityHTTPMapper struct{}

func (authSessionIdentityHTTPMapper) Map(err error) (code httperrors.ErrorCode, msg string, ok bool) {
	if err == nil {
		return "", "", false
	}
	switch {
	case errors.Is(err, authentication.ErrInvalidTOTPToken),
		errors.Is(err, authentication.ErrPasswordDoesNotMatch):
		return httperrors.ErrValidatingRequestInput, "invalid credentials", true
	case errors.Is(err, sessions.ErrAuthenticationNotFound):
		return httperrors.ErrFetchingSessionContextData, "session not found", true
	case errors.Is(err, identitymanager.ErrInvalidIDProvided),
		errors.Is(err, identitymanager.ErrNilInputParameter),
		errors.Is(err, identitymanager.ErrEmptyInputProvided):
		return httperrors.ErrValidatingRequestInput, "invalid input", true
	default:
		return "", "", false
	}
}
