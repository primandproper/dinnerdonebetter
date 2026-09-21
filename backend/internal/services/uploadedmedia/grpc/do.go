package grpc

import (
	uploadedmediasvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/platform-go/v14/metering"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"

	"github.com/samber/do/v2"
)

// RegisterUploadedMediaService registers the uploaded media gRPC service with the injector.
func RegisterUploadedMediaService(i do.Injector) {
	do.Provide[UploadedMediaMethodPermissions](i, func(i do.Injector) (UploadedMediaMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[uploadedmediasvc.UploadedMediaServiceServer](i, func(i do.Injector) (uploadedmediasvc.UploadedMediaServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[mediaregistry.Store](i),
			do.MustInvoke[uploads.UploadManager](i),
			do.MustInvoke[metering.Recorder](i),
		), nil
	})
}
