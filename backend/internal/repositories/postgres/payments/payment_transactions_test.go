package payments

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests ---

func TestCreatePaymentTransaction(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreatePaymentTransaction(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestGetPaymentTransactionsForAccount(T *testing.T) {
	T.Parallel()

	T.Run("with empty account id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetPaymentTransactionsForAccount(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests ---

func TestQuerier_Integration_PaymentTransactions(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	input := &payments.PaymentTransactionDatabaseCreationInput{
		ID:                    identifiers.New(),
		BelongsToAccount:      account.ID,
		ExternalTransactionID: "ext_txn_" + identifiers.New(),
		AmountCents:           1999,
		Currency:              "usd",
		Status:                payments.PaymentTransactionStatusSucceeded,
	}

	created, err := dbc.CreatePaymentTransaction(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, input.ID, created.ID)
	assert.Equal(t, account.ID, created.BelongsToAccount)
	assert.Equal(t, payments.PaymentTransactionStatusSucceeded, created.Status)

	// List for account
	transactions, err := dbc.GetPaymentTransactionsForAccount(ctx, account.ID, nil)
	require.NoError(t, err)
	assert.Len(t, transactions.Data, 1)
	assert.Equal(t, created.ID, transactions.Data[0].ID)
}

func TestQuerier_Integration_PaymentTransactions_WithSubscription(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	product := createProductForTest(t, ctx, nil, dbc)

	subInput := &payments.SubscriptionDatabaseCreationInput{
		ID:                     identifiers.New(),
		BelongsToAccount:       account.ID,
		ProductID:              product.ID,
		ExternalSubscriptionID: "ext_sub_" + identifiers.New(),
		Status:                 payments.SubscriptionStatusActive,
		CurrentPeriodStart:     fake.BuildFakeTime(),
		CurrentPeriodEnd:       fake.BuildFakeTime().AddDate(0, 1, 0),
	}
	subscription, err := dbc.CreateSubscription(ctx, subInput)
	require.NoError(t, err)
	require.NotNil(t, subscription)

	subID := subscription.ID
	input := &payments.PaymentTransactionDatabaseCreationInput{
		ID:                    identifiers.New(),
		BelongsToAccount:      account.ID,
		SubscriptionID:        &subID,
		ExternalTransactionID: "ext_txn_" + identifiers.New(),
		AmountCents:           1999,
		Currency:              "usd",
		Status:                payments.PaymentTransactionStatusSucceeded,
	}

	created, err := dbc.CreatePaymentTransaction(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, input.ID, created.ID)
	assert.NotNil(t, created.SubscriptionID)
	assert.Equal(t, subscription.ID, *created.SubscriptionID)
}
