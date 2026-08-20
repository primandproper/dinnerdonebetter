package grpc

import (
	commentsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	uploadedmediamanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/manager"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	"github.com/primandproper/platform-go/v12/observability/logging"
	"github.com/primandproper/platform-go/v12/observability/tracing"
	"github.com/primandproper/platform-go/v12/uploads"

	"github.com/samber/do/v2"
)

// RegisterMealPlanningService registers the meal planning gRPC service with the injector.
func RegisterMealPlanningService(i do.Injector) {
	do.Provide[MealPlanningMethodPermissions](i, func(i do.Injector) (MealPlanningMethodPermissions, error) {
		return ProvideMethodPermissions(), nil
	})

	do.Provide[mealplanningsvc.MealPlanningServiceServer](i, func(i do.Injector) (mealplanningsvc.MealPlanningServiceServer, error) {
		return NewService(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[managers.MealPlanningManager](i),
			do.MustInvoke[*mealplanfinalization.Starter](i),
			do.MustInvoke[commentsmanager.CommentsDataManager](i),
			do.MustInvoke[uploadedmediamanager.UploadedMediaManager](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
