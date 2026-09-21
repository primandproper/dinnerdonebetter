package grpc

import (
	"testing"

	mockmanagers "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers/mock"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"
	"github.com/primandproper/dinnerdonebetter/backend/internal/testutils"

	commentsmock "github.com/primandproper/platform-go/v14/comments/mock"
	registrymock "github.com/primandproper/platform-go/v14/mediaregistry/mock"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	mockuploads "github.com/primandproper/primitives-go/v2/uploads/mock"

	"github.com/stretchr/testify/assert"
)

func TestNewService(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		logger := loggingnoop.NewLogger()
		tracerProvider := tracingnoop.NewTracerProvider()
		mealPlanningManager := &mockmanagers.MealPlanningManagerMock{}
		mealPlanFinalizationStarter := &mealplanfinalization.Starter{}
		commentStore := &commentsmock.StoreMock{}
		uploadsRegistry := &registrymock.StoreMock{}
		uploadManager := &mockuploads.UploadManagerMock{}

		service := NewService(
			logger,
			tracerProvider,
			testutils.MockDatabaseClient(),
			mealPlanningManager,
			mealPlanFinalizationStarter,
			commentStore,
			uploadsRegistry,
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
		assert.Equal(t, commentStore, impl.comments)
	})
}
