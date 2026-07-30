package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"

	"github.com/primandproper/platform-go/v8/identifiers"
)

// ConvertServiceSettingCreationRequestInputToServiceSettingDatabaseCreationInput creates a ServiceSettingDatabaseCreationInput from a ServiceSettingCreationRequestInput.
func ConvertServiceSettingCreationRequestInputToServiceSettingDatabaseCreationInput(input *types.ServiceSettingCreationRequestInput) *types.ServiceSettingDatabaseCreationInput {
	x := &types.ServiceSettingDatabaseCreationInput{
		ID:           identifiers.New(),
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		DefaultValue: input.DefaultValue,
		AdminsOnly:   input.AdminsOnly,
		Enumeration:  input.Enumeration,
	}

	return x
}

// ConvertServiceSettingToServiceSettingCreationRequestInput builds a ServiceSettingCreationRequestInput from a ServiceSetting.
func ConvertServiceSettingToServiceSettingCreationRequestInput(input *types.ServiceSetting) *types.ServiceSettingCreationRequestInput {
	return &types.ServiceSettingCreationRequestInput{
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		DefaultValue: input.DefaultValue,
		AdminsOnly:   input.AdminsOnly,
		Enumeration:  input.Enumeration,
	}
}

// ConvertServiceSettingToServiceSettingDatabaseCreationInput builds a ServiceSettingDatabaseCreationInput from a ServiceSetting.
func ConvertServiceSettingToServiceSettingDatabaseCreationInput(input *types.ServiceSetting) *types.ServiceSettingDatabaseCreationInput {
	return &types.ServiceSettingDatabaseCreationInput{
		ID:           input.ID,
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		DefaultValue: input.DefaultValue,
		AdminsOnly:   input.AdminsOnly,
		Enumeration:  input.Enumeration,
	}
}
