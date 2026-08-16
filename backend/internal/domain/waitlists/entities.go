package waitlists

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: Waitlist{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "ValidUntil", Expr: `time.Now().Add(24 * time.Hour).UTC().Truncate(time.Second)`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: WaitlistCreationRequestInput{}, Converter: "ConvertWaitlistToWaitlistCreationRequestInput"},
					{Type: WaitlistUpdateRequestInput{}, Converter: "ConvertWaitlistToWaitlistUpdateRequestInput"},
				},
			},
		},
		{
			Type: WaitlistSignup{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: WaitlistSignupCreationRequestInput{}, Converter: "ConvertWaitlistSignupToWaitlistSignupCreationRequestInput"},
					{Type: WaitlistSignupUpdateRequestInput{}, Converter: "ConvertWaitlistSignupToWaitlistSignupUpdateRequestInput"},
				},
			},
		},
	},
}
