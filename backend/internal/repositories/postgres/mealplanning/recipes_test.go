package mealplanning

import (
	"context"
	"fmt"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildRecipeForTestCreation(t *testing.T, ctx context.Context, userID string, dbc *repository) *mealplanning.Recipe {
	t.Helper()

	if userID == "" {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		userID = user.ID
	}

	recipeStepID := identifiers.New()
	exampleRecipe := fakes.BuildFakeRecipe()

	exampleRecipeMedia := fakes.BuildFakeRecipeMedia()
	exampleRecipeMedia.BelongsToRecipe = &exampleRecipe.ID
	exampleRecipe.Media = []*mealplanning.RecipeMedia{
		exampleRecipeMedia,
	}

	exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
	exampleRecipePrepTask.BelongsToRecipe = exampleRecipe.ID
	exampleRecipePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
		{
			ID:                      identifiers.New(),
			BelongsToRecipeStep:     recipeStepID,
			BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
			SatisfiesRecipeStep:     true,
		},
	}
	exampleRecipe.PrepTasks = []*mealplanning.RecipePrepTask{
		exampleRecipePrepTask,
	}

	exampleRecipe.Steps = []*mealplanning.RecipeStep{
		buildRecipeStepForTestCreation(t, ctx, exampleRecipe.ID, dbc),
	}
	exampleRecipePrepTask.TaskSteps[0].BelongsToRecipeStep = exampleRecipe.Steps[0].ID
	exampleRecipe.CreatedByUser = userID

	exampleRecipe.Media = []*mealplanning.RecipeMedia{}
	for i := range exampleRecipe.Steps {
		exampleRecipe.Steps[i].Media = []*mealplanning.RecipeMedia{}
		exampleRecipe.Steps[i].CompletionConditions = nil // []*mealplanning.RecipeStepCompletionCondition{}
	}

	return exampleRecipe
}

// buildRecipeForTestCreationWithInstrument builds a recipe with a step that has a required instrument.
// Use this when testing SearchForRecipesWithInstrumentOwnership.
func buildRecipeForTestCreationWithInstrument(t *testing.T, ctx context.Context, userID string, instrument *mealplanning.ValidInstrument, dbc *repository) *mealplanning.Recipe {
	t.Helper()

	if userID == "" {
		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
		userID = user.ID
	}

	exampleRecipe := fakes.BuildFakeRecipe()

	exampleRecipeMedia := fakes.BuildFakeRecipeMedia()
	exampleRecipeMedia.BelongsToRecipe = &exampleRecipe.ID
	exampleRecipe.Media = []*mealplanning.RecipeMedia{
		exampleRecipeMedia,
	}

	exampleRecipePrepTask := fakes.BuildFakeRecipePrepTask()
	exampleRecipePrepTask.BelongsToRecipe = exampleRecipe.ID
	exampleRecipe.Steps = []*mealplanning.RecipeStep{
		buildRecipeStepForTestCreationWithInstrument(t, ctx, exampleRecipe.ID, instrument, dbc),
	}
	exampleRecipePrepTask.TaskSteps = []*mealplanning.RecipePrepTaskStep{
		{
			ID:                      identifiers.New(),
			BelongsToRecipeStep:     exampleRecipe.Steps[0].ID,
			BelongsToRecipePrepTask: exampleRecipePrepTask.ID,
			SatisfiesRecipeStep:     true,
		},
	}
	exampleRecipe.PrepTasks = []*mealplanning.RecipePrepTask{
		exampleRecipePrepTask,
	}
	exampleRecipe.CreatedByUser = userID

	exampleRecipe.Media = []*mealplanning.RecipeMedia{}
	for i := range exampleRecipe.Steps {
		exampleRecipe.Steps[i].Media = []*mealplanning.RecipeMedia{}
		exampleRecipe.Steps[i].CompletionConditions = nil // []*mealplanning.RecipeStepCompletionCondition{}
	}

	return exampleRecipe
}

func createRecipeForTest(t *testing.T, ctx context.Context, exampleRecipe *mealplanning.Recipe, dbc *repository, alsoCreateMeal bool) *mealplanning.Recipe {
	t.Helper()

	// create
	if exampleRecipe == nil {
		exampleRecipe = buildRecipeForTestCreation(t, ctx, "", dbc)
	}
	dbInput := converters.ConvertRecipeToRecipeDatabaseCreationInput(exampleRecipe)
	dbInput.AlsoCreateMeal = alsoCreateMeal

	created, err := dbc.CreateRecipe(ctx, dbInput)
	require.NoError(t, err)
	require.NotNil(t, created)

	recipe, err := dbc.GetRecipe(ctx, created.ID)
	require.NoError(t, err)
	require.NotNil(t, recipe)
	exampleRecipe.CreatedAt = recipe.CreatedAt

	for i := range recipe.PrepTasks {
		exampleRecipe.PrepTasks[i].CreatedAt = recipe.PrepTasks[i].CreatedAt
	}

	for i := range recipe.Steps {
		assert.Equal(t, exampleRecipe.Steps[i].Preparation.ID, recipe.Steps[i].Preparation.ID)
		exampleRecipe.Steps[i].Preparation = recipe.Steps[i].Preparation
		exampleRecipe.Steps[i].CreatedAt = recipe.Steps[i].CreatedAt

		for j := range recipe.Steps[i].Ingredients {
			exampleRecipe.Steps[i].Ingredients[j].CreatedAt = recipe.Steps[i].Ingredients[j].CreatedAt
			exampleRecipe.Steps[i].Ingredients[j].ScaleFactor = recipe.Steps[i].Ingredients[j].ScaleFactor

			assert.Equal(t, exampleRecipe.Steps[i].Ingredients[j].Ingredient.ID, recipe.Steps[i].Ingredients[j].Ingredient.ID)
			exampleRecipe.Steps[i].Ingredients[j].Ingredient = recipe.Steps[i].Ingredients[j].Ingredient

			assert.Equal(t, exampleRecipe.Steps[i].Ingredients[j].MeasurementUnit.ID, recipe.Steps[i].Ingredients[j].MeasurementUnit.ID)
			exampleRecipe.Steps[i].Ingredients[j].MeasurementUnit = recipe.Steps[i].Ingredients[j].MeasurementUnit
		}
		for j := range recipe.Steps[i].Instruments {
			exampleRecipe.Steps[i].Instruments[j].CreatedAt = recipe.Steps[i].Instruments[j].CreatedAt
			exampleRecipe.Steps[i].Instruments[j].ScaleFactor = recipe.Steps[i].Instruments[j].ScaleFactor

			assert.Equal(t, exampleRecipe.Steps[i].Instruments[j].Instrument.ID, recipe.Steps[i].Instruments[j].Instrument.ID)
			exampleRecipe.Steps[i].Instruments[j].Instrument = recipe.Steps[i].Instruments[j].Instrument
		}
		for j := range recipe.Steps[i].Vessels {
			exampleRecipe.Steps[i].Vessels[j].CreatedAt = recipe.Steps[i].Vessels[j].CreatedAt
			exampleRecipe.Steps[i].Vessels[j].ScaleFactor = recipe.Steps[i].Vessels[j].ScaleFactor

			assert.Equal(t, exampleRecipe.Steps[i].Vessels[j].Vessel.ID, recipe.Steps[i].Vessels[j].Vessel.ID)
			exampleRecipe.Steps[i].Vessels[j].Vessel = recipe.Steps[i].Vessels[j].Vessel

			if recipe.Steps[i].Vessels[j].Vessel.CapacityUnit != nil {
				assert.Equal(t, exampleRecipe.Steps[i].Vessels[j].Vessel.CapacityUnit.ID, recipe.Steps[i].Vessels[j].Vessel.CapacityUnit.ID)
				exampleRecipe.Steps[i].Vessels[j].Vessel.CapacityUnit = recipe.Steps[i].Vessels[j].Vessel.CapacityUnit
			}
		}
		for j := range recipe.Steps[i].CompletionConditions {
			exampleRecipe.Steps[i].CompletionConditions[j].CreatedAt = recipe.Steps[i].CompletionConditions[j].CreatedAt

			assert.Equal(t, exampleRecipe.Steps[i].CompletionConditions[j].IngredientState.ID, recipe.Steps[i].CompletionConditions[j].IngredientState.ID)
			exampleRecipe.Steps[i].CompletionConditions[j].IngredientState = recipe.Steps[i].CompletionConditions[j].IngredientState

			for k := range recipe.Steps[i].CompletionConditions[j].Ingredients {
				assert.Equal(t, exampleRecipe.Steps[i].CompletionConditions[j].Ingredients[k].ID, recipe.Steps[i].CompletionConditions[j].Ingredients[k].ID)
				exampleRecipe.Steps[i].CompletionConditions[j].Ingredients[k] = recipe.Steps[i].CompletionConditions[j].Ingredients[k]
			}
		}
		for j := range recipe.Steps[i].Products {
			exampleRecipe.Steps[i].Products[j].CreatedAt = recipe.Steps[i].Products[j].CreatedAt
			assert.Equal(t, exampleRecipe.Steps[i].Products[j].MeasurementUnit.ID, recipe.Steps[i].Products[j].MeasurementUnit.ID)
			exampleRecipe.Steps[i].Products[j].MeasurementUnit = recipe.Steps[i].Products[j].MeasurementUnit
		}
		for j := range recipe.Steps[i].Media {
			assert.Equal(t, exampleRecipe.Steps[i].Media[j].ID, recipe.Steps[i].Media[j].ID)
			exampleRecipe.Steps[i].Media[j] = recipe.Steps[i].Media[j]
		}

		require.Len(t, recipe.Steps[i].Products, len(exampleRecipe.Steps[i].Products))
		require.Len(t, recipe.Steps[i].Instruments, len(exampleRecipe.Steps[i].Instruments))
		require.Len(t, recipe.Steps[i].Vessels, len(exampleRecipe.Steps[i].Vessels))
		require.Len(t, recipe.Steps[i].Ingredients, len(exampleRecipe.Steps[i].Ingredients))
		require.Len(t, recipe.Steps[i].Media, len(exampleRecipe.Steps[i].Media))
		require.Len(t, recipe.Steps[i].CompletionConditions, len(exampleRecipe.Steps[i].CompletionConditions))

		expectedStep := exampleRecipe.Steps[i]
		actualStep := recipe.Steps[i]
		require.Equal(t, expectedStep, actualStep)
	}

	require.Equal(t, exampleRecipe, recipe)

	return recipe
}

func TestQuerier_Integration_Recipes(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	createdRecipes := []*mealplanning.Recipe{}

	// create
	exampleRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	createdRecipes = append(createdRecipes, createRecipeForTest(t, ctx, exampleRecipe, dbc, true))

	// update
	updatedRecipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
	updatedRecipe.ID = createdRecipes[0].ID
	require.NoError(t, dbc.UpdateRecipe(ctx, updatedRecipe))

	// create more
	for i := range exampleQuantity {
		exampleRecipe = buildRecipeForTestCreation(t, ctx, user.ID, dbc)
		exampleRecipe.Name = fmt.Sprintf("%s %d", exampleRecipe.Name, i)
		createdRecipes = append(createdRecipes, createRecipeForTest(t, ctx, exampleRecipe, dbc, true))
	}

	// fetch as list
	recipes, err := dbc.GetRecipes(ctx, mealplanning.RecipeStatusSubmitted, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, recipes.Data)
	assert.Len(t, recipes.Data, len(createdRecipes))

	needingIndexing, err := dbc.GetRecipeIDsThatNeedSearchIndexing(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, needingIndexing)

	// search
	searchResults, err := dbc.SearchForRecipes(ctx, createdRecipes[0].Name, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, searchResults)

	byIDs, err := dbc.GetRecipesWithIDs(ctx, []string{createdRecipes[0].ID})
	require.NoError(t, err)
	assert.NotEmpty(t, byIDs)
	assert.Equal(t, createdRecipes[0].ID, byIDs[0].ID)

	// delete
	for _, recipe := range createdRecipes {
		assert.NoError(t, dbc.MarkRecipeAsIndexed(ctx, recipe.ID))
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipe.ID, user.ID))

		var exists bool
		exists, err = dbc.RecipeExists(ctx, recipe.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	}
}

func TestQuerier_Integration_SearchForRecipesWithInstrumentOwnership(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	account := pgtesting.CreateAccountForTest(t, nil, user.ID, dbc.writeDB)

	// Create instrument and account ownership
	instrument := createValidInstrumentForTest(t, ctx, nil, dbc)
	ownershipInput := fakes.BuildFakeAccountInstrumentOwnership()
	ownershipInput.BelongsToAccount = account.ID
	ownershipInput.Instrument = *instrument
	createAccountInstrumentOwnershipForTest(t, ctx, ownershipInput, dbc)

	// Create recipe with required instrument (account owns it)
	recipeWithOwnedInstrument := buildRecipeForTestCreationWithInstrument(t, ctx, user.ID, instrument, dbc)
	recipeWithOwnedInstrument.Name = "Pasta Carbonara SearchTest"
	createdRecipe := createRecipeForTest(t, ctx, recipeWithOwnedInstrument, dbc, false)

	// Search with account that owns the instrument - should find the recipe
	results, err := dbc.SearchForRecipesWithInstrumentOwnership(ctx, account.ID, "Pasta Carbonara", nil)
	require.NoError(t, err)
	require.NotEmpty(t, results.Data)
	found := false
	for _, r := range results.Data {
		if r.ID == createdRecipe.ID {
			found = true
			break
		}
	}
	assert.True(t, found, "recipe with owned instrument should appear in search results")

	// Create recipe with different instrument (account does NOT own it)
	otherInstrument := createValidInstrumentForTest(t, ctx, nil, dbc)
	recipeWithUnownedInstrument := buildRecipeForTestCreationWithInstrument(t, ctx, user.ID, otherInstrument, dbc)
	recipeWithUnownedInstrument.Name = "Excluded Recipe SearchTest"
	excludedRecipe := createRecipeForTest(t, ctx, recipeWithUnownedInstrument, dbc, false)

	// Search again - excluded recipe should NOT appear (account doesn't own its instrument)
	results2, err := dbc.SearchForRecipesWithInstrumentOwnership(ctx, account.ID, "SearchTest", nil)
	require.NoError(t, err)
	foundExcluded := false
	for _, r := range results2.Data {
		if r.ID == excludedRecipe.ID {
			foundExcluded = true
			break
		}
	}
	assert.False(t, foundExcluded, "recipe with unowned instrument should not appear in search results")

	// Cleanup
	assert.NoError(t, dbc.ArchiveRecipe(ctx, createdRecipe.ID, user.ID))
	assert.NoError(t, dbc.ArchiveRecipe(ctx, excludedRecipe.ID, user.ID))
}

func TestQuerier_Integration_GetRecipesWithIDs(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)
	defer func() {
	}()

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	r1 := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)
	r2 := createRecipeForTest(t, ctx, buildRecipeForTestCreation(t, ctx, user.ID, dbc), dbc, false)

	res, err := dbc.GetRecipesWithIDs(ctx, []string{r1.ID, r2.ID})
	require.NoError(t, err)
	require.Len(t, res, 2)

	found := map[string]*mealplanning.Recipe{}
	for _, rec := range res {
		found[rec.ID] = rec
		require.NotEmpty(t, rec.Steps)
	}

	assert.Contains(t, found, r1.ID)
	assert.Contains(t, found, r2.ID)
}

func TestQuerier_RecipeExists(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		c := buildInertClientForTest(t)

		actual, err := c.RecipeExists(ctx, "")
		require.Error(t, err)
		assert.False(t, actual)
	})
}

func TestQuerier_CreateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateRecipe(ctx, nil)
		require.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.UpdateRecipe(ctx, nil))
	})
}

func TestQuerier_ArchiveRecipe(T *testing.T) {
	T.Parallel()

	T.Run("with invalid recipe MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleAccountID := fakes.BuildFakeID()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipe(ctx, "", exampleAccountID))
	})

	T.Run("with invalid account MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		exampleRecipe := fakes.BuildFakeRecipe()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveRecipe(ctx, exampleRecipe.ID, ""))
	})
}

func Test_findCreatedRecipeStepProductsForIngredients(T *testing.T) {
	T.Parallel()

	T.Run("sopa de frijol", func(t *testing.T) {
		t.Parallel()

		soak := fakes.BuildFakeValidPreparation()
		water := fakes.BuildFakeValidIngredient()
		pintoBeans := fakes.BuildFakeValidIngredient()
		garlicPaste := fakes.BuildFakeValidIngredient()
		productName := "soaked pinto beans"

		exampleRecipeInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:          identifiers.New(),
			Name:        "sopa de frijol",
			Description: "",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                fakes.BuildFakeID(),
							Name:              productName,
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "first step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							IngredientID:      &pintoBeans.ID,
							Name:              "pinto beans",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       500,
						},
						{
							IngredientID:      &water.ID,
							Name:              "water",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       500,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							Name:              "final output",
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "second step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							Name:                            productName,
							MeasurementUnitID:               fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:                     1000,
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							ProductOfRecipeStepIndex:        new(uint64(0)),
						},
						{
							IngredientID:      &garlicPaste.ID,
							Name:              "garlic paste",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       10,
						},
					},
					Index: 1,
				},
			},
		}

		ctx := t.Context()
		c := buildInertClientForTest(t)
		err := c.findCreatedRecipeStepProductsForIngredients(ctx, exampleRecipeInput)
		require.NoError(t, err)

		require.NotNil(t, exampleRecipeInput.Steps[1].Ingredients[0].RecipeStepProductID)
		assert.Equal(t, exampleRecipeInput.Steps[0].Products[0].ID, *exampleRecipeInput.Steps[1].Ingredients[0].RecipeStepProductID)
	})

	T.Run("slightly more complicated recipe", func(t *testing.T) {
		t.Parallel()

		soak := fakes.BuildFakeValidPreparation()
		water := fakes.BuildFakeValidIngredient()
		pintoBeans := fakes.BuildFakeValidIngredient()
		garlicPaste := fakes.BuildFakeValidIngredient()
		productName := "soaked pinto beans"

		exampleRecipeInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:          identifiers.New(),
			Name:        "sopa de frijol",
			Description: "",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                fakes.BuildFakeID(),
							Name:              productName,
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "first step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							IngredientID:      &pintoBeans.ID,
							Name:              "pinto beans",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       500,
						},
						{
							IngredientID:      &water.ID,
							Name:              "water",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       5,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							Name:              "pressure cooked beans",
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "second step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							Name:                            productName,
							MeasurementUnitID:               fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:                     1000,
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
						},
						{
							IngredientID:      &garlicPaste.ID,
							Name:              "garlic paste",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       10,
						},
					},
					Index: 1,
				},
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                fakes.BuildFakeID(),
							Name:              productName,
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "third step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							IngredientID:      &pintoBeans.ID,
							Name:              "pinto beans",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       500,
						},
						{
							IngredientID:      &water.ID,
							Name:              "water",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       5,
						},
					},
					Index: 2,
				},
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							Name:              "final output",
							MeasurementUnitID: &fakes.BuildFakeValidMeasurementUnit().ID,
							Type:              mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes:         "fourth step",
					PreparationID: soak.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							Name:                            productName,
							MeasurementUnitID:               fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:                     1000,
							ProductOfRecipeStepIndex:        new(uint64(2)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
						},
						{
							Name:              "pressure cooked beans",
							MeasurementUnitID: fakes.BuildFakeValidMeasurementUnit().ID,
							MinQuantity:       10,
						},
					},
					Index: 3,
				},
			},
		}

		ctx := t.Context()
		c := buildInertClientForTest(t)
		err := c.findCreatedRecipeStepProductsForIngredients(ctx, exampleRecipeInput)
		require.NoError(t, err)

		require.NotNil(t, exampleRecipeInput.Steps[1].Ingredients[0].RecipeStepProductID)
		assert.Equal(t, exampleRecipeInput.Steps[0].Products[0].ID, *exampleRecipeInput.Steps[1].Ingredients[0].RecipeStepProductID)
		require.NotNil(t, exampleRecipeInput.Steps[3].Ingredients[0].RecipeStepProductID)
		assert.Equal(t, exampleRecipeInput.Steps[2].Products[0].ID, *exampleRecipeInput.Steps[3].Ingredients[0].RecipeStepProductID)
	})
}

func Test_findCreatedRecipeStepProductsForInstruments(T *testing.T) {
	T.Parallel()

	T.Run("example", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		bake := fakes.BuildFakeValidPreparation()
		line := fakes.BuildFakeValidPreparation()
		bakingSheet := fakes.BuildFakeValidInstrument()
		aluminumFoil := fakes.BuildFakeValidIngredient()
		asparagus := fakes.BuildFakeValidIngredient()
		grams := fakes.BuildFakeValidMeasurementUnit()
		sheet := fakes.BuildFakeValidMeasurementUnit()

		productName := "lined baking sheet"

		exampleRecipeInput := &mealplanning.RecipeDatabaseCreationInput{
			Name:        "example",
			Description: "",
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:   fakes.BuildFakeID(),
							Name: productName,
							Type: mealplanning.RecipeStepProductInstrumentType,
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							InstrumentID:        &bakingSheet.ID,
							RecipeStepProductID: nil,
							Name:                "baking sheet",
						},
					},
					Notes:         "first step",
					PreparationID: line.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							RecipeStepProductID: nil,
							IngredientID:        &aluminumFoil.ID,
							Name:                "aluminum foil",
							MeasurementUnitID:   sheet.ID,
							MinQuantity:         1,
						},
					},
					Index: 0,
				},
				{
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:   fakes.BuildFakeID(),
							Name: "roasted asparagus",
							Type: mealplanning.RecipeStepProductInstrumentType,
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							InstrumentID:                    &bakingSheet.ID,
							RecipeStepProductID:             nil,
							Name:                            productName,
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
						},
					},
					Notes:         "second step",
					PreparationID: bake.ID,
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							RecipeStepProductID: nil,
							IngredientID:        &asparagus.ID,
							Name:                "asparagus",
							MeasurementUnitID:   grams.ID,
							MinQuantity:         1000,
						},
					},
					Index: 1,
				},
			},
		}

		c := buildInertClientForTest(t)
		c.findCreatedRecipeStepProductsForInstruments(ctx, exampleRecipeInput)

		require.NotNil(t, exampleRecipeInput.Steps[1].Instruments[0].RecipeStepProductID)
		assert.Equal(t, exampleRecipeInput.Steps[0].Products[0].ID, *exampleRecipeInput.Steps[1].Instruments[0].RecipeStepProductID)
	})
}

func TestQuerier_MarkRecipeAsIndexed(T *testing.T) {
	T.Parallel()

	T.Run("with invalid MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.MarkRecipeAsIndexed(ctx, ""))
	})
}

func TestQuerier_Integration_Recipes_CursorBasedPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[mealplanning.Recipe]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "recipe",
		CreateItem: func(ctx context.Context, i int) *mealplanning.Recipe {
			recipe := buildRecipeForTestCreation(t, ctx, user.ID, dbc)
			recipe.Name = fmt.Sprintf("Recipe %02d", i)
			return createRecipeForTest(t, ctx, recipe, dbc, false)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
			return dbc.GetRecipes(ctx, mealplanning.RecipeStatusSubmitted, filter)
		},
		GetID: func(recipe *mealplanning.Recipe) string {
			return recipe.ID
		},
		CleanupItem: func(ctx context.Context, recipe *mealplanning.Recipe) error {
			return dbc.ArchiveRecipe(ctx, recipe.ID, user.ID)
		},
	})
}

func TestQuerier_GetRecipe_AssociatedRecipes(T *testing.T) {
	T.Parallel()

	T.Run("recipe with cross-recipe reference populates AssociatedRecipes", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)
		defer func() {
		}()

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

		// Create first recipe (base recipe with a product)
		preparation1 := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnit1 := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredient1 := createValidIngredientForTest(t, ctx, nil, dbc)
		instrument1 := createValidInstrumentForTest(t, ctx, nil, dbc)

		vip1Input := fakes.BuildFakeValidIngredientPreparation()
		vip1Input.Ingredient = *ingredient1
		vip1Input.Preparation = *preparation1
		vip1 := createValidIngredientPreparationForTest(t, ctx, vip1Input, dbc)

		vimu1Input := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimu1Input.Ingredient = *ingredient1
		vimu1Input.MeasurementUnit = *measurementUnit1
		vimu1 := createValidIngredientMeasurementUnitForTest(t, ctx, vimu1Input, dbc)

		vpi1Input := fakes.BuildFakeValidPreparationInstrument()
		vpi1Input.Preparation = *preparation1
		vpi1Input.Instrument = *instrument1
		vpi1 := createValidPreparationInstrumentForTest(t, ctx, vpi1Input, dbc)

		firstRecipeInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Base Sauce Recipe",
			Slug:                 "base-sauce-recipe",
			Description:          "A base sauce recipe",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesUnspecified,
			PortionName:          "cup",
			PluralPortionName:    "cups",
			MinEstimatedPortions: 1,
			EligibleForMeals:     false,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparation1.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi1.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "base ingredient",
							ValidIngredientPreparationID:     &vip1.ID,
							ValidIngredientMeasurementUnitID: &vimu1.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "base sauce",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit1.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparation1.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi1.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final sauce",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit1.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		firstRecipe, err := dbc.CreateRecipe(ctx, firstRecipeInput)
		require.NoError(t, err)
		require.NotNil(t, firstRecipe)

		// Create second recipe that references the first recipe's product
		preparation2 := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnit2 := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredient2a := createValidIngredientForTest(t, ctx, nil, dbc) // Different ingredient for first ingredient in step
		instrument2 := createValidInstrumentForTest(t, ctx, nil, dbc)

		vip2Input := fakes.BuildFakeValidIngredientPreparation()
		vip2Input.Ingredient = *ingredient2a
		vip2Input.Preparation = *preparation2
		vip2 := createValidIngredientPreparationForTest(t, ctx, vip2Input, dbc)

		vimu2Input := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimu2Input.Ingredient = *ingredient2a
		vimu2Input.MeasurementUnit = *measurementUnit2
		vimu2 := createValidIngredientMeasurementUnitForTest(t, ctx, vimu2Input, dbc)

		vpi2Input := fakes.BuildFakeValidPreparationInstrument()
		vpi2Input.Preparation = *preparation2
		vpi2Input.Instrument = *instrument2
		vpi2 := createValidPreparationInstrumentForTest(t, ctx, vpi2Input, dbc)

		secondRecipeInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe Using Base Sauce",
			Slug:                 "recipe-using-base-sauce",
			Description:          "A recipe that uses the base sauce",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			EligibleForMeals:     true,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi2.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "other ingredient",
							IngredientID:                     &ingredient2a.ID,
							ValidIngredientPreparationID:     &vip2.ID,
							ValidIngredientMeasurementUnitID: &vimu2.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
						{
							ID: identifiers.New(),
							// Reference the product from the first recipe
							// The product "base sauce" is from step 0 (index 0), product index 0
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							RecipeStepProductRecipeID:       &firstRecipe.ID,
							Name:                            "base sauce",
							IngredientID:                    nil,                 // No ingredient ID when referencing a product from another recipe
							MeasurementUnitID:               measurementUnit2.ID, // Use MeasurementUnitID directly, not ValidIngredientMeasurementUnitID
							MinQuantity:                     0.5,
							Index:                           1,
							OptionIndex:                     0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final dish",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit2.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparation2.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi2.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit2.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		secondRecipe, err := dbc.CreateRecipe(ctx, secondRecipeInput)
		require.NoError(t, err)
		require.NotNil(t, secondRecipe)

		// Retrieve the second recipe and validate AssociatedRecipes
		retrievedSecond, err := dbc.GetRecipe(ctx, secondRecipe.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedSecond)

		// Validate that the first recipe is in the second's AssociatedRecipes
		require.Len(t, retrievedSecond.AssociatedRecipes, 1, "second recipe should have one associated recipe")
		assert.Equal(t, firstRecipe.ID, retrievedSecond.AssociatedRecipes[0].ID, "associated recipe should be the first recipe")
		assert.Equal(t, firstRecipe.Name, retrievedSecond.AssociatedRecipes[0].Name, "associated recipe should have the correct name")

		// Retrieve the first recipe and validate it has empty AssociatedRecipes
		retrievedFirst, err := dbc.GetRecipe(ctx, firstRecipe.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedFirst)

		assert.Empty(t, retrievedFirst.AssociatedRecipes, "first recipe should have no associated recipes")

		// Cleanup
		assert.NoError(t, dbc.ArchiveRecipe(ctx, secondRecipe.ID, user.ID))
		assert.NoError(t, dbc.ArchiveRecipe(ctx, firstRecipe.ID, user.ID))
	})

	T.Run("cyclical recipe relationship terminates without infinite recursion", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)
		defer func() {
		}()

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

		// Create preparation, measurement unit, ingredient, and instrument for both recipes
		preparation := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnit := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredient := createValidIngredientForTest(t, ctx, nil, dbc)
		instrument := createValidInstrumentForTest(t, ctx, nil, dbc)

		vipInput := fakes.BuildFakeValidIngredientPreparation()
		vipInput.Ingredient = *ingredient
		vipInput.Preparation = *preparation
		vip := createValidIngredientPreparationForTest(t, ctx, vipInput, dbc)

		vimuInput := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimuInput.Ingredient = *ingredient
		vimuInput.MeasurementUnit = *measurementUnit
		vimu := createValidIngredientMeasurementUnitForTest(t, ctx, vimuInput, dbc)

		vpiInput := fakes.BuildFakeValidPreparationInstrument()
		vpiInput.Preparation = *preparation
		vpiInput.Instrument = *instrument
		vpi := createValidPreparationInstrumentForTest(t, ctx, vpiInput, dbc)

		// Create Recipe A
		recipeAInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe A",
			Slug:                 "recipe-a",
			Description:          "Recipe A",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 2,
			EligibleForMeals:     true,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "ingredient",
							IngredientID:                     &ingredient.ID,
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "product A",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		recipeA, err := dbc.CreateRecipe(ctx, recipeAInput)
		require.NoError(t, err)
		require.NotNil(t, recipeA)

		// Create Recipe B that references Recipe A
		recipeBInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe B",
			Slug:                 "recipe-b",
			Description:          "Recipe B",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 2,
			EligibleForMeals:     true,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "ingredient",
							IngredientID:                     &ingredient.ID,
							ValidIngredientPreparationID:     &vip.ID,
							ValidIngredientMeasurementUnitID: &vimu.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
						{
							ID: identifiers.New(),
							// Reference Recipe A's product
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							RecipeStepProductRecipeID:       &recipeA.ID,
							Name:                            "product A",
							IngredientID:                    nil,                // No ingredient ID when referencing a product from another recipe
							MeasurementUnitID:               measurementUnit.ID, // Use MeasurementUnitID directly, not ValidIngredientMeasurementUnitID
							MinQuantity:                     0.5,
							Index:                           1,
							OptionIndex:                     0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "product B",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparation.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpi.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnit.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		recipeB, err := dbc.CreateRecipe(ctx, recipeBInput)
		require.NoError(t, err)
		require.NotNil(t, recipeB)

		// Now manually create a cycle by directly updating Recipe A to reference Recipe B
		// This simulates the edge case where a cycle exists (even though validation should prevent it)
		// We'll use raw SQL to bypass validation for testing purposes
		_, err = dbc.writeDB.ExecContext(ctx, `
			UPDATE recipe_step_ingredients 
			SET recipe_step_product_recipe_id = $1
			WHERE belongs_to_recipe_step IN (
				SELECT id FROM recipe_steps WHERE belongs_to_recipe = $2
			)
			AND id = (
				SELECT id FROM recipe_step_ingredients 
				WHERE belongs_to_recipe_step IN (
					SELECT id FROM recipe_steps WHERE belongs_to_recipe = $2
				)
				LIMIT 1
			)
		`, recipeB.ID, recipeA.ID)
		require.NoError(t, err)

		// Now try to retrieve Recipe A - this should terminate without infinite recursion
		// The visited set should prevent infinite loops
		retrievedA, err := dbc.GetRecipe(ctx, recipeA.ID)
		require.NoError(t, err, "GetRecipe should terminate even with cyclical dependency")
		require.NotNil(t, retrievedA)

		// Verify that Recipe A was retrieved (even if AssociatedRecipes might be incomplete due to cycle)
		assert.Equal(t, recipeA.ID, retrievedA.ID)
		assert.Equal(t, recipeA.Name, retrievedA.Name)

		// Verify that Recipe B can also be retrieved
		retrievedB, err := dbc.GetRecipe(ctx, recipeB.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedB)
		assert.Equal(t, recipeB.ID, retrievedB.ID)

		// Cleanup
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipeB.ID, user.ID))
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipeA.ID, user.ID))
	})

	T.Run("nested associated recipes are flattened", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		dbc, _ := buildDatabaseClientForTest(t)
		defer func() {
		}()

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

		// Create Recipe C (base recipe, no dependencies)
		preparationC := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnitC := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredientC := createValidIngredientForTest(t, ctx, nil, dbc)
		instrumentC := createValidInstrumentForTest(t, ctx, nil, dbc)

		vipCInput := fakes.BuildFakeValidIngredientPreparation()
		vipCInput.Ingredient = *ingredientC
		vipCInput.Preparation = *preparationC
		vipC := createValidIngredientPreparationForTest(t, ctx, vipCInput, dbc)

		vimuCInput := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimuCInput.Ingredient = *ingredientC
		vimuCInput.MeasurementUnit = *measurementUnitC
		vimuC := createValidIngredientMeasurementUnitForTest(t, ctx, vimuCInput, dbc)

		vpiCInput := fakes.BuildFakeValidPreparationInstrument()
		vpiCInput.Preparation = *preparationC
		vpiCInput.Instrument = *instrumentC
		vpiC := createValidPreparationInstrumentForTest(t, ctx, vpiCInput, dbc)

		recipeCInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe C",
			Slug:                 "recipe-c",
			Description:          "Base recipe with no dependencies",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesUnspecified,
			PortionName:          "cup",
			PluralPortionName:    "cups",
			MinEstimatedPortions: 1,
			EligibleForMeals:     false,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparationC.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiC.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "ingredient C",
							ValidIngredientPreparationID:     &vipC.ID,
							ValidIngredientMeasurementUnitID: &vimuC.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "product C",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitC.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparationC.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiC.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final product C",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitC.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		recipeC, err := dbc.CreateRecipe(ctx, recipeCInput)
		require.NoError(t, err)
		require.NotNil(t, recipeC)

		// Create Recipe B (references Recipe C)
		preparationB := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnitB := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredientB := createValidIngredientForTest(t, ctx, nil, dbc)
		instrumentB := createValidInstrumentForTest(t, ctx, nil, dbc)

		vipBInput := fakes.BuildFakeValidIngredientPreparation()
		vipBInput.Ingredient = *ingredientB
		vipBInput.Preparation = *preparationB
		vipB := createValidIngredientPreparationForTest(t, ctx, vipBInput, dbc)

		vimuBInput := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimuBInput.Ingredient = *ingredientB
		vimuBInput.MeasurementUnit = *measurementUnitB
		vimuB := createValidIngredientMeasurementUnitForTest(t, ctx, vimuBInput, dbc)

		vpiBInput := fakes.BuildFakeValidPreparationInstrument()
		vpiBInput.Preparation = *preparationB
		vpiBInput.Instrument = *instrumentB
		vpiB := createValidPreparationInstrumentForTest(t, ctx, vpiBInput, dbc)

		recipeBInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe B",
			Slug:                 "recipe-b",
			Description:          "Recipe that uses Recipe C",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesUnspecified,
			PortionName:          "cup",
			PluralPortionName:    "cups",
			MinEstimatedPortions: 1,
			EligibleForMeals:     false,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparationB.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiB.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "ingredient B",
							ValidIngredientPreparationID:     &vipB.ID,
							ValidIngredientMeasurementUnitID: &vimuB.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
						{
							ID: identifiers.New(),
							// Reference the product from Recipe C
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							RecipeStepProductRecipeID:       &recipeC.ID,
							Name:                            "product C",
							IngredientID:                    nil,
							MeasurementUnitID:               measurementUnitB.ID,
							MinQuantity:                     0.5,
							Index:                           1,
							OptionIndex:                     0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "product B",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitB.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparationB.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiB.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final product B",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitB.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		recipeB, err := dbc.CreateRecipe(ctx, recipeBInput)
		require.NoError(t, err)
		require.NotNil(t, recipeB)

		// Create Recipe A (references Recipe B)
		preparationA := createValidPreparationForTest(t, ctx, nil, dbc)
		measurementUnitA := createValidMeasurementUnitForTest(t, ctx, nil, dbc)
		ingredientA := createValidIngredientForTest(t, ctx, nil, dbc)
		instrumentA := createValidInstrumentForTest(t, ctx, nil, dbc)

		vipAInput := fakes.BuildFakeValidIngredientPreparation()
		vipAInput.Ingredient = *ingredientA
		vipAInput.Preparation = *preparationA
		vipA := createValidIngredientPreparationForTest(t, ctx, vipAInput, dbc)

		vimuAInput := fakes.BuildFakeValidIngredientMeasurementUnit()
		vimuAInput.Ingredient = *ingredientA
		vimuAInput.MeasurementUnit = *measurementUnitA
		vimuA := createValidIngredientMeasurementUnitForTest(t, ctx, vimuAInput, dbc)

		vpiAInput := fakes.BuildFakeValidPreparationInstrument()
		vpiAInput.Preparation = *preparationA
		vpiAInput.Instrument = *instrumentA
		vpiA := createValidPreparationInstrumentForTest(t, ctx, vpiAInput, dbc)

		recipeAInput := &mealplanning.RecipeDatabaseCreationInput{
			ID:                   identifiers.New(),
			Name:                 "Recipe A",
			Slug:                 "recipe-a",
			Description:          "Recipe that uses Recipe B (which uses Recipe C)",
			CreatedByUser:        user.ID,
			YieldsComponentType:  mealplanning.MealComponentTypesMain,
			PortionName:          "serving",
			PluralPortionName:    "servings",
			MinEstimatedPortions: 4,
			EligibleForMeals:     true,
			Steps: []*mealplanning.RecipeStepDatabaseCreationInput{
				{
					ID:            identifiers.New(),
					PreparationID: preparationA.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiA.ID,
						},
					},
					Ingredients: []*mealplanning.RecipeStepIngredientDatabaseCreationInput{
						{
							ID:                               identifiers.New(),
							Name:                             "ingredient A",
							ValidIngredientPreparationID:     &vipA.ID,
							ValidIngredientMeasurementUnitID: &vimuA.ID,
							MinQuantity:                      1,
							Index:                            0,
							OptionIndex:                      0,
						},
						{
							ID: identifiers.New(),
							// Reference the product from Recipe B
							ProductOfRecipeStepIndex:        new(uint64(0)),
							ProductOfRecipeStepProductIndex: new(uint64(0)),
							RecipeStepProductRecipeID:       &recipeB.ID,
							Name:                            "product B",
							IngredientID:                    nil,
							MeasurementUnitID:               measurementUnitA.ID,
							MinQuantity:                     0.5,
							Index:                           1,
							OptionIndex:                     0,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final dish",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitA.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 0,
				},
				{
					ID:            identifiers.New(),
					PreparationID: preparationA.ID,
					Instruments: []*mealplanning.RecipeStepInstrumentDatabaseCreationInput{
						{
							ID:                           identifiers.New(),
							Name:                         "spoon",
							ValidPreparationInstrumentID: &vpiA.ID,
						},
					},
					Products: []*mealplanning.RecipeStepProductDatabaseCreationInput{
						{
							ID:                     identifiers.New(),
							Name:                   "final output",
							Type:                   mealplanning.RecipeStepProductIngredientType,
							MeasurementUnitID:      &measurementUnitA.ID,
							MinMeasurementQuantity: new(float32(1)),
						},
					},
					Index: 1,
				},
			},
		}

		recipeA, err := dbc.CreateRecipe(ctx, recipeAInput)
		require.NoError(t, err)
		require.NotNil(t, recipeA)

		// Retrieve Recipe A and validate flattened AssociatedRecipes
		retrievedA, err := dbc.GetRecipe(ctx, recipeA.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedA)

		// Recipe A should have both B and C in its AssociatedRecipes (flattened)
		require.Len(t, retrievedA.AssociatedRecipes, 2, "Recipe A should have both B and C in its AssociatedRecipes")

		// Find B and C in the AssociatedRecipes
		var foundB, foundC bool
		for _, associated := range retrievedA.AssociatedRecipes {
			if associated.ID == recipeB.ID {
				foundB = true
				assert.Equal(t, recipeB.Name, associated.Name, "Recipe B should have correct name")
				// Recipe B should have empty AssociatedRecipes (C should be flattened out)
				assert.Empty(t, associated.AssociatedRecipes, "Recipe B should have empty AssociatedRecipes after flattening")
			}
			if associated.ID == recipeC.ID {
				foundC = true
				assert.Equal(t, recipeC.Name, associated.Name, "Recipe C should have correct name")
				// Recipe C should have empty AssociatedRecipes
				assert.Empty(t, associated.AssociatedRecipes, "Recipe C should have empty AssociatedRecipes")
			}
		}
		assert.True(t, foundB, "Recipe B should be in Recipe A's AssociatedRecipes")
		assert.True(t, foundC, "Recipe C should be in Recipe A's AssociatedRecipes (flattened)")

		// Retrieve Recipe B directly and validate it has C in its AssociatedRecipes
		// (when fetched directly, it should not be flattened)
		retrievedB, err := dbc.GetRecipe(ctx, recipeB.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedB)
		require.Len(t, retrievedB.AssociatedRecipes, 1, "Recipe B should have C in its AssociatedRecipes when fetched directly")
		assert.Equal(t, recipeC.ID, retrievedB.AssociatedRecipes[0].ID, "Recipe B should reference Recipe C when fetched directly")

		// Retrieve Recipe C directly and validate it has empty AssociatedRecipes
		retrievedC, err := dbc.GetRecipe(ctx, recipeC.ID)
		require.NoError(t, err)
		require.NotNil(t, retrievedC)
		assert.Empty(t, retrievedC.AssociatedRecipes, "Recipe C should have no associated recipes")

		// Cleanup
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipeA.ID, user.ID))
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipeB.ID, user.ID))
		assert.NoError(t, dbc.ArchiveRecipe(ctx, recipeC.ID, user.ID))
	})
}
