package errors

import (
	"errors"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"

	"github.com/primandproper/platform-go/v14/authentication/signin"
	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/errors/grpc"

	"google.golang.org/grpc/codes"
)

func init() {
	grpc.RegisterGRPCErrorMapper(authSessionIdentityGRPCMapper{})

	// platform's directory maps its own refusals, and it exports the mapper rather than
	// registering it: which registry a consumer keeps its mappers in is the consumer's,
	// and a package that registered itself on import would be deciding that for them.
	//
	// It is registered here rather than beside the identity surface because a mapper is
	// not a surface's — the same errors reach the wire from the authentication
	// interceptor, which resolves a principal on every request and is not part of that
	// service at all.
	grpc.RegisterGRPCErrorMapper(platformidentity.GRPCMapper)

	// And platform's sign-in orchestration, which now owns the refusals this application's
	// login door used to name itself. It is a separate registration from the directory's
	// because they answer about different things: identity refuses a principal read,
	// signin refuses a credential — and signin tells suspended, terminated and unverified
	// apart where the directory has one sentinel for all three.
	grpc.RegisterGRPCErrorMapper(signin.GRPCMapper)
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
	default:
		return codes.Unknown, false
	}
}
