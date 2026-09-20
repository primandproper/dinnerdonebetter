package grpc

import (
	authentication2 "github.com/primandproper/dinnerdonebetter/backend/internal/authentication"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/auth/managers"
	authsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/auth"

	platformidentity "github.com/primandproper/platform-go/v14/identity"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/encoding"
	"github.com/primandproper/primitives-go/v2/featureflags"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
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
		directory             platformidentity.Store
		db                    database.Client
		authenticationManager authentication2.Manager
		authManager           managers.AuthManagerInterface
		featureFlagManager    featureflags.FeatureFlagManager
		passkeyService        *webauthn.Service
		jsonEncoder           encoding.ServerEncoderDecoder
	}
)

func NewAuthService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	directory platformidentity.Store,
	db database.Client,
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
		directory:             directory,
		db:                    db,
		authManager:           authManager,
		authenticationManager: authenticationManager,
		featureFlagManager:    featureFlagManager,
		passkeyService:        passkeyService,
		jsonEncoder:           passkeyJSONEncoder,
	}
}
