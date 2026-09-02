package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformsettings "github.com/primandproper/platform-go/v13/settings"

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
