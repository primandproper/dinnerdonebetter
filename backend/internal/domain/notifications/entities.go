package notifications

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: UserDeviceToken{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "DeviceToken", Expr: `"a1b2c3d4e5f6789012345678901234567890abcdef1234567890abcdef123456"`},
				},
				List: &entitydecl.List{},
			},
		},
		{
			Type: UserDeviceTokenDatabaseCreationInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `token := BuildFakeUserDeviceToken()`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `token.ID`},
					{Name: "DeviceToken", Expr: `token.DeviceToken`},
					{Name: "Platform", Expr: `token.Platform`},
					{Name: "BelongsToUser", Expr: `token.BelongsToUser`},
				},
			},
		},
		{
			Type: UserNotification{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: UserNotificationUpdateRequestInput{}, Converter: "ConvertUserNotificationToUserNotificationUpdateRequestInput"},
					{Type: UserNotificationCreationRequestInput{}, Converter: "ConvertUserNotificationToUserNotificationCreationRequestInput"},
				},
			},
		},
	},
}
