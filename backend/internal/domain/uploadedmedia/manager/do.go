package manager

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterUploadedMediaManager registers the uploaded media manager with the injector.
func RegisterUploadedMediaManager(i do.Injector) {
	do.Provide[UploadedMediaManager](i, func(i do.Injector) (UploadedMediaManager, error) {
		return NewUploadedMediaDataManager(
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[uploadedmedia.Repository](i),
		), nil
	})
}
