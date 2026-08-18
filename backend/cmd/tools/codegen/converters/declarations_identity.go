package main

// Conversions declared for the identity domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("identity", []*Conversion{
		{Name: "ConvertAccountInvitationToAccountInvitationCreationInput", From: Param{Name: "accountInvitation", Type: "AccountInvitation"}, To: "AccountInvitationCreationRequestInput"},
		{Name: "ConvertAccountInvitationToAccountInvitationUpdateInput", From: Param{Name: "accountInvitation", Type: "AccountInvitation"}, To: "AccountInvitationUpdateRequestInput"},
		{Name: "ConvertAccountInvitationToAccountInvitationDatabaseCreationInput", From: Param{Name: "accountInvitation", Type: "AccountInvitation"}, To: "AccountInvitationDatabaseCreationInput",
			Fields: map[string]Rule{
				"DestinationAccountID": NestedID("DestinationAccount"),
				"FromUser":             NestedID("FromUser"),
			},
		},
		{Name: "ConvertAccountUserMembershipToAccountUserMembershipDatabaseCreationInput", From: Param{Name: "membership", Type: "AccountUserMembership"}, To: "AccountUserMembershipDatabaseCreationInput",
			Fields: map[string]Rule{
				"AccountID": Rename("BelongsToAccount"),
				"ID":        NewID(),
				"Reason":    Skip("The reason a membership was granted is supplied by whoever grants it. A membership read back does not carry one."),
				"UserID":    Rename("BelongsToUser"),
			},
		},
		{Name: "ConvertAccountCreationInputToAccountDatabaseCreationInput", From: Param{Name: "input", Type: "AccountCreationRequestInput"}, To: "AccountDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID":                   NewID(),
				"WebhookEncryptionKey": Skip("The HMAC key an account signs its webhook deliveries with. It is generated for the account, not requested, and the repository writes it as webhook_hmac_secret."),
			},
		},
		{Name: "ConvertAccountToAccountUpdateRequestInput", From: Param{Name: "input", Type: "Account"}, To: "AccountUpdateRequestInput"},
		{Name: "ConvertAccountToAccountCreationRequestInput", From: Param{Name: "account", Type: "Account"}, To: "AccountCreationRequestInput",
			Fields: map[string]Rule{
				"BelongsToUser": Skip("Account carries this and the converter this replaced did not copy it. Preserved rather than corrected, so that generating these converters is not also a behavior change."),
			},
		},
		{Name: "ConvertAccountToAccountDatabaseCreationInput", From: Param{Name: "account", Type: "Account"}, To: "AccountDatabaseCreationInput"},
		{Name: "ConvertUserToUserDatabaseCreationInput", From: Param{Name: "user", Type: "User"}, To: "UserDatabaseCreationInput",
			Fields: map[string]Rule{
				"AcceptedPrivacyPolicy": Skip("Registration input; see AcceptedTOS."),
				"AcceptedTOS":           Skip("Registration input. A stored User records neither the acceptance nor when it happened, so there is nothing to carry back."),
				"AccountName":           Skip("Names the household account created alongside the user. It belongs to the registration request, not to the user it created."),
				"DestinationAccountID":  Skip("Set only when the registration is redeeming an invitation; a stored User does not record which one."),
				"InvitationToken":       Skip("The invitation being redeemed, consumed during registration and not kept on the user."),
			},
		},
		{Name: "ConvertUserDetailsUpdateRequestInputToUserDetailsUpdateInput", From: Param{Name: "x", Type: "UserDetailsUpdateRequestInput"}, To: "UserDetailsDatabaseUpdateInput"},
	})
}
