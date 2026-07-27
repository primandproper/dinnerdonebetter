package managers

import (
	"context"
	"fmt"

	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/recipeanalysis"
	eatingindexing "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/indexing"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/mealplanning/workers"

	"github.com/primandproper/platform-go/v7/filtering"
	"github.com/primandproper/platform-go/v7/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v7/messagequeue/config"
	"github.com/primandproper/platform-go/v7/observability"
	"github.com/primandproper/platform-go/v7/observability/logging"
	"github.com/primandproper/platform-go/v7/observability/metrics"
	"github.com/primandproper/platform-go/v7/observability/tracing"
	textsearch "github.com/primandproper/platform-go/v7/search/text"
	textsearchcfg "github.com/primandproper/platform-go/v7/search/text/config"
)

const (
	mealPlannerName = "meal_planning_manager"
)

type (
	MealPlanningManager interface {
		// Meals
		ListMeals(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error)
		CreateMeal(ctx context.Context, creatorID string, input *types.MealCreationRequestInput) (*types.Meal, error)
		ReadMeal(ctx context.Context, mealID string) (*types.Meal, error)
		SearchMeals(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Meal], error)
		ArchiveMeal(ctx context.Context, mealID, ownerID string) error
		AddMealImage(ctx context.Context, mealID, uploadedMediaID, uploadedByUser string) error

		// Meal plans
		ListMealPlans(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlan], error)
		CreateMealPlan(ctx context.Context, ownerID, creatorID string, input *types.MealPlanCreationRequestInput) (*types.MealPlan, error)
		ReadMealPlan(ctx context.Context, mealPlanID, ownerID string) (*types.MealPlan, error)
		UpdateMealPlan(ctx context.Context, mealPlanID, ownerID string, input *types.MealPlanUpdateRequestInput) error
		ArchiveMealPlan(ctx context.Context, mealPlanID, ownerID string) error
		FinalizeMealPlan(ctx context.Context, mealPlanID, ownerID string) (bool, error)

		// Meal plan events
		ListMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanEvent], error)
		CreateMealPlanEvent(ctx context.Context, mealPlanID string, input *types.MealPlanEventCreationRequestInput) (*types.MealPlanEvent, error)
		ReadMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*types.MealPlanEvent, error)
		UpdateMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string, input *types.MealPlanEventUpdateRequestInput) error
		SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error
		ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error

		// Meal plan options
		ListMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOption], error)
		CreateMealPlanOption(ctx context.Context, input *types.MealPlanOptionCreationRequestInput) (*types.MealPlanOption, error)
		CreateMealPlanOptionWithEventID(ctx context.Context, mealPlanEventID string, input *types.MealPlanOptionCreationRequestInput) (*types.MealPlanOption, error)
		ReadMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*types.MealPlanOption, error)
		MealPlanOptionBelongsToAccount(ctx context.Context, mealPlanOptionID, accountID string) (bool, error)
		UpdateMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, input *types.MealPlanOptionUpdateRequestInput) error
		ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) error

		// Meal plan option votes
		ListMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanOptionVote], error)
		CreateMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, creatorID string, input *types.MealPlanOptionVoteCreationRequestInput) ([]*types.MealPlanOptionVote, error)
		ReadMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*types.MealPlanOptionVote, error)
		UpdateMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string, input *types.MealPlanOptionVoteUpdateRequestInput) error
		ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error

		// Meal plan tasks
		ListMealPlanTasksByMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanTask], error)
		ReadMealPlanTask(ctx context.Context, mealPlanID, mealPlanTaskID string) (*types.MealPlanTask, error)
		CreateMealPlanTask(ctx context.Context, input *types.MealPlanTaskCreationRequestInput) (*types.MealPlanTask, error)
		MealPlanTaskStatusChange(ctx context.Context, input *types.MealPlanTaskStatusChangeRequestInput) error

		// Meal plan grocery list items
		ListMealPlanGroceryListItemsByMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanGroceryListItem], error)
		CreateMealPlanGroceryListItem(ctx context.Context, input *types.MealPlanGroceryListItemCreationRequestInput) (*types.MealPlanGroceryListItem, error)
		ReadMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (*types.MealPlanGroceryListItem, error)
		UpdateMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string, input *types.MealPlanGroceryListItemUpdateRequestInput) error
		ArchiveMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) error

		// Meal plan recipe option selections
		GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*types.MealPlanRecipeOptionSelection, error)
		GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealPlanRecipeOptionSelection], error)
		CreateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID string, input *types.MealPlanRecipeOptionSelectionCreationRequestInput) (*types.MealPlanRecipeOptionSelection, error)
		UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *types.MealPlanRecipeOptionSelectionUpdateRequestInput) error
		ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error

		// User ingredient preferences
		ReadUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) (*types.UserIngredientPreference, error)
		ListUserIngredientPreferences(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UserIngredientPreference], error)
		CreateUserIngredientPreference(ctx context.Context, ownerID string, input *types.UserIngredientPreferenceCreationRequestInput) ([]*types.UserIngredientPreference, error)
		UpdateUserIngredientPreference(ctx context.Context, ingredientPreferenceID, ownerID string, input *types.UserIngredientPreferenceUpdateRequestInput) error
		ArchiveUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) error

		// Account instrument ownerships
		ListAccountInstrumentOwnerships(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.AccountInstrumentOwnership], error)
		SearchValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error)
		CreateAccountInstrumentOwnership(ctx context.Context, ownerID string, input *types.AccountInstrumentOwnershipCreationRequestInput) (*types.AccountInstrumentOwnership, error)
		ReadAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) (*types.AccountInstrumentOwnership, error)
		UpdateAccountInstrumentOwnership(ctx context.Context, instrumentOwnershipID, ownerID string, input *types.AccountInstrumentOwnershipUpdateRequestInput) error
		ArchiveAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) error

		// Meal lists
		ListMealLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealList], error)
		CreateMealList(ctx context.Context, userID string, input *types.MealListCreationRequestInput) (*types.MealList, error)
		UpdateMealList(ctx context.Context, mealListID, userID string, input *types.MealListUpdateRequestInput) error
		ArchiveMealList(ctx context.Context, mealListID, userID string) error
		AddMealToMealList(ctx context.Context, mealListID, mealID, notes string) (*types.MealListItem, error)
		UpdateMealListItem(ctx context.Context, mealListItemID, mealListID, mealID string, input *types.MealListItemUpdateRequestInput) error
		RemoveMealFromMealList(ctx context.Context, mealListID, mealListItemID string) error
		ListMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.MealListItem], error)

		// Recipes
		ListRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error)
		CreateRecipe(ctx context.Context, creatorID string, input *types.RecipeCreationRequestInput) (*types.Recipe, error)
		ReadRecipe(ctx context.Context, recipeID string) (*types.Recipe, error)
		SearchRecipes(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error)
		SearchForMealEligibleRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error)
		SearchRecipesWithInstrumentOwnership(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.Recipe], error)
		UpdateRecipe(ctx context.Context, recipeID string, input *types.RecipeUpdateRequestInput) error
		UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error
		ArchiveRecipe(ctx context.Context, recipeID, ownerID string) error
		AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error
		RecipeEstimatedPrepSteps(ctx context.Context, recipeID string) ([]*types.MealPlanTaskDatabaseCreationEstimate, error)
		MealMermaid(ctx context.Context, meal *types.Meal) (string, error)
		RecipeMermaid(ctx context.Context, recipeID string) (string, error)
		CloneRecipe(ctx context.Context, recipeID, newOwnerID string) (*types.Recipe, error)
		RecipeImageUpload(ctx context.Context) error

		// Recipe lists
		ListRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeList], error)
		AddRecipeToRecipeList(ctx context.Context, recipeListID, recipeID, notes string) (*types.RecipeListItem, error)
		RemoveRecipeFromRecipeList(ctx context.Context, recipeListID, recipeListItemID string) error
		ListRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeListItem], error)
		CreateRecipeList(ctx context.Context, userID string, input *types.RecipeListCreationRequestInput) (*types.RecipeList, error)
		UpdateRecipeList(ctx context.Context, recipeListID, userID string, input *types.RecipeListUpdateRequestInput) error
		UpdateRecipeListItem(ctx context.Context, recipeListItemID, recipeListID, recipeID string, input *types.RecipeListItemUpdateRequestInput) error
		ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error

		// Recipe steps
		ListRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStep], error)
		CreateRecipeStep(ctx context.Context, recipeID string, input *types.RecipeStepCreationRequestInput) (*types.RecipeStep, error)
		ReadRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*types.RecipeStep, error)
		UpdateRecipeStep(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepUpdateRequestInput) error
		ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error
		AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error
		RecipeStepImageUpload(ctx context.Context) error

		// Recipe step products
		ListRecipeStepProducts(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepProduct], error)
		CreateRecipeStepProduct(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepProductCreationRequestInput) (*types.RecipeStepProduct, error)
		ReadRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (*types.RecipeStepProduct, error)
		UpdateRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string, input *types.RecipeStepProductUpdateRequestInput) error
		ArchiveRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) error

		// Recipe step instruments
		ListRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepInstrument], error)
		CreateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepInstrumentCreationRequestInput) (*types.RecipeStepInstrument, error)
		ReadRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*types.RecipeStepInstrument, error)
		UpdateRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string, input *types.RecipeStepInstrumentUpdateRequestInput) error
		ArchiveRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) error

		// Recipe step ingredients
		ListRecipeStepIngredients(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepIngredient], error)
		CreateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepIngredientCreationRequestInput) (*types.RecipeStepIngredient, error)
		ReadRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*types.RecipeStepIngredient, error)
		UpdateRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string, input *types.RecipeStepIngredientUpdateRequestInput) error
		ArchiveRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) error

		// Recipe prep tasks
		ListRecipePrepTask(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipePrepTask], error)
		CreateRecipePrepTask(ctx context.Context, recipeID string, input *types.RecipePrepTaskCreationRequestInput) (*types.RecipePrepTask, error)
		ReadRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (*types.RecipePrepTask, error)
		UpdateRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string, input *types.RecipePrepTaskUpdateRequestInput) error
		ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error

		// Recipe step completion conditions
		ListRecipeStepCompletionConditions(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepCompletionCondition], error)
		CreateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) (*types.RecipeStepCompletionCondition, error)
		ReadRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) (*types.RecipeStepCompletionCondition, error)
		UpdateRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string, input *types.RecipeStepCompletionConditionUpdateRequestInput) error
		ArchiveRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepCompletionConditionID string) error

		// Recipe step vessels
		ListRecipeStepVessels(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeStepVessel], error)
		CreateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID string, input *types.RecipeStepVesselCreationRequestInput) (*types.RecipeStepVessel, error)
		ReadRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (*types.RecipeStepVessel, error)
		UpdateRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string, input *types.RecipeStepVesselUpdateRequestInput) error
		ArchiveRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) error

		// Recipe ratings
		ListRecipeRatings(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.RecipeRating], error)
		ReadRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*types.RecipeRating, error)
		CreateRecipeRating(ctx context.Context, recipeID string, input *types.RecipeRatingCreationRequestInput) (*types.RecipeRating, error)
		UpdateRecipeRating(ctx context.Context, recipeID, recipeRatingID string, input *types.RecipeRatingUpdateRequestInput) error
		ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error

		// Valid ingredient groups
		SearchValidIngredientGroups(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error)
		ListValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientGroup], error)
		CreateValidIngredientGroup(ctx context.Context, input *types.ValidIngredientGroupCreationRequestInput) (*types.ValidIngredientGroup, error)
		ReadValidIngredientGroup(ctx context.Context, validIngredientGroupID string) (*types.ValidIngredientGroup, error)
		UpdateValidIngredientGroup(ctx context.Context, validIngredientGroupID string, input *types.ValidIngredientGroupUpdateRequestInput) (*types.ValidIngredientGroup, error)
		ArchiveValidIngredientGroup(ctx context.Context, validIngredientGroupID string) error

		// Valid ingredient measurement units
		ListValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error)
		CreateValidIngredientMeasurementUnit(ctx context.Context, input *types.ValidIngredientMeasurementUnitCreationRequestInput) (*types.ValidIngredientMeasurementUnit, error)
		ReadValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*types.ValidIngredientMeasurementUnit, error)
		UpdateValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string, input *types.ValidIngredientMeasurementUnitUpdateRequestInput) (*types.ValidIngredientMeasurementUnit, error)
		ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error
		SearchValidIngredientMeasurementUnitsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error)
		SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientMeasurementUnit], error)

		// Valid ingredient preparations
		ListValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error)
		CreateValidIngredientPreparation(ctx context.Context, input *types.ValidIngredientPreparationCreationRequestInput) (*types.ValidIngredientPreparation, error)
		ReadValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*types.ValidIngredientPreparation, error)
		UpdateValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string, input *types.ValidIngredientPreparationUpdateRequestInput) (*types.ValidIngredientPreparation, error)
		ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error
		SearchValidIngredientPreparationsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error)
		SearchValidIngredientPreparationsByPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientPreparation], error)

		// Valid prep task configs
		ListValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error)
		CreateValidPrepTaskConfig(ctx context.Context, input *types.ValidPrepTaskConfigCreationRequestInput) (*types.ValidPrepTaskConfig, error)
		ReadValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) (*types.ValidPrepTaskConfig, error)
		UpdateValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string, input *types.ValidPrepTaskConfigUpdateRequestInput) (*types.ValidPrepTaskConfig, error)
		ArchiveValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) error
		SearchValidPrepTaskConfigsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error)
		SearchValidPrepTaskConfigsByPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error)
		SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPrepTaskConfig], error)

		// Valid ingredients
		SearchValidIngredients(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error)
		ListValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error)
		CreateValidIngredient(ctx context.Context, input *types.ValidIngredientCreationRequestInput) (*types.ValidIngredient, error)
		ReadValidIngredient(ctx context.Context, validIngredientID string) (*types.ValidIngredient, error)
		RandomValidIngredient(ctx context.Context) (*types.ValidIngredient, error)
		UpdateValidIngredient(ctx context.Context, validIngredientID string, input *types.ValidIngredientUpdateRequestInput) (*types.ValidIngredient, error)
		ArchiveValidIngredient(ctx context.Context, validIngredientID string) error
		AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error
		SearchValidIngredientsByPreparationAndIngredientName(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredient], error)

		// Valid ingredient state ingredients
		ListValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error)
		CreateValidIngredientStateIngredient(ctx context.Context, input *types.ValidIngredientStateIngredientCreationRequestInput) (*types.ValidIngredientStateIngredient, error)
		ReadValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) (*types.ValidIngredientStateIngredient, error)
		UpdateValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string, input *types.ValidIngredientStateIngredientUpdateRequestInput) (*types.ValidIngredientStateIngredient, error)
		ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) error
		SearchValidIngredientStateIngredientsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error)
		SearchValidIngredientStateIngredientsByIngredientState(ctx context.Context, validIngredientStateID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientStateIngredient], error)

		// Valid ingredient states
		SearchValidIngredientStates(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error)
		ListValidIngredientStates(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidIngredientState], error)
		CreateValidIngredientState(ctx context.Context, input *types.ValidIngredientStateCreationRequestInput) (*types.ValidIngredientState, error)
		ReadValidIngredientState(ctx context.Context, validIngredientStateID string) (*types.ValidIngredientState, error)
		UpdateValidIngredientState(ctx context.Context, validIngredientStateID string, input *types.ValidIngredientStateUpdateRequestInput) (*types.ValidIngredientState, error)
		ArchiveValidIngredientState(ctx context.Context, validIngredientStateID string) error

		// Valid measurement units
		SearchValidMeasurementUnits(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error)
		SearchValidMeasurementUnitsByIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error)
		ListValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnit], error)
		CreateValidMeasurementUnit(ctx context.Context, input *types.ValidMeasurementUnitCreationRequestInput) (*types.ValidMeasurementUnit, error)
		ReadValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*types.ValidMeasurementUnit, error)
		UpdateValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string, input *types.ValidMeasurementUnitUpdateRequestInput) (*types.ValidMeasurementUnit, error)
		ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error

		// Valid instruments
		SearchValidInstruments(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error)
		ListValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidInstrument], error)
		CreateValidInstrument(ctx context.Context, input *types.ValidInstrumentCreationRequestInput) (*types.ValidInstrument, error)
		ReadValidInstrument(ctx context.Context, validInstrumentID string) (*types.ValidInstrument, error)
		RandomValidInstrument(ctx context.Context) (*types.ValidInstrument, error)
		UpdateValidInstrument(ctx context.Context, validInstrumentID string, input *types.ValidInstrumentUpdateRequestInput) (*types.ValidInstrument, error)
		ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error

		// Valid measurement unit conversions
		ValidMeasurementUnitConversionsForMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidMeasurementUnitConversion], error)
		GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*types.ValidMeasurementUnitConversion, error)
		CreateValidMeasurementUnitConversion(ctx context.Context, input *types.ValidMeasurementUnitConversionCreationRequestInput) (*types.ValidMeasurementUnitConversion, error)
		ReadValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) (*types.ValidMeasurementUnitConversion, error)
		UpdateValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string, input *types.ValidMeasurementUnitConversionUpdateRequestInput) (*types.ValidMeasurementUnitConversion, error)
		ArchiveValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) error
		GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*types.MeasurementUnitConversionMismatch, error)

		// Valid preparation instruments
		ListValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error)
		CreateValidPreparationInstrument(ctx context.Context, input *types.ValidPreparationInstrumentCreationRequestInput) (*types.ValidPreparationInstrument, error)
		ReadValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*types.ValidPreparationInstrument, error)
		UpdateValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string, input *types.ValidPreparationInstrumentUpdateRequestInput) (*types.ValidPreparationInstrument, error)
		ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error
		SearchValidPreparationInstrumentsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error)
		SearchValidPreparationInstrumentsByInstrument(ctx context.Context, validInstrumentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationInstrument], error)

		// Valid preparations
		SearchValidPreparations(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error)
		ListValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparation], error)
		CreateValidPreparation(ctx context.Context, input *types.ValidPreparationCreationRequestInput) (*types.ValidPreparation, error)
		ReadValidPreparation(ctx context.Context, validPreparationID string) (*types.ValidPreparation, error)
		RandomValidPreparation(ctx context.Context) (*types.ValidPreparation, error)
		UpdateValidPreparation(ctx context.Context, validPreparationID string, input *types.ValidPreparationUpdateRequestInput) (*types.ValidPreparation, error)
		ArchiveValidPreparation(ctx context.Context, validPreparationID string) error
		AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error

		// Valid preparation vessels
		ListValidPreparationVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error)
		CreateValidPreparationVessel(ctx context.Context, input *types.ValidPreparationVesselCreationRequestInput) (*types.ValidPreparationVessel, error)
		ReadValidPreparationVessel(ctx context.Context, validPreparationVesselID string) (*types.ValidPreparationVessel, error)
		UpdateValidPreparationVessel(ctx context.Context, validPreparationVesselID string, input *types.ValidPreparationVesselUpdateRequestInput) (*types.ValidPreparationVessel, error)
		ArchiveValidPreparationVessel(ctx context.Context, validPreparationVesselID string) error
		SearchValidPreparationVesselsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error)
		SearchValidPreparationVesselsByVessel(ctx context.Context, validVesselID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidPreparationVessel], error)

		// Valid vessels
		SearchValidVessels(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error)
		ListValidVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.ValidVessel], error)
		CreateValidVessel(ctx context.Context, input *types.ValidVesselCreationRequestInput) (*types.ValidVessel, error)
		ReadValidVessel(ctx context.Context, validVesselID string) (*types.ValidVessel, error)
		RandomValidVessel(ctx context.Context) (*types.ValidVessel, error)
		UpdateValidVessel(ctx context.Context, validVesselID string, input *types.ValidVesselUpdateRequestInput) (*types.ValidVessel, error)
		ArchiveValidVessel(ctx context.Context, validVesselID string) error
	}

	mealPlanTaskCreatorWorker interface {
		workers.Worker
	}

	mealPlanGroceryListInitializerWorker interface {
		workers.Worker
	}

	mealPlanningManager struct {
		tracer                           tracing.Tracer
		logger                           logging.Logger
		dataChangesPublisher             messagequeue.Publisher
		db                               types.Repository
		recipeAnalyzer                   recipeanalysis.RecipeAnalyzer
		groceryListInitializer           mealPlanGroceryListInitializerWorker
		taskCreator                      mealPlanTaskCreatorWorker
		mealsSearchIndex                 textsearch.IndexSearcher[eatingindexing.MealSearchSubset]
		recipeSearchIndex                textsearch.IndexSearcher[eatingindexing.RecipeSearchSubset]
		validIngredientStatesSearchIndex textsearch.IndexSearcher[eatingindexing.ValidIngredientStateSearchSubset]
		validInstrumentSearchIndex       textsearch.IndexSearcher[eatingindexing.ValidInstrumentSearchSubset]
		validMeasurementUnitSearchIndex  textsearch.IndexSearcher[eatingindexing.ValidMeasurementUnitSearchSubset]
		validIngredientSearchIndex       textsearch.IndexSearcher[eatingindexing.ValidIngredientSearchSubset]
		validPreparationsSearchIndex     textsearch.IndexSearcher[eatingindexing.ValidPreparationSearchSubset]
		validVesselsSearchIndex          textsearch.IndexSearcher[eatingindexing.ValidVesselSearchSubset]
	}
)

var (
	_ MealPlanningManager = (*mealPlanningManager)(nil)
)

func NewMealPlanningManager(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	db types.Repository,
	cfg *msgconfig.QueuesConfig,
	publisherProvider messagequeue.PublisherProvider,
	recipeAnalyzer recipeanalysis.RecipeAnalyzer,
	searchConfig *textsearchcfg.Config,
	metricsProvider metrics.Provider,
	groceryListInitializer mealPlanGroceryListInitializerWorker,
	taskCreator mealPlanTaskCreatorWorker,
) (MealPlanningManager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	mealsSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.MealSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeMeals)
	if err != nil {
		return nil, observability.PrepareError(err, nil, "initializing meals search index")
	}

	recipeSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.RecipeSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeRecipes)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeRecipes, err)
	}

	validIngredientStatesSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidIngredientStateSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidIngredientStates)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidIngredientStates, err)
	}

	validInstrumentSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidInstrumentSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidInstruments)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidInstruments, err)
	}

	validMeasurementUnitSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidMeasurementUnitSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidMeasurementUnits)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidMeasurementUnits, err)
	}

	validIngredientSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidIngredientSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidIngredients)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidIngredients, err)
	}

	validPreparationsSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidPreparationSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidPreparations)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidPreparations, err)
	}

	validVesselsSearchIndex, err := textsearchcfg.NewIndex[eatingindexing.ValidVesselSearchSubset](ctx, logger, tracerProvider, metricsProvider, searchConfig, eatingindexing.IndexTypeValidVessels)
	if err != nil {
		return nil, fmt.Errorf("failed to provide search index for %s index: %w", eatingindexing.IndexTypeValidVessels, err)
	}

	m := &mealPlanningManager{
		db:                               db,
		tracer:                           tracing.NewNamedTracer(tracerProvider, mealPlannerName),
		logger:                           logging.NewNamedLogger(logger, mealPlannerName),
		dataChangesPublisher:             dataChangesPublisher,
		recipeAnalyzer:                   recipeAnalyzer,
		groceryListInitializer:           groceryListInitializer,
		taskCreator:                      taskCreator,
		mealsSearchIndex:                 mealsSearchIndex,
		recipeSearchIndex:                recipeSearchIndex,
		validIngredientStatesSearchIndex: validIngredientStatesSearchIndex,
		validInstrumentSearchIndex:       validInstrumentSearchIndex,
		validMeasurementUnitSearchIndex:  validMeasurementUnitSearchIndex,
		validIngredientSearchIndex:       validIngredientSearchIndex,
		validPreparationsSearchIndex:     validPreparationsSearchIndex,
		validVesselsSearchIndex:          validVesselsSearchIndex,
	}

	return m, nil
}

func (m *mealPlanningManager) runPostFinalizationWorkers(ctx context.Context, logger logging.Logger, span tracing.Span) {
	if m.groceryListInitializer == nil || m.taskCreator == nil {
		return
	}
	if err := m.groceryListInitializer.Work(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "running grocery list initializer after meal plan finalization")
	}
	if err := m.taskCreator.Work(ctx); err != nil {
		observability.AcknowledgeError(err, logger, span, "running task creator after meal plan finalization")
	}
}
