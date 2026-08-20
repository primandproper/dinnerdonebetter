package authentication

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth"

	"github.com/primandproper/platform-go/v12/authentication/oauth2server"
	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
)

const (
	serviceName = "auth_service"
)

type (
	// service carries the OAuth 2.1 authorization server's HTTP surface across the
	// auth.AuthDataService interface the API router builds against.
	//
	// It is thin, and got thinner with this change: the credential checks, the token store
	// adapters and the hand-rolled RFC 7009 revocation endpoint it used to hold are all the
	// authorization server's now. What is left is four handlers and the logger the metadata
	// one writes through.
	service struct {
		logger       logging.Logger
		oauth2Server *oauth2server.Server
	}
)

// ProvideService builds a new AuthDataService.
//
// The tracer provider is taken and not held: the authorization server does its own
// instrumentation, and every handler here is a delegation to it. A span wrapped around that
// would record this function's own duration and nothing else.
func ProvideService(
	logger logging.Logger,
	oauth2Server *oauth2server.Server,
	_ tracing.Provider,
) (auth.AuthDataService, error) {
	return &service{
		logger:       logging.NewNamedLogger(logger, serviceName),
		oauth2Server: oauth2Server,
	}, nil
}
