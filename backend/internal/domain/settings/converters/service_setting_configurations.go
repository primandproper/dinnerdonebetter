package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
)

// ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput creates a ServiceSettingConfigurationUpdateRequestInput from a ServiceSettingConfiguration.
func ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput(input *settings.ServiceSettingConfiguration) *settings.ServiceSettingConfigurationUpdateRequestInput {
	x := &settings.ServiceSettingConfigurationUpdateRequestInput{
		Value:            &input.Value,
		Notes:            &input.Notes,
		ServiceSettingID: &input.ServiceSetting.ID,
	}

	return x
}

// ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput builds a ServiceSettingConfigurationCreationRequestInput from a ServiceSettingConfiguration.
func ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput(input *settings.ServiceSettingConfiguration) *settings.ServiceSettingConfigurationCreationRequestInput {
	return &settings.ServiceSettingConfigurationCreationRequestInput{
		Value:            input.Value,
		Notes:            input.Notes,
		ServiceSettingID: input.ServiceSetting.ID,
	}
}

// ConvertServiceSettingConfigurationToServiceSettingConfigurationDatabaseCreationInput builds a ServiceSettingConfigurationDatabaseCreationInput from a ServiceSettingConfiguration.
func ConvertServiceSettingConfigurationToServiceSettingConfigurationDatabaseCreationInput(input *settings.ServiceSettingConfiguration) *settings.ServiceSettingConfigurationDatabaseCreationInput {
	return &settings.ServiceSettingConfigurationDatabaseCreationInput{
		ID:               input.ID,
		Value:            input.Value,
		Notes:            input.Notes,
		ServiceSettingID: input.ServiceSetting.ID,
		BelongsToUser:    input.BelongsToUser,
		BelongsToAccount: input.BelongsToAccount,
	}
}
