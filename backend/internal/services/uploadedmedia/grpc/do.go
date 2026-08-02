package grpc

import (
	uploadedmediamanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	uploadedmediasvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/uploaded_media"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads"

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
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[uploadedmediamanager.UploadedMediaManager](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
