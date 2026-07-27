package mealplangrocerylistinitializer

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/grocerylistpreparation"

	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanGroceryListInitializer registers the meal plan grocery list initializer with the injector.
func RegisterMealPlanGroceryListInitializer(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewMealPlanGroceryListInitializer(
			do.MustInvoke[context.Context](i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[metrics.Provider](i),
			do.MustInvoke[messagequeue.PublisherProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[grocerylistpreparation.GroceryListCreator](i),
			do.MustInvoke[*msgconfig.QueuesConfig](i),
		)
	})
}
