package authorization

const (
	// CreateSettingDefinitionsPermission is a service admin permission: adding a
	// setting to the catalog.
	CreateSettingDefinitionsPermission Permission = "create.setting_definitions"
	// ReadSettingDefinitionsPermission is an account member permission. Reading
	// the catalog is what a preferences page needs before it can render
	// anything; the settings marked admin-only are filtered out of what a
	// non-admin is shown.
	ReadSettingDefinitionsPermission Permission = "read.setting_definitions"
	// UpdateSettingDefinitionsPermission is a service admin permission: editing
	// a setting's kind, default or enumeration, which decides how every answer
	// already stored against it is read.
	UpdateSettingDefinitionsPermission Permission = "update.setting_definitions"
	// ArchiveSettingDefinitionsPermission is a service admin permission.
	ArchiveSettingDefinitionsPermission Permission = "archive.setting_definitions"

	// CreateSettingValuesPermission is an account member permission: answering a
	// setting about yourself.
	//
	// It covers changing an answer as well as making one. The store's write
	// converges on the row, so the two are the same statement, and a separate
	// update permission would be a grant nothing could be given without this
	// one.
	CreateSettingValuesPermission Permission = "create.setting_values"
	// ReadSettingValuesPermission is an account member permission.
	ReadSettingValuesPermission Permission = "read.setting_values"
	// ArchiveSettingValuesPermission is an account member permission: taking
	// your answer back, which leaves you on the setting's default.
	ArchiveSettingValuesPermission Permission = "archive.setting_values"
)

var (
	// SettingsPermissions contains all settings-related permissions.
	SettingsPermissions = []Permission{
		CreateSettingDefinitionsPermission,
		ReadSettingDefinitionsPermission,
		UpdateSettingDefinitionsPermission,
		ArchiveSettingDefinitionsPermission,
		CreateSettingValuesPermission,
		ReadSettingValuesPermission,
		ArchiveSettingValuesPermission,
	}
)
