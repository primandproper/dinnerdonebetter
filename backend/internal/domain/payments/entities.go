package payments

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: Product{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `interval := int32(fake.Number(1, 12))`},
				},
				Fields: []entitydecl.Field{
					{Name: "Kind", Expr: `types.ProductKindRecurring`},
					{Name: "AmountCents", Expr: `int32(fake.Number(100, 10000))`},
					{Name: "Currency", Expr: `"usd"`},
					{Name: "BillingIntervalMonths", Expr: `&interval`},
					{Name: "ExternalProductID", Expr: `buildUniqueString()`},
				},
				List: &entitydecl.List{Name: "BuildFakeProductList"},
			},
		},
		{
			Type: ProductCreationRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `product := BuildFakeProduct()`},
					{Code: `interval := int32(1)`},
				},
				Fields: []entitydecl.Field{
					{Name: "Name", Expr: `product.Name`},
					{Name: "Description", Expr: `product.Description`},
					{Name: "Kind", Expr: `product.Kind`},
					{Name: "AmountCents", Expr: `product.AmountCents`},
					{Name: "Currency", Expr: `product.Currency`},
					{Name: "BillingIntervalMonths", Expr: `&interval`},
					{Name: "ExternalProductID", Expr: `product.ExternalProductID`},
				},
			},
		},
		{
			Type: Subscription{},
			Fake: entitydecl.Fake{
				Bespoke: true,
				BespokeWhy: "A subscription is meaningless without the account and product it belongs to, so its " +
					"builder takes both as arguments — and a builder with parameters is not something a " +
					"declaration of field expressions can describe. It also has to keep CurrentPeriodEnd after " +
					"CurrentPeriodStart, which is the sort of cross-field invariant Locals exist for but which " +
					"is not worth expressing here for one type. Declared anyway so the entity is visible to " +
					"generators that do not care how its fake is built.",
			},
		},
	},
}
