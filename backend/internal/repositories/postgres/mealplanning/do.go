package mealplanning

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	domainmealplanning "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v6/database"
	"github.com/primandproper/platform-go/v6/observability/logging"
	"github.com/primandproper/platform-go/v6/observability/tracing"

	"github.com/samber/do/v2"
)

// RegisterMealPlanningRepository registers the meal planning repository with the injector.
func RegisterMealPlanningRepository(i do.Injector) {
	do.Provide[domainmealplanning.Repository](i, func(i do.Injector) (domainmealplanning.Repository, error) {
		return ProvideMealPlanningRepository(
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.TracerProvider](i),
			do.MustInvoke[audit.Repository](i),
			do.MustInvoke[identity.Repository](i),
			do.MustInvoke[database.Client](i),
		), nil
	})

	do.Provide[domainmealplanning.ValidEnumerationDataManager](i, func(i do.Injector) (domainmealplanning.ValidEnumerationDataManager, error) {
		return ProvideValidEnumerationDataManager(do.MustInvoke[domainmealplanning.Repository](i)), nil
	})
}

func ProvideValidEnumerationDataManager(x domainmealplanning.Repository) domainmealplanning.ValidEnumerationDataManager {
	return x
}
