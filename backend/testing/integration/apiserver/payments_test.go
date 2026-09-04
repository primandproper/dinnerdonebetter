package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	paymentssvcconverters "github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc/converters"

	"github.com/primandproper/platform-go/v13/billing"
	"github.com/primandproper/platform-go/v13/capitalism"
	"github.com/primandproper/platform-go/v13/pointer"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func createProductForTest(t *testing.T) *billing.Product {
	t.Helper()
	ctx := t.Context()

	example := fakes.BuildFakeProduct()
	created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
		Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(example),
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.GetCreated())

	converted := paymentssvcconverters.ConvertGRPCProductToProduct(created.GetCreated())
	assert.Equal(t, example.Name, converted.Name)
	assert.Equal(t, example.Description, converted.Description)
	assert.Equal(t, example.Kind, converted.Kind)
	assert.Equal(t, example.AmountCents, converted.AmountCents)
	assert.Equal(t, example.BillingIntervalMonths, converted.BillingIntervalMonths)
	assert.NotEmpty(t, converted.ID)

	res, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.GetCreated().GetId()})
	require.NoError(t, err)
	require.NotNil(t, res)

	product := paymentssvcconverters.ConvertGRPCProductToProduct(res.GetResult())
	assertRoughEquality(t, converted, product, defaultIgnoredFields()...)

	return product
}

func createSubscriptionForTest(t *testing.T, productID, accountID string) *billing.Subscription {
	t.Helper()
	ctx := t.Context()

	example := fakes.BuildFakeSubscription(accountID, productID)
	created, err := adminClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{
		Input: paymentssvcconverters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(example),
	})
	require.NoError(t, err)
	require.NotNil(t, created)
	require.NotNil(t, created.GetCreated())

	res, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.GetCreated().GetId()})
	require.NoError(t, err)
	require.NotNil(t, res)

	return paymentssvcconverters.ConvertGRPCSubscriptionToSubscription(res.GetResult())
}

// requireGRPCCode is the assertion that a refusal was the right one, not merely
// a refusal: the billing store's sentinels are mapped onto codes so that a
// client can tell a malformed product from a broken server.
func requireGRPCCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, expected, status.Code(err), "expected %s, got %v", expected, err)
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

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(fakes.BuildFakeProduct()),
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input empty name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProduct()
		input.Name = ""

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(input),
		})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	// The store's rule rather than this application's: a recurring product with no
	// interval is a subscription nothing knows when to renew.
	T.Run("invalid input recurring without an interval", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProduct()
		input.BillingIntervalMonths = 0

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(input),
		})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	T.Run("invalid input currency that is not a code", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := fakes.BuildFakeProduct()
		input.Currency = "dollars"

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(input),
		})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	T.Run("a provider-side id claimed twice is refused as a duplicate", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		first := createProductForTest(t)

		input := fakes.BuildFakeProduct()
		input.ExternalProductID = first.ExternalProductID

		created, err := adminClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(input),
		})
		requireGRPCCode(t, err, codes.AlreadyExists)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		created, err := testClient.CreateProduct(ctx, &paymentsgrpc.CreateProductRequest{
			Input: paymentssvcconverters.ConvertProductToGRPCProductCreationRequestInput(fakes.BuildFakeProduct()),
		})
		require.Error(t, err)
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
		converted := paymentssvcconverters.ConvertGRPCProductToProduct(retrieved.GetResult())
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
		requireGRPCCode(t, err, codes.NotFound)
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
		for _, p := range res.GetResults() {
			if p.GetId() == created.ID {
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
		newAmount := created.AmountCents + 1

		_, err := adminClient.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: &newName, AmountCents: &newAmount},
		})
		require.NoError(t, err)

		res, err := adminClient.GetProduct(ctx, &paymentsgrpc.GetProductRequest{ProductId: created.ID})
		require.NoError(t, err)
		assert.Equal(t, newName, res.GetResult().GetName())
		assert.Equal(t, newAmount, res.GetResult().GetAmountCents())
		// A patch: what the request did not name is kept.
		assert.Equal(t, created.Description, res.GetResult().GetDescription())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: pointer.To("x")},
		})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdateProduct(ctx, &paymentsgrpc.UpdateProductRequest{
			ProductId: created.ID,
			Input:     &paymentsgrpc.ProductUpdateRequestInput{Name: pointer.To("x")},
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
		requireGRPCCode(t, err, codes.NotFound)
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

	// The store's vocabulary is capitalism's, and a word outside it is refused
	// rather than stored — which is what happens to the "cancelled" this
	// application used to spell with two Ls.
	T.Run("a status outside the vocabulary is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		input := paymentssvcconverters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(fakes.BuildFakeSubscription(accountID, product.ID))
		input.Status = "cancelled"

		created, err := adminClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{Input: input})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	T.Run("a subscription to a product nobody sells is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		created, err := adminClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{
			Input: paymentssvcconverters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(fakes.BuildFakeSubscription(accountID, nonexistentID)),
		})
		requireGRPCCode(t, err, codes.NotFound)
		assert.Nil(t, created)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{
			Input: paymentssvcconverters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(fakes.BuildFakeSubscription(accountID, product.ID)),
		})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		_, testClient := createUserAndClientForTest(T)
		created, err := testClient.CreateSubscription(ctx, &paymentsgrpc.CreateSubscriptionRequest{
			Input: paymentssvcconverters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(fakes.BuildFakeSubscription(accountID, product.ID)),
		})
		require.Error(t, err)
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
		require.NoError(t, err)
		converted := paymentssvcconverters.ConvertGRPCSubscriptionToSubscription(retrieved.GetResult())
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
		requireGRPCCode(t, err, codes.NotFound)
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
		for _, s := range res.GetResults() {
			if s.GetId() == created.ID {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	// The account read is the session's, not the request's: naming another
	// account's id shows the caller their own subscriptions, not somebody else's.
	T.Run("another account's id does not read its subscriptions", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, ownerClient := createUserAndClientForTest(t)
		ownerAccountID := getAccountIDForTest(t, ownerClient)
		created := createSubscriptionForTest(t, product.ID, ownerAccountID)

		_, otherClient := createUserAndClientForTest(t)

		res, err := otherClient.GetSubscriptionsForAccount(ctx, &paymentsgrpc.GetSubscriptionsForAccountRequest{AccountId: ownerAccountID})
		require.NoError(t, err)

		for _, s := range res.GetResults() {
			assert.NotEqual(t, created.ID, s.GetId())
		}
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

		newStatus := string(capitalism.SubscriptionStatusCanceled)
		_, err := adminClient.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: &newStatus},
		})
		require.NoError(t, err)

		res, err := adminClient.GetSubscription(ctx, &paymentsgrpc.GetSubscriptionRequest{SubscriptionId: created.ID})
		require.NoError(t, err)
		assert.Equal(t, newStatus, res.GetResult().GetStatus())

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

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: pointer.To(string(capitalism.SubscriptionStatusCanceled))},
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

		_, testClient := createUserAndClientForTest(T)
		_, err := testClient.UpdateSubscription(ctx, &paymentsgrpc.UpdateSubscriptionRequest{
			SubscriptionId: created.ID,
			Input:          &paymentsgrpc.SubscriptionUpdateRequestInput{Status: pointer.To(string(capitalism.SubscriptionStatusCanceled))},
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
		requireGRPCCode(t, err, codes.NotFound)

		AssertAuditLogContainsFuzzy(t, ctx, accountClient, accountID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "subscriptions", RelevantID: created.ID},
			{EventType: "archived", ResourceType: "subscriptions", RelevantID: created.ID},
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
