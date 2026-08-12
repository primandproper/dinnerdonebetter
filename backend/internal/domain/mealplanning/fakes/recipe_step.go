package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v10/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeRecipeStep builds a faked recipe step.
func BuildFakeRecipeStep() *types.RecipeStep {
	recipeStepID := BuildFakeID()

	var ingredients []*types.RecipeStepIngredient
	for range exampleQuantity {
		ing := BuildFakeRecipeStepIngredient()
		ing.BelongsToRecipeStep = recipeStepID
		// Index will be set from array index during recipe creation (via converter)
		// OptionIndex is already 0 from the fake builder
		ingredients = append(ingredients, ing)
	}

	var instruments []*types.RecipeStepInstrument
	for range exampleQuantity {
		ing := BuildFakeRecipeStepInstrument()
		ing.BelongsToRecipeStep = recipeStepID
		// Index will be set from array index during recipe creation (via converter)
		// OptionIndex is already 0 from the fake builder
		instruments = append(instruments, ing)
	}

	var vessels []*types.RecipeStepVessel
	for range exampleQuantity {
		ing := BuildFakeRecipeStepVessel()
		ing.BelongsToRecipeStep = recipeStepID
		// Index will be set from array index during recipe creation (via converter)
		// OptionIndex is already 0 from the fake builder
		vessels = append(vessels, ing)
	}

	var products []*types.RecipeStepProduct
	for range exampleQuantity {
		p := BuildFakeRecipeStepProduct()
		p.BelongsToRecipeStep = recipeStepID
		products = append(products, p)
	}

	completionConditionID := BuildFakeID()
	completionConditions := []*types.RecipeStepCompletionCondition{
		{
			ID:                  completionConditionID,
			BelongsToRecipeStep: recipeStepID,
			IngredientState:     types.ValidIngredientState{},
			Notes:               buildUniqueString(),
			Ingredients: []*types.RecipeStepCompletionConditionIngredient{
				{
					ID:                                     BuildFakeID(),
					BelongsToRecipeStepCompletionCondition: completionConditionID,
					RecipeStepIngredient:                   ingredients[0].ID,
				},
			},
			Optional: false,
		},
	}

	minEstimatedTime := uint32(buildFakeNumber())
	maxEstimatedTime := uint32(buildFakeNumber()) + minEstimatedTime
	minTemperature := float32(buildFakeNumber())
	maxTemperature := float32(buildFakeNumber()) + minTemperature

	return &types.RecipeStep{
		ID:                        recipeStepID,
		Index:                     fake.Uint32(),
		Preparation:               *BuildFakeValidPreparation(),
		MinEstimatedTimeInSeconds: &minEstimatedTime,
		MaxEstimatedTimeInSeconds: &maxEstimatedTime,
		MinTemperatureInCelsius:   &minTemperature,
		MaxTemperatureInCelsius:   &maxTemperature,
		Notes:                     buildUniqueString(),
		Products:                  products,
		Optional:                  false,
		CreatedAt:                 BuildFakeTime(),
		BelongsToRecipe:           BuildFakeID(),
		Ingredients:               ingredients,
		ExplicitInstructions:      buildUniqueString(),
		ConditionExpression:       buildUniqueString(),
		Instruments:               instruments,
		Vessels:                   vessels,
		CompletionConditions:      completionConditions,
		StartTimerAutomatically:   fake.Bool(),
	}
}

// BuildFakeRecipeStepsList builds a faked RecipeStepList.
func BuildFakeRecipeStepsList() *filtering.QueryFilteredResult[types.RecipeStep] {
	var examples []*types.RecipeStep
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipeStep())
	}

	return &filtering.QueryFilteredResult[types.RecipeStep]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeStepUpdateRequestInput builds a faked RecipeStepUpdateRequestInput from a recipe step.
func BuildFakeRecipeStepUpdateRequestInput() *types.RecipeStepUpdateRequestInput {
	recipeStep := BuildFakeRecipeStep()
	return converters.ConvertRecipeStepToRecipeStepUpdateRequestInput(recipeStep)
}

// BuildFakeRecipeStepCreationRequestInput builds a faked RecipeStepCreationRequestInput.
func BuildFakeRecipeStepCreationRequestInput() *types.RecipeStepCreationRequestInput {
	recipeStep := BuildFakeRecipeStep()
	return converters.ConvertRecipeStepToRecipeStepCreationRequestInput(recipeStep)
}
