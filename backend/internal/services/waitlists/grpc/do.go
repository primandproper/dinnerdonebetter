package grpc

import (
	waitlistssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/waitlists"

	waitlists "github.com/primandproper/platform-go/v14/waitlists"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

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
			do.MustInvoke[database.Client](i),
			do.MustInvoke[waitlists.Store](i),
		), nil
	})
}
