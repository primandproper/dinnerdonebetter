package keys

const (
	idSuffix = ".id"

	// SettingDefinitionIDKey is the standard key for referring to a setting
	// definition's ID.
	SettingDefinitionIDKey = "setting_definition" + idSuffix
	// SettingNameKey is the standard key for referring to a setting by the name
	// application code asks for. It is the handle every value-side call takes,
	// so it is on far more spans than the definition id is.
	SettingNameKey = "setting.name"
	// SettingValueIDKey is the standard key for referring to a stored setting
	// value's ID.
	SettingValueIDKey = "setting_value" + idSuffix
	// SettingResolutionSourceKey is the standard key for where a resolved
	// setting's value came from: the subject, the default, or neither.
	SettingResolutionSourceKey = "setting.resolution_source"
)
