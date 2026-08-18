package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/fake"
	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/pointer"
)

// BuildFakeRecipe builds a faked recipe.
func BuildFakeRecipe() *mealplanning.Recipe {
	recipe := fake.BuildFakeRecord[mealplanning.Recipe]()

	// A recipe that has been written but not approved, which is the state one arrives
	// in, and the component a recipe yields on its own.
	recipe.Status = mealplanning.RecipeStatusSubmitted
	recipe.EligibleForMeals = true
	recipe.YieldsComponentType = "main"

	// A portion range rather than two independent numbers.
	recipe.MinEstimatedPortions = float32(fake.BuildFakeNumber())
	recipe.MaxEstimatedPortions = pointer.To(recipe.MinEstimatedPortions + float32(fake.BuildFakeNumber()))

	// Everything below belongs to this recipe: steps in order, and prep tasks and media
	// pointing back at it. A child that belongs to some other recipe is one the read
	// path does not return with this one, which is a failure several layers from here.
	steps := make([]*mealplanning.RecipeStep, 0, exampleQuantity)
	for i := range exampleQuantity {
		step := BuildFakeRecipeStep()
		step.Index = uint32(i)
		step.BelongsToRecipe = recipe.ID
		steps = append(steps, step)
	}
	recipe.Steps = steps

	prepTasks := BuildFakeRecipePrepTasksList().Data
	for i := range prepTasks {
		prepTasks[i].BelongsToRecipe = recipe.ID
	}
	recipe.PrepTasks = prepTasks

	recipeMedia := BuildFakeRecipeMediaList().Data
	for i := range recipeMedia {
		recipeMedia[i].BelongsToRecipe = &recipe.ID
	}
	recipe.Media = recipeMedia

	return recipe
}

// BuildFakeRecipesList builds a faked RecipeList.
func BuildFakeRecipesList() *filtering.QueryFilteredResult[mealplanning.Recipe] {
	return fake.BuildFakePage(BuildFakeRecipe)
}

// BuildFakeRecipeUpdateRequestInput builds a faked RecipeUpdateRequestInput from a recipe.
func BuildFakeRecipeUpdateRequestInput() *mealplanning.RecipeUpdateRequestInput {
	recipe := BuildFakeRecipe()

	return converters.ConvertRecipeToRecipeUpdateRequestInput(recipe)
}

// BuildFakeRecipeCreationRequestInput builds a faked RecipeCreationRequestInput.
//
// Hand-written past the conversion: a prep task inside a creation input refers to its
// steps by index rather than by ID, because neither has one yet, and that renumbering
// is a procedure rather than a value.
func BuildFakeRecipeCreationRequestInput() *mealplanning.RecipeCreationRequestInput {
	exampleRecipe := BuildFakeRecipe()
	exampleCreationInput := converters.ConvertRecipeToRecipeCreationRequestInput(exampleRecipe)
	examplePrepTask := BuildFakeRecipePrepTask()
	examplePrepTaskInput := converters.ConvertRecipePrepTaskToRecipePrepTaskWithinRecipeCreationRequestInput(exampleRecipe, examplePrepTask)
	examplePrepTaskInput.RecipeSteps = []*mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		{
			BelongsToRecipeStepIndex: exampleCreationInput.Steps[0].Index,
			SatisfiesRecipeStep:      false,
		},
	}
	exampleCreationInput.PrepTasks = []*mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		examplePrepTaskInput,
	}

	return exampleCreationInput
}
