package recipeanalysis

import (
	"testing"
	"time"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"

	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	"github.com/primandproper/platform-go/v5/observability/tracing"

	"github.com/stretchr/testify/assert"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

func newAnalyzerForTest(t *testing.T) *recipeAnalyzer {
	t.Helper()

	return &recipeAnalyzer{
		tracer: tracing.NewTracerForTest(t.Name()),
		logger: loggingnoop.NewLogger(),
	}
}

func TestRecipeGrapher_makeGraphForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)

		ctx := t.Context()
		r := &mealplanning.Recipe{
			Steps: []*mealplanning.RecipeStep{
				{},
			},
		}

		actual, err := g.MakeGraphForRecipe(ctx, r)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestRecipeGrapher_makeDAGForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)

		ctx := t.Context()
		r := &mealplanning.Recipe{
			Steps: []*mealplanning.RecipeStep{
				{},
			},
		}

		actual, err := g.makeDAGForRecipe(ctx, r)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
	})
}

func TestRecipeAnalyzer_GenerateMealPlanTasksForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("creates thaw tasks for frozen ingredients", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		exampleMeal := fakes.BuildFakeMeal()
		exampleMealPlan := fakes.BuildFakeMealPlan()

		exampleMealPlanEvent := fakes.BuildFakeMealPlanEvent()
		exampleMealPlanEvent.BelongsToMealPlan = exampleMealPlan.ID
		now := time.Now().Add(0).Truncate(time.Second).UTC()
		inThreeDays := now.Add((time.Hour * 24) * 3).Add(0).Truncate(time.Second).UTC()
		inOneWeek := now.Add((time.Hour * 24) * 7).Add(0).Truncate(time.Second).UTC()
		exampleMealPlanEvent.StartsAt = inThreeDays
		exampleMealPlanEvent.EndsAt = inOneWeek

		exampleMealPlanOption := fakes.BuildFakeMealPlanOption()
		exampleMealPlanOption.BelongsToMealPlanEvent = exampleMealPlanEvent.ID
		exampleMealPlanOption.Meal = *exampleMeal

		recipeStepID := fakes.BuildFakeID()

		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipe := &mealplanning.Recipe{
			Name: "Recipe 1",
			ID:   exampleRecipeID,
			Steps: []*mealplanning.RecipeStep{
				{
					BelongsToRecipe: exampleRecipeID,
					ID:              recipeStepID,
					Preparation:     mealplanning.ValidPreparation{Name: "dice"},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: &mealplanning.ValidIngredient{
								// -5°C is below the corrected frozen threshold (<= 0°C), so this ingredient
								// genuinely qualifies as frozen.
								MinStorageTemperatureInCelsius: new(float32(-5)),
								PluralName:                     "chicken breasts",
								StorageInstructions:            "keep frozen",
								Name:                           "chicken breast",
								ID:                             fakes.BuildFakeID(),
							},
							Name:                "chicken breast",
							ID:                  fakes.BuildFakeID(),
							BelongsToRecipeStep: recipeStepID,
							MeasurementUnit:     mealplanning.ValidMeasurementUnit{Name: "gram", PluralName: "grams"},
							MinQuantity:         900,

							MaxQuantity: new(float32(900)),
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                "diced chicken breast",
							Type:                mealplanning.RecipeStepProductIngredientType,
							BelongsToRecipeStep: recipeStepID,
							ID:                  fakes.BuildFakeID(),
							MeasurementUnit:     &mealplanning.ValidMeasurementUnit{},
						},
					},
				},
			},
		}

		actual, err := g.GenerateMealPlanTasksForRecipe(ctx, exampleMealPlanOption.ID, exampleRecipe)
		assert.NoError(t, err)

		// The frozen chicken breast in step 1 yields exactly one ad-hoc thaw task. Thaw tasks have
		// no backing recipe prep task (belongs_to_recipe_prep_task is nullable as of migration 00023).
		assert.Len(t, actual, 1)
		assert.Equal(t, exampleMealPlanOption.ID, actual[0].MealPlanOptionID)
		assert.Empty(t, actual[0].RecipePrepTaskID)
		assert.Equal(t, buildThawStepCreationExplanation(1, 0), actual[0].CreationExplanation)
		assert.NotEmpty(t, actual[0].ID)
	})

	T.Run("does not create thaw tasks for refrigerated ingredients", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		exampleMealPlanOptionID := fakes.BuildFakeID()
		recipeStepID := fakes.BuildFakeID()
		exampleRecipeID := fakes.BuildFakeID()
		exampleRecipe := &mealplanning.Recipe{
			Name: "Recipe 1",
			ID:   exampleRecipeID,
			Steps: []*mealplanning.RecipeStep{
				{
					BelongsToRecipe: exampleRecipeID,
					ID:              recipeStepID,
					Preparation:     mealplanning.ValidPreparation{Name: "dice"},
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: &mealplanning.ValidIngredient{
								// 2.5°C is fridge temperature, above the frozen threshold (<= 0°C).
								MinStorageTemperatureInCelsius: new(float32(2.5)),
								PluralName:                     "chicken breasts",
								StorageInstructions:            "keep refrigerated",
								Name:                           "chicken breast",
								ID:                             fakes.BuildFakeID(),
							},
							Name:                "chicken breast",
							ID:                  fakes.BuildFakeID(),
							BelongsToRecipeStep: recipeStepID,
							MeasurementUnit:     mealplanning.ValidMeasurementUnit{Name: "gram", PluralName: "grams"},
							MinQuantity:         900,
							MaxQuantity:         new(float32(900)),
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							Name:                "diced chicken breast",
							Type:                mealplanning.RecipeStepProductIngredientType,
							BelongsToRecipeStep: recipeStepID,
							ID:                  fakes.BuildFakeID(),
							MeasurementUnit:     &mealplanning.ValidMeasurementUnit{},
						},
					},
				},
			},
		}

		actual, err := g.GenerateMealPlanTasksForRecipe(ctx, exampleMealPlanOptionID, exampleRecipe)
		assert.NoError(t, err)
		assert.Empty(t, actual)
	})
}

func Test_recipeAnalyzer_RenderMermaidDiagramForRecipe(T *testing.T) {
	T.Parallel()

	T.Run("basic", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		g := newAnalyzerForTest(t)

		dice := fakes.BuildFakeValidPreparation()
		dice.Name = "dice"
		sautee := fakes.BuildFakeValidPreparation()
		sautee.Name = "sautee"

		step1ID := fakes.BuildFakeID()
		step2ID := fakes.BuildFakeID()
		step3ID := fakes.BuildFakeID()
		step4ID := fakes.BuildFakeID()
		dicedOnionRecipeStepProductID := fakes.BuildFakeID()
		dicedCarrotRecipeStepProductID := fakes.BuildFakeID()
		dicedCeleryRecipeStepProductID := fakes.BuildFakeID()

		recipe := &mealplanning.Recipe{
			Name: "example recipe",
			Steps: []*mealplanning.RecipeStep{
				{
					ID:          step1ID,
					Preparation: *dice,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: fakes.BuildFakeValidIngredient(),
							Name:       "onion",
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							ID:   dicedOnionRecipeStepProductID,
							Name: "diced onion",
							Type: mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes: "first step",
					Index: 0,
				},
				{
					ID:          step2ID,
					Preparation: *dice,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: fakes.BuildFakeValidIngredient(),
							Name:       "carrot",
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							ID:   dicedCarrotRecipeStepProductID,
							Name: "diced carrot",
							Type: mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes: "second step",
					Index: 1,
				},
				{
					ID:          step3ID,
					Preparation: *dice,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Ingredient: fakes.BuildFakeValidIngredient(),
							Name:       "celery",
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							ID:   dicedCeleryRecipeStepProductID,
							Name: "diced celery",
							Type: mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes: "third step",
					Index: 2,
				},
				{
					ID:          step4ID,
					Preparation: *sautee,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Name:                "diced onion",
							RecipeStepProductID: new(dicedOnionRecipeStepProductID),
						},
						{
							Name:                "diced carrot",
							RecipeStepProductID: new(dicedCarrotRecipeStepProductID),
						},
						{
							Name:                "diced celery",
							RecipeStepProductID: new(dicedCeleryRecipeStepProductID),
						},
					},
					Products: []*mealplanning.RecipeStepProduct{
						{
							ID:   dicedOnionRecipeStepProductID,
							Name: "sauteed mire poix",
							Type: mealplanning.RecipeStepProductIngredientType,
						},
					},
					Notes: "fourth step",
					Index: 3,
				},
			},
		}

		expected := `flowchart TD;
	Step1["Step #1 (dice)"];
	Step2["Step #2 (dice)"];
	Step3["Step #3 (dice)"];
	Step4["Step #4 (sautee)"];
	Step1 -->|ingredient| Step4;
	Step2 -->|ingredient| Step4;
	Step3 -->|ingredient| Step4;
`
		actual := g.RenderMermaidDiagramForRecipe(ctx, recipe)

		assert.Equal(t, expected, actual)
	})

	T.Run("with associated recipe", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		g := newAnalyzerForTest(t)

		dice := fakes.BuildFakeValidPreparation()
		dice.Name = "dice"
		top := fakes.BuildFakeValidPreparation()
		top.Name = "top"

		breadcrumbProductID := fakes.BuildFakeID()
		assocStepID := fakes.BuildFakeID()
		assocRecipe := &mealplanning.Recipe{
			ID:   "breadcrumbs-recipe-id",
			Name: "Caesar Breadcrumbs",
			Steps: []*mealplanning.RecipeStep{
				{
					ID:          assocStepID,
					Preparation: *dice,
					Products: []*mealplanning.RecipeStepProduct{
						{ID: breadcrumbProductID, Name: "caesar breadcrumbs", Type: mealplanning.RecipeStepProductIngredientType},
					},
					Index: 0,
				},
			},
		}

		mainStepID := fakes.BuildFakeID()
		recipe := &mealplanning.Recipe{
			Name: "Caesar Roasted Broccoli",
			Steps: []*mealplanning.RecipeStep{
				{
					ID:          mainStepID,
					Preparation: *top,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							Name:                      "caesar breadcrumbs",
							RecipeStepProductID:       &breadcrumbProductID,
							RecipeStepProductRecipeID: func() *string { s := assocRecipe.ID; return &s }(),
						},
					},
					Index: 0,
				},
			},
			AssociatedRecipes: []*mealplanning.Recipe{assocRecipe},
		}

		actual := g.RenderMermaidDiagramForRecipe(ctx, recipe)

		assert.Contains(t, actual, "Caesar Breadcrumbs")
		assert.Contains(t, actual, "Step1")
		assert.Contains(t, actual, "Step11001") // first associated recipe's first step
		assert.Contains(t, actual, "ingredient")
	})
}

func TestRecipeAnalyzer_MakeGraphForRecipe_WithAssociatedRecipe(T *testing.T) {
	T.Parallel()

	T.Run("builds graph with associated recipe steps", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		breadcrumbProductID := fakes.BuildFakeID()
		assocRecipe := &mealplanning.Recipe{
			ID:   "breadcrumbs-recipe-id",
			Name: "Caesar Breadcrumbs",
			Steps: []*mealplanning.RecipeStep{
				{
					ID:    fakes.BuildFakeID(),
					Index: 0,
					Products: []*mealplanning.RecipeStepProduct{
						{ID: breadcrumbProductID, Type: mealplanning.RecipeStepProductIngredientType},
					},
				},
			},
		}

		recipe := &mealplanning.Recipe{
			Steps: []*mealplanning.RecipeStep{
				{ID: fakes.BuildFakeID(), Index: 0},
				{
					ID:    fakes.BuildFakeID(),
					Index: 1,
					Ingredients: []*mealplanning.RecipeStepIngredient{
						{
							RecipeStepProductID:       &breadcrumbProductID,
							RecipeStepProductRecipeID: func() *string { s := assocRecipe.ID; return &s }(),
						},
					},
				},
			},
			AssociatedRecipes: []*mealplanning.Recipe{assocRecipe},
		}

		actual, err := g.MakeGraphForRecipe(ctx, recipe)
		assert.NoError(t, err)
		assert.NotNil(t, actual)

		// Should have 3 nodes: 1 from assoc + 2 from main
		assert.Equal(t, 3, actual.Nodes().Len())
	})
}

func TestRecipeAnalyzer_MakeGraphForMeal(T *testing.T) {
	T.Parallel()

	T.Run("combines two recipe graphs", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		dice := fakes.BuildFakeValidPreparation()
		dice.Name = "dice"
		mainRecipe := &mealplanning.Recipe{
			ID:   "main-id",
			Name: "Roasted Chicken",
			Steps: []*mealplanning.RecipeStep{
				{ID: fakes.BuildFakeID(), Preparation: *dice, Index: 0},
				{ID: fakes.BuildFakeID(), Preparation: *dice, Index: 1},
			},
		}
		sideRecipe := &mealplanning.Recipe{
			ID:   "side-id",
			Name: "Caesar Broccoli",
			Steps: []*mealplanning.RecipeStep{
				{ID: fakes.BuildFakeID(), Preparation: *dice, Index: 0},
			},
		}

		meal := &mealplanning.Meal{
			Name: "Chicken Dinner",
			Components: []*mealplanning.MealComponent{
				{ComponentType: mealplanning.MealComponentTypesMain, Recipe: *mainRecipe},
				{ComponentType: mealplanning.MealComponentTypesSide, Recipe: *sideRecipe},
			},
		}

		actual, err := g.MakeGraphForMeal(ctx, meal)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		// 2 steps from main + 1 from side = 3 nodes
		assert.Equal(t, 3, actual.Nodes().Len())
	})
}

func TestRecipeAnalyzer_RenderMermaidDiagramForMeal(T *testing.T) {
	T.Parallel()

	T.Run("renders meal with multiple components", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		dice := fakes.BuildFakeValidPreparation()
		dice.Name = "dice"
		mainRecipe := &mealplanning.Recipe{
			ID:   "main-id",
			Name: "Roasted Chicken",
			Steps: []*mealplanning.RecipeStep{
				{ID: fakes.BuildFakeID(), Preparation: *dice, Index: 0},
			},
		}
		sideRecipe := &mealplanning.Recipe{
			ID:   "side-id",
			Name: "Caesar Broccoli",
			Steps: []*mealplanning.RecipeStep{
				{ID: fakes.BuildFakeID(), Preparation: *dice, Index: 0},
			},
		}

		meal := &mealplanning.Meal{
			Name: "Chicken Dinner",
			Components: []*mealplanning.MealComponent{
				{ComponentType: mealplanning.MealComponentTypesMain, Recipe: *mainRecipe},
				{ComponentType: mealplanning.MealComponentTypesSide, Recipe: *sideRecipe},
			},
		}

		actual := g.RenderMermaidDiagramForMeal(ctx, meal)

		assert.Contains(t, actual, "Roasted Chicken")
		assert.Contains(t, actual, "Caesar Broccoli")
		assert.Contains(t, actual, "main:")
		assert.Contains(t, actual, "side:")
		assert.Contains(t, actual, "Step1")
		assert.Contains(t, actual, "Step100001") // second component's first step
	})
}

func TestRecipeAnalyzer_ValidateRecipeCreationRequestInputIsDAG(T *testing.T) {
	T.Parallel()

	T.Run("valid DAG with no dependencies", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Simple Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{Index: 0},
				{Index: 1},
				{Index: 2},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("valid DAG with linear dependencies via ingredients", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Linear Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{Name: "raw ingredient", MinQuantity: 1},
					},
				},
				{
					Index: 1,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "product from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("valid DAG with multiple dependencies", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)
		step2Index := uint64(2)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Multi-dependency Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{Index: 0},
				{Index: 1},
				{Index: 2},
				{
					Index: 3,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
						{
							Name:                     "from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                     "instrument from step 2",
							ProductOfRecipeStepIndex: &step2Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("valid DAG with vessel dependencies", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Vessel Dependency Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{Index: 0},
				{
					Index: 1,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "vessel from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("invalid DAG with direct cycle via ingredients", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Cyclic Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 1,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNotAcyclic)
	})

	T.Run("invalid DAG with three-step cycle", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)
		step2Index := uint64(2)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Three-Step Cycle",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 2",
							ProductOfRecipeStepIndex: &step2Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 1,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 2,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNotAcyclic)
	})

	T.Run("invalid DAG with cycle via instruments", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Instrument Cycle",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                     "instrument from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 1,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                     "instrument from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNotAcyclic)
	})

	T.Run("invalid DAG with cycle via vessels", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Vessel Cycle",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "vessel from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 1,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "vessel from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNotAcyclic)
	})

	T.Run("invalid step index reference in ingredient", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		invalidIndex := uint64(99)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Invalid Index Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "invalid reference",
							ProductOfRecipeStepIndex: &invalidIndex,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "references invalid step array index")
	})

	T.Run("invalid step index reference in instrument", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		invalidIndex := uint64(50)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Invalid Instrument Index",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                     "invalid instrument",
							ProductOfRecipeStepIndex: &invalidIndex,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "references invalid step array index")
	})

	T.Run("invalid step index reference in vessel", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		invalidIndex := uint64(10)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Invalid Vessel Index",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Vessels: []*mealplanning.RecipeStepVesselCreationRequestInput{
						{
							Name:                     "invalid vessel",
							ProductOfRecipeStepIndex: &invalidIndex,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "references invalid step array index")
	})

	T.Run("empty steps", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		input := &mealplanning.RecipeCreationRequestInput{
			Name:  "Empty Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("single step with no dependencies", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Single Step Recipe",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{Index: 0},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})

	T.Run("valid DAG with self-reference prevention", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Self-Reference Test",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{
					Index: 0,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "self-reference",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.Error(t, err)
		assert.ErrorIs(t, err, errNotAcyclic)
	})

	T.Run("valid DAG with complex branching", func(t *testing.T) {
		t.Parallel()

		g := newAnalyzerForTest(t)
		ctx := t.Context()

		step0Index := uint64(0)
		step1Index := uint64(1)
		step2Index := uint64(2)

		input := &mealplanning.RecipeCreationRequestInput{
			Name: "Complex Branching",
			Steps: []*mealplanning.RecipeStepCreationRequestInput{
				{Index: 0},
				{Index: 1},
				{Index: 2},
				{
					Index: 3,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 0",
							ProductOfRecipeStepIndex: &step0Index,
							MinQuantity:              1,
						},
						{
							Name:                     "from step 1",
							ProductOfRecipeStepIndex: &step1Index,
							MinQuantity:              1,
						},
					},
				},
				{
					Index: 4,
					Ingredients: []*mealplanning.RecipeStepIngredientCreationRequestInput{
						{
							Name:                     "from step 2",
							ProductOfRecipeStepIndex: &step2Index,
							MinQuantity:              1,
						},
					},
					Instruments: []*mealplanning.RecipeStepInstrumentCreationRequestInput{
						{
							Name:                     "from step 3",
							ProductOfRecipeStepIndex: new(uint64(3)),
							MinQuantity:              1,
						},
					},
				},
			},
		}

		err := g.ValidateRecipeCreationRequestInputIsDAG(ctx, input)
		assert.NoError(t, err)
	})
}

func TestFindUnitigsByLength(T *testing.T) {
	T.Parallel()

	T.Run("example graph", func(t *testing.T) {
		t.Parallel()

		g := simple.NewDirectedGraph()

		// Helper to add edges using character aliases
		addEdge := func(u, v rune) {
			uid, vid := int64(u), int64(v)
			if g.Node(uid) == nil {
				g.AddNode(simple.Node(uid))
			}
			if g.Node(vid) == nil {
				g.AddNode(simple.Node(vid))
			}
			g.SetEdge(g.NewEdge(g.Node(uid), g.Node(vid)))
		}

		// Reconstruct the Mermaid graph
		// Path 1
		addEdge('A', 'B')
		addEdge('B', 'C')
		addEdge('C', 'G')
		// Path 2
		addEdge('D', 'E')
		addEdge('E', 'F')
		addEdge('F', 'G')
		// Central Chain
		addEdge('G', 'H')
		addEdge('H', 'I')
		addEdge('I', 'J')
		addEdge('J', 'K')
		// Split 1
		addEdge('K', 'L')
		addEdge('L', 'M')
		addEdge('M', 'N')
		addEdge('N', 'R')
		// Split 2
		addEdge('K', 'O')
		addEdge('O', 'P')
		addEdge('P', 'Q')
		addEdge('Q', 'R')

		res := FindUnitigsByLength(g)

		// Length 3: [A-B-C-G], [D-E-F-G] (2 chains)
		// Length 4: [G-H-I-J-K], [K-L-M-N-R], [K-O-P-Q-R] (3 chains)
		expected := map[int][][]graph.Node{
			5: {
				{
					g.Node(int64('G')),
					g.Node(int64('H')),
					g.Node(int64('I')),
					g.Node(int64('J')),
					g.Node(int64('K')),
				},
			},
			4: {
				{
					g.Node(int64('A')),
					g.Node(int64('B')),
					g.Node(int64('C')),
					g.Node(int64('G')),
				},
				{
					g.Node(int64('D')),
					g.Node(int64('E')),
					g.Node(int64('F')),
					g.Node(int64('G')),
				},
			},
		}

		assert.Equal(t, expected, res)
	})
}
