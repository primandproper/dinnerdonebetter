package indexing

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"
	"github.com/primandproper/platform-go/v3/reflection"
	textsearch "github.com/primandproper/platform-go/v3/search/text"
	mocksearch "github.com/primandproper/platform-go/v3/search/text/mock"

	"github.com/stretchr/testify/assert"
)

func TestHandleIndexRequest(T *testing.T) {
	T.Parallel()

	T.Run("recipe index type", func(t *testing.T) {
		t.Parallel()

		exampleRecipe := fakes.BuildFakeRecipe()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetRecipe), testutils.ContextMatcher, exampleRecipe.ID).Return(exampleRecipe, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkRecipeAsIndexed), testutils.ContextMatcher, exampleRecipe.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleRecipe.ID,
			IndexType: IndexTypeRecipes,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("meal index type", func(t *testing.T) {
		t.Parallel()

		exampleMeal := fakes.BuildFakeMeal()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetMeal), testutils.ContextMatcher, exampleMeal.ID).Return(exampleMeal, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkMealAsIndexed), testutils.ContextMatcher, exampleMeal.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}

		mim := &mocksearch.IndexMock[MealSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleMeal.ID,
			IndexType: IndexTypeMeals,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid vessel index type", func(t *testing.T) {
		t.Parallel()

		exampleValidVessel := fakes.BuildFakeValidVessel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidVessel), testutils.ContextMatcher, exampleValidVessel.ID).Return(exampleValidVessel, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidVesselAsIndexed), testutils.ContextMatcher, exampleValidVessel.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}

		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidVessel.ID,
			IndexType: IndexTypeValidVessels,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid ingredient index type", func(t *testing.T) {
		t.Parallel()

		exampleValidIngredient := fakes.BuildFakeValidIngredient()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidIngredient), testutils.ContextMatcher, exampleValidIngredient.ID).Return(exampleValidIngredient, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidIngredientAsIndexed), testutils.ContextMatcher, exampleValidIngredient.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidIngredient.ID,
			IndexType: IndexTypeValidIngredients,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid instrument index type", func(t *testing.T) {
		t.Parallel()

		exampleValidInstrument := fakes.BuildFakeValidInstrument()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidInstrument), testutils.ContextMatcher, exampleValidInstrument.ID).Return(exampleValidInstrument, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidInstrumentAsIndexed), testutils.ContextMatcher, exampleValidInstrument.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}

		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidInstrument.ID,
			IndexType: IndexTypeValidInstruments,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid preparation index type", func(t *testing.T) {
		t.Parallel()

		exampleValidPreparation := fakes.BuildFakeValidPreparation()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidPreparation), testutils.ContextMatcher, exampleValidPreparation.ID).Return(exampleValidPreparation, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidPreparationAsIndexed), testutils.ContextMatcher, exampleValidPreparation.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}

		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidPreparation.ID,
			IndexType: IndexTypeValidPreparations,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid measurement unit index type", func(t *testing.T) {
		t.Parallel()

		exampleValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidMeasurementUnit), testutils.ContextMatcher, exampleValidMeasurementUnit.ID).Return(exampleValidMeasurementUnit, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidMeasurementUnitAsIndexed), testutils.ContextMatcher, exampleValidMeasurementUnit.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}

		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}
		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{}
		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidMeasurementUnit.ID,
			IndexType: IndexTypeValidMeasurementUnits,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})

	T.Run("valid ingredient state index type", func(t *testing.T) {
		t.Parallel()

		exampleValidIngredientState := fakes.BuildFakeValidIngredientState()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.Repository{}
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.GetValidIngredientState), testutils.ContextMatcher, exampleValidIngredientState.ID).Return(exampleValidIngredientState, nil)
		mealPlanningRepo.On(reflection.GetMethodName(mealPlanningRepo.MarkValidIngredientStateAsIndexed), testutils.ContextMatcher, exampleValidIngredientState.ID).Return(nil)

		rim := &mocksearch.IndexMock[RecipeSearchSubset]{}
		mim := &mocksearch.IndexMock[MealSearchSubset]{}
		vinm := &mocksearch.IndexMock[ValidIngredientSearchSubset]{}
		vism := &mocksearch.IndexMock[ValidInstrumentSearchSubset]{}
		vmuim := &mocksearch.IndexMock[ValidMeasurementUnitSearchSubset]{}
		vpim := &mocksearch.IndexMock[ValidPreparationSearchSubset]{}

		visim := &mocksearch.IndexMock[ValidIngredientStateSearchSubset]{
			IndexFunc: func(_ context.Context, _ string, _ any) error { return nil },
		}

		vvim := &mocksearch.IndexMock[ValidVesselSearchSubset]{}

		cdi := NewMealPlanningDataIndexer(
			logger,
			tracingnoop.NewTracerProvider(),
			mealPlanningRepo,
			rim,
			mim,
			vinm,
			vism,
			vmuim,
			vpim,
			visim,
			vvim,
		)

		indexReq := &textsearch.IndexRequest{
			RowID:     exampleValidIngredientState.ID,
			IndexType: IndexTypeValidIngredientStates,
			Delete:    false,
		}

		assert.NoError(t, cdi.HandleIndexRequest(ctx, indexReq))
	})
}
