package main

// Conversions declared for the settings domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("settings", []*Conversion{
		{Name: "ConvertServiceSettingConfigurationToServiceSettingConfigurationUpdateRequestInput", From: Param{Name: "input", Type: "ServiceSettingConfiguration"}, To: "ServiceSettingConfigurationUpdateRequestInput",
			Fields: map[string]Rule{
				"ServiceSettingID": NestedID("ServiceSetting"),
			},
		},
		{Name: "ConvertServiceSettingConfigurationToServiceSettingConfigurationCreationRequestInput", From: Param{Name: "input", Type: "ServiceSettingConfiguration"}, To: "ServiceSettingConfigurationCreationRequestInput",
			Fields: map[string]Rule{
				"ServiceSettingID": NestedID("ServiceSetting"),
			},
		},
		{Name: "ConvertServiceSettingConfigurationToServiceSettingConfigurationDatabaseCreationInput", From: Param{Name: "input", Type: "ServiceSettingConfiguration"}, To: "ServiceSettingConfigurationDatabaseCreationInput",
			Fields: map[string]Rule{
				"ServiceSettingID": NestedID("ServiceSetting"),
			},
		},
		{Name: "ConvertServiceSettingCreationRequestInputToServiceSettingDatabaseCreationInput", From: Param{Name: "input", Type: "ServiceSettingCreationRequestInput"}, To: "ServiceSettingDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID": NewID(),
			},
		},
		{Name: "ConvertServiceSettingToServiceSettingCreationRequestInput", From: Param{Name: "input", Type: "ServiceSetting"}, To: "ServiceSettingCreationRequestInput"},
		{Name: "ConvertServiceSettingToServiceSettingDatabaseCreationInput", From: Param{Name: "input", Type: "ServiceSetting"}, To: "ServiceSettingDatabaseCreationInput"},
	})
}
