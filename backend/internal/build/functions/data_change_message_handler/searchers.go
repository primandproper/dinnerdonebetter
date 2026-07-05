package datachangemessagehandler

import (
	"context"

	identityindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/identity/indexing"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"

	"github.com/primandproper/platform-go/v3/observability/logging"
	"github.com/primandproper/platform-go/v3/observability/metrics"
	"github.com/primandproper/platform-go/v3/observability/tracing"
	textsearchcfg "github.com/primandproper/platform-go/v3/search/text/config"

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
	return textsearchcfg.ProvideIndex[identityindexing.UserSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		identityindexing.IndexTypeUsers,
	)
}

func ProvideRecipeTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.RecipeTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.RecipeSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeRecipes,
	)
}

func ProvideMealTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.MealTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.MealSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeMeals,
	)
}

func ProvideValidIngredientTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidIngredientTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidIngredientSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidIngredients,
	)
}

func ProvideValidInstrumentTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidInstrumentTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidInstrumentSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidInstruments,
	)
}

func ProvideValidMeasurementUnitTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidMeasurementUnitTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidMeasurementUnitSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidMeasurementUnits,
	)
}

func ProvideValidPreparationTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidPreparationTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidPreparationSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidPreparations,
	)
}

func ProvideValidIngredientStateTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidIngredientStateTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidIngredientStateSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidIngredientStates,
	)
}

func ProvideValidVesselTextSearcher(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	cfg *textsearchcfg.Config,
) (eatingindexing.ValidVesselTextSearcher, error) {
	return textsearchcfg.ProvideIndex[eatingindexing.ValidVesselSearchSubset](
		ctx,
		logger,
		tracerProvider, metricsProvider,
		cfg,
		eatingindexing.IndexTypeValidVessels,
	)
}
