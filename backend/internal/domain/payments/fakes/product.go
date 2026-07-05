package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v3/database/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeProduct builds a faked product.
func BuildFakeProduct() *types.Product {
	interval := int32(fake.Number(1, 12))
	return &types.Product{
		ID:                    BuildFakeID(),
		Name:                  buildUniqueString(),
		Description:           buildUniqueString(),
		Kind:                  types.ProductKindRecurring,
		AmountCents:           int32(fake.Number(100, 10000)),
		Currency:              "usd",
		BillingIntervalMonths: &interval,
		ExternalProductID:     buildUniqueString(),
		CreatedAt:             BuildFakeTime(),
	}
}

// BuildFakeProductList builds a faked Product list.
func BuildFakeProductList() *filtering.QueryFilteredResult[types.Product] {
	var examples []*types.Product
	for range exampleQuantity {
		examples = append(examples, BuildFakeProduct())
	}

	return &filtering.QueryFilteredResult[types.Product]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeProductCreationRequestInput builds a faked ProductCreationRequestInput.
func BuildFakeProductCreationRequestInput() *types.ProductCreationRequestInput {
	product := BuildFakeProduct()
	interval := int32(1)
	return &types.ProductCreationRequestInput{
		Name:                  product.Name,
		Description:           product.Description,
		Kind:                  product.Kind,
		AmountCents:           product.AmountCents,
		Currency:              product.Currency,
		BillingIntervalMonths: &interval,
		ExternalProductID:     product.ExternalProductID,
	}
}
