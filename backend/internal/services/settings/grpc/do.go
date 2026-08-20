package grpc

import (
	settingsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/manager"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterSettingsService registers the settings gRPC service with the injector.
func RegisterSettingsService(i do.Injector) {
	do.Provide[SettingsMethodPermissions](i, func(i do.Injector) (SettingsMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[settingssvc.SettingsServiceServer](i, func(i do.Injector) (settingssvc.SettingsServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[settingsmanager.SettingsDataManager](i),
		), nil
	})
}
