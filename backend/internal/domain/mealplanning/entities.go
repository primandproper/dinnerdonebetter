package mealplanning

import "github.com/primandproper/dinnerdonebetter/backend/internal/domain/entitydecl"

// Entities declares this domain's entities for code generation.
var Entities = entitydecl.Domain{
	Entities: []entitydecl.Entity{
		{
			Type: AccountInstrumentOwnership{},
			Fake: entitydecl.Fake{
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
					{
						Name: "ComponentType",
						Expr: `types.MealComponentTypesMain`,
						Why: "A meal is rejected unless one of its components is the main, and a meal's " +
							"components are faked independently — so a random type per component leaves " +
							"a meal with no main most of the time.",
					},
				},
			},
		},
		{
			Type: MealPlan{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{
						Code: `votingDeadline := time.Now().Add(5 * time.Minute).Truncate(time.Second).UTC()`,
						Why:  "The voting deadline must be in the future but before every event's start time (events start in ten minutes, see BuildFakeMealPlanEvent), so the meal plan passes MealPlanCreationRequestInput validation.",
					},
				},
				Fields: []entitydecl.Field{
					{Name: "Status", Expr: `string(types.MealPlanStatusAwaitingVotes)`},
					{Name: "VotingDeadline", Expr: `votingDeadline`},
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
					{Code: `now := time.Now().Add(0).Truncate(time.Second).UTC()`},
					{Code: `inTenMinutes := now.Add(time.Minute * 10).Add(0).Truncate(time.Second).UTC()`},
					{Code: `inOneWeek := now.Add((time.Hour * 24) * 7).Add(0).Truncate(time.Second).UTC()`},
				},
				Fields: []entitydecl.Field{
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
					{Code: `meal := BuildFakeMeal()`},
					{Code: `meal.Components = nil`},
				},
				Fields: []entitydecl.Field{
					{Name: "AssignedCook", Expr: `func(s string) *string { return &s }(BuildFakeID())`},
					{Name: "Meal", Expr: `*meal`},
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
			Fake: entitydecl.Fake{},
		},
		{
			Type: MealPlanTask{},
			Fake: entitydecl.Fake{
				List: &entitydecl.List{},
				Inputs: []entitydecl.Input{
					{Type: MealPlanTaskCreationRequestInput{}, Converter: "ConvertMealPlanTaskToMealPlanTaskCreationRequestInput"},
				},
			},
		},
		{
			Type: MealPlanTaskStatusChangeRequestInput{},
			Fake: entitydecl.Fake{},
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
				Fields: []entitydecl.Field{
					{Name: "Status", Expr: `types.RecipeStatusSubmitted`},
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
			Fake: entitydecl.Fake{},
		},
		{
			Type: RecipePrepTaskCreationRequestInput{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()`},
					{Code: `minBuf, maxBuf := BuildFakeUint32WithOptionalMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "StorageType", Expr: `types.RecipePrepTaskStorageTypeUncovered`},
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
					{Code: `minTemp, maxTemp := BuildFakeOptionalFloat32MinMax()`},
					{Code: `minBuf, maxBuf := BuildFakeOptionalUint32MinMax()`},
				},
				Fields: []entitydecl.Field{
					{Name: "StorageType", Expr: `pointer.To(types.RecipePrepTaskStorageTypeUncovered)`},
					{Name: "MinTimeBufferBeforeRecipeInSeconds", Expr: `minBuf`},
					{Name: "MaxTimeBufferBeforeRecipeInSeconds", Expr: `maxBuf`},
					{Name: "MinStorageTemperatureInCelsius", Expr: `minTemp`},
					{Name: "MaxStorageTemperatureInCelsius", Expr: `maxTemp`},
				},
			},
		},
		{
			Type: RecipeStep{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{Code: `minEstimatedTime := uint32(buildFakeNumber())`},
					{Code: `maxEstimatedTime := uint32(buildFakeNumber()) + minEstimatedTime`},
					{Code: `minTemperature := float32(buildFakeNumber())`},
					{Code: `maxTemperature := float32(buildFakeNumber()) + minTemperature`},
				},
				Fields: []entitydecl.Field{
					{Name: "MinEstimatedTimeInSeconds", Expr: `&minEstimatedTime`},
					{Name: "MaxEstimatedTimeInSeconds", Expr: `&maxEstimatedTime`},
					{Name: "MinTemperatureInCelsius", Expr: `&minTemperature`},
					{Name: "MaxTemperatureInCelsius", Expr: `&maxTemperature`},
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
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
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
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
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
					{Name: "MinQuantity", Expr: `minQty`},
					{Name: "MaxQuantity", Expr: `maxQty`},
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
			Fake: entitydecl.Fake{},
		},
		{
			Type: ValidMeasurementUnit{},
			Fake: entitydecl.Fake{
				Locals: []entitydecl.Local{
					{
						Code: `metric := fake.Bool()`,
						Why:  "A unit is metric or imperial. Validation rejects one claiming to be both, which two independent fakes produce a quarter of the time.",
					},
				},
				Fields: []entitydecl.Field{
					{Name: "Metric", Expr: `metric`},
					{Name: "Imperial", Expr: `!metric`},
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
			Fake: entitydecl.Fake{},
		},
		{
			Type: InitializeMealPlanGroceryListRequest{},
			Fake: entitydecl.Fake{},
		},
		{
			Type: InitializeMealPlanGroceryListResponse{},
			Fake: entitydecl.Fake{},
		},
	},
}
