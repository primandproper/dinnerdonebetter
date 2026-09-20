package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	httperrors "github.com/primandproper/primitives-go/v2/errors/http"
)

func init() {
	httperrors.RegisterHTTPErrorMapper(authSessionIdentityHTTPMapper{})

	// platform's directory, for the reason its gRPC mapper is registered beside it.
	httperrors.RegisterHTTPErrorMapper(platformidentity.HTTPMapper)
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
	default:
		return "", "", false
	}
}
