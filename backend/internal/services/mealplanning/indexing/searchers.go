package indexing

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex"

	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	textsearch "github.com/primandproper/primitives-go/v2/search/text"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterSearchers registers the text index client behind each of this domain's eight indexes.
//
// They are what RegisterIndexes resolves, one per index, so a process that registers the indexes
// registers these beside them. The clients are built from the one *textsearchcfg.Config the
// process provides — which backend, which credentials — and differ only in the index they name.
func RegisterSearchers(i do.Injector) {
	do.Provide(i, func(i do.Injector) (searchindex.RecipeTextSearcher, error) {
		index, err := newIndex[searchindex.RecipeSearchSubset](i, searchindex.IndexTypeRecipes)

		return searchindex.RecipeTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.MealTextSearcher, error) {
		index, err := newIndex[searchindex.MealSearchSubset](i, searchindex.IndexTypeMeals)

		return searchindex.MealTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidIngredientTextSearcher, error) {
		index, err := newIndex[searchindex.ValidIngredientSearchSubset](i, searchindex.IndexTypeValidIngredients)

		return searchindex.ValidIngredientTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidInstrumentTextSearcher, error) {
		index, err := newIndex[searchindex.ValidInstrumentSearchSubset](i, searchindex.IndexTypeValidInstruments)

		return searchindex.ValidInstrumentTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidMeasurementUnitTextSearcher, error) {
		index, err := newIndex[searchindex.ValidMeasurementUnitSearchSubset](i, searchindex.IndexTypeValidMeasurementUnits)

		return searchindex.ValidMeasurementUnitTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidPreparationTextSearcher, error) {
		index, err := newIndex[searchindex.ValidPreparationSearchSubset](i, searchindex.IndexTypeValidPreparations)

		return searchindex.ValidPreparationTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidIngredientStateTextSearcher, error) {
		index, err := newIndex[searchindex.ValidIngredientStateSearchSubset](i, searchindex.IndexTypeValidIngredientStates)

		return searchindex.ValidIngredientStateTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidVesselTextSearcher, error) {
		index, err := newIndex[searchindex.ValidVesselSearchSubset](i, searchindex.IndexTypeValidVessels)

		return searchindex.ValidVesselTextSearcher(index), err
	})
}

// newIndex builds the text index client for one index, over the process's search config and
// pillars.
func newIndex[T any](i do.Injector, indexName string) (textsearch.Index[T], error) {
	return textsearchcfg.NewIndex[T](
		do.MustInvoke[context.Context](i),
		do.MustInvoke[*textsearchcfg.Config](i),
		indexName,
		textsearchcfg.WithLogger(do.MustInvoke[logging.Logger](i)),
		textsearchcfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		textsearchcfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
	)
}
