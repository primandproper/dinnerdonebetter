package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v12/fake"
	"github.com/primandproper/platform-go/v12/filtering"
	"github.com/primandproper/platform-go/v12/pointer"
)

// BuildFakeRecipeStep builds a faked recipe step.
func BuildFakeRecipeStep() *types.RecipeStep {
	step := fake.BuildFakeRecord[types.RecipeStep]()

	step.Preparation = *BuildFakeValidPreparation()

	// Two ranges rather than four independent numbers: a step that ends before it
	// starts, or cools as it heats, is one no timer or thermometer check can satisfy.
	minEstimatedTime := uint32(fake.BuildFakeNumber())
	step.MinEstimatedTimeInSeconds = &minEstimatedTime
	step.MaxEstimatedTimeInSeconds = pointer.To(minEstimatedTime + uint32(fake.BuildFakeNumber()))

	minTemperature := float32(fake.BuildFakeNumber())
	step.MinTemperatureInCelsius = &minTemperature
	step.MaxTemperatureInCelsius = pointer.To(minTemperature + float32(fake.BuildFakeNumber()))

	// A step every test has to account for, rather than one half of them may skip.
	step.Optional = false

	// Everything below belongs to this step. Their indices are assigned from their
	// position by the converter during recipe creation, so they are left alone here.
	ingredients := make([]*types.RecipeStepIngredient, 0, exampleQuantity)
	for range exampleQuantity {
		ingredient := BuildFakeRecipeStepIngredient()
		ingredient.BelongsToRecipeStep = step.ID
		ingredients = append(ingredients, ingredient)
	}
	step.Ingredients = ingredients

	instruments := make([]*types.RecipeStepInstrument, 0, exampleQuantity)
	for range exampleQuantity {
		instrument := BuildFakeRecipeStepInstrument()
		instrument.BelongsToRecipeStep = step.ID
		instruments = append(instruments, instrument)
	}
	step.Instruments = instruments

	vessels := make([]*types.RecipeStepVessel, 0, exampleQuantity)
	for range exampleQuantity {
		vessel := BuildFakeRecipeStepVessel()
		vessel.BelongsToRecipeStep = step.ID
		vessels = append(vessels, vessel)
	}
	step.Vessels = vessels

	products := make([]*types.RecipeStepProduct, 0, exampleQuantity)
	for range exampleQuantity {
		product := BuildFakeRecipeStepProduct()
		product.BelongsToRecipeStep = step.ID
		products = append(products, product)
	}
	step.Products = products

	// A condition naming one of the step's own ingredients, since a condition about an
	// ingredient the step does not have is one that can never be met.
	completionCondition := fake.BuildFakeRecord[types.RecipeStepCompletionCondition]()
	completionCondition.BelongsToRecipeStep = step.ID
	completionCondition.IngredientState = types.ValidIngredientState{}
	completionCondition.Optional = false
	conditionIngredient := fake.BuildFakeRecord[types.RecipeStepCompletionConditionIngredient]()
	conditionIngredient.BelongsToRecipeStepCompletionCondition = completionCondition.ID
	conditionIngredient.RecipeStepIngredient = ingredients[0].ID
	completionCondition.Ingredients = []*types.RecipeStepCompletionConditionIngredient{conditionIngredient}
	step.CompletionConditions = []*types.RecipeStepCompletionCondition{completionCondition}

	return step
}

// BuildFakeRecipeStepsList builds a faked RecipeStepList.
func BuildFakeRecipeStepsList() *filtering.QueryFilteredResult[types.RecipeStep] {
	return fake.BuildFakePage(BuildFakeRecipeStep)
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
