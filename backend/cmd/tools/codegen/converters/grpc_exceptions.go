package main

// grpcConversions are the gRPC conversions that exist. This is the only list on the gRPC axis:
// which directions a message is converted in is a fact about the handlers and tests that use it,
// not about its name.
var grpcConversions = [][]grpcDeclaration{
	both("AccountInstrumentOwnership"),
	both("AccountInstrumentOwnershipCreationRequestInput"),
	fromGRPC("AccountInstrumentOwnershipUpdateRequestInput"),
	both("Meal"),
	documented(mealSummaryDoc, toGRPCAs("Meal", "MealSummary")),
	documented(mealSummaryInflationDoc, fromGRPCAs("MealSummary", "Meal")),
	both("MealComponent"),
	toGRPCAs("MealComponent", "MealComponentSummary"),
	fromGRPCAs("MealComponentSummary", "MealComponent"),
	both("MealComponentCreationRequestInput"),
	both("MealCreationRequestInput"),
	toGRPC("MealList"),
	fromGRPC("MealListCreationRequestInput"),
	toGRPC("MealListItem"),
	fromGRPC("MealListItemCreationRequestInput"),
	fromGRPC("MealListItemUpdateRequestInput"),
	fromGRPC("MealListUpdateRequestInput"),
	both("MealPlan"),
	fromGRPCAs("MealPlanSummary", "MealPlan"),
	both("MealPlanCreationRequestInput"),
	both("MealPlanEvent"),
	toGRPCAs("MealPlanEvent", "MealPlanEventSummary"),
	fromGRPCAs("MealPlanEventSummary", "MealPlanEvent"),
	both("MealPlanEventCreationRequestInput"),
	both("MealPlanEventUpdateRequestInput"),
	both("MealPlanGroceryListItem"),
	both("MealPlanGroceryListItemCreationRequestInput"),
	fromGRPC("MealPlanGroceryListItemUpdateRequestInput"),
	both("MealPlanOption"),
	both("MealPlanOptionCreationRequestInput"),
	both("MealPlanOptionUpdateRequestInput"),
	both("MealPlanOptionVote"),
	both("MealPlanOptionVoteCreationInput"),
	both("MealPlanOptionVoteCreationRequestInput"),
	both("MealPlanOptionVoteUpdateRequestInput"),
	both("MealPlanRecipeOptionSelection"),
	both("MealPlanRecipeOptionSelectionCreationRequestInput"),
	fromGRPC("MealPlanRecipeOptionSelectionUpdateRequestInput"),
	both("MealPlanTask"),
	both("MealPlanTaskCreationRequestInput"),
	toGRPCAs("MealPlanTaskDatabaseCreationEstimate", "MealPlanTask"),
	fromGRPC("MealPlanTaskStatusChangeRequestInput"),
	fromGRPC("MealPlanUpdateRequestInput"),
	toGRPC("MeasurementUnitConversionMismatch"),
	both("Recipe"),
	documented(recipeSummaryDoc, toGRPCAs("Recipe", "RecipeSummary")),
	documented(recipeSummaryInflationDoc, fromGRPCAs("RecipeSummary", "Recipe")),
	both("RecipeCreationRequestInput"),
	toGRPC("RecipeList"),
	fromGRPC("RecipeListCreationRequestInput"),
	toGRPC("RecipeListItem"),
	fromGRPC("RecipeListItemCreationRequestInput"),
	fromGRPC("RecipeListItemUpdateRequestInput"),
	fromGRPC("RecipeListUpdateRequestInput"),
	both("RecipeMedia"),
	both("RecipeMediaCreationRequestInput"),
	both("RecipePrepTask"),
	both("RecipePrepTaskCreationRequestInput"),
	both("RecipePrepTaskStep"),
	both("RecipePrepTaskStepCreationRequestInput"),
	both("RecipePrepTaskStepUpdateRequestInput"),
	both("RecipePrepTaskStepWithinRecipeCreationRequestInput"),
	both("RecipePrepTaskUpdateRequestInput"),
	both("RecipePrepTaskWithinRecipeCreationRequestInput"),
	both("RecipeRating"),
	both("RecipeRatingCreationRequestInput"),
	both("RecipeRatingUpdateRequestInput"),
	both("RecipeStep"),
	both("RecipeStepCompletionCondition"),
	both("RecipeStepCompletionConditionCreationRequestInput"),
	both("RecipeStepCompletionConditionForExistingRecipeCreationRequestInput"),
	both("RecipeStepCompletionConditionIngredient"),
	toGRPCAs("RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput", "RecipeStepCompletionConditionIngredient"),
	both("RecipeStepCompletionConditionUpdateRequestInput"),
	both("RecipeStepCreationRequestInput"),
	both("RecipeStepIngredient"),
	both("RecipeStepIngredientCreationRequestInput"),
	both("RecipeStepIngredientUpdateRequestInput"),
	both("RecipeStepInstrument"),
	both("RecipeStepInstrumentCreationRequestInput"),
	both("RecipeStepInstrumentUpdateRequestInput"),
	both("RecipeStepProduct"),
	both("RecipeStepProductCreationRequestInput"),
	both("RecipeStepProductUpdateRequestInput"),
	both("RecipeStepUpdateRequestInput"),
	both("RecipeStepVessel"),
	both("RecipeStepVesselCreationRequestInput"),
	both("RecipeStepVesselUpdateRequestInput"),
	both("RecipeUpdateRequestInput"),
	both("UserIngredientPreference"),
	both("UserIngredientPreferenceCreationRequestInput"),
	fromGRPC("UserIngredientPreferenceUpdateRequestInput"),
	both("ValidIngredient"),
	both("ValidIngredientCreationRequestInput"),
	both("ValidIngredientGroup"),
	both("ValidIngredientGroupCreationRequestInput"),
	both("ValidIngredientGroupUpdateRequestInput"),
	both("ValidIngredientMeasurementUnit"),
	both("ValidIngredientMeasurementUnitCreationRequestInput"),
	both("ValidIngredientMeasurementUnitUpdateRequestInput"),
	both("ValidIngredientPreparation"),
	both("ValidIngredientPreparationCreationRequestInput"),
	both("ValidIngredientPreparationUpdateRequestInput"),
	both("ValidIngredientState"),
	both("ValidIngredientStateCreationRequestInput"),
	both("ValidIngredientStateIngredient"),
	both("ValidIngredientStateIngredientCreationRequestInput"),
	both("ValidIngredientStateIngredientUpdateRequestInput"),
	both("ValidIngredientStateUpdateRequestInput"),
	both("ValidIngredientUpdateRequestInput"),
	both("ValidInstrument"),
	both("ValidInstrumentCreationRequestInput"),
	both("ValidInstrumentUpdateRequestInput"),
	both("ValidMeasurementUnit"),
	both("ValidMeasurementUnitConversion"),
	both("ValidMeasurementUnitConversionCreationRequestInput"),
	both("ValidMeasurementUnitConversionUpdateRequestInput"),
	both("ValidMeasurementUnitCreationRequestInput"),
	both("ValidMeasurementUnitUpdateRequestInput"),
	both("ValidPreparation"),
	both("ValidPreparationCreationRequestInput"),
	both("ValidPreparationInstrument"),
	both("ValidPreparationInstrumentCreationRequestInput"),
	both("ValidPreparationInstrumentUpdateRequestInput"),
	both("ValidPreparationUpdateRequestInput"),
	both("ValidPreparationVessel"),
	both("ValidPreparationVesselCreationRequestInput"),
	both("ValidPreparationVesselUpdateRequestInput"),
	both("ValidPrepTaskConfig"),
	both("ValidPrepTaskConfigCreationRequestInput"),
	both("ValidPrepTaskConfigUpdateRequestInput"),
	both("ValidVessel"),
	both("ValidVesselCreationRequestInput"),
	both("ValidVesselUpdateRequestInput"),
}

const (
	mealSummaryDoc = "The components survive — a meal without them says almost nothing — but each carries a RecipeSummary rather than a whole Recipe, which is what keeps a max-limit page inside the 4 MiB gRPC message bound. See the MealSummary comment in mealplanning_messages.proto."

	mealSummaryInflationDoc = "The result's components carry recipes with empty nested collections. Fetch the meal by ID for components carrying full recipes."

	recipeSummaryDoc = "Everything that hangs off the recipe — steps, prep tasks, media, associated recipes — is dropped, which is what keeps a max-limit page inside the 4 MiB gRPC message bound. See the RecipeSummary comment in mealplanning_messages.proto."

	recipeSummaryInflationDoc = "The summary carries no steps, prep tasks, media, or associated recipes, so neither does the result — fetch the recipe by ID for those."
)

// grpcFieldExceptions are the destination fields of a gRPC conversion that the derivation does not
// answer, or answers differently from the converter this replaced, keyed by conversion and field.
var grpcFieldExceptions = map[string]map[string]Rule{
	"ConvertGRPCMealPlanSummaryToMealPlan": {
		fieldSelections: Skip(summaryOmits),
		"Events":        EmptySlice(summaryInflatesEmpty),
	},
	"ConvertGRPCMealPlanEventSummaryToMealPlanEvent": {
		"Options": Expr("[]*mealplanning.MealPlanOption{}", summaryInflatesEmpty),
	},
	"ConvertMealPlanEventToGRPCMealPlanEventSummary": {
		"ChosenMealName": Skip("The domain event has no chosen meal name of its own. ConvertMealPlanToGRPCMealPlanSummary stamps it from the summary annotations after this."),
	},
	"ConvertMealPlanRecipeOptionSelectionCreationRequestInputToGRPCMealPlanRecipeOptionSelectionCreationRequestInput": {
		fieldBelongsToMealPlanOption: Skip("The domain creation input has no owning option to read it from. " + parentStampedByCaller),
	},
	"ConvertGRPCMealPlanTaskToMealPlanTask": {
		"NotificationSentAt": Skip("The message has no field for when a task's notification went out, so there is nothing to read it from."),
	},
	"ConvertMealPlanTaskDatabaseCreationEstimateToGRPCMealPlanTask": {
		"RecipePrepTask":    Skip(estimateOnlyExplains),
		"CreatedAt":         Skip(estimateOnlyExplains),
		"LastUpdatedAt":     Skip(estimateOnlyExplains),
		"CompletedAt":       Skip(estimateOnlyExplains),
		"AssignedToUser":    Skip(estimateOnlyExplains),
		"Id":                Skip(estimateOnlyExplains),
		"Status":            Skip(estimateOnlyExplains),
		"StatusExplanation": Skip(estimateOnlyExplains),
		"MealPlanOption":    Skip(estimateOnlyExplains),
	},
	"ConvertGRPCMealPlanTaskStatusChangeRequestInputToMealPlanTaskStatusChangeRequestInput": {
		"MealPlanTaskID": From("Id", "The message names the task it changes Id."),
	},
	"ConvertGRPCRecipeToRecipe": {
		"SealOfApproval":    Skip(sealOfApprovalNotOnTheWire),
		"Steps":             EmptySlice(recipeCollectionsEmpty),
		fieldMedia:          EmptySlice(recipeCollectionsEmpty),
		"PrepTasks":         EmptySlice(recipeCollectionsEmpty),
		"AssociatedRecipes": EmptySlice(recipeCollectionsEmpty),
	},
	"ConvertGRPCRecipeSummaryToRecipe": {
		"SealOfApproval":    Skip(sealOfApprovalNotOnTheWire),
		"Steps":             Expr("[]*mealplanning.RecipeStep{}", summaryInflatesEmpty),
		fieldMedia:          Expr("[]*mealplanning.RecipeMedia{}", summaryInflatesEmpty),
		"PrepTasks":         Expr("[]*mealplanning.RecipePrepTask{}", summaryInflatesEmpty),
		"AssociatedRecipes": Expr("[]*mealplanning.Recipe{}", summaryInflatesEmpty),
	},
	"ConvertRecipeRatingToGRPCRecipeRating": {
		fieldMessageRecipeID: From(fieldBelongsToRecipe, ratingRecipeRenamed),
		"ByUser":             From(fieldCreatedByUser, ratingUserRenamed),
	},
	"ConvertGRPCRecipeRatingToRecipeRating": {
		fieldBelongsToRecipe: From(fieldMessageRecipeID, ratingRecipeRenamed),
		fieldCreatedByUser:   From("ByUser", ratingUserRenamed),
	},
	"ConvertRecipeRatingCreationRequestInputToGRPCRecipeRatingCreationRequestInput": {
		fieldMessageRecipeID: From(fieldBelongsToRecipe, ratingRecipeRenamed),
		"ByUser":             From(fieldCreatedByUser, ratingUserRenamed),
	},
	"ConvertGRPCRecipeRatingCreationRequestInputToRecipeRatingCreationRequestInput": {
		fieldBelongsToRecipe: From(fieldMessageRecipeID, ratingRecipeRenamed),
		fieldCreatedByUser:   From("ByUser", ratingUserRenamed),
	},
	"ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput": {
		fieldMessageRecipeID: From(fieldBelongsToRecipe, ratingRecipeRenamed),
	},
	"ConvertGRPCRecipeRatingUpdateRequestInputToRecipeRatingUpdateRequestInput": {
		fieldBelongsToRecipe: From(fieldMessageRecipeID, ratingRecipeRenamed),
	},
	"ConvertGRPCRecipeStepToRecipeStep": {
		"StepImages": Skip(uploadedMediaOneWay),
	},
	"ConvertGRPCValidIngredientToValidIngredient": {
		fieldMedia: Skip(uploadedMediaOneWay),
	},
	"ConvertGRPCValidPreparationToValidPreparation": {
		fieldMedia: Skip(uploadedMediaOneWay),
	},
	"ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionIngredient": {
		"CreatedAt":                              Skip(conditionIngredientInputNamesOnlyTheIngredient),
		"ArchivedAt":                             Skip(conditionIngredientInputNamesOnlyTheIngredient),
		"LastUpdatedAt":                          Skip(conditionIngredientInputNamesOnlyTheIngredient),
		"Id":                                     Skip(conditionIngredientInputNamesOnlyTheIngredient),
		"BelongsToRecipeStepCompletionCondition": Skip(conditionIngredientInputNamesOnlyTheIngredient),
	},
	"ConvertRecipeStepIngredientCreationRequestInputToGRPCRecipeStepIngredientCreationRequestInput": {
		"RecipeStepProductRecipeSlug": Skip("The domain input names a product recipe by ID only and has no slug to read. Nothing on the server reads the message's slug either."),
	},
	"ConvertGRPCMealPlanOptionCreationRequestInputToMealPlanOptionCreationRequestInput": {
		fieldSelections: EmptySlice(emptyAsBefore),
	},
	"ConvertGRPCMealSummaryToMeal": {
		"Components": EmptySlice(summaryInflatesEmpty),
	},
	"ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInput": {
		"Ingredients": EmptySlice(emptyAsBefore),
	},
	"ConvertMealPlanToGRPCMealPlan": {
		fieldSelections: Skip(leftUnsetBefore),
	},
	"ConvertGRPCMealPlanToMealPlan": {
		fieldSelections: Skip(leftUnsetBefore),
	},
	"ConvertMealPlanOptionCreationRequestInputToGRPCMealPlanOptionCreationRequestInput": {
		fieldSelections: Skip(leftUnsetBefore),
	},
	"ConvertRecipeMediaToGRPCRecipeMedia": {
		fieldBelongsToRecipeStep: Skip(leftUnsetBefore),
	},
	"ConvertGRPCRecipeMediaToRecipeMedia": {
		fieldBelongsToRecipeStep: Skip(leftUnsetBefore),
	},
	"ConvertGRPCRecipeMediaCreationRequestInputToRecipeMediaCreationRequestInput": {
		fieldBelongsToRecipe:     Expr("new(x.BelongsToRecipe)", pointsAtACopyOfTheMessage),
		fieldBelongsToRecipeStep: Expr("new(x.BelongsToRecipeStep)", pointsAtACopyOfTheMessage),
	},
	"ConvertGRPCRecipeStepIngredientCreationRequestInputToRecipeStepIngredientCreationRequestInput": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.GetScaleFactor())", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentCreationRequestInput": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.GetScaleFactor())", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepVesselCreationRequestInputToRecipeStepVesselCreationRequestInput": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.GetScaleFactor())", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepIngredientToRecipeStepIngredient": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.ScaleFactor)", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.ScaleFactor)", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepVesselToRecipeStepVessel": {
		fieldScaleFactor: Expr("scaleFactorOrOne(x.ScaleFactor)", scaleFactorDefaultedToOne),
	},
	"ConvertGRPCRecipeStepIngredientUpdateRequestInputToRecipeStepIngredientUpdateRequestInput": {
		fieldScaleFactor: Expr("scaleFactorPointerOrOne(x.ScaleFactor)", scaleFactorUpdateDefaultedToOne),
	},
	"ConvertGRPCRecipeStepInstrumentUpdateRequestInputToRecipeStepInstrumentUpdateRequestInput": {
		fieldScaleFactor: Expr("scaleFactorPointerOrOne(x.ScaleFactor)", scaleFactorUpdateDefaultedToOne),
	},
	"ConvertGRPCRecipeStepVesselUpdateRequestInputToRecipeStepVesselUpdateRequestInput": {
		fieldScaleFactor: Expr("scaleFactorPointerOrOne(x.ScaleFactor)", scaleFactorUpdateDefaultedToOne),
	},
	"ConvertGRPCRecipeStepProductToRecipeStepProduct": {
		"MeasurementUnit": Unguarded(unguardedBefore),
	},
	"ConvertGRPCRecipeStepUpdateRequestInputToRecipeStepUpdateRequestInput": {
		"Preparation": Unguarded(unguardedBefore),
	},
	"ConvertRecipeStepUpdateRequestInputToGRPCRecipeStepUpdateRequestInput": {
		"Preparation": Unguarded(unguardedBefore),
	},
	"ConvertGRPCUserIngredientPreferenceUpdateRequestInputToUserIngredientPreferenceUpdateRequestInput": {
		"Rating": Expr("new(int8(pointer.Dereference(x.Rating)))", "An absent rating becomes a rating of zero rather than staying absent, so an update that does not mention the rating sets it to zero. Preserved rather than corrected, so that generating this converter is not also a behavior change."),
	},
	"ConvertUserIngredientPreferenceToGRPCUserIngredientPreference": {
		"BelongsToUser": From(fieldCreatedByUser, preferenceUserRenamed),
	},
	"ConvertGRPCUserIngredientPreferenceToUserIngredientPreference": {
		fieldCreatedByUser: From("BelongsToUser", preferenceUserRenamed),
	},
}

const (
	fieldBelongsToRecipe = "BelongsToRecipe"
	// fieldMessageRecipeID is the message's spelling, which protoc-gen-go gives an initialism.
	fieldMessageRecipeID = "RecipeId"
)

const (
	leftUnsetBefore                 = "The converter this replaced left it unset. Preserved rather than corrected, so that generating these converters is not also a behavior change."
	emptyAsBefore                   = "Built from an empty slice rather than a nil one, as the converter this replaced built it."
	unguardedBefore                 = "Read without a nil check, as the converter this replaced read it: an absent one panics. Preserved rather than corrected, so that generating these converters is not also a behavior change."
	pointsAtACopyOfTheMessage       = "Points at a copy rather than into the message, as the converter this replaced did."
	scaleFactorDefaultedToOne       = "A scale factor that is absent, zero or negative means an unscaled one, and scaleFactorOrOne is where that is decided."
	scaleFactorUpdateDefaultedToOne = "A scale factor that is absent, zero or negative becomes one rather than staying absent, so an update that does not mention it resets it. Preserved rather than corrected, so that generating these converters is not also a behavior change."
)

const (
	summaryOmits                                   = "A summary does not carry it, which is what keeps a page of summaries inside the gRPC message bound."
	summaryInflatesEmpty                           = "A summary does not carry the collection, so the result has an empty one rather than a nil one, as the converter this replaced built it."
	recipeCollectionsEmpty                         = "Built from an empty slice rather than a nil one, as the converter this replaced built it: a recipe with no steps reads back with an empty collection, not a missing one."
	estimateOnlyExplains                           = "An estimate is a task that has not been created yet, and the only thing it carries is the explanation of why it would be."
	sealOfApprovalNotOnTheWire                     = "The message has no seal of approval field, so there is nothing to read it from."
	ratingRecipeRenamed                            = "The message names the rated recipe RecipeId; the domain calls it BelongsToRecipe."
	ratingUserRenamed                              = "The message names the rating's author ByUser; the domain calls it CreatedByUser."
	preferenceUserRenamed                          = "The message names the preference's owner BelongsToUser; the domain calls it CreatedByUser."
	uploadedMediaOneWay                            = "Platform's mediaregistry/grpc renders an uploaded object for the wire but has no conversion back, and the converter this replaced left it unset rather than write one here."
	conditionIngredientInputNamesOnlyTheIngredient = "A condition ingredient for an existing recipe is only the recipe step ingredient it names; the rest belongs to a row that does not exist yet."
)

// grpcHandWritten are the declared gRPC conversions this tool does not generate, and why. Each is
// in converters_manual.go beside the generated file. The bar is the domain axis's: a body that is
// not a set of field assignments.
var grpcHandWritten = map[string]string{
	"ConvertGRPCMealListUpdateRequestInputToMealListUpdateRequestInput":                                                                                 returnsNilForNil,
	"ConvertGRPCMealListItemUpdateRequestInputToMealListItemUpdateRequestInput":                                                                         returnsNilForNil,
	"ConvertGRPCRecipeListUpdateRequestInputToRecipeListUpdateRequestInput":                                                                             returnsNilForNil,
	"ConvertGRPCRecipeListItemUpdateRequestInputToRecipeListItemUpdateRequestInput":                                                                     returnsNilForNil,
	"ConvertGRPCMealPlanRecipeOptionSelectionUpdateRequestInputToMealPlanRecipeOptionSelectionUpdateRequestInput":                                       returnsNilForNil,
	"ConvertMeasurementUnitConversionMismatchToGRPCMeasurementUnitConversionMismatch":                                                                   returnsNilForNil,
	"ConvertValidIngredientGroupToGRPCValidIngredientGroup":                                                                                             inlineGRPCGroupMembers,
	"ConvertGRPCValidIngredientGroupToValidIngredientGroup":                                                                                             inlineGRPCGroupMembers,
	"ConvertValidIngredientGroupCreationRequestInputToGRPCValidIngredientGroupCreationRequestInput":                                                     inlineGRPCGroupMembers,
	"ConvertGRPCValidIngredientGroupCreationRequestInputToValidIngredientGroupCreationRequestInput":                                                     inlineGRPCGroupMembers,
	"ConvertGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput": "It builds each ingredient inline, from a message that has no conversion of its own into the domain's ingredient input.",
}

const (
	inlineGRPCGroupMembers = "It builds each member inline into a slice made to the members' length, so an empty group is an empty slice rather than a nil one, and no member has a conversion of its own."
)
