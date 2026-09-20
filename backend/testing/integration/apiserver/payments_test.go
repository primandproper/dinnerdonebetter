package integration

import (
	"testing"

	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v14/billing"
	billingpb "github.com/primandproper/platform-go/v14/billing/billingpb"
	"github.com/primandproper/primitives-go/v2/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The billing surface is platform's, and it is two RPCs shorter than the one it replaced.
//
// CreateSubscription and UpdateSubscription are gone, and deliberately: a Subscription here
// mirrors what a payment provider says is currently paid for, and billing/standing reads those
// rows to decide entitlement. A subscription created over the wire would grant paid features
// with no payment behind them. The rows still have to exist for the reads below, so they are
// written the way the webhook handler writes them — through the store, on a transaction.
//
// One read changed name rather than meaning: GetPaymentHistoryForAccount is
// ListTransactionsForAccount.

// productInputForTest builds what a client sends to add a product to the catalog.
func productInputForTest() *billingpb.ProductCreationInput {
	example := fakes.BuildFakeProduct()

	return &billingpb.ProductCreationInput{
		Name:                  example.Name,
		Description:           example.Description,
		Kind:                  billingpb.ProductKind_PRODUCT_KIND_RECURRING,
		Currency:              example.Currency,
		ExternalProductId:     example.ExternalProductID,
		AmountCents:           example.AmountCents,
		BillingIntervalMonths: example.BillingIntervalMonths,
	}
}

func createProductForTest(t *testing.T) *billingpb.Product {
	t.Helper()
	ctx := t.Context()

	input := productInputForTest()

	created, err := adminClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: input})
	require.NoError(t, err)
	require.NotNil(t, created.GetResult())

	assert.Equal(t, input.GetName(), created.GetResult().GetName())
	assert.Equal(t, input.GetDescription(), created.GetResult().GetDescription())
	assert.Equal(t, input.GetKind(), created.GetResult().GetKind())
	assert.Equal(t, input.GetAmountCents(), created.GetResult().GetAmountCents())
	assert.Equal(t, input.GetBillingIntervalMonths(), created.GetResult().GetBillingIntervalMonths())
	assert.NotEmpty(t, created.GetResult().GetId())

	res, err := adminClient.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: created.GetResult().GetId()})
	require.NoError(t, err)
	require.NotNil(t, res.GetResult())
	assert.Equal(t, created.GetResult().GetId(), res.GetResult().GetId())

	return res.GetResult()
}

// createSubscriptionForTest writes a subscription the only way there is to write one.
//
// Through the store rather than over the wire, because the RPC that used to do this is gone
// — see the note at the top of this file. The transaction is the caller's now, so this opens
// one, which is what the webhook handler does when a provider tells it about an agreement.
func createSubscriptionForTest(t *testing.T, productID, accountID string) *billing.Subscription {
	t.Helper()
	ctx := t.Context()

	example := fakes.BuildFakeSubscription(accountID, productID)

	var created *billing.Subscription

	require.NoError(t, databaseClient.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		created, writeErr = billingStore.CreateSubscription(ctx, tx, ddbpayments.Scope(), example)

		return writeErr
	}))
	require.NotNil(t, created)

	return created
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
		created, err := c.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: productInputForTest()})
		require.Error(t, err)
		assert.Nil(t, created)
	})

	T.Run("invalid input empty name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := productInputForTest()
		input.Name = ""

		created, err := adminClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: input})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	// The store's rule rather than this application's: a recurring product with no
	// interval is a subscription nothing knows when to renew.
	T.Run("invalid input recurring without an interval", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := productInputForTest()
		input.BillingIntervalMonths = 0

		created, err := adminClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: input})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	T.Run("invalid input currency that is not a code", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		input := productInputForTest()
		input.Currency = "dollars"

		created, err := adminClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: input})
		requireGRPCCode(t, err, codes.InvalidArgument)
		assert.Nil(t, created)
	})

	T.Run("a provider-side id claimed twice is refused as a duplicate", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		first := createProductForTest(t)

		input := productInputForTest()
		input.ExternalProductId = first.GetExternalProductId()

		created, err := adminClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: input})
		requireGRPCCode(t, err, codes.AlreadyExists)
		assert.Nil(t, created)
	})

	T.Run("non-admin users are forbidden from creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(T)

		created, err := testClient.CreateProduct(ctx, &billingpb.CreateProductRequest{Input: productInputForTest()})
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

		retrieved, err := adminClient.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: created.GetId()})
		require.NoError(t, err)
		require.NotNil(t, retrieved.GetResult())
		assert.Equal(t, created.GetId(), retrieved.GetResult().GetId())
		assert.Equal(t, created.GetName(), retrieved.GetResult().GetName())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: created.GetId()})
		assert.Error(t, err)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: nonexistentID})
		requireGRPCCode(t, err, codes.NotFound)
	})
}

func TestPayments_GetProducts(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)

		res, err := adminClient.ListProducts(ctx, &billingpb.ListProductsRequest{})
		require.NoError(t, err)
		require.NotNil(t, res)

		var found bool
		for _, p := range res.GetResults() {
			if p.GetId() == created.GetId() {
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
		_, err := c.ListProducts(ctx, &billingpb.ListProductsRequest{})
		assert.Error(t, err)
	})
}

func TestPayments_UpdateProduct(T *testing.T) {
	T.Parallel()

	// UpdateProduct replaces rather than patches: its input carries every field, so a caller
	// restates the ones it is keeping. The local RPC took pointer fields and merged.
	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)

		const newName = "updated product name"
		newAmount := created.GetAmountCents() + 1

		_, err := adminClient.UpdateProduct(ctx, &billingpb.UpdateProductRequest{
			ProductId: created.GetId(),
			Input: &billingpb.ProductUpdateInput{
				Name:                  newName,
				Description:           created.GetDescription(),
				Kind:                  created.GetKind(),
				Currency:              created.GetCurrency(),
				ExternalProductId:     created.GetExternalProductId(),
				AmountCents:           newAmount,
				BillingIntervalMonths: created.GetBillingIntervalMonths(),
			},
		})
		require.NoError(t, err)

		res, err := adminClient.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: created.GetId()})
		require.NoError(t, err)
		assert.Equal(t, newName, res.GetResult().GetName())
		assert.Equal(t, newAmount, res.GetResult().GetAmountCents())
		assert.Equal(t, created.GetDescription(), res.GetResult().GetDescription())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.UpdateProduct(ctx, &billingpb.UpdateProductRequest{
			ProductId: created.GetId(),
			Input:     &billingpb.ProductUpdateInput{Name: "x"},
		})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.UpdateProduct(ctx, &billingpb.UpdateProductRequest{
			ProductId: created.GetId(),
			Input:     &billingpb.ProductUpdateInput{Name: "x"},
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

		_, err := adminClient.ArchiveProduct(ctx, &billingpb.ArchiveProductRequest{ProductId: created.GetId()})
		require.NoError(t, err)

		res, err := adminClient.GetProduct(ctx, &billingpb.GetProductRequest{ProductId: created.GetId()})
		assert.Nil(t, res)
		requireGRPCCode(t, err, codes.NotFound)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		c := buildUnauthenticatedGRPCClientForTest(t)

		_, err := c.ArchiveProduct(ctx, &billingpb.ArchiveProductRequest{ProductId: created.GetId()})
		assert.Error(t, err)
	})

	T.Run("non-admin forbidden", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		created := createProductForTest(t)
		_, testClient := createUserAndClientForTest(T)

		_, err := testClient.ArchiveProduct(ctx, &billingpb.ArchiveProductRequest{ProductId: created.GetId()})
		assert.Error(t, err)
	})
}

// TestPayments_SubscriptionsAreNotWritableOverTheWire is the ruling, pinned.
//
// There is no RPC to write one, so there is nothing here to call — which is the assertion.
// A client that wants an account subscribed drives the provider, and the provider's webhook
// is what reaches the store. See internal/domain/payments/manager.
func TestPayments_SubscriptionsAreNotWritableOverTheWire(T *testing.T) {
	T.Parallel()

	T.Run("the surface offers no create or update", func(t *testing.T) {
		t.Parallel()

		// Every method the billing service declares, none of which writes a subscription.
		// A rename upstream reds this rather than silently reopening the hole.
		for _, method := range []string{
			billingpb.BillingService_CreateProduct_FullMethodName,
			billingpb.BillingService_UpdateProduct_FullMethodName,
			billingpb.BillingService_ArchiveSubscription_FullMethodName,
		} {
			assert.NotContains(t, method, "CreateSubscription")
			assert.NotContains(t, method, "UpdateSubscription")
		}
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
		created := createSubscriptionForTest(t, product.GetId(), accountID)

		retrieved, err := testClient.GetSubscription(ctx, &billingpb.GetSubscriptionRequest{SubscriptionId: created.ID})
		require.NoError(t, err)
		assert.Equal(t, created.ID, retrieved.GetResult().GetId())
		assert.Equal(t, created.BelongsToAccount, retrieved.GetResult().GetBelongsToAccount())
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		accountID := getAccountIDForTest(t, testClient)
		created := createSubscriptionForTest(t, product.GetId(), accountID)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.GetSubscription(ctx, &billingpb.GetSubscriptionRequest{SubscriptionId: created.ID})
		assert.Error(t, err)
	})

	T.Run("nonexistent ID", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, err := adminClient.GetSubscription(ctx, &billingpb.GetSubscriptionRequest{SubscriptionId: nonexistentID})
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
		created := createSubscriptionForTest(t, product.GetId(), accountID)

		res, err := accountClient.ListSubscriptionsForAccount(ctx, &billingpb.ListSubscriptionsForAccountRequest{AccountId: accountID})
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

	// Naming another account's id is refused rather than quietly answered with the
	// caller's own, which is what this used to assert.
	//
	// The refusal is the better answer of the two. Silently substituting the session's
	// account means a client that got the id wrong is handed a correct-looking page of
	// somebody else's subscriptions — no, of its own, labelled with an id it did not ask
	// for — and cannot tell that its request was ignored.
	T.Run("another account's id is refused", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		product := createProductForTest(t)
		_, ownerClient := createUserAndClientForTest(t)
		ownerAccountID := getAccountIDForTest(t, ownerClient)
		createSubscriptionForTest(t, product.GetId(), ownerAccountID)

		_, otherClient := createUserAndClientForTest(t)

		_, err := otherClient.ListSubscriptionsForAccount(ctx, &billingpb.ListSubscriptionsForAccountRequest{AccountId: ownerAccountID})
		require.Error(t, err)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListSubscriptionsForAccount(ctx, &billingpb.ListSubscriptionsForAccountRequest{AccountId: accountID})
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
		created := createSubscriptionForTest(t, product.GetId(), accountID)

		_, err := adminClient.ArchiveSubscription(ctx, &billingpb.ArchiveSubscriptionRequest{SubscriptionId: created.ID})
		require.NoError(t, err)

		res, err := adminClient.GetSubscription(ctx, &billingpb.GetSubscriptionRequest{SubscriptionId: created.ID})
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
		created := createSubscriptionForTest(t, product.GetId(), accountID)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveSubscription(ctx, &billingpb.ArchiveSubscriptionRequest{SubscriptionId: created.ID})
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

		res, err := accountClient.ListPurchasesForAccount(ctx, &billingpb.ListPurchasesForAccountRequest{AccountId: accountID})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListPurchasesForAccount(ctx, &billingpb.ListPurchasesForAccountRequest{AccountId: accountID})
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

		res, err := accountClient.ListTransactionsForAccount(ctx, &billingpb.ListTransactionsForAccountRequest{AccountId: accountID})
		require.NoError(t, err)
		require.NotNil(t, res)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, accountClient := createUserAndClientForTest(t)
		accountID := getAccountIDForTest(t, accountClient)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ListTransactionsForAccount(ctx, &billingpb.ListTransactionsForAccountRequest{AccountId: accountID})
		assert.Error(t, err)
	})
}
