package settings

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: ServiceSetting{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `defaultValue := buildUniqueString()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Type", Expr: `"user"`},
					{Name: "Enumeration", Expr: `[]string{defaultValue}`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ServiceSettingCreationRequestInput{}, Converter: "ConvertServiceSettingToServiceSettingCreationRequestInput"},
				},
			},
		},
		{
			Type: ServiceSettingConfiguration{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ServiceSettingConfigurationUpdateRequestInput{}, Converter: "ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput"},
					{Type: ServiceSettingConfigurationCreationRequestInput{}, Converter: "ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput"},
				},
			},
		},
	},
}
