package mockmanagers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v4/filtering"
)

// Recipe-related mock methods for MockMealPlanningManager. Struct is defined in meal_planning_manager.go.

func (m *MockMealPlanningManager) ListRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, status, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipe(ctx context.Context, creatorID string, input *mealplanning.RecipeCreationRequestInput) (*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, creatorID, input)

	return returnValues.Get(0).(*mealplanning.Recipe), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipe(ctx context.Context, recipeID string) (*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, recipeID)

	return returnValues.Get(0).(*mealplanning.Recipe), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchRecipes(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchRecipesWithInstrumentOwnership(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, accountID, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchForMealEligibleRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, query, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipe(ctx context.Context, recipeID string, input *mealplanning.RecipeUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error {
	returnValues := m.Called(ctx, recipeID, newStatus)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipe(ctx context.Context, recipeID, ownerID string) error {
	returnValues := m.Called(ctx, recipeID, ownerID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error {
	returnValues := m.Called(ctx, recipeID, uploadedMediaID, uploadedByUser)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error {
	returnValues := m.Called(ctx, recipeStepID, uploadedMediaID, uploadedByUser)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) RecipeEstimatedPrepSteps(ctx context.Context, recipeID string) ([]*mealplanning.MealPlanTaskDatabaseCreationEstimate, error) {
	returnValues := m.Called(ctx, recipeID)

	return returnValues.Get(0).([]*mealplanning.MealPlanTaskDatabaseCreationEstimate), returnValues.Error(1)
}

func (m *MockMealPlanningManager) MealMermaid(ctx context.Context, meal *mealplanning.Meal) (string, error) {
	returnArgs := m.Called(ctx, meal)
	return returnArgs.String(0), returnArgs.Error(1)
}

func (m *MockMealPlanningManager) RecipeMermaid(ctx context.Context, recipeID string) (string, error) {
	returnValues := m.Called(ctx, recipeID)

	return returnValues.Get(0).(string), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CloneRecipe(ctx context.Context, recipeID, newOwnerID string) (*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, recipeID, newOwnerID)

	return returnValues.Get(0).(*mealplanning.Recipe), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RecipeImageUpload(ctx context.Context) error {
	returnValues := m.Called(ctx)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeList], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeList]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeList(ctx context.Context, userID string, input *mealplanning.RecipeListCreationRequestInput) (*mealplanning.RecipeList, error) {
	returnValues := m.Called(ctx, userID, input)

	return returnValues.Get(0).(*mealplanning.RecipeList), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeList(ctx context.Context, recipeListID, userID string, input *mealplanning.RecipeListUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeListID, userID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error {
	returnValues := m.Called(ctx, recipeListID, userID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddRecipeToRecipeList(ctx context.Context, recipeListID, recipeID, notes string) (*mealplanning.RecipeListItem, error) {
	returnValues := m.Called(ctx, recipeListID, recipeID, notes)

	return returnValues.Get(0).(*mealplanning.RecipeListItem), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RemoveRecipeFromRecipeList(ctx context.Context, recipeListID, recipeListItemID string) error {
	returnValues := m.Called(ctx, recipeListID, recipeListItemID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) UpdateRecipeListItem(ctx context.Context, recipeListItemID, recipeListID, recipeID string, input *mealplanning.RecipeListItemUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeListItemID, recipeListID, recipeID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeListItem], error) {
	returnValues := m.Called(ctx, recipeListID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeListItem]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStep], error) {
	returnValues := m.Called(ctx, recipeID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStep]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStep(ctx context.Context, recipeID string, input *mealplanning.RecipeStepCreationRequestInput) (*mealplanning.RecipeStep, error) {
	returnValues := m.Called(ctx, recipeID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStep), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*mealplanning.RecipeStep, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID)

	return returnValues.Get(0).(*mealplanning.RecipeStep), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStep(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) RecipeStepImageUpload(ctx context.Context) error {
	returnValues := m.Called(ctx)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeStepProducts(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepProduct], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepProduct]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStepProduct(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepProductCreationRequestInput) (*mealplanning.RecipeStepProduct, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStepProduct), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (*mealplanning.RecipeStepProduct, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepProductID)

	return returnValues.Get(0).(*mealplanning.RecipeStepProduct), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string, input *mealplanning.RecipeStepProductUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepProductID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepProductID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepInstrumentCreationRequestInput) (*mealplanning.RecipeStepInstrument, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStepInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*mealplanning.RecipeStepInstrument, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepInstrumentID)

	return returnValues.Get(0).(*mealplanning.RecipeStepInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string, input *mealplanning.RecipeStepInstrumentUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepInstrumentID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepInstrumentID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeStepIngredients(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepIngredientCreationRequestInput) (*mealplanning.RecipeStepIngredient, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStepIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*mealplanning.RecipeStepIngredient, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)

	return returnValues.Get(0).(*mealplanning.RecipeStepIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string, input *mealplanning.RecipeStepIngredientUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipePrepTask(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipePrepTask], error) {
	returnValues := m.Called(ctx, recipeID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipePrepTask]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipePrepTask(ctx context.Context, recipeID string, input *mealplanning.RecipePrepTaskCreationRequestInput) (*mealplanning.RecipePrepTask, error) {
	returnValues := m.Called(ctx, recipeID, input)

	return returnValues.Get(0).(*mealplanning.RecipePrepTask), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (*mealplanning.RecipePrepTask, error) {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID)

	return returnValues.Get(0).(*mealplanning.RecipePrepTask), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string, input *mealplanning.RecipePrepTaskUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeStepCompletionConditions(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) (*mealplanning.RecipeStepCompletionCondition, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStepCompletionCondition), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) (*mealplanning.RecipeStepCompletionCondition, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepCompletionConditionID)

	return returnValues.Get(0).(*mealplanning.RecipeStepCompletionCondition), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string, input *mealplanning.RecipeStepCompletionConditionUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepCompletionConditionID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepCompletionConditionID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeStepVessels(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepVessel], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID string, input *mealplanning.RecipeStepVesselCreationRequestInput) (*mealplanning.RecipeStepVessel, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, input)

	return returnValues.Get(0).(*mealplanning.RecipeStepVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (*mealplanning.RecipeStepVessel, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepVesselID)

	return returnValues.Get(0).(*mealplanning.RecipeStepVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string, input *mealplanning.RecipeStepVesselUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepVesselID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) error {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepVesselID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListRecipeRatings(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
	returnValues := m.Called(ctx, recipeID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeRating]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*mealplanning.RecipeRating, error) {
	returnValues := m.Called(ctx, recipeID, recipeRatingID)

	return returnValues.Get(0).(*mealplanning.RecipeRating), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateRecipeRating(ctx context.Context, recipeID string, input *mealplanning.RecipeRatingCreationRequestInput) (*mealplanning.RecipeRating, error) {
	returnValues := m.Called(ctx, recipeID, input)

	return returnValues.Get(0).(*mealplanning.RecipeRating), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateRecipeRating(ctx context.Context, recipeID, recipeRatingID string, input *mealplanning.RecipeRatingUpdateRequestInput) error {
	returnValues := m.Called(ctx, recipeID, recipeRatingID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error {
	returnValues := m.Called(ctx, recipeID, recipeRatingID)

	return returnValues.Error(0)
}
