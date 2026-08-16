package comments

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: Comment{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "TargetType", Expr: `"recipes"`},
				},
			},
		},
		{
			Type: CommentCreationRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "TargetType", Expr: `"recipes"`},
				},
			},
		},
		{
			Type: CommentDatabaseCreationInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "TargetType", Expr: `"recipes"`},
				},
			},
		},
	},
}
