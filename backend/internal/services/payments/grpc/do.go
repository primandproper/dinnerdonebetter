package grpc

import (
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsService registers the payments gRPC service with the injector.
func RegisterPaymentsService(i do.Injector) {
	do.Provide[PaymentsMethodPermissions](i, func(i do.Injector) (PaymentsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[paymentssvc.PaymentsServiceServer](i, func(i do.Injector) (paymentssvc.PaymentsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[billing.Store](i),
		), nil
	})
}
