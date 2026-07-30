package grpc

import (
	waitlistsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/waitlists/manager"
	waitlistssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterWaitlistsService registers the waitlists gRPC service with the injector.
func RegisterWaitlistsService(i do.Injector) {
	do.Provide[WaitlistsMethodPermissions](i, func(i do.Injector) (WaitlistsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[waitlistssvc.WaitlistsServiceServer](i, func(i do.Injector) (waitlistssvc.WaitlistsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[waitlistsmanager.WaitlistsDataManager](i),
		), nil
	})
}
