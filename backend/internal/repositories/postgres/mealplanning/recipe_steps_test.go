package mealplanning

import (
	"context"
	"database/sql"
	"testing"
	"time"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v8/filtering"
	"github.com/primandproper/platform-go/v8/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRecipeStepForTestCreation(t *testing.T, ctx context.Context, recipeID string, dbc *repository) *types.RecipeStep {
	t.Helper()

	recipeStepID := identifiers.New()

	validIngredientState := createValidIngredientStateForTest(t, ctx, nil, dbc)

	preparation := createValidPreparationForTest(t, ctx, nil, dbc)
	ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
	measurementUnit1 := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
	recipeStepIngredient := fakes.BuildFakeRecipeStepIngredient()
	recipeStepIngredient.Ingredient = ingredient
	recipeStepIngredient.MeasurementUnit = *measurementUnit1
	recipeStepIngredient.BelongsToRecipeStep = recipeStepID

	instrument := createValidInstrumentForTest(t, ctx, nil, dbc)
	recipeStepInstrument := fakes.BuildFakeRecipeStepInstrument()
	recipeStepInstrument.Instrument = instrument
	recipeStepInstrument.BelongsToRecipeStep = recipeStepID

	measurementUnit2 := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
	exampleVessel := fakes.BuildFakeValidVessel()
	exampleVessel.CapacityUnit = measurementUnit2
	vessel := createValidVesselForTest(t, ctx, exampleVessel, dbc)
	recipeStepVessel := fakes.BuildFakeRecipeStepVessel()
	recipeStepVessel.Vessel = vessel
	recipeStepVessel.BelongsToRecipeStep = recipeStepID

	exampleRecipeMedia := fakes.BuildFakeRecipeMedia()
	exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
	exampleRecipePrepTask.BelongsToRecipe = recipeID
	exampleRecipePrepTask.TaskSteps = []*types.RecipePrepTaskStep{
		{
			ID:                      identifiers.New(),
			BelongsToRecipeStep:     recipeStepID,
			BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
			SatisfiesRecipeStep:     true,
		},
	}

	exampleRecipeStepProduct := fakes.BuildFakeRecipeStepProduct()
	exampleRecipeStepProduct.BelongsToRecipeStep = recipeStepID
	exampleRecipeStepProduct.MeasurementUnit = measurementUnit1

	exampleRecipeStepCompletionCondition := fakes.BuildFakeRecipeStepCompletionCondition()
	exampleRecipeStepCompletionCondition.BelongsToRecipeStep = recipeStepID
	exampleRecipeStepCompletionCondition.IngredientState = *validIngredientState
	exampleRecipeStepCompletionCondition.Ingredients = []*types.RecipeStepCompletionConditionIngredient{
		{
			ID:                                     identifiers.New(),
			BelongsToRecipeStepCompletionCondition: exampleRecipeStepCompletionCondition.ID,
			RecipeStepIngredient:                   recipeStepIngredient.ID,
		},
	}

	exampleRecipeStep := fakes.BuildFakeRecipeStep()
	exampleRecipeStep.ID = recipeStepID
	exampleRecipeStep.Index = 0
	exampleRecipeStep.BelongsToRecipe = recipeID
	exampleRecipeStep.Preparation = *preparation
	exampleRecipeStep.Ingredients = []*types.RecipeStepIngredient{
		recipeStepIngredient,
	}
	exampleRecipeStep.Instruments = []*types.RecipeStepInstrument{
		recipeStepInstrument,
	}
	exampleRecipeStep.Vessels = []*types.RecipeStepVessel{
		recipeStepVessel,
	}
	exampleRecipeStep.Media = []*types.RecipeMedia{
		exampleRecipeMedia,
	}
	exampleRecipeStep.Products = []*types.RecipeStepProduct{
		exampleRecipeStepProduct,
	}
	exampleRecipeStep.CompletionConditions = []*types.RecipeStepCompletionCondition{
		exampleRecipeStepCompletionCondition,
	}

	return exampleRecipeStep
}

// buildRecipeStepForTestCreationWithInstrument builds a recipe step with a required (optional=false) instrument.
// Use this when testing SearchForRecipesWithInstrumentOwnership.
func buildRecipeStepForTestCreationWithInstrument(t *testing.T, ctx context.Context, recipeID string, instrument *types.ValidInstrument, dbc *repository) *types.RecipeStep {
	t.Helper()

	step := buildRecipeStepForTestCreation(t, ctx, recipeID, dbc)
	// Override with the passed instrument and ensure it's required (optional=false)
	step.Instruments = []*types.RecipeStepInstrument{
		{
			ID:                  identifiers.New(),
			Instrument:          instrument,
			Name:                instrument.Name,
			BelongsToRecipeStep: step.ID,
			Optional:            false, // Required for SearchForRecipesWithInstrumentOwnership
			MinQuantity:         1,
		},
	}
	return step
}

func createRecipeStepForTest(t *testing.T, ctx context.Context, recipeID string, exampleRecipeStep *types.RecipeStep, dbc *repository) *types.RecipeStep {
	t.Helper()

	// create
	if exampleRecipeStep == nil {
		exampleRecipeStep = fakes.BuildFakeRecipeStep()
	}
	dbInput := converters.ConvertRecipeStepToRecipeStepDatabaseCreationInput(exampleRecipeStep)

	created, err := dbc.CreateRecipeStep(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleRecipeStep.Media = nil
	exampleRecipeStep.CreatedAt = created.CreatedAt
	exampleRecipeStep.Preparation = types.ValidPreparation{ID: created.Preparation.ID}

	for j := range created.Ingredients {
		exampleRecipeStep.Ingredients[j].CreatedAt = created.Ingredients[j].CreatedAt

		assert.Equal(t, exampleRecipeStep.Ingredients[j].Ingredient.ID, created.Ingredients[j].Ingredient.ID)
		exampleRecipeStep.Ingredients[j].Ingredient = created.Ingredients[j].Ingredient

		assert.Equal(t, exampleRecipeStep.Ingredients[j].MeasurementUnit.ID, created.Ingredients[j].MeasurementUnit.ID)
		exampleRecipeStep.Ingredients[j].MeasurementUnit = created.Ingredients[j].MeasurementUnit
	}

	for j := range created.Instruments {
		exampleRecipeStep.Instruments[j].CreatedAt = created.Instruments[j].CreatedAt

		assert.Equal(t, exampleRecipeStep.Instruments[j].Instrument.ID, created.Instruments[j].Instrument.ID)
		exampleRecipeStep.Instruments[j].Instrument = created.Instruments[j].Instrument
	}

	for j := range created.Vessels {
		exampleRecipeStep.Vessels[j].CreatedAt = created.Vessels[j].CreatedAt

		assert.Equal(t, exampleRecipeStep.Vessels[j].Vessel.ID, created.Vessels[j].Vessel.ID)
		exampleRecipeStep.Vessels[j].Vessel = created.Vessels[j].Vessel

		if created.Vessels[j].Vessel.CapacityUnit != nil {
			assert.Equal(t, exampleRecipeStep.Vessels[j].Vessel.CapacityUnit.ID, created.Vessels[j].Vessel.CapacityUnit.ID)
			exampleRecipeStep.Vessels[j].Vessel.CapacityUnit = created.Vessels[j].Vessel.CapacityUnit
		}
	}

	for j := range created.CompletionConditions {
		exampleRecipeStep.CompletionConditions[j].CreatedAt = created.CompletionConditions[j].CreatedAt

		assert.Equal(t, exampleRecipeStep.CompletionConditions[j].IngredientState.ID, created.CompletionConditions[j].IngredientState.ID)
		exampleRecipeStep.CompletionConditions[j].IngredientState = created.CompletionConditions[j].IngredientState

		// Update the CreatedAt for each ingredient in the completion condition
		for k := range created.CompletionConditions[j].Ingredients {
			if k < len(exampleRecipeStep.CompletionConditions[j].Ingredients) {
				exampleRecipeStep.CompletionConditions[j].Ingredients[k].CreatedAt = created.CompletionConditions[j].Ingredients[k].CreatedAt
			}
		}
	}

	for j := range created.Products {
		exampleRecipeStep.Products[j].CreatedAt = created.Products[j].CreatedAt
		assert.Equal(t, exampleRecipeStep.Products[j].MeasurementUnit.ID, created.Products[j].MeasurementUnit.ID)
		exampleRecipeStep.Products[j].MeasurementUnit = created.Products[j].MeasurementUnit
	}

	for j := range created.Media {
		assert.Equal(t, exampleRecipeStep.Media[j].ID, created.Media[j].ID)
		exampleRecipeStep.Media[j] = created.Media[j]
	}

	assert.Equal(t, exampleRecipeStep.Products, created.Products)
	assert.Equal(t, exampleRecipeStep.Instruments, created.Instruments)
	assert.Equal(t, exampleRecipeStep.Vessels, created.Vessels)
	assert.Equal(t, exampleRecipeStep.Ingredients, created.Ingredients)
	assert.Equal(t, exampleRecipeStep.Media, created.Media)
	assert.Equal(t, exampleRecipeStep.CompletionConditions, created.CompletionConditions)

	// Test functional equality for created recipe step - compare all fields except timestamps
	// which are expected to differ between app time and database time
	assert.Equal(t, exampleRecipeStep.ID, created.ID)
	assert.Equal(t, exampleRecipeStep.Index, created.Index)
	assert.Equal(t, exampleRecipeStep.BelongsToRecipe, created.BelongsToRecipe)
	assert.Equal(t, exampleRecipeStep.Notes, created.Notes)
	assert.Equal(t, exampleRecipeStep.ExplicitInstructions, created.ExplicitInstructions)
	assert.Equal(t, exampleRecipeStep.ConditionExpression, created.ConditionExpression)
	assert.Equal(t, exampleRecipeStep.Optional, created.Optional)
	assert.Equal(t, exampleRecipeStep.StartTimerAutomatically, created.StartTimerAutomatically)
	assert.Equal(t, exampleRecipeStep.MinEstimatedTimeInSeconds, created.MinEstimatedTimeInSeconds)
	assert.Equal(t, exampleRecipeStep.MaxEstimatedTimeInSeconds, created.MaxEstimatedTimeInSeconds)
	assert.Equal(t, exampleRecipeStep.MinTemperatureInCelsius, created.MinTemperatureInCelsius)
	assert.Equal(t, exampleRecipeStep.MaxTemperatureInCelsius, created.MaxTemperatureInCelsius)

	// Check timestamps are within reasonable tolerance (5 seconds)
	assert.WithinDuration(t, exampleRecipeStep.CreatedAt, created.CreatedAt, 5*time.Second)

	recipeStep, err := dbc.GetRecipeStep(ctx, recipeID, created.ID)
	require.NoError(t, err)
	require.NotNil(t, recipeStep)

	assert.Equal(t, exampleRecipeStep.Preparation.ID, recipeStep.Preparation.ID)
	exampleRecipeStep.Preparation = recipeStep.Preparation

	// Test functional equality - compare all fields except timestamps
	// which are expected to differ between app time and database time
	assert.Equal(t, exampleRecipeStep.ID, recipeStep.ID)
	assert.Equal(t, exampleRecipeStep.Index, recipeStep.Index)
	assert.Equal(t, exampleRecipeStep.BelongsToRecipe, recipeStep.BelongsToRecipe)
	assert.Equal(t, exampleRecipeStep.Notes, recipeStep.Notes)
	assert.Equal(t, exampleRecipeStep.ExplicitInstructions, recipeStep.ExplicitInstructions)
	assert.Equal(t, exampleRecipeStep.ConditionExpression, recipeStep.ConditionExpression)
	assert.Equal(t, exampleRecipeStep.Optional, recipeStep.Optional)
	assert.Equal(t, exampleRecipeStep.StartTimerAutomatically, recipeStep.StartTimerAutomatically)
	assert.Equal(t, exampleRecipeStep.MinEstimatedTimeInSeconds, recipeStep.MinEstimatedTimeInSeconds)
	assert.Equal(t, exampleRecipeStep.MaxEstimatedTimeInSeconds, recipeStep.MaxEstimatedTimeInSeconds)
	assert.Equal(t, exampleRecipeStep.MinTemperatureInCelsius, recipeStep.MinTemperatureInCelsius)
	assert.Equal(t, exampleRecipeStep.MaxTemperatureInCelsius, recipeStep.MaxTemperatureInCelsius)

	// Check timestamps are within reasonable tolerance (5 seconds)
	assert.WithinDuration(t, exampleRecipeStep.CreatedAt, recipeStep.CreatedAt, 5*time.Second)

	// Compare nested collections by checking their lengths and non-timestamp fields
	assert.Equal(t, len(exampleRecipeStep.Ingredients), len(recipeStep.Ingredients))
	for i := range recipeStep.Ingredients {
		if i > len(exampleRecipeStep.Ingredients) {
			continue
		}

		assert.Equal(t, exampleRecipeStep.Ingredients[i].ID, recipeStep.Ingredients[i].ID)
		assert.Equal(t, exampleRecipeStep.Ingredients[i].BelongsToRecipeStep, recipeStep.Ingredients[i].BelongsToRecipeStep)
		assert.Equal(t, exampleRecipeStep.Ingredients[i].Name, recipeStep.Ingredients[i].Name)
		assert.Equal(t, exampleRecipeStep.Ingredients[i].MeasurementUnit.ID, recipeStep.Ingredients[i].MeasurementUnit.ID)
		assert.WithinDuration(t, exampleRecipeStep.Ingredients[i].CreatedAt, recipeStep.Ingredients[i].CreatedAt, 5*time.Second)
	}

	assert.Equal(t, len(exampleRecipeStep.Instruments), len(recipeStep.Instruments))
	for i := range recipeStep.Instruments {
		if i > len(exampleRecipeStep.Instruments) {
			continue
		}

		assert.Equal(t, exampleRecipeStep.Instruments[i].ID, recipeStep.Instruments[i].ID)
		assert.Equal(t, exampleRecipeStep.Instruments[i].BelongsToRecipeStep, recipeStep.Instruments[i].BelongsToRecipeStep)
		assert.WithinDuration(t, exampleRecipeStep.Instruments[i].CreatedAt, recipeStep.Instruments[i].CreatedAt, 5*time.Second)
	}

	assert.Equal(t, len(exampleRecipeStep.Vessels), len(recipeStep.Vessels))
	for i := range recipeStep.Vessels {
		if i > len(exampleRecipeStep.Vessels) {
			continue
		}

		assert.Equal(t, exampleRecipeStep.Vessels[i].ID, recipeStep.Vessels[i].ID)
		assert.Equal(t, exampleRecipeStep.Vessels[i].BelongsToRecipeStep, recipeStep.Vessels[i].BelongsToRecipeStep)
		assert.WithinDuration(t, exampleRecipeStep.Vessels[i].CreatedAt, recipeStep.Vessels[i].CreatedAt, 5*time.Second)
	}

	assert.Equal(t, len(exampleRecipeStep.Products), len(recipeStep.Products))
	for i := range recipeStep.Products {
		if i > len(exampleRecipeStep.Products) {
			continue
		}

		assert.Equal(t, exampleRecipeStep.Products[i].ID, recipeStep.Products[i].ID)
		assert.Equal(t, exampleRecipeStep.Products[i].BelongsToRecipeStep, recipeStep.Products[i].BelongsToRecipeStep)
		assert.Equal(t, exampleRecipeStep.Products[i].Name, recipeStep.Products[i].Name)
		assert.WithinDuration(t, exampleRecipeStep.Products[i].CreatedAt, recipeStep.Products[i].CreatedAt, 5*time.Second)
	}

	assert.Equal(t, len(exampleRecipeStep.CompletionConditions), len(recipeStep.CompletionConditions))
	for i := range recipeStep.CompletionConditions {
		if i > len(exampleRecipeStep.CompletionConditions) {
			continue
		}

		assert.Equal(t, exampleRecipeStep.CompletionConditions[i].ID, recipeStep.CompletionConditions[i].ID)
		assert.Equal(t, exampleRecipeStep.CompletionConditions[i].BelongsToRecipeStep, recipeStep.CompletionConditions[i].BelongsToRecipeStep)
		assert.WithinDuration(t, exampleRecipeStep.CompletionConditions[i].CreatedAt, recipeStep.CompletionConditions[i].CreatedAt, 5*time.Second)
	}

	assert.NoError(t, err)

	return created
}

func TestQuerier_Integration_RecipeSteps(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipe := createRecipeForTest(t, ctx, exampleRecipe, dbc, true)

	exampleRecipeStep := buildRecipeStepForTestCreation(t, ctx, createdRecipe.ID, dbc)
	exampleRecipeStep.BelongsToRecipe = createdRecipe.ID
	createdRecipeSteps := []*types.RecipeStep{
		exampleRecipe.Steps[0],
	}

	// create
	createdRecipeSteps = append(createdRecipeSteps, createRecipeStepForTest(t, ctx, exampleRecipe.ID, exampleRecipeStep, dbc))

	// create more
	for range exampleQuantity {
		input := buildRecipeStepForTestCreation(t, ctx, createdRecipe.ID, dbc)
		input.BelongsToRecipe = createdRecipe.ID
		createdRecipeSteps = append(createdRecipeSteps, createRecipeStepForTest(t, ctx, exampleRecipe.ID, input, dbc))
	}

	// fetch as list
	recipeSteps, err := dbc.GetRecipeSteps(ctx, exampleRecipe.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, recipeSteps.Data)
	assert.Equal(t, len(createdRecipeSteps), len(recipeSteps.Data))

	// delete
	for _, recipeStep := range createdRecipeSteps {
		assert.NoError(t, dbc.ArchiveRecipeStep(ctx, exampleRecipe.ID, recipeStep.ID))

		var exists bool
		exists, err = dbc.RecipeStepExists(ctx, exampleRecipe.ID, recipeStep.ID)
		assert.NoError(t, err)
		assert.False(t, exists)

		var y *types.RecipeStep
		y, err = dbc.GetRecipeStep(ctx, exampleRecipe.ID, recipeStep.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_RecipeStepExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeStep := fakes.BuildFakeRecipeStep()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepExists(ctx, "", exampleRecipeStep.ID)
		assert.Error(t, err)
		assert.False(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		exampleRecipeID := fakes.BuildFakeID()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeStepExists(ctx, exampleRecipeID, "")
		assert.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_GetRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStep := fakes.BuildFakeRecipeStep()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStep(ctx, "", exampleRecipeStep.ID)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeStep(ctx, exampleRecipeID, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_getRecipeStepByID(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.getRecipeStepByID(ctx, c.writeDB, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetRecipeSteps(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		filter := filtering.DefaultQueryFilter()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetRecipeSteps(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_CreateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipeStep(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestSQLQuerier_createRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.createRecipeStep(ctx, c.writeDB, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipeStep(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipeStep(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeStep := fakes.BuildFakeRecipeStep()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStep(ctx, "", exampleRecipeStep.ID))
	})

	T.Run("with invalid recipe step MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipeID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipeStep(ctx, exampleRecipeID, ""))
	})
}

func TestQuerier_Integration_RecipeSteps_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	recipeStruct := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	// Remove the default step so we start with an empty recipe
	recipeStruct.Steps = nil
	// Also clear prep tasks since they reference recipe steps that won't exist
	recipeStruct.PrepTasks = nil
	recipe := createRecipeForTest(t, ctx, recipeStruct, dbc, false)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.RecipeStep]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe step",
		CreateItem: func(ctx context.Context, i int) *types.RecipeStep {
			recipeStep := buildRecipeStepForTestCreation(t, ctx, recipe.ID, dbc)
			recipeStep.Index = uint32(i)
			return createRecipeStepForTest(t, ctx, recipe.ID, recipeStep, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStep], error) {
			return dbc.GetRecipeSteps(ctx, recipe.ID, filter)
		},
		GetID: func(recipeStep *types.RecipeStep) string {
			return recipeStep.ID
		},
		CleanupItem: func(ctx context.Context, recipeStep *types.RecipeStep) error {
			return dbc.ArchiveRecipeStep(ctx, recipe.ID, recipeStep.ID)
		},
	})
}
