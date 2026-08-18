package main

// Conversions declared for the notifications domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("notifications", []*Conversion{
		{Name: "ConvertUserDeviceTokenToUserDeviceTokenDatabaseCreationInput", From: Param{Name: "x", Type: "UserDeviceToken"}, To: "UserDeviceTokenDatabaseCreationInput"},
		{Name: "ConvertUserNotificationToUserNotificationUpdateRequestInput", From: Param{Name: "x", Type: "UserNotification"}, To: "UserNotificationUpdateRequestInput"},
		{Name: "ConvertUserNotificationCreationRequestInputToUserNotificationDatabaseCreationInput", From: Param{Name: "x", Type: "UserNotificationCreationRequestInput"}, To: "UserNotificationDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID": NewID(),
			},
		},
		{Name: "ConvertUserNotificationToUserNotificationCreationRequestInput", From: Param{Name: "x", Type: "UserNotification"}, To: "UserNotificationCreationRequestInput"},
		{Name: "ConvertUserNotificationToUserNotificationDatabaseCreationInput", From: Param{Name: "x", Type: "UserNotification"}, To: "UserNotificationDatabaseCreationInput"},
	})
}
