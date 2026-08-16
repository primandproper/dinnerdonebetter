package audit

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: AuditLogEntry{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "ResourceType", Expr: `"example"`},
					{Name: "EventType", Expr: `types.AuditLogEventTypeOther`},
					{Name: "ActorType", Expr: `""`},
					{Name: "ActorIP", Expr: `""`},
					{Name: "Scope", Expr: `""`},
					{Name: "PrevHash", Expr: `""`},
					{Name: "Hash", Expr: `""`},
					{Name: "Seq", Expr: `0`},
				},
				List: &entitydecl.List{Name: "BuildFakeAuditLogEntriesList"},
			},
		},
	},
}
