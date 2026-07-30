package searchdataindexscheduler

import (
	"maps"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/identity"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	coreindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v8/search/text/indexing"
)

func ProvideIndexFunctions(identityRepo identity.Repository, mealPlanningRepo mealplanning.Repository) map[string]indexing.Function {
	outputMap := map[string]indexing.Function{}
	coreMap := coreindexing.BuildCoreDataIndexingFunctions(identityRepo)
	eatingMap := eatingindexing.BuildEatingDataIndexingFunctions(mealPlanningRepo)

	maps.Copy(outputMap, coreMap)

	maps.Copy(outputMap, eatingMap)

	return outputMap
}
