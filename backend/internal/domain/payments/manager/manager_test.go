package manager

import (
	"context"
	"testing"

	identitymock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/identity/manager/mock"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/fakes"
	paymentsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/payments/mock"

	"github.com/primandproper/platform-go/v11/fake"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildPaymentsManagerForTest(t *testing.T) *paymentsManager {
	t.Helper()

	ctx := t.Context()
	m, err := NewPaymentsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&paymentsmock.RepositoryMock{},
		&identitymock.IdentityDataManagerMock{},
	)
	require.NoError(t, err)

	return m.(*paymentsManager)
}

// attachRepositoryToPaymentsManager wires a configured repository mock
// into the manager under test.
func attachRepositoryToPaymentsManager(manager *paymentsManager, repo *paymentsmock.RepositoryMock) {
	manager.repo = repo
}

func TestPaymentsManager_CreateProduct(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		input := fakes.BuildFakeProductCreationRequestInput()
		expected := fakes.BuildFakeProduct()

		repo := &paymentsmock.RepositoryMock{
			CreateProductFunc: func(_ context.Context, _ *payments.ProductDatabaseCreationInput) (*payments.Product, error) {
				return expected, nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		actual, err := pm.CreateProduct(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateProductCalls(), 1)
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

		repo := &paymentsmock.RepositoryMock{
			GetProductFunc: func(_ context.Context, id string) (*payments.Product, error) {
				assert.Equal(t, productID, id)

				return product, nil
			},
			UpdateProductFunc: func(_ context.Context, p *payments.Product) error {
				assert.Equal(t, productID, p.ID)
				assert.Equal(t, name, p.Name)

				return nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		err := pm.UpdateProduct(ctx, productID, input)
		require.NoError(t, err)

		assert.Len(t, repo.GetProductCalls(), 1)
		assert.Len(t, repo.UpdateProductCalls(), 1)
	})
}

func TestPaymentsManager_ArchiveProduct(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		productID := fake.BuildFakeID()

		repo := &paymentsmock.RepositoryMock{
			ArchiveProductFunc: func(_ context.Context, id string) error {
				assert.Equal(t, productID, id)

				return nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		err := pm.ArchiveProduct(ctx, productID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveProductCalls(), 1)
	})
}

func TestPaymentsManager_CreateSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		accountID := fake.BuildFakeID()
		productID := fake.BuildFakeID()
		input := fakes.BuildFakeSubscriptionCreationRequestInput(accountID, productID)
		expected := fakes.BuildFakeSubscription(accountID, productID)

		repo := &paymentsmock.RepositoryMock{
			CreateSubscriptionFunc: func(_ context.Context, _ *payments.SubscriptionDatabaseCreationInput) (*payments.Subscription, error) {
				return expected, nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		actual, err := pm.CreateSubscription(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateSubscriptionCalls(), 1)
	})
}

func TestPaymentsManager_UpdateSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		accountID := fake.BuildFakeID()
		productID := fake.BuildFakeID()
		sub := fakes.BuildFakeSubscription(accountID, productID)
		subID := sub.ID
		status := payments.SubscriptionStatusCancelled
		input := &payments.SubscriptionUpdateRequestInput{Status: &status}

		repo := &paymentsmock.RepositoryMock{
			GetSubscriptionFunc: func(_ context.Context, id string) (*payments.Subscription, error) {
				assert.Equal(t, subID, id)

				return sub, nil
			},
			UpdateSubscriptionFunc: func(_ context.Context, s *payments.Subscription) error {
				assert.Equal(t, subID, s.ID)
				assert.Equal(t, status, s.Status)

				return nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		err := pm.UpdateSubscription(ctx, subID, input)
		require.NoError(t, err)

		assert.Len(t, repo.GetSubscriptionCalls(), 1)
		assert.Len(t, repo.UpdateSubscriptionCalls(), 1)
	})
}

func TestPaymentsManager_ArchiveSubscription(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		pm := buildPaymentsManagerForTest(t)

		subID := fake.BuildFakeID()

		repo := &paymentsmock.RepositoryMock{
			ArchiveSubscriptionFunc: func(_ context.Context, id string) error {
				assert.Equal(t, subID, id)

				return nil
			},
		}
		attachRepositoryToPaymentsManager(pm, repo)

		err := pm.ArchiveSubscription(ctx, subID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveSubscriptionCalls(), 1)
	})
}
