package payments

import (
	"context"
	"testing"

	identityfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v15/billing"
	billingmock "github.com/primandproper/platform-go/v15/billing/mock"
	platformidentity "github.com/primandproper/platform-go/v15/identity"
	identitymock "github.com/primandproper/platform-go/v15/identity/mock"
	"github.com/primandproper/primitives-go/v2/capitalism"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/fake"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catalogOf is a product store that sells exactly product, under its external id.
func catalogOf(product *billing.Product) *billingmock.StoreMock {
	return &billingmock.StoreMock{
		GetProductByExternalIDFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, externalID string) (*billing.Product, error) {
			if externalID != product.ExternalProductID {
				return nil, billing.ErrProductNotFound
			}

			return product, nil
		},
	}
}

// directoryOf is a directory holding exactly account, readable by its id and by the
// processor customer it holds.
func directoryOf(account *platformidentity.Account) *identitymock.StoreMock {
	return &identitymock.StoreMock{
		GetAccountFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, accountID string) (*platformidentity.Account, error) {
			if accountID != account.ID {
				return nil, platformidentity.ErrAccountNotFound
			}

			return account, nil
		},
		GetAccountByPaymentProcessorCustomerIDFunc: func(_ context.Context, _ database.SQLQueryExecutor, _ tenancy.Scope, customerID string) (*platformidentity.Account, error) {
			if customerID != account.PaymentProcessorCustomerID {
				return nil, platformidentity.ErrAccountNotFound
			}

			return account, nil
		},
	}
}

func TestStripePlace(T *testing.T) {
	T.Parallel()

	T.Run("places on the account holding the customer and the product sold at the price", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()
		account.PaymentProcessorCustomerID = fake.BuildFakeID()

		placement, err := StripePlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: account.PaymentProcessorCustomerID,
			PriceID:    product.ExternalProductID,
			ProductID:  fake.BuildFakeID(),
		})
		require.NoError(t, err)

		assert.Equal(t, account.ID, placement.AccountID)
		assert.Equal(t, product.ID, placement.ProductID)
	})

	T.Run("a customer no account holds is refused", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()
		account.PaymentProcessorCustomerID = fake.BuildFakeID()

		// The customer id is not an account id, so an account whose id happens to be the
		// customer's is not the one this is for.
		placement, err := StripePlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: account.ID,
			PriceID:    product.ExternalProductID,
		})

		require.ErrorIs(t, err, platformidentity.ErrAccountNotFound)
		assert.Nil(t, placement)
	})

	T.Run("a price nobody sells is refused", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()
		account.PaymentProcessorCustomerID = fake.BuildFakeID()

		// Stripe's product id is not what this catalog is keyed on for Stripe.
		placement, err := StripePlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: account.PaymentProcessorCustomerID,
			PriceID:    fake.BuildFakeID(),
			ProductID:  product.ExternalProductID,
		})

		require.ErrorIs(t, err, billing.ErrProductNotFound)
		assert.Nil(t, placement)
	})
}

func TestRevenueCatPlace(T *testing.T) {
	T.Parallel()

	T.Run("places on the account the app logged in as and the store product", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()

		placement, err := RevenueCatPlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: account.ID,
			ProductID:  product.ExternalProductID,
		})
		require.NoError(t, err)

		assert.Equal(t, account.ID, placement.AccountID)
		assert.Equal(t, product.ID, placement.ProductID)
	})

	T.Run("an app user who is not an account here is refused", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()

		// What a purchase made before the app logged in to RevenueCat is filed under.
		placement, err := RevenueCatPlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: "$RCAnonymousID:" + fake.BuildFakeID(),
			ProductID:  product.ExternalProductID,
		})

		require.ErrorIs(t, err, platformidentity.ErrAccountNotFound)
		assert.Nil(t, placement)
	})

	T.Run("a store product nobody sells is refused", func(t *testing.T) {
		t.Parallel()

		product := fakes.BuildFakeProduct()
		account := identityfakes.BuildFakeAccount()

		placement, err := RevenueCatPlace(catalogOf(product), directoryOf(account))(t.Context(), nil, tenancy.Global(), &capitalism.SubscriptionState{
			ID:         fake.BuildFakeID(),
			CustomerID: account.ID,
			ProductID:  fake.BuildFakeID(),
		})

		require.ErrorIs(t, err, billing.ErrProductNotFound)
		assert.Nil(t, placement)
	})
}
