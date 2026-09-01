package grpc

import (
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	// Registers the upload registry's sentinels with the gRPC error mapper, so an
	// archived object reads as NotFound rather than as a server fault.
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/errors"

	"github.com/primandproper/platform-go/v13/metering"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

const (
	o11yName = "uploaded_media_service"
)

var _ uploadedmediasvc.UploadedMediaServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		uploadedmediasvc.UnimplementedUploadedMediaServiceServer
		tracer tracing.Tracer
		logger logging.Logger

		// registry holds the rows, uploadManager holds the bytes. They stay separate
		// because registry.StoreAndRecord is a free function over the two rather than
		// a method on either: an object that arrived through a signed URL was stored
		// by somebody else and still needs a row.
		registry      registry.Store
		uploadManager uploads.UploadManager

		// usageRecorder counts bytes accepted by Upload. It is a Recorder rather than an
		// Enforcer deliberately: nothing here refuses an upload for being over a limit, and
		// holding the interface that could would invite it to start.
		usageRecorder metering.Recorder
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	registryStore registry.Store,
	uploadManager uploads.UploadManager,
	usageRecorder metering.Recorder,
) uploadedmediasvc.UploadedMediaServiceServer {
	return &serviceImpl{
		logger:        logging.NewNamedLogger(logger, o11yName),
		tracer:        tracing.NewNamedTracer(tracerProvider, o11yName),
		registry:      registryStore,
		uploadManager: uploadManager,
		usageRecorder: usageRecorder,
	}
}
