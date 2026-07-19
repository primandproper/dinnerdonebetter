package manager

import (
	"context"
	"testing"

	identitymock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/keys"
	paymentsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/payments/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/payments/adapters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v5/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildPaymentsManagerForTest(t *testing.T) *paymentsManager {
	t.Helper()

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	stub := adapters.NewStubPaymentProcessor(nil)
	registry := payments.NewMapProcessorRegistry(map[string]payments.PaymentProcessor{
		"stripe":     stub,
		"revenuecat": stub,
	})
	m, err := NewPaymentsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&paymentsmock.Repository{},
		registry,
		&identitymock.IdentityDataManager{},
		queueCfg,
		mpp,
	)
	require.NoError(t, err)

	return m.(*paymentsManager)
}

func setupExpectationsForPaymentsManager(
	manager *paymentsManager,
	repoSetupFunc func(repo *paymentsmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &paymentsmock.Repository{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.repo = repo

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{repo}
}

func TestPaymentsManager_CreateProduct(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		input := fakes.BuildFakeProductCreationRequestInput()
		expected := fakes.BuildFakeProduct()

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateProduct), testutils.ContextMatcher, testutils.MatchType[*payments.ProductDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				payments.ProductCreatedServiceEventType: {paymentskeys.ProductIDKey},
			},
		)

		actual, err := pm.CreateProduct(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestPaymentsManager_UpdateProduct(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		product := fakes.BuildFakeProduct()
		productID := product.ID
		name := "Updated Name"
		input := &payments.ProductUpdateRequestInput{Name: &name}

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.GetProduct), testutils.ContextMatcher, productID).Return(product, nil)
				repo.On(reflection.GetMethodName(repo.UpdateProduct), testutils.ContextMatcher, mock.MatchedBy(func(p *payments.Product) bool {
					return p.ID == productID && p.Name == name
				})).Return(nil)
			},
			map[string][]string{
				payments.ProductUpdatedServiceEventType: {paymentskeys.ProductIDKey},
			},
		)

		err := pm.UpdateProduct(ctx, productID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestPaymentsManager_ArchiveProduct(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		productID := fakes.BuildFakeID()

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveProduct), testutils.ContextMatcher, productID).Return(nil)
			},
			map[string][]string{
				payments.ProductArchivedServiceEventType: {paymentskeys.ProductIDKey},
			},
		)

		err := pm.ArchiveProduct(ctx, productID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestPaymentsManager_CreateSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		accountID := fakes.BuildFakeID()
		productID := fakes.BuildFakeID()
		input := fakes.BuildFakeSubscriptionCreationRequestInput(accountID, productID)
		expected := fakes.BuildFakeSubscription(accountID, productID)

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateSubscription), testutils.ContextMatcher, testutils.MatchType[*payments.SubscriptionDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				payments.SubscriptionCreatedServiceEventType: {paymentskeys.SubscriptionIDKey},
			},
		)

		actual, err := pm.CreateSubscription(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestPaymentsManager_UpdateSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		accountID := fakes.BuildFakeID()
		productID := fakes.BuildFakeID()
		sub := fakes.BuildFakeSubscription(accountID, productID)
		subID := sub.ID
		status := payments.SubscriptionStatusCancelled
		input := &payments.SubscriptionUpdateRequestInput{Status: &status}

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.GetSubscription), testutils.ContextMatcher, subID).Return(sub, nil)
				repo.On(reflection.GetMethodName(repo.UpdateSubscription), testutils.ContextMatcher, mock.MatchedBy(func(s *payments.Subscription) bool {
					return s.ID == subID && s.Status == status
				})).Return(nil)
			},
			map[string][]string{
				payments.SubscriptionUpdatedServiceEventType: {paymentskeys.SubscriptionIDKey},
			},
		)

		err := pm.UpdateSubscription(ctx, subID, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestPaymentsManager_ArchiveSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		subID := fakes.BuildFakeID()

		expectations := setupExpectationsForPaymentsManager(
			pm,
			func(repo *paymentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveSubscription), testutils.ContextMatcher, subID).Return(nil)
			},
			map[string][]string{
				payments.SubscriptionArchivedServiceEventType: {paymentskeys.SubscriptionIDKey},
			},
		)

		err := pm.ArchiveSubscription(ctx, subID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
