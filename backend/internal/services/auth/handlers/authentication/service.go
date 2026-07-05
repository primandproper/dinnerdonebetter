package authentication

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/oauth"

	"github.com/primandproper/platform-go/v3/authentication/totp"
	perrors "github.com/primandproper/platform-go/v3/errors"
	"github.com/primandproper/platform-go/v3/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v3/messagequeue/config"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"

	"github.com/go-oauth2/oauth2/v4/server"
)

const (
	serviceName = "auth_service"
)

type (
	// service handles passwords service-wide.
	service struct {
		logger               logging.Logger
		authenticator        authentication.Authenticator
		totpVerifier         totp.Verifier
		tracer               tracing.Tracer
		dataChangesPublisher messagequeue.Publisher
		oauth2Server         *server.Server
		oauthRepo            oauth.Repository
	}
)

// ProvideService builds a new AuthDataService.
func ProvideService(
	ctx context.Context,
	logger logging.Logger,
	cfg *Config,
	authenticator authentication.Authenticator,
	totpVerifier totp.Verifier,
	oauthRepo oauth.Repository,
	identityDataManager identitymanager.IdentityDataManager,
	tracerProvider tracing.TracerProvider,
	publisherProvider messagequeue.PublisherProvider,
	queuesConfig *msgconfig.QueuesConfig,
) (auth.AuthDataService, error) {
	if queuesConfig == nil {
		return nil, perrors.ErrNilInputProvided
	}

	dataChangesPublisher, publisherProviderErr := publisherProvider.ProvidePublisher(ctx, queuesConfig.DataChangesTopicName)
	if publisherProviderErr != nil {
		return nil, fmt.Errorf("setting up %s data changes publisher: %w", serviceName, publisherProviderErr)
	}

	signer, err := cfg.Tokens.ProvideTokenIssuer(logger, tracerProvider)
	if err != nil {
		return nil, fmt.Errorf("creating json web token signer: %w", err)
	}

	manager := ProvideOAuth2ClientManager(logger, tracerProvider, &cfg.OAuth2, oauthRepo)

	svc := &service{
		logger:               logging.NewNamedLogger(logger, serviceName),
		authenticator:        authenticator,
		totpVerifier:         totpVerifier,
		tracer:               tracing.NewNamedTracer(tracerProvider, serviceName),
		dataChangesPublisher: dataChangesPublisher,
		oauth2Server:         ProvideOAuth2ServerImplementation(logger, tracerProvider, identityDataManager, authenticator, signer, manager),
		oauthRepo:            oauthRepo,
	}

	return svc, nil
}
