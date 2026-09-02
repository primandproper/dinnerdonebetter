package grpc

import (
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	waitlists "github.com/primandproper/platform-go/v13/waitlists"

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
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[waitlists.Store](i),
		), nil
	})
}
