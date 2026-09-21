package grpc

import (
	"context"

	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"
	// Registers the upload registry's sentinels with the gRPC error mapper, so an
	// archived object reads as NotFound rather than as a server fault.
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/uploadedmedia/errors"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/platform-go/v14/metering"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"
)

const (
	o11yName = "uploaded_media_service"
)

var _ uploadedmediasvc.UploadedMediaServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		uploadedmediasvc.UnimplementedUploadedMediaServiceServer
		tracer tracing.Tracer
		db     database.Client
		logger logging.Logger

		// registry holds the rows, uploadManager holds the bytes. They stay separate
		// because mediaregistry.StoreAndRecord is a free function over the two rather than
		// a method on either: an object that arrived through a signed URL was stored
		// by somebody else and still needs a row.
		registry      mediaregistry.Store
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
	db database.Client,
	registryStore mediaregistry.Store,
	uploadManager uploads.UploadManager,
	usageRecorder metering.Recorder,
) uploadedmediasvc.UploadedMediaServiceServer {
	return &serviceImpl{
		logger:        logging.NewNamedLogger(logger, o11yName),
		tracer:        tracing.NewNamedTracer(tracerProvider, o11yName),
		db:            db,
		registry:      registryStore,
		uploadManager: uploadManager,
		usageRecorder: usageRecorder,
	}
}

// inTransaction runs one store write on a transaction of its own.
//
// As of platform-go v14 a store holds no database handle: a write takes the
// caller's database.Tx. Every write in this service is a single store call, so
// each gets one transaction — which is exactly what the store opened for itself
// before the caller was required to supply it. A handler that ever writes twice
// should take one transaction across both rather than call this twice.
func inTransaction[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
