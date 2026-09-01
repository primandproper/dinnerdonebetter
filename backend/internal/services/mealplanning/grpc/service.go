package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/errors"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/errors"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	comments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads"
	"github.com/primandproper/platform-go/v13/uploads/registry"
)

var _ mealplanningsvc.MealPlanningServiceServer = (*serviceImpl)(nil)

const (
	o11yName = "mealplanning_service"
)

type (
	serviceImpl struct {
		mealplanningsvc.UnimplementedMealPlanningServiceServer
		tracer              tracing.Tracer
		logger              logging.Logger
		mealPlanningManager managers.MealPlanningManager
		// One starter, where there used to be three workers. All three of the admin RPCs that
		// ran those on demand reach this: finalizing, creating tasks, and building the grocery
		// list are one saga now, and the only part of it left to run on demand is entering
		// plans into it.
		mealPlanFinalizationStarter *mealplanfinalization.Starter
		comments                    comments.Store
		registry                    registry.Store
		uploadManager               uploads.UploadManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	mealPlanningManager managers.MealPlanningManager,
	mealPlanFinalizationStarter *mealplanfinalization.Starter,
	commentStore comments.Store,
	registryStore registry.Store,
	uploadManager uploads.UploadManager,
) mealplanningsvc.MealPlanningServiceServer {
	return &serviceImpl{
		logger:                      logging.NewNamedLogger(logger, o11yName),
		tracer:                      tracing.NewNamedTracer(tracerProvider, o11yName),
		mealPlanningManager:         mealPlanningManager,
		mealPlanFinalizationStarter: mealPlanFinalizationStarter,
		comments:                    commentStore,
		registry:                    registryStore,
		uploadManager:               uploadManager,
	}
}
