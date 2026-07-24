package dbcleaner

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops"

	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/metrics"
	"github.com/primandproper/platform-go/v6/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterDBCleaner registers the DB cleaner with the injector.
func RegisterDBCleaner(i do.Injector) {
	do.Provide[*Job](i, func(i do.Injector) (*Job, error) {
		return NewDBCleaner(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[internalops.InternalOpsDataManager](i),
		)
	})
}
