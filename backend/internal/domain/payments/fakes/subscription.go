package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v7/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeSubscription builds a faked subscription.
func BuildFakeSubscription(accountID, productID string) *types.Subscription {
	start := BuildFakeTime()
	end := start.AddDate(0, fake.Number(1, 12), 0)
	return &types.Subscription{
		ID:                     BuildFakeID(),
		BelongsToAccount:       accountID,
		ProductID:              productID,
		ExternalSubscriptionID: buildUniqueString(),
		Status:                 types.SubscriptionStatusActive,
		CurrentPeriodStart:     start,
		CurrentPeriodEnd:       end,
		CreatedAt:              start,
	}
}

// BuildFakeSubscriptionList builds a faked Subscription list.
func BuildFakeSubscriptionList(accountID, productID string) *filtering.QueryFilteredResult[types.Subscription] {
	var examples []*types.Subscription
	for range exampleQuantity {
		examples = append(examples, BuildFakeSubscription(accountID, productID))
	}

	return &filtering.QueryFilteredResult[types.Subscription]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeSubscriptionCreationRequestInput builds a faked SubscriptionCreationRequestInput.
func BuildFakeSubscriptionCreationRequestInput(accountID, productID string) *types.SubscriptionCreationRequestInput {
	sub := BuildFakeSubscription(accountID, productID)
	return &types.SubscriptionCreationRequestInput{
		BelongsToAccount:       accountID,
		ProductID:              productID,
		ExternalSubscriptionID: sub.ExternalSubscriptionID,
		Status:                 sub.Status,
		CurrentPeriodStart:     sub.CurrentPeriodStart,
		CurrentPeriodEnd:       sub.CurrentPeriodEnd,
	}
}
