package internalops

import (
	domaininternalops "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"

	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterInternalOpsRepository registers the internal ops repository with the injector.
func RegisterInternalOpsRepository(i do.Injector) {
	do.Provide[domaininternalops.InternalOpsDataManager](i, func(i do.Injector) (domaininternalops.InternalOpsDataManager, error) {
		return ProvideInternalOpsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[database.Client](i),
		), nil
	})
}
