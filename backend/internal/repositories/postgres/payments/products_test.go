package payments

import (
	"context"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const exampleQuantity = 3

func createProductForTest(t *testing.T, ctx context.Context, input *payments.ProductDatabaseCreationInput, dbc *repository) *payments.Product {
	t.Helper()

	if input == nil {
		product := fakes.BuildFakeProduct()
		input = &payments.ProductDatabaseCreationInput{
			ID:                    product.ID,
			Name:                  product.Name,
			Description:           product.Description,
			Kind:                  product.Kind,
			AmountCents:           product.AmountCents,
			Currency:              product.Currency,
			BillingIntervalMonths: product.BillingIntervalMonths,
			ExternalProductID:     product.ExternalProductID,
		}
	}

	created, err := dbc.CreateProduct(ctx, input)
	require.NoError(t, err)
	require.NotNil(t, created)
	return created
}

// --- Unit tests ---

func TestCreateProduct(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateProduct(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestGetProduct(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetProduct(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestGetProductByExternalID(T *testing.T) {
	T.Parallel()

	T.Run("with empty external id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetProductByExternalID(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestUpdateProduct(T *testing.T) {
	T.Parallel()

	T.Run("with nil product", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateProduct(ctx, nil)
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrNilInputProvided)
	})
}

func TestArchiveProduct(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveProduct(ctx, "")
		assert.Error(t, err)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

func TestProductExists(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()
		c := buildInertClientForTest(t)

		exists, err := c.ProductExists(ctx, "")
		assert.Error(t, err)
		assert.False(t, exists)
		assert.ErrorIs(t, err, platformerrors.ErrInvalidIDProvided)
	})
}

// --- Integration tests ---

func TestQuerier_Integration_Products(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	product := fakes.BuildFakeProduct()
	input := &payments.ProductDatabaseCreationInput{
		ID:                    product.ID,
		Name:                  product.Name,
		Description:           product.Description,
		Kind:                  product.Kind,
		AmountCents:           product.AmountCents,
		Currency:              product.Currency,
		BillingIntervalMonths: product.BillingIntervalMonths,
		ExternalProductID:     product.ExternalProductID,
	}

	created := createProductForTest(t, ctx, input, dbc)
	assert.Equal(t, product.ID, created.ID)
	assert.Equal(t, product.Name, created.Name)

	// Get by ID
	fetched, err := dbc.GetProduct(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)

	// Get by external ID
	byExt, err := dbc.GetProductByExternalID(ctx, product.ExternalProductID)
	require.NoError(t, err)
	require.NotNil(t, byExt)
	assert.Equal(t, created.ID, byExt.ID)

	// Product exists
	exists, err := dbc.ProductExists(ctx, created.ID)
	require.NoError(t, err)
	assert.True(t, exists)

	// Update
	created.Name = "Updated Name"
	err = dbc.UpdateProduct(ctx, created)
	require.NoError(t, err)

	updated, err := dbc.GetProduct(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)

	// List
	products, err := dbc.GetProducts(ctx, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, products.Data)

	// Archive
	err = dbc.ArchiveProduct(ctx, created.ID)
	require.NoError(t, err)

	exists, err = dbc.ProductExists(ctx, created.ID)
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestQuerier_Integration_Products_GetProducts(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	for i := range exampleQuantity {
		product := fakes.BuildFakeProduct()
		product.Name = fmt.Sprintf("Product %d", i)
		input := &payments.ProductDatabaseCreationInput{
			ID:                    identifiers.New(),
			Name:                  product.Name,
			Description:           product.Description,
			Kind:                  product.Kind,
			AmountCents:           product.AmountCents,
			Currency:              product.Currency,
			BillingIntervalMonths: product.BillingIntervalMonths,
			ExternalProductID:     product.ExternalProductID,
		}
		createProductForTest(t, ctx, input, dbc)
	}

	result, err := dbc.GetProducts(ctx, &filtering.QueryFilter{MaxResponseSize: new(uint16(10))})
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.GreaterOrEqual(t, len(result.Data), exampleQuantity)
}
