package internalops

import (
	domaininternalops "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/internalops"

	"github.com/primandproper/platform-go/v3/database"
	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterInternalOpsRepository registers the internal ops repository with the injector.
func RegisterInternalOpsRepository(i do.Injector) {
	do.Provide[domaininternalops.InternalOpsDataManager](i, func(i do.Injector) (domaininternalops.InternalOpsDataManager, error) {
		return ProvideInternalOpsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[database.Client](i),
		), nil
	})
}
