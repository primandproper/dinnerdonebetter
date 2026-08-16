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
					{Name: "DefaultValue", Expr: `new(defaultValue)`},
					{Name: "AdminsOnly", Expr: `true`},
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
				Fields: []entitydecl.Field{
					{Name: "BelongsToUser", Expr: `buildUniqueString()`},
					{Name: "BelongsToAccount", Expr: `buildUniqueString()`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ServiceSettingConfigurationUpdateRequestInput{}, Converter: "ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput"},
					{Type: ServiceSettingConfigurationCreationRequestInput{}, Converter: "ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput"},
				},
			},
		},
	},
}
