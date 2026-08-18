package grpc

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"
	uploadedmediamanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	identitysvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/identity"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/types"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	"github.com/primandproper/platform-go/v11/uploads"
)

const (
	o11yName = "identity_service"
)

var _ identitysvc.IdentityServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		identitysvc.UnimplementedIdentityServiceServer
		tracer               tracing.Tracer
		logger               logging.Logger
		identityDataManager  manager.IdentityDataManager
		uploadedMediaManager uploadedmediamanager.UploadedMediaManager
		uploadManager        uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	identityDataManager manager.IdentityDataManager,
	uploadedMediaManager uploadedmediamanager.UploadedMediaManager,
	uploadManager uploads.UploadManager,
) identitysvc.IdentityServiceServer {
	return &serviceImpl{
		logger:               logging.NewNamedLogger(logger, o11yName),
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		identityDataManager:  identityDataManager,
		uploadedMediaManager: uploadedMediaManager,
		uploadManager:        uploadManager,
	}
}

func (s *serviceImpl) buildResponseDetails(ctx context.Context, span tracing.Span) *types.ResponseDetails {
	out := &types.ResponseDetails{}
	if span != nil {
		out.TraceId = span.SpanContext().TraceID().String()
	}

	// Response details are built for unauthenticated routes too, so absence is expected here.
	out.CurrentAccountId = sessions.FromContext(ctx).GetActiveAccountID()

	return out
}
