package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	httperrors "github.com/primandproper/primitives-go/v2/errors/http"
)

func init() {
	httperrors.RegisterHTTPErrorMapper(authSessionIdentityHTTPMapper{})

	// platform's directory, for the reason its gRPC mapper is registered beside it.
	httperrors.RegisterHTTPErrorMapper(platformidentity.HTTPMapper)

	// And platform's sign-in orchestration, which now owns the refusals this application's
	// login door used to name itself. It is a separate registration from the directory's
	// because they answer about different things: identity refuses a principal read,
	// signin refuses a credential — and signin tells suspended, terminated and unverified
	// apart where the directory has one sentinel for all three.
	httperrors.RegisterHTTPErrorMapper(signin.HTTPMapper)
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
