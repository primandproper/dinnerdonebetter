package indexing

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	mealplanningmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/mocks"

	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"
	textsearch "github.com/primandproper/platform-go/v9/search/text"
	mocksearch "github.com/primandproper/platform-go/v9/search/text/mock"

	"github.com/stretchr/testify/assert"
)

func TestHandleIndexRequest(T *testing.T) {
	T.Parallel()

	T.Run("recipe index type", func(t *testing.T) {
		t.Parallel()

		exampleRecipe := fakes.BuildFakeRecipe()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetRecipeFunc: func(_ context.Context, id string) (*mealplanning.Recipe, error) {
				assert.Equal(t, exampleRecipe.ID, id)
				return exampleRecipe, nil
			},
			MarkRecipeAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleRecipe.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetRecipeCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkRecipeAsIndexedCalls(), 1)
	})

	T.Run("meal index type", func(t *testing.T) {
		t.Parallel()

		exampleMeal := fakes.BuildFakeMeal()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetMealFunc: func(_ context.Context, id string) (*mealplanning.Meal, error) {
				assert.Equal(t, exampleMeal.ID, id)
				return exampleMeal, nil
			},
			MarkMealAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleMeal.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetMealCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkMealAsIndexedCalls(), 1)
	})

	T.Run("valid vessel index type", func(t *testing.T) {
		t.Parallel()

		exampleValidVessel := fakes.BuildFakeValidVessel()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidVesselFunc: func(_ context.Context, id string) (*mealplanning.ValidVessel, error) {
				assert.Equal(t, exampleValidVessel.ID, id)
				return exampleValidVessel, nil
			},
			MarkValidVesselAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidVessel.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidVesselCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidVesselAsIndexedCalls(), 1)
	})

	T.Run("valid ingredient index type", func(t *testing.T) {
		t.Parallel()

		exampleValidIngredient := fakes.BuildFakeValidIngredient()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidIngredientFunc: func(_ context.Context, id string) (*mealplanning.ValidIngredient, error) {
				assert.Equal(t, exampleValidIngredient.ID, id)
				return exampleValidIngredient, nil
			},
			MarkValidIngredientAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidIngredient.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidIngredientCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidIngredientAsIndexedCalls(), 1)
	})

	T.Run("valid instrument index type", func(t *testing.T) {
		t.Parallel()

		exampleValidInstrument := fakes.BuildFakeValidInstrument()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidInstrumentFunc: func(_ context.Context, id string) (*mealplanning.ValidInstrument, error) {
				assert.Equal(t, exampleValidInstrument.ID, id)
				return exampleValidInstrument, nil
			},
			MarkValidInstrumentAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidInstrument.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidInstrumentCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidInstrumentAsIndexedCalls(), 1)
	})

	T.Run("valid preparation index type", func(t *testing.T) {
		t.Parallel()

		exampleValidPreparation := fakes.BuildFakeValidPreparation()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidPreparationFunc: func(_ context.Context, id string) (*mealplanning.ValidPreparation, error) {
				assert.Equal(t, exampleValidPreparation.ID, id)
				return exampleValidPreparation, nil
			},
			MarkValidPreparationAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidPreparation.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidPreparationCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidPreparationAsIndexedCalls(), 1)
	})

	T.Run("valid measurement unit index type", func(t *testing.T) {
		t.Parallel()

		exampleValidMeasurementUnit := fakes.BuildFakeValidMeasurementUnit()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidMeasurementUnitFunc: func(_ context.Context, id string) (*mealplanning.ValidMeasurementUnit, error) {
				assert.Equal(t, exampleValidMeasurementUnit.ID, id)
				return exampleValidMeasurementUnit, nil
			},
			MarkValidMeasurementUnitAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidMeasurementUnit.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidMeasurementUnitCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidMeasurementUnitAsIndexedCalls(), 1)
	})

	T.Run("valid ingredient state index type", func(t *testing.T) {
		t.Parallel()

		exampleValidIngredientState := fakes.BuildFakeValidIngredientState()

		ctx := t.Context()
		logger := loggingnoop.NewLogger()

		mealPlanningRepo := &mealplanningmock.RepositoryMock{
			GetValidIngredientStateFunc: func(_ context.Context, id string) (*mealplanning.ValidIngredientState, error) {
				assert.Equal(t, exampleValidIngredientState.ID, id)
				return exampleValidIngredientState, nil
			},
			MarkValidIngredientStateAsIndexedFunc: func(_ context.Context, id string) error {
				assert.Equal(t, exampleValidIngredientState.ID, id)
				return nil
			},
		}

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

		assert.Len(t, mealPlanningRepo.GetValidIngredientStateCalls(), 1)
		assert.Len(t, mealPlanningRepo.MarkValidIngredientStateAsIndexedCalls(), 1)
	})
}
