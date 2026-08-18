package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	uploadedmediamock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v11/filtering"
	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"
	mockuploads "github.com/primandproper/platform-go/v11/uploads/mock"

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
		mealPlanFinalizationStarter := &mealplanfinalization.Starter{}
		commentsManager := &noopCommentsManager{}
		uploadedMediaManager := &uploadedmediamock.RepositoryMock{}
		uploadManager := &mockuploads.UploadManagerMock{}

		service := NewService(
			logger,
			tracerProvider,
			mealPlanningManager,
			mealPlanFinalizationStarter,
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
		assert.Equal(t, mealPlanFinalizationStarter, impl.mealPlanFinalizationStarter)
		assert.Equal(t, commentsManager, impl.commentsManager)
	})
}
