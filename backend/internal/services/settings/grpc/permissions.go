package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
)

// SettingsMethodPermissions is a named type for Wire dependency injection.
// It allows Wire to distinguish between different services' permission maps.
type SettingsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the settings service's method permissions.
// This uses the generated FullMethodName constants from the gRPC generated code to ensure
// type safety and compile-time verification.
func ProvideMethodPermissions() SettingsMethodPermissions {
	return SettingsMethodPermissions{
		// ServiceSettings methods
		settingssvc.SettingsService_CreateServiceSetting_FullMethodName: {
			authorization.CreateServiceSettingsPermission,
		},
		settingssvc.SettingsService_GetServiceSetting_FullMethodName: {
			authorization.ReadServiceSettingsPermission,
		},
		settingssvc.SettingsService_GetServiceSettings_FullMethodName: {
			authorization.ReadServiceSettingsPermission,
		},
		settingssvc.SettingsService_SearchForServiceSettings_FullMethodName: {
			authorization.ReadServiceSettingsPermission,
		},
		settingssvc.SettingsService_ArchiveServiceSetting_FullMethodName: {
			authorization.ArchiveServiceSettingsPermission,
		},

		// ServiceSettingConfigurations methods
		settingssvc.SettingsService_CreateServiceSettingConfiguration_FullMethodName: {
			authorization.CreateServiceSettingConfigurationsPermission,
		},
		settingssvc.SettingsService_GetServiceSettingConfigurationByName_FullMethodName: {
			authorization.ReadServiceSettingConfigurationsPermission,
		},
		settingssvc.SettingsService_GetServiceSettingConfigurationsForAccount_FullMethodName: {
			authorization.ReadServiceSettingConfigurationsPermission,
		},
		settingssvc.SettingsService_GetServiceSettingConfigurationsForUser_FullMethodName: {
			authorization.ReadServiceSettingConfigurationsPermission,
		},
		settingssvc.SettingsService_UpdateServiceSettingConfiguration_FullMethodName: {
			authorization.UpdateServiceSettingConfigurationsPermission,
		},
		settingssvc.SettingsService_ArchiveServiceSettingConfiguration_FullMethodName: {
			authorization.ArchiveServiceSettingConfigurationsPermission,
		},
	}
}
