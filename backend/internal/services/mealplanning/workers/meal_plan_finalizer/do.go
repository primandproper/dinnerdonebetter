package mealplanfinalizer

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanFinalizer registers the meal plan finalizer with the injector.
func RegisterMealPlanFinalizer(i do.Injector) {
	do.Provide[*Worker](i, func(i do.Injector) (*Worker, error) {
		return NewMealPlanFinalizer(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[mealplanning.Repository](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}
