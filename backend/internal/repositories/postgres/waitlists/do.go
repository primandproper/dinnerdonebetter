package waitlists

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformwaitlists "github.com/primandproper/platform-go/v13/waitlists"

	"github.com/samber/do/v2"
)

// RegisterWaitlistsRepository registers the waitlists store with the injector.
func RegisterWaitlistsRepository(i do.Injector) {
	do.Provide[platformwaitlists.Store](i, func(i do.Injector) (platformwaitlists.Store, error) {
		return ProvideWaitlistsRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		)
	})
}
