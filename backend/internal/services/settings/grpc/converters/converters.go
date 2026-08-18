package converters

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	grpcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/converters"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v11/identifiers"
)

func ConvertGPRCServiceSettingCreationRequestInputToServiceSettingDatabaseCreationInput(input *settingssvc.ServiceSettingCreationRequestInput) *settings.ServiceSettingDatabaseCreationInput {
	return &settings.ServiceSettingDatabaseCreationInput{
		ID:           identifiers.New(),
		DefaultValue: input.DefaultValue,
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		Enumeration:  input.Enumeration,
		AdminsOnly:   input.AdminsOnly,
	}
}

func ConvertServiceSettingToGRPCServiceSetting(input *settings.ServiceSetting) *settingssvc.ServiceSetting {
	return &settingssvc.ServiceSetting{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		Name:          input.Name,
		Id:            input.ID,
		Type:          input.Type,
		Description:   input.Description,
		Enumeration:   input.Enumeration,
		AdminsOnly:    input.AdminsOnly,
		DefaultValue:  input.DefaultValue,
	}
}

func ConvertGRPCServiceSettingToServiceSetting(input *settingssvc.ServiceSetting) *settings.ServiceSetting {
	return &settings.ServiceSetting{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		ArchivedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		Name:          input.Name,
		ID:            input.Id,
		Type:          input.Type,
		Description:   input.Description,
		Enumeration:   input.Enumeration,
		AdminsOnly:    input.AdminsOnly,
		DefaultValue:  input.DefaultValue,
	}
}

func ConvertGRPCServiceSettingConfigurationCreationRequestInputToServiceSettingConfigurationDatabaseCreationInput(input *settingssvc.ServiceSettingConfigurationCreationRequestInput, userID, accountID string) *settings.ServiceSettingConfigurationDatabaseCreationInput {
	return &settings.ServiceSettingConfigurationDatabaseCreationInput{
		ID:               identifiers.New(),
		Value:            input.Value,
		Notes:            input.Notes,
		ServiceSettingID: input.ServiceSettingId,
		BelongsToUser:    userID,
		BelongsToAccount: accountID,
	}
}

func ConvertServiceSettingConfigurationCreationRequestInputToGRPCServiceSettingConfigurationCreationRequestInput(input *settings.ServiceSettingConfigurationCreationRequestInput) *settingssvc.ServiceSettingConfigurationCreationRequestInput {
	return &settingssvc.ServiceSettingConfigurationCreationRequestInput{
		Value:            input.Value,
		Notes:            input.Notes,
		ServiceSettingId: input.ServiceSettingID,
	}
}

func ConvertServiceSettingCreationRequestInputToGRPCServiceSettingCreationRequestInput(input *settings.ServiceSettingCreationRequestInput) *settingssvc.ServiceSettingCreationRequestInput {
	return &settingssvc.ServiceSettingCreationRequestInput{
		DefaultValue: input.DefaultValue,
		Name:         input.Name,
		Type:         input.Type,
		Description:  input.Description,
		Enumeration:  input.Enumeration,
		AdminsOnly:   input.AdminsOnly,
	}
}

func ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(input *settings.ServiceSettingConfiguration) *settingssvc.ServiceSettingConfiguration {
	return &settingssvc.ServiceSettingConfiguration{
		CreatedAt:        grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Id:               input.ID,
		Value:            input.Value,
		Notes:            input.Notes,
		BelongsToUser:    input.BelongsToUser,
		BelongsToAccount: input.BelongsToAccount,
		ServiceSetting:   ConvertServiceSettingToGRPCServiceSetting(&input.ServiceSetting),
	}
}

func ConvertGRPCServiceSettingConfigurationToServiceSettingConfiguration(input *settingssvc.ServiceSettingConfiguration) *settings.ServiceSettingConfiguration {
	return &settings.ServiceSettingConfiguration{
		CreatedAt:        grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:    grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:       grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		ID:               input.Id,
		Value:            input.Value,
		Notes:            input.Notes,
		BelongsToUser:    input.BelongsToUser,
		BelongsToAccount: input.BelongsToAccount,
		ServiceSetting:   *ConvertGRPCServiceSettingToServiceSetting(input.ServiceSetting),
	}
}

func ConvertGRPCServiceSettingConfigurationUpdateRequestInputToServiceSettingConfigurationUpdateRequestInputTo(input *settingssvc.ServiceSettingConfigurationUpdateRequestInput) *settings.ServiceSettingConfigurationUpdateRequestInput {
	return &settings.ServiceSettingConfigurationUpdateRequestInput{
		Value:            input.Value,
		Notes:            input.Notes,
		ServiceSettingID: input.ServiceSettingId,
	}
}

// ConvertUserDataCollectionToGRPCDataCollection converts a domain settings UserDataCollection to a proto DataCollection.
func ConvertUserDataCollectionToGRPCDataCollection(input *settings.UserDataCollection) *settingssvc.DataCollection {
	result := &settingssvc.DataCollection{
		ServiceSettingConfigurations: make(map[string]*settingssvc.ServiceSettingConfiguration),
	}

	for i := range input.AccountSettings {
		result.ServiceSettingConfigurations[input.AccountSettings[i].ID] = ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(&input.AccountSettings[i])
	}

	for i := range input.UserSettings {
		result.UserServiceSettingConfigurations = append(result.UserServiceSettingConfigurations, ConvertServiceSettingConfigurationToGRPCServiceSettingConfiguration(&input.UserSettings[i]))
	}

	return result
}
