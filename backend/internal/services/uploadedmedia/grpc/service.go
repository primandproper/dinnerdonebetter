package grpc

import (
	uploadedmediamanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads"
)

const (
	o11yName = "uploaded_media_service"
)

var _ uploadedmediasvc.UploadedMediaServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		uploadedmediasvc.UnimplementedUploadedMediaServiceServer
		tracer               tracing.Tracer
		logger               logging.Logger
		uploadedMediaManager uploadedmediamanager.UploadedMediaManager
		uploadManager        uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	uploadedMediaManager uploadedmediamanager.UploadedMediaManager,
	uploadManager uploads.UploadManager,
) uploadedmediasvc.UploadedMediaServiceServer {
	return &serviceImpl{
		logger:               logging.NewNamedLogger(logger, o11yName),
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		uploadedMediaManager: uploadedMediaManager,
		uploadManager:        uploadManager,
	}
}
