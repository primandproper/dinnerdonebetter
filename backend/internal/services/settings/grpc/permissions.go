package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
)

// SettingsMethodPermissions is a named type for dependency injection.
// It allows the injector to distinguish between different services' permission maps.
type SettingsMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns the settings service's method permissions.
// This uses the generated FullMethodName constants from the gRPC generated code to ensure
// type safety and compile-time verification.
//
// The map is what the interceptor checks, and it decides the capability rather
// than who may aim it. Two of this service's rules are not capabilities and are
// enforced in the handlers instead:
//
//   - An admin-only setting is refused to a non-admin, on every method that
//     names one. That is per-row rather than per-method — the same call reads a
//     setting anybody may see and one only an administrator may — so no entry
//     here could express it.
//   - Listing everyone who has answered a setting is a service admin's read.
//     Its capability is the same read.setting_values every member holds for
//     their own answers; what differs is whose rows it returns.
func ProvideMethodPermissions() SettingsMethodPermissions {
	return SettingsMethodPermissions{
		// The catalog. Writing it is administrative — a setting's kind, default and
		// enumeration decide how every stored answer is read — while reading it is
		// every member's, because a preferences page cannot render without it.
		settingssvc.SettingsService_CreateSettingDefinition_FullMethodName: {
			authorization.CreateSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_GetSettingDefinition_FullMethodName: {
			authorization.ReadSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_GetSettingDefinitionByName_FullMethodName: {
			authorization.ReadSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_GetSettingDefinitions_FullMethodName: {
			authorization.ReadSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_UpdateSettingDefinition_FullMethodName: {
			authorization.UpdateSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_ArchiveSettingDefinition_FullMethodName: {
			authorization.ArchiveSettingDefinitionsPermission,
		},

		// The answers. Every one of these is about the requester's own settings:
		// the subject comes from the session, so there is no way to spell somebody
		// else's here.
		//
		// Setting a value is one permission rather than a create/update pair,
		// because the store's write is one write — it converges on the row, so a
		// first answer and a changed one are the same statement, and a permission
		// that separated them would be a line in the role grid nothing could ever
		// grant on its own.
		settingssvc.SettingsService_SetSettingValue_FullMethodName: {
			authorization.CreateSettingValuesPermission,
		},
		settingssvc.SettingsService_GetSettingValue_FullMethodName: {
			authorization.ReadSettingValuesPermission,
		},
		settingssvc.SettingsService_GetSettingValues_FullMethodName: {
			authorization.ReadSettingValuesPermission,
		},
		settingssvc.SettingsService_GetSettingValuesForDefinition_FullMethodName: {
			authorization.ReadSettingValuesPermission,
		},
		settingssvc.SettingsService_ClearSettingValue_FullMethodName: {
			authorization.ArchiveSettingValuesPermission,
		},

		// Resolution reads both halves — the catalog and the requester's answers —
		// and is gated on the catalog's permission, which is the one a member holds
		// in order to be shown a setting at all.
		settingssvc.SettingsService_ResolveSetting_FullMethodName: {
			authorization.ReadSettingDefinitionsPermission,
		},
		settingssvc.SettingsService_ResolveSettings_FullMethodName: {
			authorization.ReadSettingDefinitionsPermission,
		},
	}
}
