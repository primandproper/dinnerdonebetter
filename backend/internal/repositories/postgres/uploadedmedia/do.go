package uploadedmedia

import (
	"github.com/primandproper/platform-go/v15/mediaregistry"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterUploadedMediaRepository registers the upload registry with the injector.
func RegisterUploadedMediaRepository(i do.Injector) {
	do.Provide[mediaregistry.Store](i, func(i do.Injector) (mediaregistry.Store, error) {
		return ProvideUploadedMediaRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*platformrecording.Recorder](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
