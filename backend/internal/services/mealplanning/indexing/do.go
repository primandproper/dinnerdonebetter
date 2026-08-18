package indexing

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v11/observability/logging"
	"github.com/primandproper/platform-go/v11/observability/metrics"
	"github.com/primandproper/platform-go/v11/observability/tracing"
	searchsync "github.com/primandproper/platform-go/v11/search/sync"
	syncsource "github.com/primandproper/platform-go/v11/search/sync/source"
	textsearch "github.com/primandproper/platform-go/v11/search/text"

	"github.com/samber/do/v2"
)

// o11yName names the loggers, spans and metrics of the search sync sources built here. It
// keeps the name the deleted internal/search/syncsource used, so nothing downstream of a log
// query has to change.
const o11yName = "search_sync_source"

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
	source func(mealplanning.Repository) (*syncsource.Source[E, T], error),
	index func(do.Injector) textsearch.IndexManager,
) {
	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[T], error) {
		src, err := source(do.MustInvoke[mealplanning.Repository](i))
		if err != nil {
			return nil, err
		}

		return syncsource.NewSyncer(src, index(i), o11yOptions(i)...)
	})

	do.Provide(i, func(i do.Injector) (*searchsync.Reindexer[T], error) {
		src, err := source(do.MustInvoke[mealplanning.Repository](i))
		if err != nil {
			return nil, err
		}

		return syncsource.NewReindexer(src, index(i), o11yOptions(i)...)
	})
}

// o11yOptions is the three pillars as syncsource options. They are resolved individually
// rather than as an observability.Pillars because that is how the container holds them.
func o11yOptions(i do.Injector) []syncsource.Option {
	return []syncsource.Option{
		syncsource.WithLogger(logging.NewNamedLogger(do.MustInvoke[logging.Logger](i), o11yName)),
		syncsource.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		syncsource.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	}
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
