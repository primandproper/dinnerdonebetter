package indexing

import (
	"context"

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
	do.Provide(i, func(i do.Injector) (RecipeTextSearcher, error) {
		index, err := newIndex[RecipeSearchSubset](i, IndexTypeRecipes)

		return RecipeTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (MealTextSearcher, error) {
		index, err := newIndex[MealSearchSubset](i, IndexTypeMeals)

		return MealTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidIngredientTextSearcher, error) {
		index, err := newIndex[ValidIngredientSearchSubset](i, IndexTypeValidIngredients)

		return ValidIngredientTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidInstrumentTextSearcher, error) {
		index, err := newIndex[ValidInstrumentSearchSubset](i, IndexTypeValidInstruments)

		return ValidInstrumentTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidMeasurementUnitTextSearcher, error) {
		index, err := newIndex[ValidMeasurementUnitSearchSubset](i, IndexTypeValidMeasurementUnits)

		return ValidMeasurementUnitTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidPreparationTextSearcher, error) {
		index, err := newIndex[ValidPreparationSearchSubset](i, IndexTypeValidPreparations)

		return ValidPreparationTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidIngredientStateTextSearcher, error) {
		index, err := newIndex[ValidIngredientStateSearchSubset](i, IndexTypeValidIngredientStates)

		return ValidIngredientStateTextSearcher(index), err
	})
	do.Provide(i, func(i do.Injector) (ValidVesselTextSearcher, error) {
		index, err := newIndex[ValidVesselSearchSubset](i, IndexTypeValidVessels)

		return ValidVesselTextSearcher(index), err
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
