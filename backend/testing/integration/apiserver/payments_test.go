package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"

	"github.com/primandproper/platform-go/v15/billing"
	billingpb "github.com/primandproper/platform-go/v15/billing/billingpb"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// The billing surface's behavior — the catalog, subscription reads, confinement to an account and
// a tenant — is asserted by platform's billing conformance suite, run against this deployment in
// conformance_test.go, and its reach without a caller by the anonymous suite. What remains here
// is this application's own: that the catalog is a service admin's to write, the product
// validation the suite does not reach, and the audit entries its store records.
//
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
		created, writeErr = billingStore.CreateSubscription(ctx, tx, tenancy.Global(), example)

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
			{EventType: "created", ResourceType: billing.ResourceTypeSubscription, RelevantID: created.ID},
			{EventType: "archived", ResourceType: billing.ResourceTypeSubscription, RelevantID: created.ID},
		})
	})
}
