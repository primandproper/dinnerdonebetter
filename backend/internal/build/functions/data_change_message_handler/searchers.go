package datachangemessagehandler

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/searchindex"
	identityindexing "github.com/primandproper/dinnerdonebetter/backend/internal/services/identity/indexing"

	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	textsearchcfg "github.com/primandproper/primitives-go/v2/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterSearchers registers all text searcher providers with the injector.
func RegisterSearchers(i do.Injector) {
	do.Provide(i, func(i do.Injector) (identityindexing.UserTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideUserTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.RecipeTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideRecipeTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.MealTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideMealTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidIngredientTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidIngredientTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidInstrumentTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidInstrumentTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidMeasurementUnitTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidMeasurementUnitTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidPreparationTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidPreparationTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidIngredientStateTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidIngredientStateTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (searchindex.ValidVesselTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.Provider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidVesselTextSearcher(ctx, logger, tp, mp, cfg)
	})
}

func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (identityindexing.UserTextSearcher, error) {
	return textsearchcfg.NewIndex[identityindexing.UserSearchSubset](
		ctx,
		cfg,
		identityindexing.IndexTypeUsers,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideRecipeTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.RecipeTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.RecipeSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeRecipes,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideMealTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.MealTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.MealSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeMeals,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidIngredientTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidIngredientTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidIngredientSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidIngredients,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidInstrumentTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidInstrumentTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidInstrumentSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidInstruments,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidMeasurementUnitTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidMeasurementUnitTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidMeasurementUnitSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidMeasurementUnits,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidPreparationTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidPreparationTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidPreparationSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidPreparations,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidIngredientStateTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidIngredientStateTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidIngredientStateSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidIngredientStates,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidVesselTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (searchindex.ValidVesselTextSearcher, error) {
	return textsearchcfg.NewIndex[searchindex.ValidVesselSearchSubset](
		ctx,
		cfg,
		searchindex.IndexTypeValidVessels,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}
