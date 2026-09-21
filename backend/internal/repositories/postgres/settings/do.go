package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	platformsettings "github.com/primandproper/platform-go/v14/settings"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSettingsRepository registers the settings store with the injector.
func RegisterSettingsRepository(i do.Injector) {
	do.Provide[platformsettings.Store](i, func(i do.Injector) (platformsettings.Store, error) {
		return ProvideSettingsRepository(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[database.Client](i),
			do.MustInvoke[*events.Emitter](i),
		)
	})
}
