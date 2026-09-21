package authorization

import (
	settingsgrpc "github.com/primandproper/platform-go/v14/settings/grpc"
)

// The settings permissions are platform's, re-exported under the names this
// application's policy already spells. See comments_permissions.go for why
// re-exporting rather than declaring is what keeps a grant and the method that
// requires it from drifting.
//
// WriteAdminSettingValuesPermission has no local predecessor. It is the grant
// platform asks inside the handler for a setting the catalog marked AdminOnly,
// and it replaces the IsServiceAdmin() check the deleted service performed. It
// belongs to service admins and to nobody else.
const (
	// CreateSettingDefinitionsPermission is a permission.
	CreateSettingDefinitionsPermission = settingsgrpc.PermissionCreateDefinitions
	// ReadSettingDefinitionsPermission is a permission.
	ReadSettingDefinitionsPermission = settingsgrpc.PermissionReadDefinitions
	// UpdateSettingDefinitionsPermission is a permission.
	UpdateSettingDefinitionsPermission = settingsgrpc.PermissionUpdateDefinitions
	// ArchiveSettingDefinitionsPermission is a permission.
	ArchiveSettingDefinitionsPermission = settingsgrpc.PermissionArchiveDefinitions
	// WriteSettingValuesPermission covers setting and clearing a value.
	WriteSettingValuesPermission = settingsgrpc.PermissionWriteValues
	// ReadSettingValuesPermission is a permission.
	ReadSettingValuesPermission = settingsgrpc.PermissionReadValues
	// ReadAllSettingValuesPermission covers the administrative read of who has
	// overridden a setting.
	ReadAllSettingValuesPermission = settingsgrpc.PermissionReadAllValues
	// WriteAdminSettingValuesPermission covers writing a setting marked AdminOnly.
	WriteAdminSettingValuesPermission = settingsgrpc.PermissionWriteAdminValues
)
