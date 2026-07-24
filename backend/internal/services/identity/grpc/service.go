package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	uploadedmediamanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	identitysvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/types"

	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"
	"github.com/primandproper/platform-go/v6/uploads"
)

const (
	o11yName = "identity_service"
)

var _ identitysvc.IdentityServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		identitysvc.UnimplementedIdentityServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(ctx context.Context) (*sessions.ContextData, error)
		identityDataManager       manager.IdentityDataManager
		uploadedMediaManager      uploadedmediamanager.UploadedMediaManager
		uploadManager             uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	sessionContextDataFetcher func(ctx context.Context) (*sessions.ContextData, error),
	identityDataManager manager.IdentityDataManager,
	uploadedMediaManager uploadedmediamanager.UploadedMediaManager,
	uploadManager uploads.UploadManager,
) identitysvc.IdentityServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessionContextDataFetcher,
		identityDataManager:       identityDataManager,
		uploadedMediaManager:      uploadedMediaManager,
		uploadManager:             uploadManager,
	}
}

func (s *serviceImpl) buildResponseDetails(ctx context.Context, span tracing.Span) *types.ResponseDetails {
	out := &types.ResponseDetails{}
	if span != nil {
		out.TraceId = span.SpanContext().TraceID().String()
	}

	if sessionContextData, err := s.sessionContextDataFetcher(ctx); err == nil && sessionContextData != nil {
		out.CurrentAccountId = sessionContextData.GetActiveAccountID()
	}

	return out
}
