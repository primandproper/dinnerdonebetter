package issuereports

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: IssueReport{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "IssueType", Expr: `fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})`},
					{Name: "Details", Expr: `fake.Sentence(20)`},
					{Name: "RelevantTable", Expr: `fake.RandomString([]string{"users", "accounts"})`},
				},
			},
		},
		{
			Type: IssueReportCreationRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "IssueType", Expr: `fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})`},
					{Name: "Details", Expr: `fake.Sentence(20)`},
					{Name: "RelevantTable", Expr: `fake.RandomString([]string{"users", "accounts"})`},
				},
			},
		},
		{
			Type: IssueReportDatabaseCreationInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "IssueType", Expr: `fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})`},
					{Name: "Details", Expr: `fake.Sentence(20)`},
					{Name: "RelevantTable", Expr: `fake.RandomString([]string{"users", "accounts"})`},
				},
			},
		},
		{
			Type: IssueReportUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `issueType := fake.RandomString([]string{"bug", "feature_request", "data_quality", "performance", "other"})`},
					{Code: `details := fake.Sentence(20)`},
					{Code: `relevantTable := fake.RandomString([]string{"users", "accounts"})`},
					{Code: `relevantRecordID := BuildFakeID()`},
				},
				Fields: []entitydecl.Field{
					{Name: "IssueType", Expr: `&issueType`},
					{Name: "Details", Expr: `&details`},
					{Name: "RelevantTable", Expr: `&relevantTable`},
					{Name: "RelevantRecordID", Expr: `&relevantRecordID`},
				},
			},
		},
	},
}
