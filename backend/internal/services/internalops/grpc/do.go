package grpc

import (
	domaininternalops "github.com/primandproper/dinnerdonebetter/backend/internal/domain/internalops"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/internalops"

	msgconfig "github.com/primandproper/platform-go/v9/messagequeue/config"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterInternalOpsService registers the internal ops gRPC service with the injector.
func RegisterInternalOpsService(i do.Injector) {
	do.Provide[InternalOpsMethodPermissions](i, func(i do.Injector) (InternalOpsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[settingssvc.InternalOperationsServer](i, func(i do.Injector) (settingssvc.InternalOperationsServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[*msgconfig.Config](i),
			do.MustInvoke[domaininternalops.InternalOpsDataManager](i),
		), nil
	})
}
