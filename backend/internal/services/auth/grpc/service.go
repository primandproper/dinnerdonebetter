package grpc

import (
	authentication2 "github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers"
	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	"github.com/primandproper/platform-go/v9/encoding"
	"github.com/primandproper/platform-go/v9/featureflags"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
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
	passkeyJSONEncoder := encoding.NewServerEncoderDecoder(encoding.ContentTypeJSON, encoding.WithLogger(logger), encoding.WithTracerProvider(tracerProvider))

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
