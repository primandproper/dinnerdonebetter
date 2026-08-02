package datachangemessagehandler

import (
	"context"

	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v9/search/text/config"

	"github.com/samber/do/v2"
)

// RegisterSearchers registers all text searcher providers with the injector.
func RegisterSearchers(i do.Injector) {
	do.Provide(i, func(i do.Injector) (identityindexing.UserTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideUserTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.RecipeTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideRecipeTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.MealTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideMealTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidIngredientTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidIngredientTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidInstrumentTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidInstrumentTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidMeasurementUnitTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidMeasurementUnitTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidPreparationTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidPreparationTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidIngredientStateTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidIngredientStateTextSearcher(ctx, logger, tp, mp, cfg)
	})
	do.Provide(i, func(i do.Injector) (eatingindexing.ValidVesselTextSearcher, error) {
		ctx := do.MustInvoke[context.Context](i)
		logger := do.MustInvoke[logging.Logger](i)
		tp := do.MustInvoke[tracing.TracerProvider](i)
		mp := do.MustInvoke[metrics.Provider](i)
		cfg := do.MustInvoke[*textsearchcfg.Config](i)
		return ProvideValidVesselTextSearcher(ctx, logger, tp, mp, cfg)
	})
}

func ProvideUserTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
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
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.RecipeTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.RecipeSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeRecipes,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideMealTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.MealTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.MealSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeMeals,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidIngredientTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidIngredientTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidIngredientSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidIngredients,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidInstrumentTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidInstrumentTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidInstrumentSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidInstruments,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidMeasurementUnitTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidMeasurementUnitTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidMeasurementUnitSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidMeasurementUnits,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidPreparationTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidPreparationTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidPreparationSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidPreparations,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidIngredientStateTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidIngredientStateTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidIngredientStateSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidIngredientStates,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}

func ProvideValidVesselTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidVesselTextSearcher, error) {
	return textsearchcfg.NewIndex[eatingindexing.ValidVesselSearchSubset](
		ctx,
		cfg,
		eatingindexing.IndexTypeValidVessels,
		textsearchcfg.WithLogger(logger),
		textsearchcfg.WithTracerProvider(tracerProvider),
		textsearchcfg.WithMetricsProvider(metricsProvider),
	)
}
