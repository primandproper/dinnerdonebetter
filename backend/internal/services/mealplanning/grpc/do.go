package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/managers"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	mealplanfinalization "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/workers/meal_plan_finalization"

	comments "github.com/primandproper/platform-go/v14/comments"
	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"

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
			do.MustInvoke[database.Client](i),
			do.MustInvoke[managers.MealPlanningManager](i),
			do.MustInvoke[*mealplanfinalization.Starter](i),
			do.MustInvoke[comments.Store](i),
			do.MustInvoke[mediaregistry.Store](i),
			do.MustInvoke[uploads.UploadManager](i),
		), nil
	})
}
