package mealplanning

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: AccountInstrumentOwnership{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "BelongsToAccount", Expr: `buildUniqueString()`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: AccountInstrumentOwnershipUpdateRequestInput{}, Converter: "ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipUpdateRequestInput"},
					{Type: AccountInstrumentOwnershipCreationRequestInput{}, Converter: "ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput"},
				},
			},
		},
		{
			Type: Meal{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `recipes := []*types.MealComponent{}`},
					{Code: `for range exampleQuantity {
	recipes = append(recipes, BuildFakeMealComponent())
}`},
				},
				Fields: []entitydecl.Field{
					{Name: "Components", Expr: `recipes`},
					{Name: "EligibleForMealPlans", Expr: `true`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealCreationRequestInput{}, Converter: "ConvertMealToMealCreationRequestInput"},
				},
			},
		},
		{
			Type: MealComponent{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "RecipeScale", Expr: `float32(1.0)`},
					{Name: "ComponentType", Expr: `types.MealComponentTypesMain`},
				},
			},
		},
		{
			Type: MealPlan{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `mealPlanID := BuildFakeID()`},
					{Code: `var events []*types.MealPlanEvent`},
					{Code: `for range exampleQuantity {
	event := BuildFakeMealPlanEvent()
	event.BelongsToMealPlan = mealPlanID
	events = append(events, event)
}`},
					{
						Code: `votingDeadline := time.Now().Add(5 * time.Minute).Truncate(time.Second).UTC()`,
						Why:  "The voting deadline must be in the future but before every event's start time (events start in ten minutes, see BuildFakeMealPlanEvent), so the meal plan passes MealPlanCreationRequestInput validation.",
					},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `mealPlanID`},
					{Name: "Status", Expr: `string(types.MealPlanStatusAwaitingVotes)`},
					{Name: "VotingDeadline", Expr: `votingDeadline`},
					{Name: "BelongsToAccount", Expr: `fake.UUID()`},
					{Name: "TasksCreated", Expr: `false`},
					{Name: "GroceryListInitialized", Expr: `false`},
					{Name: "ElectionMethod", Expr: `types.MealPlanElectionMethodSchulze`},
					{Name: "Events", Expr: `events`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanUpdateRequestInput{}, Converter: "ConvertMealPlanToMealPlanUpdateRequestInput"},
					{Type: MealPlanCreationRequestInput{}, Converter: "ConvertMealPlanToMealPlanCreationRequestInput"},
				},
			},
		},
		{
			Type: MealPlanEvent{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `mealPlanEventID := BuildFakeID()`},
					{Code: `now := time.Now().Add(0).Truncate(time.Second).UTC()`},
					{Code: `inTenMinutes := now.Add(time.Minute * 10).Add(0).Truncate(time.Second).UTC()`},
					{Code: `inOneWeek := now.Add((time.Hour * 24) * 7).Add(0).Truncate(time.Second).UTC()`},
					{Code: `options := []*types.MealPlanOption{}`},
					{Code: `for _, opt := range BuildFakeMealPlanOptionsList().Data {
	opt.BelongsToMealPlanEvent = mealPlanEventID
	options = append(options, opt)
}`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `mealPlanEventID`},
					{Name: "StartsAt", Expr: `inTenMinutes`},
					{Name: "EndsAt", Expr: `inOneWeek`},
					{Name: "MealName", Expr: `fake.RandomString([]string{
	types.BreakfastMealName,
	types.SecondBreakfastMealName,
	types.BrunchMealName,
	types.LunchMealName,
	types.SupperMealName,
	types.DinnerMealName,
})`},
					{Name: "Options", Expr: `options`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanEventCreationRequestInput{}, Converter: "ConvertMealPlanEventToMealPlanEventCreationRequestInput"},
				},
			},
		},
		{
			Type: MealPlanEventUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `mealPlanEvent := BuildFakeMealPlanEvent()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&mealPlanEvent.Notes`},
					{Name: "StartsAt", Expr: `&mealPlanEvent.StartsAt`},
					{Name: "EndsAt", Expr: `&mealPlanEvent.EndsAt`},
					{Name: "MealName", Expr: `&mealPlanEvent.MealName`},
					{Name: "BelongsToMealPlan", Expr: `mealPlanEvent.BelongsToMealPlan`},
				},
			},
		},
		{
			Type: MealPlanGroceryListItem{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minQty, maxQty := BuildFakeFloat32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinQuantityNeeded", Expr: `minQty`},
					{Name: "MaxQuantityNeeded", Expr: `maxQty`},
					{Name: "Status", Expr: `types.MealPlanGroceryListItemStatusUnknown`},
					{
						Name: "BelongsToMealPlanOption",
						Expr: `nil`,
						Why:  "Recipe context fields (optional - only set when item is part of a choice group)",
					},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanGroceryListItemCreationRequestInput{}, Converter: "ConvertMealPlanGroceryListItemToMealPlanGroceryListItemCreationRequestInput"},
					{Type: MealPlanGroceryListItemUpdateRequestInput{}, Converter: "ConvertMealPlanGroceryListItemToMealPlanGroceryListItemUpdateRequestInput"},
				},
			},
		},
		{
			Type: MealPlanOption{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `var examples []*types.MealPlanOptionVote`},
					{Code: `for range exampleQuantity {
	examples = append(examples, BuildFakeMealPlanOptionVote())
}`},
					{Code: `meal := BuildFakeMeal()`},
					{Code: `meal.Components = nil`},
				},
				Fields: []entitydecl.Field{
					{Name: "AssignedCook", Expr: `func(s string) *string { return &s }(BuildFakeID())`},
					{Name: "Meal", Expr: `*meal`},
					{Name: "Votes", Expr: `examples`},
					{Name: "Chosen", Expr: `false`},
					{Name: "BelongsToMealPlanEvent", Expr: `fake.UUID()`},
					{Name: "MealScale", Expr: `0`},
					{Name: "TieBroken", Expr: `false`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanOptionUpdateRequestInput{}, Converter: "ConvertMealPlanOptionToMealPlanOptionUpdateRequestInput"},
					{Type: MealPlanOptionCreationRequestInput{}, Converter: "ConvertMealPlanOptionToMealPlanOptionCreationRequestInput"},
				},
			},
		},
		{
			Type: MealPlanOptionVote{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Rank", Expr: `uint8(fake.Number(1, math.MaxUint8))`},
					{Name: "BelongsToMealPlanOption", Expr: `fake.UUID()`},
					{Name: "ByUser", Expr: `""`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanOptionVoteCreationRequestInput{}, Converter: "ConvertMealPlanOptionVoteToMealPlanOptionVoteCreationRequestInput"},
					{Type: MealPlanOptionVotesDatabaseCreationInput{}, Converter: "ConvertMealPlanOptionVoteToMealPlanOptionVotesDatabaseCreationInput", Name: "BuildFakeMealPlanOptionVoteDatabaseCreationInput"},
				},
			},
		},
		{
			Type: MealPlanOptionVoteUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `mealPlanOptionVote := BuildFakeMealPlanOptionVote()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Rank", Expr: `&mealPlanOptionVote.Rank`},
					{Name: "Abstain", Expr: `&mealPlanOptionVote.Abstain`},
					{Name: "Notes", Expr: `&mealPlanOptionVote.Notes`},
					{Name: "BelongsToMealPlanOption", Expr: `mealPlanOptionVote.BelongsToMealPlanOption`},
				},
			},
		},
		{
			Type: MealPlanRecipeOptionSelection{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "IngredientIndex", Expr: `fake.Uint16()`},
					{Name: "SelectedOptionIndex", Expr: `fake.Uint16()`},
					{Name: "SelectionType", Expr: `types.MealPlanRecipeOptionSelectionTypeIngredient`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanRecipeOptionSelectionDatabaseCreationInput{}, Converter: "ConvertMealPlanRecipeOptionSelectionToMealPlanRecipeOptionSelectionDatabaseCreationInput"},
				},
			},
		},
		{
			Type: MealPlanRecipeOptionSelectionUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `selectedOptionIndex := fake.Uint16()`},
				},
				Fields: []entitydecl.Field{
					{Name: "SelectedOptionIndex", Expr: `&selectedOptionIndex`},
				},
			},
		},
		{
			Type: MealPlanRecipeOptionSelectionCreationRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "SelectionType", Expr: `types.MealPlanRecipeOptionSelectionTypeIngredient`},
					{Name: "IngredientIndex", Expr: `0`},
					{Name: "SelectedOptionIndex", Expr: `0`},
				},
			},
		},
		{
			Type: MealPlanTask{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Status", Expr: `"unfinished"`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanTaskCreationRequestInput{}, Converter: "ConvertMealPlanTaskToMealPlanTaskCreationRequestInput"},
				},
			},
		},
		{
			Type: MealPlanTaskStatusChangeRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Status", Expr: `new("unfinished")`},
				},
			},
		},
		{
			Type: RecipeRating{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeRatingUpdateRequestInput{}, Converter: "ConvertRecipeRatingToRecipeRatingUpdateRequestInput"},
					{Type: RecipeRatingCreationRequestInput{}, Converter: "ConvertRecipeRatingToRecipeRatingCreationRequestInput"},
				},
			},
		},
		{
			Type: Recipe{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `recipeID := BuildFakeID()`},
					{Code: `var steps []*types.RecipeStep`},
					{Code: `for i := range exampleQuantity {
	step := BuildFakeRecipeStep()
	step.Index = uint32(i)
	step.BelongsToRecipe = recipeID
	steps = append(steps, step)
}`},
					{Code: `prepTasks := BuildFakeRecipePrepTasksList().Data`},
					{Code: `for i := range prepTasks {
	prepTasks[i].BelongsToRecipe = recipeID
}`},
					{Code: `recipeMedia := BuildFakeRecipeMediaList().Data`},
					{Code: `for i := range recipeMedia {
	recipeMedia[i].BelongsToRecipe = &recipeID
}`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `recipeID`},
					{Name: "Steps", Expr: `steps`},
					{Name: "PrepTasks", Expr: `prepTasks`},
					{Name: "Status", Expr: `types.RecipeStatusSubmitted`},
					{Name: "Media", Expr: `recipeMedia`},
					{Name: "MaxEstimatedPortions", Expr: `new(float32(buildFakeNumber()))`},
					{Name: "EligibleForMeals", Expr: `true`},
					{Name: "YieldsComponentType", Expr: `"main"`},
					{Name: "SourceISBN", Expr: `""`},
					{Name: "SealOfApproval", Expr: `false`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeUpdateRequestInput{}, Converter: "ConvertRecipeToRecipeUpdateRequestInput"},
				},
			},
		},
		{
			Type: RecipeMedia{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "MimeType", Expr: `fake.FileMimeType()`},
					{Name: "InternalPath", Expr: `fmt.Sprintf("%s.%s", buildFakePassword(), fake.FileExtension())`},
					{Name: "ExternalPath", Expr: `""`},
					{Name: "Index", Expr: `0`},
				},
				List: &entitydecl.List{Name: "BuildFakeRecipeMediaList"},
				Inputs: []entitydecl.Input{
					{Type: RecipeMediaCreationRequestInput{}, Converter: "ConvertRecipeMediaToRecipeMediaCreationRequestInput"},
				},
			},
		},
		{
			Type: RecipeMediaUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validPreparation := BuildFakeRecipeMedia()`},
				},
				Fields: []entitydecl.Field{
					{Name: "BelongsToRecipe", Expr: `validPreparation.BelongsToRecipe`},
					{Name: "BelongsToRecipeStep", Expr: `validPreparation.BelongsToRecipeStep`},
					{Name: "MimeType", Expr: `&validPreparation.MimeType`},
					{Name: "InternalPath", Expr: `&validPreparation.InternalPath`},
					{Name: "ExternalPath", Expr: `&validPreparation.ExternalPath`},
				},
			},
		},
		{
			Type: RecipePrepTask{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `recipePrepTaskSteps := []*types.RecipePrepTaskStep{}`},
					{Code: `for range exampleQuantity {
	recipePrepTaskSteps = append(recipePrepTaskSteps, BuildFakeRecipePrepTaskStep())
}`},
					{Code: `minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()`},
					{Code: `minBuf, maxBuf := BuildFakeUint32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "StorageType", Expr: `fake.RandomString([]string{
	types.RecipePrepTaskStorageTypeUncovered,
	types.RecipePrepTaskStorageTypeCovered,
	types.RecipePrepTaskStorageTypeAirtightContainer,
	types.RecipePrepTaskStorageTypeWireRack,
})`},
					{Name: "TaskSteps", Expr: `recipePrepTaskSteps`},
					{Name: "MinStorageTemperatureInCelsius", Expr: `minTemp`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxTemp`},
					{Name: "MinTimeBufferBeforeRecipeInSeconds", Expr: `minBuf`},
					{Name: "MaxTimeBufferBeforeRecipeInSeconds", Expr: `maxBuf`},
				},
			},
		},
		{
			Type: RecipePrepTaskStep{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: RecipePrepTaskStepCreationRequestInput{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: RecipePrepTaskStepUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "BelongsToRecipeStep", Expr: `new(BuildFakeID())`},
					{Name: "BelongsToRecipePrepTask", Expr: `new(BuildFakeID())`},
					{Name: "SatisfiesRecipeStep", Expr: `new(fake.Bool())`},
				},
			},
		},
		{
			Type: RecipePrepTaskCreationRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `taskSteps := []*types.RecipePrepTaskStepCreationRequestInput{}`},
					{Code: `for range exampleQuantity {
	taskSteps = append(taskSteps, BuildFakeRecipePrepTaskStepCreationRequestInput())
}`},
					{Code: `minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()`},
					{Code: `minBuf, maxBuf := BuildFakeUint32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "StorageType", Expr: `types.RecipePrepTaskStorageTypeUncovered`},
					{Name: "RecipeSteps", Expr: `taskSteps`},
					{Name: "MinTimeBufferBeforeRecipeInSeconds", Expr: `minBuf`},
					{Name: "MaxTimeBufferBeforeRecipeInSeconds", Expr: `maxBuf`},
					{Name: "MinStorageTemperatureInCelsius", Expr: `minTemp`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxTemp`},
				},
			},
		},
		{
			Type: RecipePrepTaskUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `taskSteps := []*types.RecipePrepTaskStepUpdateRequestInput{}`},
					{Code: `for range exampleQuantity {
	taskSteps = append(taskSteps, BuildFakeRecipePrepTaskStepUpdateRequestInput())
}`},
					{Code: `minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()`},
					{Code: `minBuf, maxBuf := BuildFakeOptionalUint32MinMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `new(buildUniqueString())`},
					{Name: "ExplicitStorageInstructions", Expr: `new(buildUniqueString())`},
					{Name: "Name", Expr: `new(buildUniqueString())`},
					{Name: "Description", Expr: `new(buildUniqueString())`},
					{Name: "Optional", Expr: `new(fake.Bool())`},
					{Name: "StorageType", Expr: `pointer.To(types.RecipePrepTaskStorageTypeUncovered)`},
					{Name: "BelongsToRecipe", Expr: `new(BuildFakeID())`},
					{Name: "MinTimeBufferBeforeRecipeInSeconds", Expr: `minBuf`},
					{Name: "MaxTimeBufferBeforeRecipeInSeconds", Expr: `maxBuf`},
					{Name: "MinStorageTemperatureInCelsius", Expr: `minTemp`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxTemp`},
					{Name: "TaskSteps", Expr: `taskSteps`},
				},
			},
		},
		{
			Type: RecipeStep{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `recipeStepID := BuildFakeID()`},
					{Code: `var ingredients []*types.RecipeStepIngredient`},
					{Code: `for range exampleQuantity {
	ing := BuildFakeRecipeStepIngredient()
	ing.BelongsToRecipeStep = recipeStepID

	ingredients = append(ingredients, ing)
}`},
					{Code: `var instruments []*types.RecipeStepInstrument`},
					{Code: `for range exampleQuantity {
	ing := BuildFakeRecipeStepInstrument()
	ing.BelongsToRecipeStep = recipeStepID

	instruments = append(instruments, ing)
}`},
					{Code: `var vessels []*types.RecipeStepVessel`},
					{Code: `for range exampleQuantity {
	ing := BuildFakeRecipeStepVessel()
	ing.BelongsToRecipeStep = recipeStepID

	vessels = append(vessels, ing)
}`},
					{Code: `var products []*types.RecipeStepProduct`},
					{Code: `for range exampleQuantity {
	p := BuildFakeRecipeStepProduct()
	p.BelongsToRecipeStep = recipeStepID
	products = append(products, p)
}`},
					{Code: `completionConditionID := BuildFakeID()`},
					{Code: `completionConditions := []*types.RecipeStepCompletionCondition{
	{
		ID:			completionConditionID,
		BelongsToRecipeStep:	recipeStepID,
		IngredientState:	types.ValidIngredientState{},
		Notes:			buildUniqueString(),
		Ingredients: []*types.RecipeStepCompletionConditionIngredient{
			{
				ID:					BuildFakeID(),
				BelongsToRecipeStepCompletionCondition:	completionConditionID,
				RecipeStepIngredient:			ingredients[0].ID,
			},
		},
		Optional:	false,
	},
}`},
					{Code: `minEstimatedTime := uint32(buildFakeNumber())`},
					{Code: `maxEstimatedTime := uint32(buildFakeNumber()) + minEstimatedTime`},
					{Code: `minTemperature := float32(buildFakeNumber())`},
					{Code: `maxTemperature := float32(buildFakeNumber()) + minTemperature`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `recipeStepID`},
					{Name: "Index", Expr: `fake.Uint32()`},
					{Name: "MinEstimatedTimeInSeconds", Expr: `&minEstimatedTime`},
					{Name: "MaxEstimatedTimeInSeconds", Expr: `&maxEstimatedTime`},
					{Name: "MinTemperatureInCelsius", Expr: `&minTemperature`},
					{Name: "MaxTemperatureInCelsius", Expr: `&maxTemperature`},
					{Name: "Products", Expr: `products`},
					{Name: "Optional", Expr: `false`},
					{Name: "Ingredients", Expr: `ingredients`},
					{Name: "Instruments", Expr: `instruments`},
					{Name: "Vessels", Expr: `vessels`},
					{Name: "CompletionConditions", Expr: `completionConditions`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepUpdateRequestInput{}, Converter: "ConvertRecipeStepToRecipeStepUpdateRequestInput"},
					{Type: RecipeStepCreationRequestInput{}, Converter: "ConvertRecipeStepToRecipeStepCreationRequestInput"},
				},
			},
		},
		{
			Type: RecipeStepCompletionCondition{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `id := BuildFakeID()`},
					{Code: `var ingredients []*types.RecipeStepCompletionConditionIngredient`},
					{Code: `for range exampleQuantity {
	ingredient := BuildFakeRecipeStepCompletionConditionIngredient()
	ingredient.BelongsToRecipeStepCompletionCondition = id
	ingredients = append(ingredients, ingredient)
}`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `id`},
					{Name: "Ingredients", Expr: `ingredients`},
					{Name: "CreatedAt", Expr: `time.Time{}`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{}, Converter: "ConvertRecipeStepCompletionConditionToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput"},
				},
			},
		},
		{
			Type: RecipeStepCompletionConditionIngredient{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "RecipeStepIngredient", Expr: `BuildFakeID()`},
					{Name: "CreatedAt", Expr: `time.Time{}`},
				},
			},
		},
		{
			Type: RecipeStepCompletionConditionUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `recipeStepIngredient := BuildFakeRecipeStepCompletionCondition()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Optional", Expr: `&recipeStepIngredient.Optional`},
					{Name: "BelongsToRecipeStep", Expr: `&recipeStepIngredient.BelongsToRecipeStep`},
					{Name: "IngredientStateID", Expr: `&recipeStepIngredient.IngredientState.ID`},
					{Name: "Notes", Expr: `&recipeStepIngredient.Notes`},
				},
			},
		},
		{
			Type: RecipeStepIngredient{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minQty, maxQty := BuildFakeFloat32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Ingredient", Expr: `BuildFakeValidIngredient()`},
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
					{
						Name: "Index",
						Expr: `0`,
						Why:  "Will be set from array index during recipe creation (via converter)",
					},
					{
						Name: "OptionIndex",
						Expr: `0`,
						Why:  "Default to 0 for single-option items",
					},
					{Name: "VesselIndex", Expr: `new(fake.Uint16())`},
					{Name: "ProductPercentageToUse", Expr: `new(float32(buildFakeNumber()))`},
					{Name: "ScaleFactor", Expr: `1.0`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepIngredientUpdateRequestInput{}, Converter: "ConvertRecipeStepIngredientToRecipeStepIngredientUpdateRequestInput"},
				},
			},
		},
		{
			Type: RecipeStepInstrument{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minQty, maxQty := BuildFakeUint32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Instrument", Expr: `BuildFakeValidInstrument()`},
					{Name: "PreferenceRank", Expr: `fake.Uint8()`},
					{Name: "BelongsToRecipeStep", Expr: `fake.UUID()`},
					{Name: "Index", Expr: `0`},
					{Name: "OptionIndex", Expr: `0`},
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
					{Name: "ScaleFactor", Expr: `1.0`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepInstrumentUpdateRequestInput{}, Converter: "ConvertRecipeStepInstrumentToRecipeStepInstrumentUpdateRequestInput"},
				},
			},
		},
		{
			Type: RecipeStepProduct{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{
						Code: `measurementMin := float32(buildFakeNumber())`,
						Why:  "Each max is built from its min so that it exceeds it, which is what the range validation on this type requires.",
					},
					{Code: `measurementMax := measurementMin + float32(buildFakeNumber())`},
					{Code: `itemMin := float32(buildFakeNumber())`},
					{Code: `itemMax := itemMin + float32(buildFakeNumber())`},
					{Code: `storageTempMin := float32(buildFakeNumber())`},
					{Code: `storageTempMax := storageTempMin + float32(buildFakeNumber())`},
					{Code: `storageDurationMax := uint32(buildFakeNumber())`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinStorageTemperatureInCelsius", Expr: `&storageTempMin`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `&storageTempMax`},
					{Name: "MaxStorageDurationInSeconds", Expr: `&storageDurationMax`},
					{Name: "MinMeasurementQuantity", Expr: `&measurementMin`},
					{Name: "MaxMeasurementQuantity", Expr: `&measurementMax`},
					{Name: "MinItemQuantity", Expr: `&itemMin`},
					{Name: "MaxItemQuantity", Expr: `&itemMax`},
					{Name: "MeasurementUnit", Expr: `BuildFakeValidMeasurementUnit()`},
					{Name: "ContainedInVesselIndex", Expr: `new(fake.Uint16())`},
					{Name: "BelongsToRecipeStep", Expr: `fake.UUID()`},
					{Name: "Type", Expr: `types.RecipeStepProductIngredientType`},
					{Name: "Index", Expr: `fake.Uint16()`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepProductUpdateRequestInput{}, Converter: "ConvertRecipeStepProductToRecipeStepProductUpdateRequestInput"},
					{Type: RecipeStepProductCreationRequestInput{}, Converter: "ConvertRecipeStepProductToRecipeStepProductCreationRequestInput"},
				},
			},
		},
		{
			Type: RecipeStepVessel{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minQty, maxQty := BuildFakeUint16WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Vessel", Expr: `BuildFakeValidVessel()`},
					{Name: "BelongsToRecipeStep", Expr: `fake.UUID()`},
					{
						Name: "Index",
						Expr: `0`,
						Why:  "Will be set from array index during recipe creation",
					},
					{
						Name: "OptionIndex",
						Expr: `0`,
						Why:  "Default to 0 for single-option items",
					},
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
					{Name: "ScaleFactor", Expr: `1.0`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: RecipeStepVesselUpdateRequestInput{}, Converter: "ConvertRecipeStepVesselToRecipeStepVesselUpdateRequestInput"},
				},
			},
		},
		{
			Type: UserIngredientPreference{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Rating", Expr: `1`},
					{Name: "CreatedByUser", Expr: `""`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: UserIngredientPreferenceUpdateRequestInput{}, Converter: "ConvertUserIngredientPreferenceToUserIngredientPreferenceUpdateRequestInput"},
					{Type: UserIngredientPreferenceCreationRequestInput{}, Converter: "ConvertUserIngredientPreferenceToUserIngredientPreferenceCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredient{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minST, maxST := BuildFakeOptionalFloat32MinMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinStorageTemperatureInCelsius", Expr: `minST`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxST`},
					{Name: "ContaminatesEquipment", Expr: `false`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientUpdateRequestInput{}, Converter: "ConvertValidIngredientToValidIngredientUpdateRequestInput"},
					{Type: ValidIngredientCreationRequestInput{}, Converter: "ConvertValidIngredientToValidIngredientCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientGroup{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `groupID := BuildFakeID()`},
					{Code: `var members []*types.ValidIngredientGroupMember`},
					{Code: `for range exampleQuantity {
	newMember := BuildFakeValidIngredientGroupMember()
	newMember.BelongsToGroup = groupID
	members = append(members, newMember)
}`},
				},
				Fields: []entitydecl.Field{
					{Name: "ID", Expr: `groupID`},
					{Name: "Members", Expr: `members`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientGroupUpdateRequestInput{}, Converter: "ConvertValidIngredientGroupToValidIngredientGroupUpdateRequestInput"},
					{Type: ValidIngredientGroupCreationRequestInput{}, Converter: "ConvertValidIngredientGroupToValidIngredientGroupCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientGroupMember{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: ValidIngredientMeasurementUnit{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minQty, maxQty := BuildFakeFloat32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinAllowableQuantity", Expr: `minQty`},
					{Name: "MaxAllowableQuantity", Expr: `maxQty`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientMeasurementUnitCreationRequestInput{}, Converter: "ConvertValidIngredientMeasurementUnitToValidIngredientMeasurementUnitCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientMeasurementUnitUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validIngredientMeasurementUnit := BuildFakeValidIngredientMeasurementUnit()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&validIngredientMeasurementUnit.Notes`},
					{Name: "ValidMeasurementUnitID", Expr: `&validIngredientMeasurementUnit.MeasurementUnit.ID`},
					{Name: "ValidIngredientID", Expr: `&validIngredientMeasurementUnit.Ingredient.ID`},
					{Name: "MinAllowableQuantity", Expr: `&validIngredientMeasurementUnit.MinAllowableQuantity`},
					{Name: "MaxAllowableQuantity", Expr: `validIngredientMeasurementUnit.MaxAllowableQuantity`},
				},
			},
		},
		{
			Type: ValidIngredientPreparation{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientPreparationCreationRequestInput{}, Converter: "ConvertValidIngredientPreparationToValidIngredientPreparationCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientPreparationUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validIngredientPreparation := BuildFakeValidIngredientPreparation()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&validIngredientPreparation.Notes`},
					{Name: "ValidPreparationID", Expr: `&validIngredientPreparation.Preparation.ID`},
					{Name: "ValidIngredientID", Expr: `&validIngredientPreparation.Ingredient.ID`},
				},
			},
		},
		{
			Type: ValidIngredientState{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "AttributeType", Expr: `types.ValidIngredientStateAttributeTypeOther`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientStateUpdateRequestInput{}, Converter: "ConvertValidIngredientStateToValidIngredientStateUpdateRequestInput"},
					{Type: ValidIngredientStateCreationRequestInput{}, Converter: "ConvertValidIngredientStateToValidIngredientStateCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientStateIngredient{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidIngredientStateIngredientCreationRequestInput{}, Converter: "ConvertValidIngredientStateIngredientToValidIngredientStateIngredientCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidIngredientStateIngredientUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validIngredientStateIngredient := BuildFakeValidIngredientStateIngredient()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&validIngredientStateIngredient.Notes`},
					{Name: "ValidIngredientStateID", Expr: `&validIngredientStateIngredient.IngredientState.ID`},
					{Name: "ValidIngredientID", Expr: `&validIngredientStateIngredient.Ingredient.ID`},
				},
			},
		},
		{
			Type: ValidInstrument{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidInstrumentUpdateRequestInput{}, Converter: "ConvertValidInstrumentToValidInstrumentUpdateRequestInput"},
					{Type: ValidInstrumentCreationRequestInput{}, Converter: "ConvertValidInstrumentToValidInstrumentCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidMeasurementUnitConversion{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidMeasurementUnitConversionCreationRequestInput{}, Converter: "ConvertValidMeasurementUnitConversionToValidMeasurementUnitConversionCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidMeasurementUnitConversionUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "From", Expr: `new(BuildFakeID())`},
					{Name: "To", Expr: `new(BuildFakeID())`},
					{Name: "OnlyForIngredient", Expr: `new(BuildFakeID())`},
					{Name: "Modifier", Expr: `new(float32(buildFakeNumber()))`},
					{Name: "Notes", Expr: `new(BuildFakeID())`},
				},
			},
		},
		{
			Type: ValidMeasurementUnit{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Metric", Expr: `true`},
					{Name: "Imperial", Expr: `false`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidMeasurementUnitUpdateRequestInput{}, Converter: "ConvertValidMeasurementUnitToValidMeasurementUnitUpdateRequestInput"},
					{Type: ValidMeasurementUnitCreationRequestInput{}, Converter: "ConvertValidMeasurementUnitToValidMeasurementUnitCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidPrepTaskConfig{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minSD, maxSD := BuildFakeUint32WithOptionalMax()`},
					{Code: `minST, maxST := BuildFakeOptionalFloat32MinMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinStorageDurationInSeconds", Expr: `minSD`},
					{Name: "MaxStorageDurationInSeconds", Expr: `maxSD`},
					{Name: "MinStorageTemperatureInCelsius", Expr: `minST`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxST`},
					{Name: "StorageType", Expr: `fake.RandomString([]string{
	types.RecipePrepTaskStorageTypeUncovered,
	types.RecipePrepTaskStorageTypeCovered,
	types.RecipePrepTaskStorageTypeAirtightContainer,
	types.RecipePrepTaskStorageTypeWireRack,
})`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidPrepTaskConfigUpdateRequestInput{}, Converter: "ConvertValidPrepTaskConfigToValidPrepTaskConfigUpdateRequestInput"},
					{Type: ValidPrepTaskConfigCreationRequestInput{}, Converter: "ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidPreparation{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minIngredientCount, maxIngredientCount := BuildFakeUint16WithOptionalMax()`},
					{Code: `minInstrumentCount, maxInstrumentCount := BuildFakeUint16WithOptionalMax()`},
					{Code: `minVesselCount, maxVesselCount := BuildFakeUint16WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinIngredientCount", Expr: `minIngredientCount`},
					{Name: "MaxIngredientCount", Expr: `maxIngredientCount`},
					{Name: "MinInstrumentCount", Expr: `minInstrumentCount`},
					{Name: "MaxInstrumentCount", Expr: `maxInstrumentCount`},
					{Name: "MinVesselCount", Expr: `minVesselCount`},
					{Name: "MaxVesselCount", Expr: `maxVesselCount`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidPreparationUpdateRequestInput{}, Converter: "ConvertValidPreparationToValidPreparationUpdateRequestInput"},
					{Type: ValidPreparationCreationRequestInput{}, Converter: "ConvertValidPreparationToValidPreparationCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidPreparationInstrument{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidPreparationInstrumentCreationRequestInput{}, Converter: "ConvertValidPreparationInstrumentToValidPreparationInstrumentCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidPreparationInstrumentUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validPreparationInstrument := BuildFakeValidPreparationInstrument()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&validPreparationInstrument.Notes`},
					{Name: "ValidPreparationID", Expr: `&validPreparationInstrument.Preparation.ID`},
					{Name: "ValidInstrumentID", Expr: `&validPreparationInstrument.Instrument.ID`},
				},
			},
		},
		{
			Type: ValidPreparationVessel{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidPreparationVesselCreationRequestInput{}, Converter: "ConvertValidPreparationVesselToValidPreparationVesselCreationRequestInput"},
				},
			},
		},
		{
			Type: ValidPreparationVesselUpdateRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `validPreparationVessel := BuildFakeValidPreparationVessel()`},
				},
				Fields: []entitydecl.Field{
					{Name: "Notes", Expr: `&validPreparationVessel.Notes`},
					{Name: "ValidPreparationID", Expr: `&validPreparationVessel.Preparation.ID`},
					{Name: "ValidVesselID", Expr: `&validPreparationVessel.Vessel.ID`},
				},
			},
		},
		{
			Type: ValidVessel{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "CapacityUnit", Expr: `BuildFakeValidMeasurementUnit()`},
					{Name: "Shape", Expr: `types.VesselShapeOther`},
				},
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: ValidVesselUpdateRequestInput{}, Converter: "ConvertValidVesselToValidVesselUpdateRequestInput"},
					{Type: ValidVesselCreationRequestInput{}, Converter: "ConvertValidVesselToValidVesselCreationRequestInput"},
				},
			},
		},
		{
			Type: FinalizeMealPlansRequest{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: FinalizeMealPlansResponse{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: CreateMealPlanTasksRequest{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: CreateMealPlanTasksResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Success", Expr: `true`},
				},
			},
		},
		{
			Type: InitializeMealPlanGroceryListRequest{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: InitializeMealPlanGroceryListResponse{},
			Fake: entitydecl.Fake{
				Fields: []entitydecl.Field{
					{Name: "Success", Expr: `true`},
				},
			},
		},
	},
}
