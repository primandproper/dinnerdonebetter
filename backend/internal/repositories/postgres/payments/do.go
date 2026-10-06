package payments

import (
	"context"

	"github.com/primandproper/platform-go/v15/billing"
	platformrecording "github.com/primandproper/platform-go/v15/recording"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsRepository registers the billing store with the injector.
func RegisterPaymentsRepository(i do.Injector) {
	do.Provide[billing.Store](i, func(i do.Injector) (billing.Store, error) {
		return ProvidePaymentsRepository(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[*platformrecording.Recorder](i),
			do.MustInvoke[database.Client](i),
		)
	})
}
