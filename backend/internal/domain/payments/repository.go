package payments

import (
	"context"

	"github.com/primandproper/platform-go/v10/filtering"
)

// Repository defines the data access interface for payments entities.
type Repository interface {
	CreateProduct(ctx context.Context, input *ProductDatabaseCreationInput) (*Product, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProductByExternalID(ctx context.Context, externalProductID string) (*Product, error)
	GetProducts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Product], error)
	UpdateProduct(ctx context.Context, product *Product) error
	ArchiveProduct(ctx context.Context, id string) error
	ProductExists(ctx context.Context, id string) (bool, error)

	CreateSubscription(ctx context.Context, input *SubscriptionDatabaseCreationInput) (*Subscription, error)
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetSubscriptionByExternalID(ctx context.Context, externalID string) (*Subscription, error)
	GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Subscription], error)
	UpdateSubscription(ctx context.Context, sub *Subscription) error
	UpdateSubscriptionStatus(ctx context.Context, id, status string) error
	ArchiveSubscription(ctx context.Context, id string) error

	CreatePurchase(ctx context.Context, input *PurchaseDatabaseCreationInput) (*Purchase, error)
	GetPurchase(ctx context.Context, id string) (*Purchase, error)
	GetPurchasesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[Purchase], error)

	CreatePaymentTransaction(ctx context.Context, input *PaymentTransactionDatabaseCreationInput) (*PaymentTransaction, error)
	GetPaymentTransactionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[PaymentTransaction], error)
}
