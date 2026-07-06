package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"

	"github.com/primandproper/platform-go/v3/database/filtering"

	"github.com/stretchr/testify/mock"
)

var _ payments.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

func (m *Repository) CreateProduct(ctx context.Context, input *payments.ProductDatabaseCreationInput) (*payments.Product, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Product), args.Error(1)
}

func (m *Repository) GetProduct(ctx context.Context, id string) (*payments.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Product), args.Error(1)
}

func (m *Repository) GetProductByExternalID(ctx context.Context, externalProductID string) (*payments.Product, error) {
	args := m.Called(ctx, externalProductID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Product), args.Error(1)
}

func (m *Repository) GetProducts(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Product], error) {
	args := m.Called(ctx, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[payments.Product]), args.Error(1)
}

func (m *Repository) UpdateProduct(ctx context.Context, product *payments.Product) error {
	return m.Called(ctx, product).Error(0)
}

func (m *Repository) ArchiveProduct(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *Repository) ProductExists(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

func (m *Repository) CreateSubscription(ctx context.Context, input *payments.SubscriptionDatabaseCreationInput) (*payments.Subscription, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Subscription), args.Error(1)
}

func (m *Repository) GetSubscription(ctx context.Context, id string) (*payments.Subscription, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Subscription), args.Error(1)
}

func (m *Repository) GetSubscriptionByExternalID(ctx context.Context, externalID string) (*payments.Subscription, error) {
	args := m.Called(ctx, externalID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Subscription), args.Error(1)
}

func (m *Repository) GetSubscriptionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Subscription], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[payments.Subscription]), args.Error(1)
}

func (m *Repository) UpdateSubscription(ctx context.Context, sub *payments.Subscription) error {
	return m.Called(ctx, sub).Error(0)
}

func (m *Repository) UpdateSubscriptionStatus(ctx context.Context, id, status string) error {
	return m.Called(ctx, id, status).Error(0)
}

func (m *Repository) ArchiveSubscription(ctx context.Context, id string) error {
	return m.Called(ctx, id).Error(0)
}

func (m *Repository) CreatePurchase(ctx context.Context, input *payments.PurchaseDatabaseCreationInput) (*payments.Purchase, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Purchase), args.Error(1)
}

func (m *Repository) GetPurchase(ctx context.Context, id string) (*payments.Purchase, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.Purchase), args.Error(1)
}

func (m *Repository) GetPurchasesForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.Purchase], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[payments.Purchase]), args.Error(1)
}

func (m *Repository) CreatePaymentTransaction(ctx context.Context, input *payments.PaymentTransactionDatabaseCreationInput) (*payments.PaymentTransaction, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*payments.PaymentTransaction), args.Error(1)
}

func (m *Repository) GetPaymentTransactionsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[payments.PaymentTransaction], error) {
	args := m.Called(ctx, accountID, filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*filtering.QueryFilteredResult[payments.PaymentTransaction]), args.Error(1)
}
