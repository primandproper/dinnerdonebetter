package manager

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDataPrivacyManager registers the data privacy manager with the injector.
func RegisterDataPrivacyManager(i do.Injector) {
	do.Provide[DataPrivacyManager](i, func(i do.Injector) (DataPrivacyManager, error) {
		return NewDataPrivacyManager(
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[dataprivacy.Repository](i),
		), nil
	})
}
