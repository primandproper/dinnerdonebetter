package main

// Conversions declared for the waitlists domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("waitlists", []*Conversion{
		{Name: "ConvertWaitlistToWaitlistUpdateRequestInput", From: Param{Name: "x", Type: "Waitlist"}, To: "WaitlistUpdateRequestInput"},
		{Name: "ConvertWaitlistCreationRequestInputToWaitlistDatabaseCreationInput", From: Param{Name: "x", Type: "WaitlistCreationRequestInput"}, To: "WaitlistDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID": NewID(),
			},
		},
		{Name: "ConvertWaitlistToWaitlistCreationRequestInput", From: Param{Name: "x", Type: "Waitlist"}, To: "WaitlistCreationRequestInput"},
		{Name: "ConvertWaitlistToWaitlistDatabaseCreationInput", From: Param{Name: "x", Type: "Waitlist"}, To: "WaitlistDatabaseCreationInput"},
		{Name: "ConvertWaitlistSignupToWaitlistSignupUpdateRequestInput", From: Param{Name: "x", Type: "WaitlistSignup"}, To: "WaitlistSignupUpdateRequestInput"},
		{Name: "ConvertWaitlistSignupCreationRequestInputToWaitlistSignupDatabaseCreationInput", From: Param{Name: "x", Type: "WaitlistSignupCreationRequestInput"}, To: "WaitlistSignupDatabaseCreationInput",
			Fields: map[string]Rule{
				"ID": NewID(),
			},
		},
		{Name: "ConvertWaitlistSignupToWaitlistSignupCreationRequestInput", From: Param{Name: "x", Type: "WaitlistSignup"}, To: "WaitlistSignupCreationRequestInput"},
		{Name: "ConvertWaitlistSignupToWaitlistSignupDatabaseCreationInput", From: Param{Name: "x", Type: "WaitlistSignup"}, To: "WaitlistSignupDatabaseCreationInput"},
	})
}
