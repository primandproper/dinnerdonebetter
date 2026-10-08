package authentication

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/primitives-go/v2/authentication/oauth2server"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

type (
	// service carries the OAuth 2.1 authorization server's HTTP surface across the
	// auth.AuthDataService interface the API router builds against.
	//
	// It is thin, and got thinner with this change: the credential checks, the token store
	// adapters and the hand-rolled RFC 7009 revocation endpoint it used to hold are all the
	// authorization server's now. What is left is four handlers, each a delegation.
	service struct {
		oauth2Server *oauth2server.Server
	}
)

// ProvideService builds a new AuthDataService.
//
// The logger and tracer provider are taken and not held: the authorization server does its own
// logging and instrumentation, and every handler here is a delegation to it. A span wrapped
// around that would record this function's own duration and nothing else.
func ProvideService(
	_ logging.Logger,
	oauth2Server *oauth2server.Server,
	_ tracing.Provider,
) (auth.AuthDataService, error) {
	return &service{
		oauth2Server: oauth2Server,
	}, nil
}
