package main

// Conversions declared for the issuereports domain.
//
// A conversion with no Fields is a whole-struct copy: every field of the destination is
// filled from the field of the same name, and the generator fails rather than leave one
// empty. Fields carries a rule per destination field where that is not what happens, and
// the reason it carries is rendered into the generated source. See declaration.go for the
// rules, and converters_manual.go in the domain for the conversions this cannot express.

func init() {
	register("issuereports", []*Conversion{
		{Name: "ConvertIssueReportToIssueReportUpdateRequestInput", From: Param{Name: "x", Type: "IssueReport"}, To: "IssueReportUpdateRequestInput"},
		{Name: "ConvertIssueReportCreationRequestInputToIssueReportDatabaseCreationInput", From: Param{Name: "x", Type: "IssueReportCreationRequestInput"}, Extra: []Param{{Name: "userID", Type: "string"}, {Name: "accountID", Type: "string"}}, To: "IssueReportDatabaseCreationInput",
			Fields: map[string]Rule{
				"BelongsToAccount": Expr("accountID", "Comes from the session rather than the request body, so the manager stamps it after this."),
				"CreatedByUser":    Expr("userID", "Comes from the session rather than the request body, so the manager stamps it after this."),
				"ID":               NewID(),
			},
		},
		{Name: "ConvertIssueReportToIssueReportCreationRequestInput", From: Param{Name: "x", Type: "IssueReport"}, To: "IssueReportCreationRequestInput"},
		{Name: "ConvertIssueReportToIssueReportDatabaseCreationInput", From: Param{Name: "x", Type: "IssueReport"}, To: "IssueReportDatabaseCreationInput"},
	})
}
