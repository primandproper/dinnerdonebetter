package indexing

import (
	"context"
	"errors"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/observability"
	"github.com/primandproper/platform-go/v5/observability/logging"
	"github.com/primandproper/platform-go/v5/observability/tracing"
	textsearch "github.com/primandproper/platform-go/v5/search/text"
)

const (
	o11yName = "eating_search_indexer"
)

var (
	ErrNilIndexRequest = errors.New("nil index request")
)

type MealPlanningDataIndexer struct {
	logger                          logging.Logger
	tracer                          tracing.Tracer
	mealPlanningRepo                mealplanning.Repository
	recipeSearchIndex               RecipeTextSearcher
	mealSearchIndex                 MealTextSearcher
	validIngredientSearchIndex      ValidIngredientTextSearcher
	validInstrumentSearchIndex      ValidInstrumentTextSearcher
	validMeasurementUnitSearchIndex ValidMeasurementUnitTextSearcher
	validPreparationSearchIndex     ValidPreparationTextSearcher
	validIngredientStateSearchIndex ValidIngredientStateTextSearcher
	validVesselSearchIndex          ValidVesselTextSearcher
}

func NewMealPlanningDataIndexer(
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	mealPlanningRepo mealplanning.Repository,
	recipeSearchIndex RecipeTextSearcher,
	mealSearchIndex MealTextSearcher,
	validIngredientSearchIndex ValidIngredientTextSearcher,
	validInstrumentSearchIndex ValidInstrumentTextSearcher,
	validMeasurementUnitSearchIndex ValidMeasurementUnitTextSearcher,
	validPreparationSearchIndex ValidPreparationTextSearcher,
	validIngredientStateSearchIndex ValidIngredientStateTextSearcher,
	validVesselSearchIndex ValidVesselTextSearcher,
) *MealPlanningDataIndexer {
	return &MealPlanningDataIndexer{
		logger:                          logging.NewNamedLogger(logger, o11yName),
		tracer:                          tracing.NewNamedTracer(tracerProvider, o11yName),
		mealPlanningRepo:                mealPlanningRepo,
		recipeSearchIndex:               recipeSearchIndex,
		mealSearchIndex:                 mealSearchIndex,
		validIngredientSearchIndex:      validIngredientSearchIndex,
		validInstrumentSearchIndex:      validInstrumentSearchIndex,
		validMeasurementUnitSearchIndex: validMeasurementUnitSearchIndex,
		validPreparationSearchIndex:     validPreparationSearchIndex,
		validIngredientStateSearchIndex: validIngredientStateSearchIndex,
		validVesselSearchIndex:          validVesselSearchIndex,
	}
}

func (i *MealPlanningDataIndexer) Index(indexType string) (textsearch.IndexManager, error) {
	switch indexType {
	case IndexTypeRecipes:
		return i.recipeSearchIndex, nil
	case IndexTypeMeals:
		return i.mealSearchIndex, nil
	case IndexTypeValidIngredients:
		return i.validIngredientSearchIndex, nil
	case IndexTypeValidInstruments:
		return i.validInstrumentSearchIndex, nil
	case IndexTypeValidMeasurementUnits:
		return i.validMeasurementUnitSearchIndex, nil
	case IndexTypeValidPreparations:
		return i.validPreparationSearchIndex, nil
	case IndexTypeValidIngredientStates:
		return i.validIngredientStateSearchIndex, nil
	case IndexTypeValidVessels:
		return i.validVesselSearchIndex, nil
	default:
		return nil, ErrNilIndexRequest
	}
}

func (i *MealPlanningDataIndexer) HandleIndexRequest(
	ctx context.Context,
	indexReq *textsearch.IndexRequest,
) error {
	ctx, span := i.tracer.StartSpan(ctx)
	defer span.End()

	if indexReq == nil {
		return observability.PrepareAndLogError(ErrNilIndexRequest, i.logger, span, "handling index requests")
	}

	logger := i.logger.WithValue("index_type_requested", indexReq.IndexType)

	im, err := i.Index(indexReq.IndexType)
	if err != nil {
		return fmt.Errorf("invalid index type: %s", indexReq.IndexType)
	}

	var (
		toBeIndexed       any
		markAsIndexedFunc func() error
	)

	switch indexReq.IndexType {
	case IndexTypeRecipes:
		var recipe *mealplanning.Recipe
		recipe, err = i.mealPlanningRepo.GetRecipe(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting recipe")
		}

		toBeIndexed = ConvertRecipeToRecipeSearchSubset(recipe)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkRecipeAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeMeals:
		var meal *mealplanning.Meal
		meal, err = i.mealPlanningRepo.GetMeal(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting meal")
		}

		toBeIndexed = ConvertMealToMealSearchSubset(meal)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkMealAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidIngredients:
		var validIngredient *mealplanning.ValidIngredient
		validIngredient, err = i.mealPlanningRepo.GetValidIngredient(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid ingredient")
		}

		toBeIndexed = ConvertValidIngredientToValidIngredientSearchSubset(validIngredient)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidIngredientAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidInstruments:
		var validInstrument *mealplanning.ValidInstrument
		validInstrument, err = i.mealPlanningRepo.GetValidInstrument(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid instrument")
		}

		toBeIndexed = ConvertValidInstrumentToValidInstrumentSearchSubset(validInstrument)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidInstrumentAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidMeasurementUnits:
		var validMeasurementUnit *mealplanning.ValidMeasurementUnit
		validMeasurementUnit, err = i.mealPlanningRepo.GetValidMeasurementUnit(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid measurement unit")
		}

		toBeIndexed = ConvertValidMeasurementUnitToValidMeasurementUnitSearchSubset(validMeasurementUnit)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidMeasurementUnitAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidPreparations:
		var validPreparation *mealplanning.ValidPreparation
		validPreparation, err = i.mealPlanningRepo.GetValidPreparation(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid preparation")
		}

		toBeIndexed = ConvertValidPreparationToValidPreparationSearchSubset(validPreparation)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidPreparationAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidIngredientStates:
		var validIngredientState *mealplanning.ValidIngredientState
		validIngredientState, err = i.mealPlanningRepo.GetValidIngredientState(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid ingredient state")
		}

		toBeIndexed = ConvertValidIngredientStateToValidIngredientStateSearchSubset(validIngredientState)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidIngredientStateAsIndexed(ctx, indexReq.RowID) }

	case IndexTypeValidVessels:
		var validVessel *mealplanning.ValidVessel
		validVessel, err = i.mealPlanningRepo.GetValidVessel(ctx, indexReq.RowID)
		if err != nil {
			return observability.PrepareAndLogError(err, logger, span, "getting valid vessel")
		}

		toBeIndexed = ConvertValidVesselToValidVesselSearchSubset(validVessel)
		markAsIndexedFunc = func() error { return i.mealPlanningRepo.MarkValidVesselAsIndexed(ctx, indexReq.RowID) }
	default:
		logger.Info("invalid index type specified, exiting")
		return nil
	}

	if indexReq.Delete {
		if err = im.Delete(ctx, indexReq.RowID); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "deleting data")
		}

		return nil
	} else {
		if err = im.Index(ctx, indexReq.RowID, toBeIndexed); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "indexing data")
		}

		if err = markAsIndexedFunc(); err != nil {
			return observability.PrepareAndLogError(err, logger, span, "marking data as indexed")
		}
	}

	return nil
}
