package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"
)

// BuildFakeServiceSetting builds a faked service setting.
func BuildFakeServiceSetting() *types.ServiceSetting {
	setting := fake.BuildFakeRecord[types.ServiceSetting]()

	// A setting says who it is for, and "user" is one of the few answers.
	setting.Type = "user"
	setting.AdminsOnly = true

	// The default has to be one of the values the setting enumerates — the type
	// validates exactly that — so the two are built together rather than separately.
	defaultValue := fake.BuildFakeString()
	setting.Enumeration = []string{defaultValue}
	setting.DefaultValue = pointer.To(defaultValue)

	return setting
}

// BuildFakeServiceSettingsList builds a faked ServiceSettingList.
func BuildFakeServiceSettingsList() *filtering.QueryFilteredResult[types.ServiceSetting] {
	return fake.BuildFakePage(BuildFakeServiceSetting)
}

// BuildFakeServiceSettingCreationRequestInput builds a faked ServiceSettingCreationRequestInput.
func BuildFakeServiceSettingCreationRequestInput() *types.ServiceSettingCreationRequestInput {
	serviceSetting := BuildFakeServiceSetting()

	return converters.ConvertServiceSettingToServiceSettingCreationRequestInput(serviceSetting)
}
