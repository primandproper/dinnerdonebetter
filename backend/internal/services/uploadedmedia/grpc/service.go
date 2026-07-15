package grpc

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	uploadedmediamanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"
	"github.com/primandproper/platform-go/v4/uploads"
)

const (
	o11yName = "uploaded_media_service"
)

var _ uploadedmediasvc.UploadedMediaServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		uploadedmediasvc.UnimplementedUploadedMediaServiceServer
		tracer                    tracing.Tracer
		logger                    logging.Logger
		sessionContextDataFetcher func(context.Context) (*sessions.ContextData, error)
		uploadedMediaManager      uploadedmediamanager.UploadedMediaManager
		uploadManager             uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	uploadedMediaManager uploadedmediamanager.UploadedMediaManager,
	uploadManager uploads.UploadManager,
) uploadedmediasvc.UploadedMediaServiceServer {
	return &serviceImpl{
		logger:                    logging.NewNamedLogger(logger, o11yName),
		tracer:                    tracing.NewNamedTracer(tracerProvider, o11yName),
		sessionContextDataFetcher: sessions.FetchContextDataFromContext,
		uploadedMediaManager:      uploadedMediaManager,
		uploadManager:             uploadManager,
	}
}
