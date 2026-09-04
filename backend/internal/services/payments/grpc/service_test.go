package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	ddbpayments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/services/payments/grpc/converters"

	"github.com/primandproper/platform-go/v13/billing"
	billingmock "github.com/primandproper/platform-go/v13/billing/mock"
	"github.com/primandproper/platform-go/v13/capitalism"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/pointer"
	"github.com/primandproper/platform-go/v13/tenancy"

	gofakeit "github.com/brianvoe/gofakeit/v7"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildTestService builds a service backed by the given store mock. A nil mock gets
// an unconfigured one, which panics if any of its methods are called — so a test that
// reaches the store where it did not arrange to fails loudly rather than on a nil
// result.
func buildTestService(t *testing.T, store *billingmock.StoreMock) *serviceImpl {
	t.Helper()

	if store == nil {
		store = &billingmock.StoreMock{}
	}

	return &serviceImpl{
		tracer:  tracing.NewTracerForTest(t.Name()),
		logger:  loggingnoop.NewLogger(),
		billing: store,
	}
}

// requester is who a test is making its requests as.
type requester struct {
	ctx     context.Context
	userID  string
	account string
}

// userContextForTest returns an ordinary signed-in user with an active account,
// which is who reads their own subscriptions and ledger.
func userContextForTest(t *testing.T) requester {
	t.Helper()

	r := requester{
		userID:  fake.BuildFakeID(),
		account: fake.BuildFakeID(),
	}

	r.ctx = sessions.AttachToContext(t.Context(), &sessions.ContextData{
		ActiveAccountID: r.account,
		Requester: sessions.RequesterInfo{
			UserID:             r.userID,
			EmailAddress:       gofakeit.Email(),
			ServicePermissions: authorization.NewServiceRolePermissionChecker([]string{authorization.ServiceUserRole.String()}, nil),
		},
	})

	return r
}

// requireCode asserts that a handler refused with the given gRPC code.
func requireCode(t *testing.T, err error, expected codes.Code) {
	t.Helper()

	require.Error(t, err)
	assert.Equal(t, expected, status.Code(err), "got %v", err)
}

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		service := NewService(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), &billingmock.StoreMock{})

		assert.NotNil(t, service)
		assert.Implements(t, (*paymentssvc.PaymentsServiceServer)(nil), service)
	})
}

func TestProvideMethodPermissions(t *testing.T) {
	t.Parallel()

	t.Run("covers every method the service serves", func(t *testing.T) {
		t.Parallel()

		permissions := ProvideMethodPermissions()

		// A method with no entry is refused by the interceptor, so an RPC added
		// without one is an RPC nobody can call.
		for _, method := range []string{
			paymentssvc.PaymentsService_CreateProduct_FullMethodName,
			paymentssvc.PaymentsService_GetProduct_FullMethodName,
			paymentssvc.PaymentsService_GetProducts_FullMethodName,
			paymentssvc.PaymentsService_UpdateProduct_FullMethodName,
			paymentssvc.PaymentsService_ArchiveProduct_FullMethodName,
			paymentssvc.PaymentsService_CreateSubscription_FullMethodName,
			paymentssvc.PaymentsService_GetSubscription_FullMethodName,
			paymentssvc.PaymentsService_GetSubscriptionsForAccount_FullMethodName,
			paymentssvc.PaymentsService_UpdateSubscription_FullMethodName,
			paymentssvc.PaymentsService_ArchiveSubscription_FullMethodName,
			paymentssvc.PaymentsService_GetPurchasesForAccount_FullMethodName,
			paymentssvc.PaymentsService_GetPaymentHistoryForAccount_FullMethodName,
		} {
			assert.NotEmpty(t, permissions[method], "no permissions for %s", method)
		}
	})
}

func TestService_CreateProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		example := fakes.BuildFakeProduct()

		var written *billing.Product

		service := buildTestService(t, &billingmock.StoreMock{
			CreateProductFunc: func(_ context.Context, scope tenancy.Scope, product *billing.Product) (*billing.Product, error) {
				assert.Equal(t, ddbpayments.Scope(), scope)
				written = product

				return example, nil
			},
		})

		res, err := service.CreateProduct(r.ctx, &paymentssvc.CreateProductRequest{
			Input: converters.ConvertProductToGRPCProductCreationRequestInput(example),
		})
		require.NoError(t, err)

		// What reached the store is what the request asked for, in this application's scope.
		require.NotNil(t, written)
		assert.Equal(t, example.Name, written.Name)
		assert.Equal(t, example.AmountCents, written.AmountCents)
		assert.Equal(t, ddbpayments.Scope(), written.Scope)

		assert.Equal(t, example.ID, res.GetCreated().GetId())
		assert.Equal(t, r.account, res.GetResponseDetails().GetCurrentAccountId())
	})

	T.Run("without a session", func(t *testing.T) {
		t.Parallel()

		service := buildTestService(t, nil)

		_, err := service.CreateProduct(t.Context(), &paymentssvc.CreateProductRequest{
			Input: converters.ConvertProductToGRPCProductCreationRequestInput(fakes.BuildFakeProduct()),
		})
		requireCode(t, err, codes.Unauthenticated)
	})

	T.Run("without input", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, nil)

		_, err := service.CreateProduct(r.ctx, &paymentssvc.CreateProductRequest{})
		requireCode(t, err, codes.InvalidArgument)
	})

	// The store's refusals are mapped rather than reported as an outage.
	T.Run("with a product the store refuses", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, &billingmock.StoreMock{
			CreateProductFunc: func(context.Context, tenancy.Scope, *billing.Product) (*billing.Product, error) {
				return nil, billing.ErrInvalidCurrency
			},
		})

		_, err := service.CreateProduct(r.ctx, &paymentssvc.CreateProductRequest{
			Input: converters.ConvertProductToGRPCProductCreationRequestInput(fakes.BuildFakeProduct()),
		})
		requireCode(t, err, codes.InvalidArgument)
	})

	T.Run("with a provider-side id already claimed", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, &billingmock.StoreMock{
			CreateProductFunc: func(context.Context, tenancy.Scope, *billing.Product) (*billing.Product, error) {
				return nil, billing.ErrProductExists
			},
		})

		_, err := service.CreateProduct(r.ctx, &paymentssvc.CreateProductRequest{
			Input: converters.ConvertProductToGRPCProductCreationRequestInput(fakes.BuildFakeProduct()),
		})
		requireCode(t, err, codes.AlreadyExists)
	})
}

func TestService_GetProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		example := fakes.BuildFakeProduct()

		service := buildTestService(t, &billingmock.StoreMock{
			GetProductFunc: func(_ context.Context, scope tenancy.Scope, productID string) (*billing.Product, error) {
				assert.Equal(t, ddbpayments.Scope(), scope)
				assert.Equal(t, example.ID, productID)

				return example, nil
			},
		})

		res, err := service.GetProduct(r.ctx, &paymentssvc.GetProductRequest{ProductId: example.ID})
		require.NoError(t, err)
		assert.Equal(t, example.ID, res.GetResult().GetId())
		assert.Equal(t, example.AmountCents, res.GetResult().GetAmountCents())
	})

	T.Run("with a product nobody has", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, &billingmock.StoreMock{
			GetProductFunc: func(context.Context, tenancy.Scope, string) (*billing.Product, error) {
				return nil, billing.ErrProductNotFound
			},
		})

		_, err := service.GetProduct(r.ctx, &paymentssvc.GetProductRequest{ProductId: fake.BuildFakeID()})
		requireCode(t, err, codes.NotFound)
	})
}

func TestService_GetProducts(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		page := fakes.BuildFakeProductList()

		service := buildTestService(t, &billingmock.StoreMock{
			ListProductsFunc: func(_ context.Context, scope tenancy.Scope, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Product], error) {
				assert.Equal(t, ddbpayments.Scope(), scope)

				return page, nil
			},
		})

		res, err := service.GetProducts(r.ctx, &paymentssvc.GetProductsRequest{})
		require.NoError(t, err)
		assert.Len(t, res.GetResults(), len(page.Data))
	})
}

func TestService_UpdateProduct(T *testing.T) {
	T.Parallel()

	// A patch: the fields the request set are written and the rest are kept, and
	// what reaches the store is the row it handed back with those changes applied.
	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		example := fakes.BuildFakeProduct()
		newAmount := example.AmountCents + 1

		var written *billing.Product

		service := buildTestService(t, &billingmock.StoreMock{
			GetProductFunc: func(context.Context, tenancy.Scope, string) (*billing.Product, error) {
				copied := *example

				return &copied, nil
			},
			UpdateProductFunc: func(_ context.Context, _ tenancy.Scope, product *billing.Product) error {
				written = product

				return nil
			},
		})

		_, err := service.UpdateProduct(r.ctx, &paymentssvc.UpdateProductRequest{
			ProductId: example.ID,
			Input:     &paymentssvc.ProductUpdateRequestInput{AmountCents: &newAmount},
		})
		require.NoError(t, err)

		require.NotNil(t, written)
		assert.Equal(t, newAmount, written.AmountCents)
		assert.Equal(t, example.Name, written.Name)
		assert.Equal(t, example.Kind, written.Kind)
	})

	T.Run("without input", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, nil)

		_, err := service.UpdateProduct(r.ctx, &paymentssvc.UpdateProductRequest{ProductId: fake.BuildFakeID()})
		requireCode(t, err, codes.InvalidArgument)
	})

	T.Run("with a product nobody has", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, &billingmock.StoreMock{
			GetProductFunc: func(context.Context, tenancy.Scope, string) (*billing.Product, error) {
				return nil, billing.ErrProductNotFound
			},
		})

		_, err := service.UpdateProduct(r.ctx, &paymentssvc.UpdateProductRequest{
			ProductId: fake.BuildFakeID(),
			Input:     &paymentssvc.ProductUpdateRequestInput{Name: pointer.To("x")},
		})
		requireCode(t, err, codes.NotFound)
	})
}

func TestService_ArchiveProduct(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		productID := fake.BuildFakeID()

		service := buildTestService(t, &billingmock.StoreMock{
			ArchiveProductFunc: func(_ context.Context, scope tenancy.Scope, id string) error {
				assert.Equal(t, ddbpayments.Scope(), scope)
				assert.Equal(t, productID, id)

				return nil
			},
		})

		_, err := service.ArchiveProduct(r.ctx, &paymentssvc.ArchiveProductRequest{ProductId: productID})
		require.NoError(t, err)
	})
}

func TestService_CreateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		example := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())

		var written *billing.Subscription

		service := buildTestService(t, &billingmock.StoreMock{
			CreateSubscriptionFunc: func(_ context.Context, scope tenancy.Scope, subscription *billing.Subscription) (*billing.Subscription, error) {
				assert.Equal(t, ddbpayments.Scope(), scope)
				written = subscription

				return example, nil
			},
		})

		res, err := service.CreateSubscription(r.ctx, &paymentssvc.CreateSubscriptionRequest{
			Input: converters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(example),
		})
		require.NoError(t, err)

		require.NotNil(t, written)
		assert.Equal(t, example.BelongsToAccount, written.BelongsToAccount)
		assert.Equal(t, example.ProductID, written.ProductID)
		assert.Equal(t, capitalism.SubscriptionStatusActive, written.Status)
		assert.True(t, example.CurrentPeriodEnd.Equal(written.CurrentPeriodEnd))

		assert.Equal(t, example.ID, res.GetCreated().GetId())
	})

	T.Run("with a status outside the vocabulary", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		service := buildTestService(t, &billingmock.StoreMock{
			CreateSubscriptionFunc: func(context.Context, tenancy.Scope, *billing.Subscription) (*billing.Subscription, error) {
				return nil, billing.ErrInvalidStatus
			},
		})

		input := converters.ConvertSubscriptionToGRPCSubscriptionCreationRequestInput(fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID()))
		input.Status = "cancelled"

		_, err := service.CreateSubscription(r.ctx, &paymentssvc.CreateSubscriptionRequest{Input: input})
		requireCode(t, err, codes.InvalidArgument)
	})
}

func TestService_GetSubscriptionsForAccount(T *testing.T) {
	T.Parallel()

	// The account read is the session's active one, and never the one the request
	// names: honoring the request's would let any member read another account's
	// billing by asking.
	T.Run("reads the session's account rather than the request's", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		page := fakes.BuildFakeSubscriptionList(r.account, fake.BuildFakeID())

		service := buildTestService(t, &billingmock.StoreMock{
			ListSubscriptionsForAccountFunc: func(_ context.Context, scope tenancy.Scope, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Subscription], error) {
				assert.Equal(t, ddbpayments.Scope(), scope)
				assert.Equal(t, r.account, accountID)

				return page, nil
			},
		})

		res, err := service.GetSubscriptionsForAccount(r.ctx, &paymentssvc.GetSubscriptionsForAccountRequest{AccountId: fake.BuildFakeID()})
		require.NoError(t, err)
		assert.Len(t, res.GetResults(), len(page.Data))
	})

	T.Run("without a session", func(t *testing.T) {
		t.Parallel()

		service := buildTestService(t, nil)

		_, err := service.GetSubscriptionsForAccount(t.Context(), &paymentssvc.GetSubscriptionsForAccountRequest{})
		requireCode(t, err, codes.Unauthenticated)
	})
}

func TestService_UpdateSubscription(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		example := fakes.BuildFakeSubscription(fake.BuildFakeID(), fake.BuildFakeID())

		var written *billing.Subscription

		service := buildTestService(t, &billingmock.StoreMock{
			GetSubscriptionFunc: func(context.Context, tenancy.Scope, string) (*billing.Subscription, error) {
				copied := *example

				return &copied, nil
			},
			UpdateSubscriptionFunc: func(_ context.Context, _ tenancy.Scope, subscription *billing.Subscription) error {
				written = subscription

				return nil
			},
		})

		_, err := service.UpdateSubscription(r.ctx, &paymentssvc.UpdateSubscriptionRequest{
			SubscriptionId: example.ID,
			Input:          &paymentssvc.SubscriptionUpdateRequestInput{Status: pointer.To(string(capitalism.SubscriptionStatusCanceled))},
		})
		require.NoError(t, err)

		require.NotNil(t, written)
		assert.Equal(t, capitalism.SubscriptionStatusCanceled, written.Status)
		assert.Equal(t, example.ProductID, written.ProductID)
		assert.Equal(t, example.BelongsToAccount, written.BelongsToAccount)
	})
}

func TestService_GetPaymentHistoryForAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		transaction := fakes.BuildFakeTransaction(r.account)

		service := buildTestService(t, &billingmock.StoreMock{
			ListTransactionsForAccountFunc: func(_ context.Context, _ tenancy.Scope, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Transaction], error) {
				assert.Equal(t, r.account, accountID)

				return &filtering.QueryFilteredResult[billing.Transaction]{Data: []*billing.Transaction{transaction}}, nil
			},
		})

		res, err := service.GetPaymentHistoryForAccount(r.ctx, &paymentssvc.GetPaymentHistoryForAccountRequest{})
		require.NoError(t, err)
		require.Len(t, res.GetResults(), 1)
		assert.Equal(t, transaction.ID, res.GetResults()[0].GetId())
		assert.Equal(t, transaction.Status.String(), res.GetResults()[0].GetStatus())
		assert.Equal(t, transaction.AmountCents, res.GetResults()[0].GetAmountCents())
	})
}

func TestService_GetPurchasesForAccount(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		r := userContextForTest(t)
		purchase := fakes.BuildFakePurchase(r.account, fake.BuildFakeID())

		service := buildTestService(t, &billingmock.StoreMock{
			ListPurchasesForAccountFunc: func(_ context.Context, _ tenancy.Scope, accountID string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[billing.Purchase], error) {
				assert.Equal(t, r.account, accountID)

				return &filtering.QueryFilteredResult[billing.Purchase]{Data: []*billing.Purchase{purchase}}, nil
			},
		})

		res, err := service.GetPurchasesForAccount(r.ctx, &paymentssvc.GetPurchasesForAccountRequest{})
		require.NoError(t, err)
		require.Len(t, res.GetResults(), 1)
		assert.Equal(t, purchase.ID, res.GetResults()[0].GetId())
	})
}
