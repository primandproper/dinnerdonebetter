package converters

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	fakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertUserDataCollectionToGRPCDataCollection(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		accountID := identifiers.New()
		productID := identifiers.New()

		input := &payments.UserDataCollection{
			Subscriptions:       []payments.Subscription{*fakes.BuildFakeSubscription(accountID, productID)},
			Purchases:           []payments.Purchase{{ID: identifiers.New(), BelongsToAccount: accountID, ProductID: productID}},
			PaymentTransactions: []payments.PaymentTransaction{{ID: identifiers.New(), BelongsToAccount: accountID}},
		}

		result := ConvertUserDataCollectionToGRPCDataCollection(input)

		require.NotNil(t, result)
		assert.Len(t, result.Subscriptions, 1)
		assert.Len(t, result.Purchases, 1)
		assert.Len(t, result.PaymentTransactions, 1)
	})

	T.Run("empty", func(t *testing.T) {
		t.Parallel()

		result := ConvertUserDataCollectionToGRPCDataCollection(&payments.UserDataCollection{})

		require.NotNil(t, result)
		assert.Empty(t, result.Subscriptions)
		assert.Empty(t, result.Purchases)
		assert.Empty(t, result.PaymentTransactions)
	})
}
