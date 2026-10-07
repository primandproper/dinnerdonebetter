package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/primitives-go/v2/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	// Only this application's sentinels. platform's directory and sign-in mappers are
	// installed by errormappers.Register, which service.Register calls in every process.
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
	case errors.Is(err, authentication.ErrInvalidTOTPToken):
		return codes.InvalidArgument, true
	case errors.Is(err, sessions.ErrAuthenticationNotFound):
		return codes.Unauthenticated, true
	// This application's impersonation policy refusing an operator, which signin wraps and
	// hands back. It is the operator's grants that are missing, not anything wrong with the
	// server, and signin's own mapper does not know a consumer's sentinel.
	case errors.Is(err, authentication.ErrImpersonationNotPermitted):
		return codes.PermissionDenied, true
	default:
		return codes.Unknown, false
	}
}
