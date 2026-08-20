package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	settingsrepo "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/settings"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSettingsDataManager registers the settings data manager with the injector.
func RegisterSettingsDataManager(i do.Injector) {
	// Register the repo provider (was included in wire.NewSet)
	settingsrepo.RegisterSettingsRepository(i)

	// Bind *settingsrepo.Repository to the settingsRepo interface
	do.Provide[settingsRepo](i, func(i do.Injector) (settingsRepo, error) {
		return do.MustInvoke[*settingsrepo.Repository](i), nil
	})

	do.Provide[SettingsDataManager](i, func(i do.Injector) (SettingsDataManager, error) {
		return NewSettingsDataManager(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[settingsRepo](i),
		)
	})

	// Bind SettingsDataManager to settings.Repository
	do.Provide[settings.Repository](i, func(i do.Injector) (settings.Repository, error) {
		return do.MustInvoke[SettingsDataManager](i), nil
	})
}
