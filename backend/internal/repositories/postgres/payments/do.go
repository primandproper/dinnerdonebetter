package payments

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	domainpayments "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v8/database"
	"github.com/primandproper/platform-go/v8/observability/logging"
	"github.com/primandproper/platform-go/v8/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsRepository registers the payments repository with the injector.
func RegisterPaymentsRepository(i do.Injector) {
	do.Provide[domainpayments.Repository](i, func(i do.Injector) (domainpayments.Repository, error) {
		return ProvidePaymentsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		), nil
	})
}
