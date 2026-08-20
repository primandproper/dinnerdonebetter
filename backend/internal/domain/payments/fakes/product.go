package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeProduct builds a faked product.
func BuildFakeProduct() *types.Product {
	product := fake.BuildFakeRecord[types.Product]()

	// A recurring product is the one with the extra rule — its billing interval is
	// required and has to be at least a month — so it is the kind worth defaulting to.
	product.Kind = types.ProductKindRecurring
	product.BillingIntervalMonths = pointer.To(int32(gofakeit.Number(1, 12)))

	// A currency code, which is three letters from a list rather than any string.
	product.Currency = "usd"

	product.AmountCents = int32(gofakeit.Number(100, 10000))

	return product
}

// BuildFakeProductList builds a faked Product list.
func BuildFakeProductList() *filtering.QueryFilteredResult[types.Product] {
	return fake.BuildFakePage(BuildFakeProduct)
}

// BuildFakeProductCreationRequestInput builds a faked ProductCreationRequestInput.
func BuildFakeProductCreationRequestInput() *types.ProductCreationRequestInput {
	product := BuildFakeProduct()

	return &types.ProductCreationRequestInput{
		Name:                  product.Name,
		Description:           product.Description,
		Kind:                  product.Kind,
		AmountCents:           product.AmountCents,
		Currency:              product.Currency,
		BillingIntervalMonths: product.BillingIntervalMonths,
		ExternalProductID:     product.ExternalProductID,
	}
}
