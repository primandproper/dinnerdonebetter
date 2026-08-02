package mealplangrocerylistinitializer

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanGroceryListInitializer registers the meal plan grocery list initializer with the injector.
func RegisterMealPlanGroceryListInitializer(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewMealPlanGroceryListInitializer(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[grocerylistpreparation.GroceryListCreator](i),
		)
	})
}
