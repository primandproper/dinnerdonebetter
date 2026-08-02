package grpc

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	commentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager"
	mockmanagers "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	uploadedmediamock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	mealplanningsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanfinalizer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalizer"
	mealplangrocerylistinitializer "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_grocery_list_initializer"
	mealplantaskcreator "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_task_creator"

	"github.com/primandproper/platform-go/v9/filtering"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v9/uploads/mock"

	"github.com/stretchr/testify/assert"
)

// noopCommentsManager is a stub implementation for tests that only need service construction.
type noopCommentsManager struct{}

func (n *noopCommentsManager) CreateComment(_ context.Context, _ *comments.CommentCreationRequestInput) (*comments.Comment, error) {
	return nil, nil
}
func (n *noopCommentsManager) GetComment(_ context.Context, _ string) (*comments.Comment, error) {
	return nil, nil
}
func (n *noopCommentsManager) GetCommentsForReference(_ context.Context, _, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
	return nil, nil
}
func (n *noopCommentsManager) UpdateComment(_ context.Context, _, _ string, _ *comments.CommentUpdateRequestInput) error {
	return nil
}
func (n *noopCommentsManager) ArchiveComment(_ context.Context, _ string) error {
	return nil
}
func (n *noopCommentsManager) ArchiveCommentsForReference(_ context.Context, _, _ string) error {
	return nil
}

var _ commentsmanager.CommentsDataManager = (*noopCommentsManager)(nil)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		mealPlanningManager := &mockmanagers.MealPlanningManagerMock{}
		mealPlanFinalizerWorker := &mealplanfinalizer.Worker{}
		mealPlanGroceryListInitializerWorker := &mealplangrocerylistinitializer.Worker{}
		mealPlanTaskCreatorWorker := &mealplantaskcreator.Worker{}
		commentsManager := &noopCommentsManager{}
		uploadedMediaManager := &uploadedmediamock.RepositoryMock{}
		uploadManager := &mockuploads.UploadManagerMock{}

		service := NewService(
			logger,
			tracerProvider,
			mealPlanningManager,
			mealPlanFinalizerWorker,
			mealPlanGroceryListInitializerWorker,
			mealPlanTaskCreatorWorker,
			commentsManager,
			uploadedMediaManager,
			uploadManager,
		)

		assert.NotNil(t, service)
		assert.Implements(t, (*mealplanningsvc.MealPlanningServiceServer)(nil), service)

		// Type assertion to ensure we get the correct implementation
		impl, ok := service.(*serviceImpl)
		assert.True(t, ok)
		assert.NotNil(t, impl.logger)
		assert.NotNil(t, impl.tracer)
		assert.Equal(t, mealPlanningManager, impl.mealPlanningManager)
		assert.Equal(t, mealPlanFinalizerWorker, impl.mealPlanFinalizerWorker)
		assert.Equal(t, mealPlanGroceryListInitializerWorker, impl.mealPlanGroceryListInitializerWorker)
		assert.Equal(t, mealPlanTaskCreatorWorker, impl.mealPlanTaskCreatorWorker)
		assert.Equal(t, commentsManager, impl.commentsManager)
		assert.NotNil(t, impl.sessionContextDataFetcher)
	})
}
