package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"

	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeServiceSettingConfiguration builds a faked service setting.
func BuildFakeServiceSettingConfiguration() *types.ServiceSettingConfiguration {
	return &types.ServiceSettingConfiguration{
		ID:               BuildFakeID(),
		Value:            buildUniqueString(),
		Notes:            buildUniqueString(),
		ServiceSetting:   *BuildFakeServiceSetting(),
		BelongsToUser:    buildUniqueString(),
		BelongsToAccount: buildUniqueString(),
		CreatedAt:        BuildFakeTime(),
	}
}

// BuildFakeServiceSettingConfigurationsList builds a faked ServiceSettingConfigurationList.
func BuildFakeServiceSettingConfigurationsList() *filtering.QueryFilteredResult[types.ServiceSettingConfiguration] {
	var examples []*types.ServiceSettingConfiguration
	for range exampleQuantity {
		examples = append(examples, BuildFakeServiceSettingConfiguration())
	}

	return &filtering.QueryFilteredResult[types.ServiceSettingConfiguration]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeServiceSettingConfigurationUpdateRequestInput builds a faked ServiceSettingConfigurationUpdateRequestInput from a service setting.
func BuildFakeServiceSettingConfigurationUpdateRequestInput() *types.ServiceSettingConfigurationUpdateRequestInput {
	serviceSetting := BuildFakeServiceSettingConfiguration()
	return converters.ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput(serviceSetting)
}

// BuildFakeServiceSettingConfigurationCreationRequestInput builds a faked ServiceSettingConfigurationCreationRequestInput.
func BuildFakeServiceSettingConfigurationCreationRequestInput() *types.ServiceSettingConfigurationCreationRequestInput {
	serviceSetting := BuildFakeServiceSettingConfiguration()
	return converters.ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput(serviceSetting)
}
