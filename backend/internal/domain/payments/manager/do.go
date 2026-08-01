package manager

import (
	"context"

	identitymanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsDataManager registers the payments data manager with the injector.
func RegisterPaymentsDataManager(i do.Injector) {
	do.Provide[PaymentsDataManager](i, func(i do.Injector) (PaymentsDataManager, error) {
		return NewPaymentsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[payments.Repository](i),
			do.MustInvoke[payments.PaymentProcessorRegistry](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
		)
	})
}
