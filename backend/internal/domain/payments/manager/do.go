package manager

import (
	"context"

	identitymanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager"

	"github.com/primandproper/platform-go/v14/billing"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterPaymentsDataManager registers the payments data manager with the injector.
func RegisterPaymentsDataManager(i do.Injector) {
	do.Provide[PaymentsDataManager](i, func(i do.Injector) (PaymentsDataManager, error) {
		return NewPaymentsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[billing.Store](i),
			do.MustInvoke[identitymanager.IdentityDataManager](i),
		)
	})
}
