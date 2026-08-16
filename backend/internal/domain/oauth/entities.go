package oauth

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: OAuth2Client{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Name", Expr: `fake.Password(true, true, true, false, false, 32)`},
					{Name: "ClientSecret", Expr: `buildFakePassword()`},
				},
				List: &entitydecl.List{},
			},
		},
		{
			Type: OAuth2ClientToken{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "RedirectURI", Expr: `fake.URL()`},
					{Name: "CodeChallengeMethod", Expr: `"S256"`},
				},
			},
		},
		{
			Type: OAuth2ClientCreationResponse{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `client := BuildFakeOAuth2Client()`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `client.ID`},
					{Name: "ClientID", Expr: `client.ClientID`},
					{Name: "ClientSecret", Expr: `client.ClientSecret`},
				},
			},
		},
		{
			Type: OAuth2ClientCreationRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `client := BuildFakeOAuth2Client()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Name", Expr: `client.Name`},
					{Name: "Description", Expr: `client.Description`},
				},
			},
		},
	},
}
