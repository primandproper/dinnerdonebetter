package grpc

import (
	"context"
	"errors"

	authentication2 "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/auth/managers"
	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	authsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v4/encoding"
	"github.com/primandproper/platform-go/v4/featureflags"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

const (
	o11yName = "auth_service"
)

var _ authsvc.AuthServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		authsvc.UnimplementedAuthServiceServer
		tracer                tracing.Tracer
		logger                logging.Logger
		identityDataManager   identitymanager.IdentityDataManager
		authenticationManager authentication2.Manager
		authManager           managers.AuthManagerInterface
		featureFlagManager    featureflags.FeatureFlagManager
		passkeyService        *webauthn.Service
		jsonEncoder           encoding.ServerEncoderDecoder
	}
)

func NewAuthService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	identityDataManager identitymanager.IdentityDataManager,
	authManager managers.AuthManagerInterface,
	authenticationManager authentication2.Manager,
	featureFlagManager featureflags.FeatureFlagManager,
	passkeyService *webauthn.Service,
) authsvc.AuthServiceServer {
	// Passkey options are always JSON; create a dedicated encoder rather than relying on
	// a potentially non-JSON encoder from wire.
	passkeyJSONEncoder := encoding.NewServerEncoderDecoder(logger, tracerProvider, encoding.ContentTypeJSON)

	return &serviceImpl{
		logger:                logging.NewNamedLogger(logger, o11yName),
		tracer:                tracing.NewNamedTracer(tracerProvider, o11yName),
		identityDataManager:   identityDataManager,
		authManager:           authManager,
		authenticationManager: authenticationManager,
		featureFlagManager:    featureFlagManager,
		passkeyService:        passkeyService,
		jsonEncoder:           passkeyJSONEncoder,
	}
}

func (s *serviceImpl) fetchSessionContext(ctx context.Context) (*sessions.ContextData, error) {
	sessionContext, ok := ctx.Value(sessions.SessionContextDataKey).(*sessions.ContextData)
	if !ok {
		return nil, errors.New("session context not found")
	}

	return sessionContext, nil
}
