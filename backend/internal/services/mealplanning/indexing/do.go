package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/search/syncsource"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v10/search/sync"
	textsearch "github.com/primandproper/platform-go/v10/search/text"

	"github.com/samber/do/v2"
)

// registerPair registers the Syncer and Reindexer for one entity over one index.
//
// The two are registered together because they are the two halves of keeping one index right,
// and registering one without the other is the mistake worth making impossible: a Syncer with
// no Reindexer has no way back from an index that was already wrong, and a Reindexer with no
// Syncer is the sampler this replaced, just on a longer timer.
//
// Which of them a given process actually resolves is still that process's business — the API
// server resolves neither, the consumer resolves Syncers, the scheduler resolves Reindexers —
// because do is lazy and nothing is constructed until something asks for it.
func registerPair[E, T any](
	i do.Injector,
	source func(mealplanning.Repository) *syncsource.Source[E, T],
	index func(do.Injector) textsearch.IndexManager,
) {
	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[T], error) {
		return syncsource.NewSyncer(
			source(do.MustInvoke[mealplanning.Repository](i)),
			index(i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})

	do.Provide(i, func(i do.Injector) (*searchsync.Reindexer[T], error) {
		return syncsource.NewReindexer(
			source(do.MustInvoke[mealplanning.Repository](i)),
			index(i),
			do.MustInvoke[logging.Logger](i),
			do.MustInvoke[tracing.Provider](i),
			do.MustInvoke[metrics.Provider](i),
		)
	})
}

// RegisterIndexSyncers registers the Syncer and Reindexer for all eight meal planning indexes.
func RegisterIndexSyncers(i do.Injector) {
	registerPair(i, NewMealSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[MealTextSearcher](i)
	})
	registerPair(i, NewRecipeSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[RecipeTextSearcher](i)
	})
	registerPair(i, NewValidIngredientSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidIngredientTextSearcher](i)
	})
	registerPair(i, NewValidInstrumentSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidInstrumentTextSearcher](i)
	})
	registerPair(i, NewValidMeasurementUnitSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidMeasurementUnitTextSearcher](i)
	})
	registerPair(i, NewValidPreparationSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidPreparationTextSearcher](i)
	})
	registerPair(i, NewValidIngredientStateSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidIngredientStateTextSearcher](i)
	})
	registerPair(i, NewValidVesselSource, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidVesselTextSearcher](i)
	})
}
