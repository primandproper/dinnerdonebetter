package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeSubscription builds a faked subscription.
//
// It takes the account and the product because a subscription is the row joining them,
// and a subscription pointing at rows no test created is one no read path returns.
func BuildFakeSubscription(accountID, productID string) *types.Subscription {
	subscription := fake.BuildFakeRecord[types.Subscription]()

	subscription.BelongsToAccount = accountID
	subscription.ProductID = productID

	// One of the statuses the domain lists, and the one an unqualified "a subscription"
	// means everywhere else.
	subscription.Status = types.SubscriptionStatusActive

	// A period rather than two arbitrary instants: the end has to follow the start, and
	// the row was created when its first period began.
	subscription.CurrentPeriodStart = fake.BuildFakeTime()
	subscription.CurrentPeriodEnd = subscription.CurrentPeriodStart.AddDate(0, gofakeit.Number(1, 12), 0)
	subscription.CreatedAt = subscription.CurrentPeriodStart

	return subscription
}

// BuildFakeSubscriptionList builds a faked Subscription list.
func BuildFakeSubscriptionList(accountID, productID string) *filtering.QueryFilteredResult[types.Subscription] {
	return fake.BuildFakePage(func() *types.Subscription {
		return BuildFakeSubscription(accountID, productID)
	})
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
