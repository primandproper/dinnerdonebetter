package payments

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformerrors "github.com/primandproper/platform-go/v11/errors"
	"github.com/primandproper/platform-go/v11/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Unit tests ---

func TestCreatePurchase(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreatePurchase(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
	})
}

func TestGetPurchase(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetPurchase(ctx, "")
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetPurchasesForAccount(T *testing.T) {
	T.Parallel()

	T.Run("with empty account id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetPurchasesForAccount(ctx, "", nil)
		require.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests ---

func TestQuerier_Integration_Purchases(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)
	product := createProductForTest(t, ctx, nil, dbc)

	input := &payments.PurchaseDatabaseCreationInput{
		ID:                    identifiers.New(),
		BelongsToAccount:      account.ID,
		ProductID:             product.ID,
		AmountCents:           1999,
		Currency:              "usd",
		ExternalTransactionID: "ext_txn_" + identifiers.New(),
	}

	created, err := dbc.CreatePurchase(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, input.ID, created.ID)
	assert.Equal(t, account.ID, created.BelongsToAccount)
	assert.Equal(t, product.ID, created.ProductID)

	// Get by ID
	fetched, err := dbc.GetPurchase(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)

	// List for account
	purchases, err := dbc.GetPurchasesForAccount(ctx, account.ID, nil)
	require.NoError(t, err)
	assert.Len(t, purchases.Data, 1)
}
