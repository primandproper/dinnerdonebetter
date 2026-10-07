package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	httperrors "github.com/primandproper/primitives-go/v2/errors/http"
)

func init() {
	// Only this application's sentinels, for the reason the gRPC mapper gives.
	httperrors.RegisterHTTPErrorMapper(authSessionIdentityHTTPMapper{})
}

type authSessionIdentityHTTPMapper struct{}

func (authSessionIdentityHTTPMapper) Map(err error) (code httperrors.ErrorCode, msg string, ok bool) {
	if err == nil {
		return "", "", false
	}
	switch {
	case errors.Is(err, authentication.ErrInvalidTOTPToken):
		return httperrors.ErrValidatingRequestInput, "invalid credentials", true
	case errors.Is(err, sessions.ErrAuthenticationNotFound):
		return httperrors.ErrFetchingSessionContextData, "session not found", true
	default:
		return "", "", false
	}
}
