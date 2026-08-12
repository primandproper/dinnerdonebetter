package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v10/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(authSessionIdentityGRPCMapper{})
}

type authSessionIdentityGRPCMapper struct{}

func (authSessionIdentityGRPCMapper) Map(err error) (code codes.Code, ok bool) {
	if err == nil {
		return codes.Unknown, false
	}
	switch {
	case errors.Is(err, authentication.ErrTOTPRequired):
		return codes.Unauthenticated, true
	case errors.Is(err, authentication.ErrInvalidTOTPToken),
		errors.Is(err, authentication.ErrPasswordDoesNotMatch):
		return codes.InvalidArgument, true
	case errors.Is(err, sessions.ErrAuthenticationNotFound):
		return codes.Unauthenticated, true
	case errors.Is(err, identitymanager.ErrInvalidIDProvided),
		errors.Is(err, identitymanager.ErrNilInputParameter),
		errors.Is(err, identitymanager.ErrEmptyInputProvided):
		return codes.InvalidArgument, true
	default:
		return codes.Unknown, false
	}
}
