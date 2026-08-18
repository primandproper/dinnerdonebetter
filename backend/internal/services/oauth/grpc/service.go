package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/oauth/manager"
	oauthsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/oauth"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
)

const (
	o11yName = "oauth_service"
)

var _ oauthsvc.OAuthServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		oauthsvc.UnimplementedOAuthServiceServer
		tracer           tracing.Tracer
		logger           logging.Logger
		oauthDataManager manager.OAuth2Manager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	oauthDataManager manager.OAuth2Manager,
) oauthsvc.OAuthServiceServer {
	return &serviceImpl{
		logger:           logging.NewNamedLogger(logger, o11yName),
		tracer:           tracing.NewNamedTracer(tracerProvider, o11yName),
		oauthDataManager: oauthDataManager,
	}
}
