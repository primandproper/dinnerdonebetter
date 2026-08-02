package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	paymentssvcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createProductForTest(t *testing.T) *payments.Product {
	t.Helper()
	ctx := t.Context()

	exampleInput := fakes.BuildFakeProductCreationRequestInput()
	grpcInput := paymentssvcconverters.ConvertProductCreationRequestInputToGRPC(exampleInput)
	created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
		Input: grpcInput,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.Created)

	converted := paymentssvcconverters.ConvertGRPCProductToProduct(created.Created)
	assert.Equal(t, exampleInput.Name, converted.Name)
	assert.Equal(t, exampleInput.Description, converted.Description)
	assert.Equal(t, exampleInput.Kind, converted.Kind)
	assert.NotEmpty(t, converted.ID)

	res, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, res)

	product := paymentssvcconverters.ConvertGRPCProductToProduct(res.Result)
	assertRoughEquality(t, converted, product, defaultIgnoredFields()...)

	return product
}

func createSubscriptionForTest(t *testing.T, productID, accountID string) *payments.Subscription {
	t.Helper()
	ctx := t.Context()

	exampleInput := fakes.BuildFakeSubscriptionCreationRequestInput(accountID, productID)
	grpcInput := paymentssvcconverters.ConvertSubscriptionCreationRequestInputToGRPC(exampleInput)
	created, err := adminClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{
		Input: grpcInput,
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.Created)

	res, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.Created.Id})
	require.NoError(t, err)
	require.NotNil(t, res)

	return paymentssvcconverters.ConvertGRPCSubscriptionToSubscription(res.Result)
}

func TestPayments_CreateProduct(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		createProductForTest(t)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProductCreationRequestInput()
		grpcInput := paymentssvcconverters.ConvertProductCreationRequestInputToGRPC(input)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input empty name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProductCreationRequestInput()
		input.Name = ""
		grpcInput := paymentssvcconverters.ConvertProductCreationRequestInputToGRPC(input)

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input empty description", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProductCreationRequestInput()
		input.Description = ""
		grpcInput := paymentssvcconverters.ConvertProductCreationRequestInputToGRPC(input)

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)
		input := fakes.BuildFakeProductCreationRequestInput()
		grpcInput := paymentssvcconverters.ConvertProductCreationRequestInputToGRPC(input)

		created, err := testClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestPayments_GetProduct(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)

		retrieved, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.ID})
		require.NoError(t, err)
		require.NotNil(t, retrieved)
		converted := paymentssvcconverters.ConvertGRPCProductToProduct(retrieved.Result)
		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.ID})
		assert.Error(t, err)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestPayments_GetProducts(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)

		res, err := adminClient.GetProducts(ctx, &paymentsgrpc.GetProductsRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)

		var found bool
		for _, p := range res.Results {
			if p.Id == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetProducts(ctx, &paymentsgrpc.GetProductsRequest{})
		assert.Error(t, err)
	})
}

func TestPayments_UpdateProduct(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		newName := "updated product name"

		_, err := adminClient.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: &newName},
		})
		require.NoError(t, err)

		res, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.ID})
		require.NoError(t, err)
		assert.Equal(t, newName, res.Result.Name)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		newName := "x"
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: &newName},
		})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		_, testClient := createUserAndClientForTest(T)
		newName := "x"

		_, err := testClient.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: &newName},
		})
		assert.Error(t, err)
	})
}

func TestPayments_ArchiveProduct(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)

		_, err := adminClient.ArchiveProduct(ctx, &paymentsgrpc.ArchiveProductRequest{ProductId: created.ID})
		require.NoError(t, err)

		res, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.ID})
		assert.Nil(t, res)
		assert.Error(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveProduct(ctx, &paymentsgrpc.ArchiveProductRequest{ProductId: created.ID})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveProduct(ctx, &paymentsgrpc.ArchiveProductRequest{ProductId: created.ID})
		assert.Error(t, err)
	})
}

func TestPayments_CreateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		created := createSubscriptionForTest(t, product.ID, accountID)

		AssertAuditLogContainsFuzzy(t, ctx, accountClient, accountID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "subscriptions", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		input := fakes.BuildFakeSubscriptionCreationRequestInput(accountID, product.ID)
		grpcInput := paymentssvcconverters.ConvertSubscriptionCreationRequestInputToGRPC(input)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		input := fakes.BuildFakeSubscriptionCreationRequestInput(accountID, product.ID)
		grpcInput := paymentssvcconverters.ConvertSubscriptionCreationRequestInputToGRPC(input)

		_, testClient := createUserAndClientForTest(T)
		created, err := testClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{Input: grpcInput})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestPayments_GetSubscription(T *testing.T) {
	T.Parallel()

	_, testClient := createUserAndClientForTest(T)

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		accountID := getAccountIDForTest(t, testClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		retrieved, err := testClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.ID})
		assert.NoError(t, err)
		converted := paymentssvcconverters.ConvertGRPCSubscriptionToSubscription(retrieved.Result)
		assertRoughEquality(t, created, converted, defaultIgnoredFields()...)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		accountID := getAccountIDForTest(t, testClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.ID})
		assert.Error(t, err)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: nonexistentID})
		assert.Error(t, err)
	})
}

func TestPayments_GetSubscriptionsForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		res, err := accountClient.GetSubscriptionsForAccount(ctx, &paymentsgrpc.GetSubscriptionsForAccountRequest{AccountId: accountID})
		require.NoError(t, err)

		var found bool
		for _, s := range res.Results {
			if s.Id == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetSubscriptionsForAccount(ctx, &paymentsgrpc.GetSubscriptionsForAccountRequest{AccountId: accountID})
		assert.Error(t, err)
	})
}

func TestPayments_UpdateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		newStatus := payments.SubscriptionStatusCancelled
		_, err := adminClient.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: &newStatus},
		})
		require.NoError(t, err)

		res, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.ID})
		require.NoError(t, err)
		assert.Equal(t, newStatus, res.Result.Status)

		AssertAuditLogContainsFuzzy(t, ctx, accountClient, accountID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "subscriptions", RelevantID: created.ID},
			{EventType: "updated", ResourceType: "subscriptions", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		newStatus := "cancelled"
		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: &newStatus},
		})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		newStatus := "cancelled"
		_, testClient := createUserAndClientForTest(T)
		_, err := testClient.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: &newStatus},
		})
		assert.Error(t, err)
	})
}

func TestPayments_ArchiveSubscription(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		_, err := adminClient.ArchiveSubscription(ctx, &paymentsgrpc.ArchiveSubscriptionRequest{SubscriptionId: created.ID})
		require.NoError(t, err)

		res, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.ID})
		assert.Nil(t, res)
		assert.Error(t, err)

		AssertAuditLogContainsFuzzy(t, ctx, accountClient, accountID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "subscriptions", RelevantID: created.ID},
		})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)
		created := createSubscriptionForTest(t, product.ID, accountID)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveSubscription(ctx, &paymentsgrpc.ArchiveSubscriptionRequest{SubscriptionId: created.ID})
		assert.Error(t, err)
	})
}

func TestPayments_GetPurchasesForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path may be empty", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		res, err := accountClient.GetPurchasesForAccount(ctx, &paymentsgrpc.GetPurchasesForAccountRequest{AccountId: accountID})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetPurchasesForAccount(ctx, &paymentsgrpc.GetPurchasesForAccountRequest{AccountId: accountID})
		assert.Error(t, err)
	})
}

func TestPayments_GetPaymentHistoryForAccount(T *testing.T) {
	T.Parallel()

	T.Run("happy path may be empty", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		res, err := accountClient.GetPaymentHistoryForAccount(ctx, &paymentsgrpc.GetPaymentHistoryForAccountRequest{AccountId: accountID})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetPaymentHistoryForAccount(ctx, &paymentsgrpc.GetPaymentHistoryForAccountRequest{AccountId: accountID})
		assert.Error(t, err)
	})
}
