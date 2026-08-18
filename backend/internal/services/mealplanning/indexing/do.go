package indexing

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/indexstamp"

	"github.com/primandproper/platform-go/v11/batching"
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

// registerPair registers the Syncer and Reindexer for one entity over one index, plus the
// Stamper the Syncer writes through.
//
// The two are registered together because they are the two halves of keeping one index right,
// and registering one without the other is the mistake worth making impossible: a Syncer with
// no Reindexer has no way back from an index that was already wrong, and a Reindexer with no
// Syncer is the sampler this replaced, just on a longer timer.
//
// Which of them a given process actually resolves is still that process's business — the API
// server resolves neither, the consumer resolves Syncers, the scheduler resolves Reindexers —
// because do is lazy and nothing is constructed until something asks for it. That laziness is
// also why the Stamper registered here costs the scheduler nothing: only the Syncer asks for
// one.
//
// mark is spelled as a method expression at the call sites — mealplanning.Repository.MarkXAsIndexed
// — so the entity's stamping method is named next to its source and its index, and a pair
// wired to the wrong one does not compile.
func registerPair[E, T any](
	i do.Injector,
	name string,
	source func(mealplanning.Repository) (*syncsource.Source[E, T], error),
	mark func(mealplanning.Repository, context.Context, string) error,
	index func(do.Injector) textsearch.IndexManager,
) {
	// Named for its index, so that the consumer can resolve the same Stamper the Syncer got in
	// order to close it. do provides one instance per name, which is what makes those the same
	// object rather than two buffers over one index.
	do.ProvideNamed(i, name, func(i do.Injector) (*indexstamp.Stamper, error) {
		repo := do.MustInvoke[mealplanning.Repository](i)

		return indexstamp.New(index(i), func(ctx context.Context, id string) error {
			return mark(repo, ctx, id)
		}, stampOptions(i)...)
	})

	do.Provide(i, func(i do.Injector) (*searchsync.Syncer[T], error) {
		src, err := source(do.MustInvoke[mealplanning.Repository](i))
		if err != nil {
			return nil, err
		}

		return syncsource.NewSyncer(src, do.MustInvokeNamed[*indexstamp.Stamper](i, name), o11yOptions(i)...)
	})

	// The Reindexer writes to the searcher directly rather than through the Stamper. A reindex
	// writes every document there is, so stamping it would turn last_indexed_at into a record
	// of when the last rebuild ran — the same value on every row — instead of how current each
	// document is.
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

// stampOptions is the same three pillars as batching options, for the Stamper's buffer.
func stampOptions(i do.Injector) []batching.Option {
	return []batching.Option{
		batching.WithLogger(logging.NewNamedLogger(do.MustInvoke[logging.Logger](i), o11yName)),
		batching.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		batching.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	}
}

// RegisterIndexSyncers registers the Syncer and Reindexer for all eight meal planning indexes.
func RegisterIndexSyncers(i do.Injector) {
	registerPair(i, IndexTypeMeals, NewMealSource, mealplanning.Repository.MarkMealAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[MealTextSearcher](i)
	})
	registerPair(i, IndexTypeRecipes, NewRecipeSource, mealplanning.Repository.MarkRecipeAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[RecipeTextSearcher](i)
	})
	registerPair(i, IndexTypeValidIngredients, NewValidIngredientSource, mealplanning.Repository.MarkValidIngredientAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidIngredientTextSearcher](i)
	})
	registerPair(i, IndexTypeValidInstruments, NewValidInstrumentSource, mealplanning.Repository.MarkValidInstrumentAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidInstrumentTextSearcher](i)
	})
	registerPair(i, IndexTypeValidMeasurementUnits, NewValidMeasurementUnitSource, mealplanning.Repository.MarkValidMeasurementUnitAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidMeasurementUnitTextSearcher](i)
	})
	registerPair(i, IndexTypeValidPreparations, NewValidPreparationSource, mealplanning.Repository.MarkValidPreparationAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidPreparationTextSearcher](i)
	})
	registerPair(i, IndexTypeValidIngredientStates, NewValidIngredientStateSource, mealplanning.Repository.MarkValidIngredientStateAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidIngredientStateTextSearcher](i)
	})
	registerPair(i, IndexTypeValidVessels, NewValidVesselSource, mealplanning.Repository.MarkValidVesselAsIndexed, func(i do.Injector) textsearch.IndexManager {
		return do.MustInvoke[ValidVesselTextSearcher](i)
	})
}
