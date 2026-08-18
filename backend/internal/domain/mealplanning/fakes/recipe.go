package fakes

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/filtering"
)

// BuildFakeRecipe builds a faked recipe.
func BuildFakeRecipe() *mealplanning.Recipe {
	recipeID := BuildFakeID()

	var steps []*mealplanning.RecipeStep
	for i := range exampleQuantity {
		step := BuildFakeRecipeStep()
		step.Index = uint32(i)
		step.BelongsToRecipe = recipeID
		steps = append(steps, step)
	}

	prepTasks := BuildFakeRecipePrepTasksList().Data
	for i := range prepTasks {
		prepTasks[i].BelongsToRecipe = recipeID
	}

	recipeMedia := BuildFakeRecipeMediaList().Data
	for i := range recipeMedia {
		recipeMedia[i].BelongsToRecipe = &recipeID
	}

	return &mealplanning.Recipe{
		ID:                   recipeID,
		Name:                 buildUniqueString(),
		Slug:                 buildUniqueString(),
		Source:               buildUniqueString(),
		Description:          buildUniqueString(),
		InspiredByRecipeID:   nil,
		CreatedAt:            BuildFakeTime(),
		CreatedByUser:        BuildFakeID(),
		Steps:                steps,
		PrepTasks:            prepTasks,
		Status:               mealplanning.RecipeStatusSubmitted,
		Media:                recipeMedia,
		MinEstimatedPortions: float32(buildFakeNumber()),
		MaxEstimatedPortions: new(float32(buildFakeNumber())),
		PortionName:          buildUniqueString(),
		PluralPortionName:    buildUniqueString(),
		EligibleForMeals:     true,
		YieldsComponentType:  "main",
	}
}

// BuildFakeRecipesList builds a faked RecipeList.
func BuildFakeRecipesList() *filtering.QueryFilteredResult[mealplanning.Recipe] {
	var examples []*mealplanning.Recipe
	for range exampleQuantity {
		examples = append(examples, BuildFakeRecipe())
	}

	return &filtering.QueryFilteredResult[mealplanning.Recipe]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeRecipeUpdateRequestInput builds a faked RecipeUpdateRequestInput from a recipe.
func BuildFakeRecipeUpdateRequestInput() *mealplanning.RecipeUpdateRequestInput {
	recipe := BuildFakeRecipe()
	return converters.ConvertRecipeToRecipeUpdateRequestInput(recipe)
}

// BuildFakeRecipeCreationRequestInput builds a faked RecipeCreationRequestInput.
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
