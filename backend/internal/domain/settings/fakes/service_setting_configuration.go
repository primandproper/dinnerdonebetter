package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeServiceSettingConfiguration builds a faked service setting configuration.
func BuildFakeServiceSettingConfiguration() *types.ServiceSettingConfiguration {
	cfg := fake.BuildFakeRecord[types.ServiceSettingConfiguration]()

	// The setting the configuration configures. BuildFakeRecord fills a nested struct
	// too, but with a setting whose default value is not in its own enumeration.
	cfg.ServiceSetting = *BuildFakeServiceSetting()

	return cfg
}

// BuildFakeServiceSettingConfigurationsList builds a faked ServiceSettingConfigurationList.
func BuildFakeServiceSettingConfigurationsList() *filtering.QueryFilteredResult[types.ServiceSettingConfiguration] {
	return fake.BuildFakePage(BuildFakeServiceSettingConfiguration)
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
