package manager

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v9/filtering"
)

// PaymentsDataManager defines the interface for payments business logic.
type PaymentsDataManager interface {
	CreateProduct(ctx context.Context, input *payments.ProductCreationRequestInput) (*payments.Product, error)
	GetProduct(ctx context.Context, id string) (*payments.Product, error)
	GetProducts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Product], error)
	UpdateProduct(ctx context.Context, id string, input *payments.ProductUpdateRequestInput) error
	ArchiveProduct(ctx context.Context, id string) error

	CreateSubscription(ctx context.Context, input *payments.SubscriptionCreationRequestInput) (*payments.Subscription, error)
	GetSubscription(ctx context.Context, id string) (*payments.Subscription, error)
	GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error)
	UpdateSubscription(ctx context.Context, id string, input *payments.SubscriptionUpdateRequestInput) error
	ArchiveSubscription(ctx context.Context, id string) error

	GetPurchasesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Purchase], error)
	GetPaymentTransactionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.PaymentTransaction], error)

	// ProcessWebhookEvent applies an already-verified, already-parsed provider event to our own
	// records. Verification and parsing are a payments.PaymentProcessor's job and happen at the
	// transport edge, where the request still exists; by the time an event reaches here it is
	// domain data, and this manager needs to know nothing about HTTP.
	ProcessWebhookEvent(ctx context.Context, provider string, event *payments.ParsedWebhookEvent, accountID string) error
}
