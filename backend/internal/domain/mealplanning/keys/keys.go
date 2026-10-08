package keys

const (
	idSuffix = ".id"

	// AccountInstrumentOwnershipKey is the standard key for referring to an account instrument ownership.
	AccountInstrumentOwnershipKey = "account_instrument_ownership"
	// AccountInstrumentOwnershipIDKey is the standard key for referring to an account instrument ownership's ID.
	AccountInstrumentOwnershipIDKey = AccountInstrumentOwnershipKey + idSuffix

	// MealKey is the standard key for referring to a meal.
	MealKey = "meal"
	// MealIDKey is the standard key for referring to a meal's ID.
	MealIDKey = MealKey + idSuffix

	// MealComponentKey is the standard key for referring to a meal component.
	MealComponentKey = "meal_component"
	// MealComponentIDKey is the standard key for referring to a meal component's ID.
	MealComponentIDKey = MealComponentKey + idSuffix

	// MealListKey is the standard key for referring to a meal list.
	MealListKey = "meal_list"
	// MealListIDKey is the standard key for referring to a meal list's ID.
	MealListIDKey = MealListKey + idSuffix

	// MealListItemKey is the standard key for referring to a meal list item.
	MealListItemKey = "meal_list_item"
	// MealListItemIDKey is the standard key for referring to a meal list item's ID.
	MealListItemIDKey = MealListItemKey + idSuffix

	// MealPlanKey is the standard key for referring to a meal plan.
	MealPlanKey = "meal_plan"
	// MealPlanIDKey is the standard key for referring to a meal plan's ID.
	MealPlanIDKey = MealPlanKey + idSuffix

	// MealPlanEventKey is the standard key for referring to a meal plan event.
	MealPlanEventKey = "meal_plan_event"
	// MealPlanEventIDKey is the standard key for referring to a meal plan event's ID.
	MealPlanEventIDKey = MealPlanEventKey + idSuffix

	// MealPlanGroceryListItemKey is the standard key for referring to a meal plan grocery list item.
	MealPlanGroceryListItemKey = "meal_plan_grocery_list_item"
	// MealPlanGroceryListItemIDKey is the standard key for referring to a meal plan grocery list item's ID.
	MealPlanGroceryListItemIDKey = MealPlanGroceryListItemKey + idSuffix

	// MealPlanOptionKey is the standard key for referring to a meal plan option.
	MealPlanOptionKey = "meal_plan_option"
	// MealPlanOptionIDKey is the standard key for referring to a meal plan option's ID.
	MealPlanOptionIDKey = MealPlanOptionKey + idSuffix

	// MealPlanOptionVoteKey is the standard key for referring to a meal plan option vote.
	MealPlanOptionVoteKey = "meal_plan_option_vote"
	// MealPlanOptionVoteIDKey is the standard key for referring to a meal plan option vote's ID.
	MealPlanOptionVoteIDKey = MealPlanOptionVoteKey + idSuffix

	// MealPlanRecipeOptionSelectionKey is the standard key for referring to a meal plan recipe option selection.
	MealPlanRecipeOptionSelectionKey = "meal_plan_recipe_option_selection"
	// MealPlanRecipeOptionSelectionIDKey is the standard key for referring to a meal plan recipe option selection's ID.
	MealPlanRecipeOptionSelectionIDKey = MealPlanRecipeOptionSelectionKey + idSuffix

	// MealPlanTaskKey is the standard key for referring to a meal plan task.
	MealPlanTaskKey = "meal_plan_task"
	// MealPlanTaskIDKey is the standard key for referring to a meal plan task's ID.
	MealPlanTaskIDKey = MealPlanTaskKey + idSuffix

	// RecipeKey is the standard key for referring to a recipe.
	RecipeKey = "recipe"
	// RecipeIDKey is the standard key for referring to a recipe's ID.
	RecipeIDKey = RecipeKey + idSuffix

	// RecipeListKey is the standard key for referring to a recipe list.
	RecipeListKey = "recipe_list"
	// RecipeListIDKey is the standard key for referring to a recipe list's ID.
	RecipeListIDKey = RecipeListKey + idSuffix

	// RecipeListItemKey is the standard key for referring to a recipe list item.
	RecipeListItemKey = "recipe_list_item"
	// RecipeListItemIDKey is the standard key for referring to a recipe list item's ID.
	RecipeListItemIDKey = RecipeListItemKey + idSuffix

	// RecipeMediaKey is the standard key for referring to a recipe media.
	RecipeMediaKey = "recipe_media"
	// RecipeMediaIDKey is the standard key for referring to a recipe media's ID.
	RecipeMediaIDKey = RecipeMediaKey + idSuffix

	// UploadedMediaKey is the standard key for referring to an uploaded media object.
	UploadedMediaKey = "uploaded_media"
	// UploadedMediaIDKey is the standard key for referring to an uploaded media object's ID.
	UploadedMediaIDKey = UploadedMediaKey + idSuffix

	// RecipePrepTaskKey is the standard key for referring to a recipe prep task.
	RecipePrepTaskKey = "recipe_prep_task"
	// RecipePrepTaskIDKey is the standard key for referring to a recipe prep task's ID.
	RecipePrepTaskIDKey = RecipePrepTaskKey + idSuffix

	// RecipePrepTaskStepKey is the standard key for referring to a recipe prep task step.
	RecipePrepTaskStepKey = "recipe_prep_task_step"
	// RecipePrepTaskStepIDKey is the standard key for referring to a recipe prep task step's ID.
	RecipePrepTaskStepIDKey = RecipePrepTaskStepKey + idSuffix

	// RecipeRatingKey is the standard key for referring to a recipe rating.
	RecipeRatingKey = "recipe_rating"
	// RecipeRatingIDKey is the standard key for referring to a recipe rating's ID.
	RecipeRatingIDKey = RecipeRatingKey + idSuffix

	// RecipeStepKey is the standard key for referring to a recipe step.
	RecipeStepKey = "recipe_step"
	// RecipeStepIDKey is the standard key for referring to a recipe step's ID.
	RecipeStepIDKey = RecipeStepKey + idSuffix

	// RecipeStepCompletionConditionKey is the standard key for referring to a recipe step completion condition.
	RecipeStepCompletionConditionKey = "recipe_step_completion_condition"
	// RecipeStepCompletionConditionIDKey is the standard key for referring to a recipe step completion condition's ID.
	RecipeStepCompletionConditionIDKey = RecipeStepCompletionConditionKey + idSuffix

	// RecipeStepCompletionConditionIngredientKey is the standard key for referring to a recipe step completion condition ingredient.
	RecipeStepCompletionConditionIngredientKey = "recipe_step_completion_condition_ingredient"
	// RecipeStepCompletionConditionIngredientIDKey is the standard key for referring to a recipe step completion condition ingredient's ID.
	RecipeStepCompletionConditionIngredientIDKey = RecipeStepCompletionConditionIngredientKey + idSuffix

	// RecipeStepIngredientKey is the standard key for referring to a recipe step ingredient.
	RecipeStepIngredientKey = "recipe_step_ingredient"
	// RecipeStepIngredientIDKey is the standard key for referring to a recipe step ingredient's ID.
	RecipeStepIngredientIDKey = RecipeStepIngredientKey + idSuffix

	// RecipeStepInstrumentKey is the standard key for referring to a recipe step instrument.
	RecipeStepInstrumentKey = "recipe_step_instrument"
	// RecipeStepInstrumentIDKey is the standard key for referring to a recipe step instrument's ID.
	RecipeStepInstrumentIDKey = RecipeStepInstrumentKey + idSuffix

	// RecipeStepProductKey is the standard key for referring to a recipe step product.
	RecipeStepProductKey = "recipe_step_product"
	// RecipeStepProductIDKey is the standard key for referring to a recipe step product's ID.
	RecipeStepProductIDKey = RecipeStepProductKey + idSuffix

	// RecipeStepVesselKey is the standard key for referring to a recipe step vessel.
	RecipeStepVesselKey = "recipe_step_vessel"
	// RecipeStepVesselIDKey is the standard key for referring to a recipe step vessel's ID.
	RecipeStepVesselIDKey = RecipeStepVesselKey + idSuffix

	// UserIngredientPreferenceKey is the standard key for referring to a user ingredient preference.
	UserIngredientPreferenceKey = "user_ingredient_preference"
	// UserIngredientPreferenceIDKey is the standard key for referring to a user ingredient preference's ID.
	UserIngredientPreferenceIDKey = UserIngredientPreferenceKey + idSuffix

	// ValidIngredientKey is the standard key for referring to a valid ingredient.
	ValidIngredientKey = "valid_ingredient"
	// ValidIngredientIDKey is the standard key for referring to a valid ingredient's ID.
	ValidIngredientIDKey = ValidIngredientKey + idSuffix

	// ValidIngredientGroupKey is the standard key for referring to a valid ingredient group.
	ValidIngredientGroupKey = "valid_ingredient_group"
	// ValidIngredientGroupIDKey is the standard key for referring to a valid ingredient group's ID.
	ValidIngredientGroupIDKey = ValidIngredientGroupKey + idSuffix

	// ValidIngredientMeasurementUnitKey is the standard key for referring to a valid ingredient measurement unit.
	ValidIngredientMeasurementUnitKey = "valid_ingredient_measurement_unit"
	// ValidIngredientMeasurementUnitIDKey is the standard key for referring to a valid ingredient measurement unit's ID.
	ValidIngredientMeasurementUnitIDKey = ValidIngredientMeasurementUnitKey + idSuffix

	// ValidIngredientPreparationKey is the standard key for referring to a valid ingredient preparation.
	ValidIngredientPreparationKey = "valid_ingredient_preparation"
	// ValidIngredientPreparationIDKey is the standard key for referring to a valid ingredient preparation's ID.
	ValidIngredientPreparationIDKey = ValidIngredientPreparationKey + idSuffix

	// ValidIngredientStateKey is the standard key for referring to a valid ingredient state.
	ValidIngredientStateKey = "valid_ingredient_state"
	// ValidIngredientStateIDKey is the standard key for referring to a valid ingredient state's ID.
	ValidIngredientStateIDKey = ValidIngredientStateKey + idSuffix

	// ValidIngredientStateIngredientKey is the standard key for referring to a valid ingredient state ingredient.
	ValidIngredientStateIngredientKey = "valid_ingredient_state_ingredient"
	// ValidIngredientStateIngredientIDKey is the standard key for referring to a valid ingredient state ingredient's ID.
	ValidIngredientStateIngredientIDKey = ValidIngredientStateIngredientKey + idSuffix

	// ValidInstrumentKey is the standard key for referring to a valid instrument.
	ValidInstrumentKey = "valid_instrument"
	// ValidInstrumentIDKey is the standard key for referring to a valid instrument's ID.
	ValidInstrumentIDKey = ValidInstrumentKey + idSuffix

	// ValidMeasurementUnitKey is the standard key for referring to a valid measurement unit.
	ValidMeasurementUnitKey = "valid_measurement_unit"
	// ValidMeasurementUnitIDKey is the standard key for referring to a valid measurement unit's ID.
	ValidMeasurementUnitIDKey = ValidMeasurementUnitKey + idSuffix

	// ValidMeasurementUnitConversionKey is the standard key for referring to a valid measurement unit conversion.
	ValidMeasurementUnitConversionKey = "valid_measurement_unit_conversion"
	// ValidMeasurementUnitConversionIDKey is the standard key for referring to a valid measurement unit conversion's ID.
	ValidMeasurementUnitConversionIDKey = ValidMeasurementUnitConversionKey + idSuffix

	// ValidPrepTaskConfigKey is the standard key for referring to a valid prep task config.
	ValidPrepTaskConfigKey = "valid_prep_task_config"
	// ValidPrepTaskConfigIDKey is the standard key for referring to a valid prep task config's ID.
	ValidPrepTaskConfigIDKey = ValidPrepTaskConfigKey + idSuffix

	// ValidPreparationKey is the standard key for referring to a valid preparation.
	ValidPreparationKey = "valid_preparation"
	// ValidPreparationIDKey is the standard key for referring to a valid preparation's ID.
	ValidPreparationIDKey = ValidPreparationKey + idSuffix

	// ValidPreparationInstrumentKey is the standard key for referring to a valid preparation instrument.
	ValidPreparationInstrumentKey = "valid_preparation_instrument"
	// ValidPreparationInstrumentIDKey is the standard key for referring to a valid preparation instrument's ID.
	ValidPreparationInstrumentIDKey = ValidPreparationInstrumentKey + idSuffix

	// ValidPreparationVesselKey is the standard key for referring to a valid preparation vessel.
	ValidPreparationVesselKey = "valid_preparation_vessel"
	// ValidPreparationVesselIDKey is the standard key for referring to a valid preparation vessel's ID.
	ValidPreparationVesselIDKey = ValidPreparationVesselKey + idSuffix

	// ValidVesselKey is the standard key for referring to a valid vessel.
	ValidVesselKey = "valid_vessel"
	// ValidVesselIDKey is the standard key for referring to a valid vessel's ID.
	ValidVesselIDKey = ValidVesselKey + idSuffix
)
