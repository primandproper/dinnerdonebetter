package grpcconverters

import (
	"log"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	grpcconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/converters"
	mealplanningsvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	uploadedmediaconverters "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/services/uploadedmedia/grpc/converters"

	"github.com/primandproper/platform-go/v8/pointer"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ConvertGRPCRecipeCreationRequestInputToRecipeCreationRequestInput(input *mealplanningsvc.RecipeCreationRequestInput) *mealplanning.RecipeCreationRequestInput {
	var steps []*mealplanning.RecipeStepCreationRequestInput
	for _, step := range input.Steps {
		steps = append(steps, ConvertGRPCRecipeStepCreationRequestInputToRecipeStepCreationRequestInput(step))
	}

	var prepTasks []*mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput
	for _, task := range input.PrepTasks {
		prepTasks = append(prepTasks, ConvertGRPCRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskWithinRecipeCreationRequestInput(task))
	}

	return &mealplanning.RecipeCreationRequestInput{
		InspiredByRecipeID:   input.InspiredByRecipeId,
		Name:                 input.Name,
		Source:               input.Source,
		SourceISBN:           input.SourceIsbn,
		Description:          input.Description,
		PluralPortionName:    input.PluralPortionName,
		PortionName:          input.PortionName,
		Slug:                 input.Slug,
		YieldsComponentType:  ConvertMealComponentTypeToString(input.YieldsComponentType),
		MinEstimatedPortions: input.MinEstimatedPortions,

		MaxEstimatedPortions: input.MaxEstimatedPortions,
		AlsoCreateMeal:       input.AlsoCreateMeal,
		EligibleForMeals:     input.EligibleForMeals,
		PrepTasks:            prepTasks,
		Steps:                steps,
	}
}

func ConvertRecipeCreationRequestInputToGRPCRecipeCreationRequestInput(input *mealplanning.RecipeCreationRequestInput) *mealplanningsvc.RecipeCreationRequestInput {
	var steps []*mealplanningsvc.RecipeStepCreationRequestInput
	for _, step := range input.Steps {
		steps = append(steps, ConvertRecipeStepCreationRequestInputToGRPCRecipeStepCreationRequestInput(step))
	}

	var prepTasks []*mealplanningsvc.RecipePrepTaskWithinRecipeCreationRequestInput
	for _, task := range input.PrepTasks {
		prepTasks = append(prepTasks, ConvertRecipePrepTaskWithinRecipeCreationRequestInputToGRPCRecipePrepTaskWithinRecipeCreationRequestInput(task))
	}

	return &mealplanningsvc.RecipeCreationRequestInput{
		InspiredByRecipeId:   input.InspiredByRecipeID,
		Name:                 input.Name,
		Source:               input.Source,
		SourceIsbn:           input.SourceISBN,
		Description:          input.Description,
		PluralPortionName:    input.PluralPortionName,
		PortionName:          input.PortionName,
		Slug:                 input.Slug,
		YieldsComponentType:  ConvertStringToMealComponentType(input.YieldsComponentType),
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
		AlsoCreateMeal:       input.AlsoCreateMeal,
		EligibleForMeals:     input.EligibleForMeals,
		PrepTasks:            prepTasks,
		Steps:                steps,
	}
}

func ConvertGRPCRecipePrepTaskWithinRecipeCreationRequestInputToRecipePrepTaskWithinRecipeCreationRequestInput(input *mealplanningsvc.RecipePrepTaskWithinRecipeCreationRequestInput) *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput {
	var prepTaskSteps []*mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput
	for _, step := range input.RecipeSteps {
		prepTaskSteps = append(prepTaskSteps, ConvertGRPCRecipePrepTaskStepWithinRecipeCreationRequestInputToRecipePrepTaskStepWithinRecipeCreationRequestInput(step))
	}

	return &mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput{
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        input.StorageType,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        prepTaskSteps,
		Optional:                           input.Optional,
	}
}

func ConvertRecipePrepTaskWithinRecipeCreationRequestInputToGRPCRecipePrepTaskWithinRecipeCreationRequestInput(input *mealplanning.RecipePrepTaskWithinRecipeCreationRequestInput) *mealplanningsvc.RecipePrepTaskWithinRecipeCreationRequestInput {
	var prepTaskSteps []*mealplanningsvc.RecipePrepTaskStepWithinRecipeCreationRequestInput
	for _, step := range input.RecipeSteps {
		prepTaskSteps = append(prepTaskSteps, ConvertRecipePrepTaskStepWithinRecipeCreationRequestInputToGRPCRecipePrepTaskStepWithinRecipeCreationRequestInput(step))
	}

	return &mealplanningsvc.RecipePrepTaskWithinRecipeCreationRequestInput{
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        input.StorageType,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        prepTaskSteps,
		Optional:                           input.Optional,
	}
}

func ConvertGRPCRecipePrepTaskCreationRequestInputToRecipePrepTaskCreationRequestInput(input *mealplanningsvc.RecipePrepTaskCreationRequestInput) *mealplanning.RecipePrepTaskCreationRequestInput {
	var steps []*mealplanning.RecipePrepTaskStepCreationRequestInput
	for _, step := range input.RecipeSteps {
		steps = append(steps, ConvertGRPCRecipePrepTaskStepCreationRequestInputToRecipePrepTaskStepCreationRequestInput(step))
	}

	return &mealplanning.RecipePrepTaskCreationRequestInput{
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        input.StorageType,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        steps,
		Optional:                           input.Optional,
	}
}

func ConvertRecipePrepTaskCreationRequestInputToGRPCRecipePrepTaskCreationRequestInput(input *mealplanning.RecipePrepTaskCreationRequestInput) *mealplanningsvc.RecipePrepTaskCreationRequestInput {
	var steps []*mealplanningsvc.RecipePrepTaskStepCreationRequestInput
	for _, step := range input.RecipeSteps {
		steps = append(steps, ConvertRecipePrepTaskStepCreationRequestInputToGRPCRecipePrepTaskStepCreationRequestInput(step))
	}

	return &mealplanningsvc.RecipePrepTaskCreationRequestInput{
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		StorageType:                        input.StorageType,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		RecipeSteps:                        steps,
		Optional:                           input.Optional,
	}
}

func ConvertGRPCRecipeRatingCreationRequestInputToRecipeRatingCreationRequestInput(input *mealplanningsvc.RecipeRatingCreationRequestInput) *mealplanning.RecipeRatingCreationRequestInput {
	return &mealplanning.RecipeRatingCreationRequestInput{
		BelongsToRecipe: input.RecipeId,
		Notes:           input.Notes,
		CreatedByUser:   input.ByUser,
		Taste:           input.Taste,
		Difficulty:      input.Difficulty,
		Cleanup:         input.Cleanup,
		Instructions:    input.Instructions,
		Overall:         input.Overall,
	}
}

func ConvertRecipeRatingCreationRequestInputToGRPCRecipeRatingCreationRequestInput(input *mealplanning.RecipeRatingCreationRequestInput) *mealplanningsvc.RecipeRatingCreationRequestInput {
	return &mealplanningsvc.RecipeRatingCreationRequestInput{
		RecipeId:     input.BelongsToRecipe,
		Notes:        input.Notes,
		ByUser:       input.CreatedByUser,
		Taste:        input.Taste,
		Difficulty:   input.Difficulty,
		Cleanup:      input.Cleanup,
		Instructions: input.Instructions,
		Overall:      input.Overall,
	}
}

func ConvertGRPCRecipePrepTaskStepCreationRequestInputToRecipePrepTaskStepCreationRequestInput(input *mealplanningsvc.RecipePrepTaskStepCreationRequestInput) *mealplanning.RecipePrepTaskStepCreationRequestInput {
	return &mealplanning.RecipePrepTaskStepCreationRequestInput{
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		SatisfiesRecipeStep: input.SatisfiesRecipeStep,
	}
}

func ConvertRecipePrepTaskStepCreationRequestInputToGRPCRecipePrepTaskStepCreationRequestInput(input *mealplanning.RecipePrepTaskStepCreationRequestInput) *mealplanningsvc.RecipePrepTaskStepCreationRequestInput {
	return &mealplanningsvc.RecipePrepTaskStepCreationRequestInput{
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		SatisfiesRecipeStep: input.SatisfiesRecipeStep,
	}
}

func ConvertGRPCRecipePrepTaskStepWithinRecipeCreationRequestInputToRecipePrepTaskStepWithinRecipeCreationRequestInput(input *mealplanningsvc.RecipePrepTaskStepWithinRecipeCreationRequestInput) *mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput {
	return &mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		BelongsToRecipeStepIndex: input.BelongsToRecipeStepIndex,
		SatisfiesRecipeStep:      input.SatisfiesRecipeStep,
	}
}

func ConvertRecipeListItemToGRPCRecipeListItem(input *mealplanning.RecipeListItem) *mealplanningsvc.RecipeListItem {
	return &mealplanningsvc.RecipeListItem{
		CreatedAt:           timestamppb.New(input.CreatedAt),
		LastUpdatedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:          grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Id:                  input.ID,
		RecipeId:            input.Recipe.ID,
		Notes:               input.Notes,
		BelongsToRecipeList: input.BelongsToRecipeList,
		Recipe:              ConvertRecipeToGRPCRecipe(&input.Recipe),
	}
}

func ConvertRecipeListToGRPCRecipeList(input *mealplanning.RecipeList) *mealplanningsvc.RecipeList {
	var items []*mealplanningsvc.RecipeListItem
	for _, item := range input.Items {
		items = append(items, ConvertRecipeListItemToGRPCRecipeListItem(item))
	}

	return &mealplanningsvc.RecipeList{
		CreatedAt:     timestamppb.New(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Id:            input.ID,
		Name:          input.Name,
		Description:   input.Description,
		BelongsToUser: input.BelongsToUser,
		Items:         items,
	}
}

func ConvertGRPCRecipeListCreationRequestInputToRecipeListCreationRequestInput(input *mealplanningsvc.RecipeListCreationRequestInput) *mealplanning.RecipeListCreationRequestInput {
	var items []*mealplanning.RecipeListItemCreationRequestInput
	for _, item := range input.Items {
		items = append(items, ConvertGRPCRecipeListItemCreationRequestInputToRecipeListItemCreationRequestInput(item))
	}

	return &mealplanning.RecipeListCreationRequestInput{
		Name:        input.Name,
		Description: input.Description,
		Items:       items,
	}
}

func ConvertGRPCRecipeListItemCreationRequestInputToRecipeListItemCreationRequestInput(input *mealplanningsvc.RecipeListItemCreationRequestInput) *mealplanning.RecipeListItemCreationRequestInput {
	return &mealplanning.RecipeListItemCreationRequestInput{
		RecipeID: input.RecipeId,
		Notes:    input.Notes,
	}
}

func ConvertGRPCRecipeListUpdateRequestInputToRecipeListUpdateRequestInput(input *mealplanningsvc.RecipeListUpdateRequestInput) *mealplanning.RecipeListUpdateRequestInput {
	if input == nil {
		return nil
	}

	var name *string
	if input.Name != nil {
		name = new(input.GetName())
	}

	var desc *string
	if input.Description != nil {
		desc = new(input.GetDescription())
	}

	return &mealplanning.RecipeListUpdateRequestInput{
		Name:        name,
		Description: desc,
	}
}

func ConvertGRPCRecipeListItemUpdateRequestInputToRecipeListItemUpdateRequestInput(input *mealplanningsvc.RecipeListItemUpdateRequestInput) *mealplanning.RecipeListItemUpdateRequestInput {
	if input == nil {
		return nil
	}

	var notes *string
	if input.Notes != nil {
		notes = new(input.GetNotes())
	}

	return &mealplanning.RecipeListItemUpdateRequestInput{
		Notes: notes,
	}
}

func ConvertRecipePrepTaskStepWithinRecipeCreationRequestInputToGRPCRecipePrepTaskStepWithinRecipeCreationRequestInput(input *mealplanning.RecipePrepTaskStepWithinRecipeCreationRequestInput) *mealplanningsvc.RecipePrepTaskStepWithinRecipeCreationRequestInput {
	return &mealplanningsvc.RecipePrepTaskStepWithinRecipeCreationRequestInput{
		BelongsToRecipeStepIndex: input.BelongsToRecipeStepIndex,
		SatisfiesRecipeStep:      input.SatisfiesRecipeStep,
	}
}

func ConvertGRPCRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentCreationRequestInput(input *mealplanningsvc.RecipeStepInstrumentCreationRequestInput) *mealplanning.RecipeStepInstrumentCreationRequestInput {
	scaleFactor := input.GetScaleFactor()
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &mealplanning.RecipeStepInstrumentCreationRequestInput{
		RecipeStepProductID:             input.RecipeStepProductId,
		ValidPreparationInstrumentID:    input.ValidPreparationInstrumentId,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		Notes:                           input.Notes,
		Name:                            input.Name,
		Index:                           grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:                     uint16(input.OptionIndex),
		Optional:                        input.Optional,
		PreferenceRank:                  uint8(input.PreferenceRank),
		MinQuantity:                     input.MinQuantity,

		MaxQuantity: input.MaxQuantity,
		ScaleFactor: scaleFactor,
	}
}

func ConvertRecipeStepInstrumentCreationRequestInputToGRPCRecipeStepInstrumentCreationRequestInput(input *mealplanning.RecipeStepInstrumentCreationRequestInput) *mealplanningsvc.RecipeStepInstrumentCreationRequestInput {
	return &mealplanningsvc.RecipeStepInstrumentCreationRequestInput{
		ValidPreparationInstrumentId:    input.ValidPreparationInstrumentID,
		RecipeStepProductId:             input.RecipeStepProductID,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		Notes:                           input.Notes,
		Name:                            input.Name,
		Index:                           grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:                     uint32(input.OptionIndex),
		Optional:                        input.Optional,
		PreferenceRank:                  uint32(input.PreferenceRank),
		MinQuantity:                     input.MinQuantity,
		MaxQuantity:                     input.MaxQuantity,
		ScaleFactor:                     &input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepVesselCreationRequestInputToRecipeStepVesselCreationRequestInput(input *mealplanningsvc.RecipeStepVesselCreationRequestInput) *mealplanning.RecipeStepVesselCreationRequestInput {
	scaleFactor := input.GetScaleFactor()
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &mealplanning.RecipeStepVesselCreationRequestInput{
		RecipeStepProductID:             input.RecipeStepProductId,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		ValidPreparationVesselID:        input.ValidPreparationVesselId,
		Name:                            input.Name,
		Notes:                           input.Notes,
		VesselPreposition:               input.VesselPreposition,
		Index:                           grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:                     uint16(input.OptionIndex),
		UnavailableAfterStep:            input.UnavailableAfterStep,
		MinQuantity:                     uint16(input.MinQuantity),

		MaxQuantity: grpcconverters.ConvertUint32PointerToUint16Pointer(input.MaxQuantity),
		ScaleFactor: scaleFactor,
	}
}

func ConvertRecipeStepVesselCreationRequestInputToGRPCRecipeStepVesselCreationRequestInput(input *mealplanning.RecipeStepVesselCreationRequestInput) *mealplanningsvc.RecipeStepVesselCreationRequestInput {
	return &mealplanningsvc.RecipeStepVesselCreationRequestInput{
		RecipeStepProductId:             input.RecipeStepProductID,
		ProductOfRecipeStepIndex:        input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex: input.ProductOfRecipeStepProductIndex,
		ValidPreparationVesselId:        input.ValidPreparationVesselID,
		Name:                            input.Name,
		Notes:                           input.Notes,
		VesselPreposition:               input.VesselPreposition,
		Index:                           grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:                     uint32(input.OptionIndex),
		UnavailableAfterStep:            input.UnavailableAfterStep,
		MinQuantity:                     uint32(input.MinQuantity),
		MaxQuantity:                     grpcconverters.ConvertUint16PointerToUint32Pointer(input.MaxQuantity),
		ScaleFactor:                     &input.ScaleFactor,
	}
}

func ConvertStringToRecipeStepProductType(s string) mealplanningsvc.RecipeStepProductType {
	switch s {
	case mealplanning.RecipeStepProductInstrumentType:
		return mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT
	case mealplanning.RecipeStepProductVesselType:
		return mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_VESSEL
	case mealplanning.RecipeStepProductIngredientType:
		return mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_INGREDIENT
	default:
		log.Printf("UNKNOWN RECIPE STEP PRODUCT TYPE: %q", s)
		return mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_INGREDIENT
	}
}

func ConvertRecipeStepProductTypeToString(s mealplanningsvc.RecipeStepProductType) string {
	switch s {
	case mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_INSTRUMENT:
		return mealplanning.RecipeStepProductInstrumentType
	case mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_VESSEL:
		return mealplanning.RecipeStepProductVesselType
	case mealplanningsvc.RecipeStepProductType_RECIPE_STEP_PRODUCT_TYPE_INGREDIENT:
		return mealplanning.RecipeStepProductIngredientType
	default:
		log.Printf("UNKNOWN RECIPE STEP PRODUCT TYPE: %q", s)
		return mealplanning.RecipeStepProductIngredientType
	}
}

func ConvertGRPCRecipeStepProductCreationRequestInputToRecipeStepProductCreationRequestInput(input *mealplanningsvc.RecipeStepProductCreationRequestInput) *mealplanning.RecipeStepProductCreationRequestInput {
	return &mealplanning.RecipeStepProductCreationRequestInput{
		MeasurementUnitID:              input.MeasurementUnitId,
		ContainedInVesselIndex:         grpcconverters.ConvertUint32PointerToUint16Pointer(input.ContainedInVesselIndex),
		QuantityNotes:                  input.QuantityNotes,
		Name:                           input.Name,
		StorageInstructions:            input.StorageInstructions,
		Type:                           ConvertRecipeStepProductTypeToString(input.Type),
		Index:                          uint16(input.Index),
		Compostable:                    input.Compostable,
		IsLiquid:                       input.IsLiquid,
		IsWaste:                        input.IsWaste,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
	}
}

func ConvertRecipeStepProductCreationRequestInputToGRPCRecipeStepProductCreationRequestInput(input *mealplanning.RecipeStepProductCreationRequestInput) *mealplanningsvc.RecipeStepProductCreationRequestInput {
	return &mealplanningsvc.RecipeStepProductCreationRequestInput{
		MeasurementUnitId:              input.MeasurementUnitID,
		ContainedInVesselIndex:         grpcconverters.ConvertUint16PointerToUint32Pointer(input.ContainedInVesselIndex),
		QuantityNotes:                  input.QuantityNotes,
		Name:                           input.Name,
		StorageInstructions:            input.StorageInstructions,
		Type:                           ConvertStringToRecipeStepProductType(input.Type),
		Index:                          uint32(input.Index),
		Compostable:                    input.Compostable,
		IsLiquid:                       input.IsLiquid,
		IsWaste:                        input.IsWaste,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
	}
}

func ConvertGRPCRecipeStepIngredientCreationRequestInputToRecipeStepIngredientCreationRequestInput(input *mealplanningsvc.RecipeStepIngredientCreationRequestInput) *mealplanning.RecipeStepIngredientCreationRequestInput {
	scaleFactor := input.GetScaleFactor()
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &mealplanning.RecipeStepIngredientCreationRequestInput{
		ValidIngredientPreparationID:     input.ValidIngredientPreparationId,
		ValidIngredientMeasurementUnitID: input.ValidIngredientMeasurementUnitId,
		ProductOfRecipeStepIndex:         input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex:  input.ProductOfRecipeStepProductIndex,
		VesselIndex:                      grpcconverters.ConvertUint32PointerToUint16Pointer(input.VesselIndex),
		ProductPercentageToUse:           input.ProductPercentageToUse,
		RecipeStepProductRecipeID:        input.RecipeStepProductRecipeId,
		IngredientNotes:                  input.IngredientNotes,
		Name:                             input.Name,
		QuantityNotes:                    input.QuantityNotes,
		Index:                            grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:                      uint16(input.OptionIndex),
		Optional:                         input.Optional,
		ToTaste:                          input.ToTaste,
		MinQuantity:                      input.MinQuantity,

		MaxQuantity: input.MaxQuantity,
		ScaleFactor: scaleFactor,
	}
}

func ConvertRecipeStepIngredientCreationRequestInputToGRPCRecipeStepIngredientCreationRequestInput(input *mealplanning.RecipeStepIngredientCreationRequestInput) *mealplanningsvc.RecipeStepIngredientCreationRequestInput {
	return &mealplanningsvc.RecipeStepIngredientCreationRequestInput{
		ValidIngredientPreparationId:     input.ValidIngredientPreparationID,
		ValidIngredientMeasurementUnitId: input.ValidIngredientMeasurementUnitID,
		ProductOfRecipeStepIndex:         input.ProductOfRecipeStepIndex,
		ProductOfRecipeStepProductIndex:  input.ProductOfRecipeStepProductIndex,
		VesselIndex:                      grpcconverters.ConvertUint16PointerToUint32Pointer(input.VesselIndex),
		ProductPercentageToUse:           input.ProductPercentageToUse,
		RecipeStepProductRecipeId:        input.RecipeStepProductRecipeID,
		IngredientNotes:                  input.IngredientNotes,
		Name:                             input.Name,
		QuantityNotes:                    input.QuantityNotes,
		Index:                            grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:                      uint32(input.OptionIndex),
		Optional:                         input.Optional,
		ToTaste:                          input.ToTaste,
		MinQuantity:                      input.MinQuantity,
		MaxQuantity:                      input.MaxQuantity,
		ScaleFactor:                      &input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionCreationRequestInput(input *mealplanningsvc.RecipeStepCompletionConditionCreationRequestInput) *mealplanning.RecipeStepCompletionConditionCreationRequestInput {
	return &mealplanning.RecipeStepCompletionConditionCreationRequestInput{
		IngredientStateID:   input.IngredientStateId,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         input.Ingredients,
		Optional:            input.Optional,
	}
}

func ConvertRecipeStepCompletionConditionCreationRequestInputToGRPCRecipeStepCompletionConditionCreationRequestInput(input *mealplanning.RecipeStepCompletionConditionCreationRequestInput) *mealplanningsvc.RecipeStepCompletionConditionCreationRequestInput {
	return &mealplanningsvc.RecipeStepCompletionConditionCreationRequestInput{
		IngredientStateId:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         input.Ingredients,
		Optional:            input.Optional,
	}
}

func ConvertGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(input *mealplanningsvc.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	ingredients := []*mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{}
	for _, ingredient := range input.Ingredients {
		ingredients = append(ingredients, &mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput{RecipeStepIngredient: ingredient.RecipeStepIngredient})
	}

	return &mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
		IngredientStateID:   input.IngredientStateId,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}
}

func ConvertRecipeStepCompletionConditionForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionForExistingRecipeCreationRequestInput(input *mealplanning.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput) *mealplanningsvc.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput {
	ingredients := []*mealplanningsvc.RecipeStepCompletionConditionIngredient{}
	for _, ingredient := range input.Ingredients {
		ingredients = append(ingredients, ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(ingredient))
	}

	return &mealplanningsvc.RecipeStepCompletionConditionForExistingRecipeCreationRequestInput{
		IngredientStateId:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Ingredients:         ingredients,
		Optional:            input.Optional,
	}
}

func ConvertGRPCRecipeStepCreationRequestInputToRecipeStepCreationRequestInput(input *mealplanningsvc.RecipeStepCreationRequestInput) *mealplanning.RecipeStepCreationRequestInput {
	var recipeStepInstrumentCreationRequestInputs []*mealplanning.RecipeStepInstrumentCreationRequestInput
	for _, instrument := range input.Instruments {
		recipeStepInstrumentCreationRequestInputs = append(recipeStepInstrumentCreationRequestInputs, ConvertGRPCRecipeStepInstrumentCreationRequestInputToRecipeStepInstrumentCreationRequestInput(instrument))
	}

	var recipeStepVesselCreationRequestInputs []*mealplanning.RecipeStepVesselCreationRequestInput
	for _, vessel := range input.Vessels {
		recipeStepVesselCreationRequestInputs = append(recipeStepVesselCreationRequestInputs, ConvertGRPCRecipeStepVesselCreationRequestInputToRecipeStepVesselCreationRequestInput(vessel))
	}

	var recipeStepProductCreationRequestInputs []*mealplanning.RecipeStepProductCreationRequestInput
	for _, product := range input.Products {
		recipeStepProductCreationRequestInputs = append(recipeStepProductCreationRequestInputs, ConvertGRPCRecipeStepProductCreationRequestInputToRecipeStepProductCreationRequestInput(product))
	}

	var recipeStepIngredientCreationRequestInputs []*mealplanning.RecipeStepIngredientCreationRequestInput
	for _, ingredient := range input.Ingredients {
		recipeStepIngredientCreationRequestInputs = append(recipeStepIngredientCreationRequestInputs, ConvertGRPCRecipeStepIngredientCreationRequestInputToRecipeStepIngredientCreationRequestInput(ingredient))
	}

	var recipeStepCompletionConditionCreationRequestInputs []*mealplanning.RecipeStepCompletionConditionCreationRequestInput
	for _, completionCondition := range input.CompletionConditions {
		recipeStepCompletionConditionCreationRequestInputs = append(recipeStepCompletionConditionCreationRequestInputs, ConvertGRPCRecipeStepCompletionConditionCreationRequestInputToRecipeStepCompletionConditionCreationRequestInput(completionCondition))
	}

	return &mealplanning.RecipeStepCreationRequestInput{
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		PreparationID:             input.PreparationId,
		Notes:                     input.Notes,
		ConditionExpression:       input.ConditionExpression,
		ExplicitInstructions:      input.ExplicitInstructions,
		Index:                     input.Index,
		Optional:                  input.Optional,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		Instruments:               recipeStepInstrumentCreationRequestInputs,
		Vessels:                   recipeStepVesselCreationRequestInputs,
		Products:                  recipeStepProductCreationRequestInputs,
		Ingredients:               recipeStepIngredientCreationRequestInputs,
		CompletionConditions:      recipeStepCompletionConditionCreationRequestInputs,
	}
}

func ConvertRecipeStepCreationRequestInputToGRPCRecipeStepCreationRequestInput(input *mealplanning.RecipeStepCreationRequestInput) *mealplanningsvc.RecipeStepCreationRequestInput {
	var recipeStepInstrumentCreationRequestInputs []*mealplanningsvc.RecipeStepInstrumentCreationRequestInput
	for _, instrument := range input.Instruments {
		recipeStepInstrumentCreationRequestInputs = append(recipeStepInstrumentCreationRequestInputs, ConvertRecipeStepInstrumentCreationRequestInputToGRPCRecipeStepInstrumentCreationRequestInput(instrument))
	}

	var recipeStepVesselCreationRequestInputs []*mealplanningsvc.RecipeStepVesselCreationRequestInput
	for _, vessel := range input.Vessels {
		recipeStepVesselCreationRequestInputs = append(recipeStepVesselCreationRequestInputs, ConvertRecipeStepVesselCreationRequestInputToGRPCRecipeStepVesselCreationRequestInput(vessel))
	}

	var recipeStepProductCreationRequestInputs []*mealplanningsvc.RecipeStepProductCreationRequestInput
	for _, product := range input.Products {
		recipeStepProductCreationRequestInputs = append(recipeStepProductCreationRequestInputs, ConvertRecipeStepProductCreationRequestInputToGRPCRecipeStepProductCreationRequestInput(product))
	}

	var recipeStepIngredientCreationRequestInputs []*mealplanningsvc.RecipeStepIngredientCreationRequestInput
	for _, ingredient := range input.Ingredients {
		recipeStepIngredientCreationRequestInputs = append(recipeStepIngredientCreationRequestInputs, ConvertRecipeStepIngredientCreationRequestInputToGRPCRecipeStepIngredientCreationRequestInput(ingredient))
	}

	var recipeStepCompletionConditionCreationRequestInputs []*mealplanningsvc.RecipeStepCompletionConditionCreationRequestInput
	for _, completionCondition := range input.CompletionConditions {
		recipeStepCompletionConditionCreationRequestInputs = append(recipeStepCompletionConditionCreationRequestInputs, ConvertRecipeStepCompletionConditionCreationRequestInputToGRPCRecipeStepCompletionConditionCreationRequestInput(completionCondition))
	}

	return &mealplanningsvc.RecipeStepCreationRequestInput{
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		PreparationId:             input.PreparationID,
		Notes:                     input.Notes,
		ConditionExpression:       input.ConditionExpression,
		ExplicitInstructions:      input.ExplicitInstructions,
		Index:                     input.Index,
		Optional:                  input.Optional,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		Instruments:               recipeStepInstrumentCreationRequestInputs,
		Vessels:                   recipeStepVesselCreationRequestInputs,
		Products:                  recipeStepProductCreationRequestInputs,
		Ingredients:               recipeStepIngredientCreationRequestInputs,
		CompletionConditions:      recipeStepCompletionConditionCreationRequestInputs,
	}
}

func ConvertRecipeToGRPCRecipe(input *mealplanning.Recipe) *mealplanningsvc.Recipe {
	var recipeSteps []*mealplanningsvc.RecipeStep
	for _, step := range input.Steps {
		recipeSteps = append(recipeSteps, ConvertRecipeStepToGRPCRecipeStep(step))
	}

	var recipeMedia []*mealplanningsvc.RecipeMedia
	for _, media := range input.Media {
		recipeMedia = append(recipeMedia, ConvertRecipeMediaToGRPCRecipeMedia(media))
	}

	var recipePrepTasks []*mealplanningsvc.RecipePrepTask
	for _, rps := range input.PrepTasks {
		recipePrepTasks = append(recipePrepTasks, ConvertRecipePrepTaskToGRPCRecipePrepTask(rps))
	}

	var associatedRecipes []*mealplanningsvc.Recipe
	for _, r := range input.AssociatedRecipes {
		associatedRecipes = append(associatedRecipes, ConvertRecipeToGRPCRecipe(r))
	}

	recipe := &mealplanningsvc.Recipe{
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
		CreatedAt:            grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:        grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:           grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Id:                   input.ID,
		YieldsComponentType:  ConvertStringToMealComponentType(input.YieldsComponentType),
		Description:          input.Description,
		Name:                 input.Name,
		PortionName:          input.PortionName,
		CreatedByUser:        input.CreatedByUser,
		Source:               input.Source,
		SourceIsbn:           input.SourceISBN,
		Slug:                 input.Slug,
		PluralPortionName:    input.PluralPortionName,
		Status:               input.Status,
		EligibleForMeals:     input.EligibleForMeals,
		Steps:                recipeSteps,
		Media:                recipeMedia,
		PrepTasks:            recipePrepTasks,
		InspiredByRecipeId:   input.InspiredByRecipeID,
		AssociatedRecipes:    associatedRecipes,
	}

	return recipe
}

func ConvertGRPCRecipeToRecipe(input *mealplanningsvc.Recipe) *mealplanning.Recipe {
	recipeSteps := []*mealplanning.RecipeStep{}
	for _, step := range input.Steps {
		recipeSteps = append(recipeSteps, ConvertGRPCRecipeStepToRecipeStep(step))
	}

	recipeMedia := []*mealplanning.RecipeMedia{}
	for _, media := range input.Media {
		recipeMedia = append(recipeMedia, ConvertGRPCRecipeMediaToRecipeMedia(media))
	}

	recipePrepTasks := []*mealplanning.RecipePrepTask{}
	for _, rps := range input.PrepTasks {
		recipePrepTasks = append(recipePrepTasks, ConvertGRPCRecipePrepTaskToRecipePrepTask(rps))
	}

	associatedRecipes := []*mealplanning.Recipe{}
	for _, r := range input.AssociatedRecipes {
		associatedRecipes = append(associatedRecipes, ConvertGRPCRecipeToRecipe(r))
	}

	recipe := &mealplanning.Recipe{
		MinEstimatedPortions: input.MinEstimatedPortions,

		MaxEstimatedPortions: input.MaxEstimatedPortions,
		CreatedAt:            grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:        grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:           grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		ID:                   input.Id,
		YieldsComponentType:  ConvertMealComponentTypeToString(input.YieldsComponentType),
		Description:          input.Description,
		Name:                 input.Name,
		PortionName:          input.PortionName,
		CreatedByUser:        input.CreatedByUser,
		Source:               input.Source,
		SourceISBN:           input.SourceIsbn,
		Slug:                 input.Slug,
		PluralPortionName:    input.PluralPortionName,
		Status:               input.Status,
		EligibleForMeals:     input.EligibleForMeals,
		Steps:                recipeSteps,
		Media:                recipeMedia,
		PrepTasks:            recipePrepTasks,
		InspiredByRecipeID:   input.InspiredByRecipeId,
		AssociatedRecipes:    associatedRecipes,
	}

	return recipe
}

func ConvertRecipeStepToGRPCRecipeStep(input *mealplanning.RecipeStep) *mealplanningsvc.RecipeStep {
	step := &mealplanningsvc.RecipeStep{
		CreatedAt:                 grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		ArchivedAt:                grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:             grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ExplicitInstructions:      input.ExplicitInstructions,
		BelongsToRecipe:           input.BelongsToRecipe,
		Notes:                     input.Notes,
		ConditionExpression:       input.ConditionExpression,
		Id:                        input.ID,
		Index:                     input.Index,
		Optional:                  input.Optional,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		Preparation:               ConvertValidPreparationToGRPCValidPreparation(&input.Preparation),
	}

	for _, media := range input.Media {
		step.Media = append(step.Media, ConvertRecipeMediaToGRPCRecipeMedia(media))
	}

	for _, instrument := range input.Instruments {
		step.Instruments = append(step.Instruments, ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(instrument))
	}

	for _, vessel := range input.Vessels {
		step.Vessels = append(step.Vessels, ConvertRecipeStepVesselToGRPCRecipeStepVessel(vessel))
	}

	for _, completionCondition := range input.CompletionConditions {
		step.CompletionConditions = append(step.CompletionConditions, ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(completionCondition))
	}

	for _, ingredient := range input.Ingredients {
		step.Ingredients = append(step.Ingredients, ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(ingredient))
	}

	for _, product := range input.Products {
		step.Products = append(step.Products, ConvertRecipeStepProductToGRPCRecipeStepProduct(product))
	}

	for _, img := range input.StepImages {
		step.StepImages = append(step.StepImages, uploadedmediaconverters.ConvertUploadedMediaToGRPCUploadedMedia(img))
	}

	return step
}

func ConvertGRPCRecipeStepToRecipeStep(input *mealplanningsvc.RecipeStep) *mealplanning.RecipeStep {
	step := &mealplanning.RecipeStep{
		CreatedAt:                 grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		ArchivedAt:                grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:             grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ExplicitInstructions:      input.ExplicitInstructions,
		BelongsToRecipe:           input.BelongsToRecipe,
		Notes:                     input.Notes,
		ConditionExpression:       input.ConditionExpression,
		ID:                        input.Id,
		Index:                     input.Index,
		Optional:                  input.Optional,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		Preparation:               *ConvertGRPCValidPreparationToValidPreparation(input.Preparation),
	}

	for _, media := range input.Media {
		step.Media = append(step.Media, ConvertGRPCRecipeMediaToRecipeMedia(media))
	}

	for _, instrument := range input.Instruments {
		step.Instruments = append(step.Instruments, ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(instrument))
	}

	for _, vessel := range input.Vessels {
		step.Vessels = append(step.Vessels, ConvertGRPCRecipeStepVesselToRecipeStepVessel(vessel))
	}

	for _, completionCondition := range input.CompletionConditions {
		step.CompletionConditions = append(step.CompletionConditions, ConvertGRPCRecipeStepCompletionConditionToRecipeStepCompletionCondition(completionCondition))
	}

	for _, ingredient := range input.Ingredients {
		step.Ingredients = append(step.Ingredients, ConvertGRPCRecipeStepIngredientToRecipeStepIngredient(ingredient))
	}

	for _, product := range input.Products {
		step.Products = append(step.Products, ConvertGRPCRecipeStepProductToRecipeStepProduct(product))
	}

	return step
}

func ConvertRecipeStepInstrumentToGRPCRecipeStepInstrument(input *mealplanning.RecipeStepInstrument) *mealplanningsvc.RecipeStepInstrument {
	var convertedInstrument *mealplanningsvc.ValidInstrument
	if input.Instrument != nil {
		convertedInstrument = ConvertValidInstrumentToGRPCValidInstrument(input.Instrument)
	}

	return &mealplanningsvc.RecipeStepInstrument{
		CreatedAt:           grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		Instrument:          convertedInstrument,
		LastUpdatedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		MinQuantity:         input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ArchivedAt:          grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Name:                input.Name,
		Notes:               input.Notes,
		Id:                  input.ID,
		RecipeStepProductId: input.RecipeStepProductID,
		Index:               uint32(input.Index),
		OptionIndex:         uint32(input.OptionIndex),
		PreferenceRank:      uint32(input.PreferenceRank),
		Optional:            input.Optional,
		ScaleFactor:         input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepInstrumentToRecipeStepInstrument(input *mealplanningsvc.RecipeStepInstrument) *mealplanning.RecipeStepInstrument {
	var convertedInstrument *mealplanning.ValidInstrument
	if input.Instrument != nil {
		convertedInstrument = ConvertGRPCValidInstrumentToValidInstrument(input.Instrument)
	}

	scaleFactor := input.ScaleFactor
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &mealplanning.RecipeStepInstrument{
		CreatedAt:     grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		Instrument:    convertedInstrument,
		LastUpdatedAt: grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		MinQuantity:   input.MinQuantity,

		MaxQuantity:         input.MaxQuantity,
		ArchivedAt:          grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Name:                input.Name,
		Notes:               input.Notes,
		ID:                  input.Id,
		RecipeStepProductID: input.RecipeStepProductId,
		Index:               uint16(input.Index),
		OptionIndex:         uint16(input.OptionIndex),
		PreferenceRank:      uint8(input.PreferenceRank),
		Optional:            input.Optional,
		ScaleFactor:         scaleFactor,
	}
}

func ConvertRecipeStepVesselToGRPCRecipeStepVessel(input *mealplanning.RecipeStepVessel) *mealplanningsvc.RecipeStepVessel {
	var validVessel *mealplanningsvc.ValidVessel
	if input.Vessel != nil {
		validVessel = ConvertValidVesselToGRPCValidVessel(input.Vessel)
	}

	return &mealplanningsvc.RecipeStepVessel{
		Vessel:               validVessel,
		MinQuantity:          uint32(input.MinQuantity),
		MaxQuantity:          grpcconverters.ConvertUint16PointerToUint32Pointer(input.MaxQuantity),
		CreatedAt:            grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt:        grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:           grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		VesselPreposition:    input.VesselPreposition,
		Notes:                input.Notes,
		RecipeStepProductId:  input.RecipeStepProductID,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		Id:                   input.ID,
		Name:                 input.Name,
		UnavailableAfterStep: input.UnavailableAfterStep,
		Index:                uint32(input.Index),
		OptionIndex:          uint32(input.OptionIndex),
		ScaleFactor:          input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepVesselToRecipeStepVessel(input *mealplanningsvc.RecipeStepVessel) *mealplanning.RecipeStepVessel {
	var validVessel *mealplanning.ValidVessel
	if input.Vessel != nil {
		validVessel = ConvertGRPCValidVesselToValidVessel(input.Vessel)
	}

	scaleFactor := input.ScaleFactor
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	return &mealplanning.RecipeStepVessel{
		Vessel:      validVessel,
		MinQuantity: uint16(input.MinQuantity),

		MaxQuantity:          grpcconverters.ConvertUint32PointerToUint16Pointer(input.MaxQuantity),
		CreatedAt:            grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:        grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:           grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		VesselPreposition:    input.VesselPreposition,
		Notes:                input.Notes,
		RecipeStepProductID:  input.RecipeStepProductId,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		ID:                   input.Id,
		Name:                 input.Name,
		UnavailableAfterStep: input.UnavailableAfterStep,
		Index:                uint16(input.Index),
		OptionIndex:          uint16(input.OptionIndex),
		ScaleFactor:          scaleFactor,
	}
}

func ConvertRecipeStepCompletionConditionToGRPCRecipeStepCompletionCondition(input *mealplanning.RecipeStepCompletionCondition) *mealplanningsvc.RecipeStepCompletionCondition {
	recipeStepCompletionCondition := &mealplanningsvc.RecipeStepCompletionCondition{
		CreatedAt:           grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		ArchivedAt:          grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:       grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		Id:                  input.ID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Optional:            input.Optional,
		IngredientState:     ConvertValidIngredientStateToGRPCValidIngredientState(&input.IngredientState),
	}

	for _, ingredient := range input.Ingredients {
		recipeStepCompletionCondition.Ingredients = append(recipeStepCompletionCondition.Ingredients, ConvertRecipeStepCompletionConditionIngredientToGRPCRecipeStepCompletionConditionIngredient(ingredient))
	}

	return recipeStepCompletionCondition
}

func ConvertGRPCRecipeStepCompletionConditionToRecipeStepCompletionCondition(input *mealplanningsvc.RecipeStepCompletionCondition) *mealplanning.RecipeStepCompletionCondition {
	recipeStepCompletionCondition := &mealplanning.RecipeStepCompletionCondition{
		CreatedAt:           grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		ArchivedAt:          grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:       grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ID:                  input.Id,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Optional:            input.Optional,
		IngredientState:     *ConvertGRPCValidIngredientStateToValidIngredientState(input.IngredientState),
	}

	for _, ingredient := range input.Ingredients {
		recipeStepCompletionCondition.Ingredients = append(recipeStepCompletionCondition.Ingredients, ConvertGRPCRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredient(ingredient))
	}

	return recipeStepCompletionCondition
}

func ConvertRecipeStepCompletionConditionIngredientToGRPCRecipeStepCompletionConditionIngredient(input *mealplanning.RecipeStepCompletionConditionIngredient) *mealplanningsvc.RecipeStepCompletionConditionIngredient {
	return &mealplanningsvc.RecipeStepCompletionConditionIngredient{
		CreatedAt:                              grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		ArchivedAt:                             grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:                          grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		Id:                                     input.ID,
		BelongsToRecipeStepCompletionCondition: input.BelongsToRecipeStepCompletionCondition,
		RecipeStepIngredient:                   input.RecipeStepIngredient,
	}
}

func ConvertRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInputToGRPCRecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput(input *mealplanning.RecipeStepCompletionConditionIngredientForExistingRecipeCreationRequestInput) *mealplanningsvc.RecipeStepCompletionConditionIngredient {
	return &mealplanningsvc.RecipeStepCompletionConditionIngredient{
		RecipeStepIngredient: input.RecipeStepIngredient,
	}
}

func ConvertGRPCRecipeStepCompletionConditionIngredientToRecipeStepCompletionConditionIngredient(input *mealplanningsvc.RecipeStepCompletionConditionIngredient) *mealplanning.RecipeStepCompletionConditionIngredient {
	return &mealplanning.RecipeStepCompletionConditionIngredient{
		CreatedAt:                              grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		ArchivedAt:                             grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:                          grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ID:                                     input.Id,
		BelongsToRecipeStepCompletionCondition: input.BelongsToRecipeStepCompletionCondition,
		RecipeStepIngredient:                   input.RecipeStepIngredient,
	}
}

func ConvertRecipeStepIngredientToGRPCRecipeStepIngredient(input *mealplanning.RecipeStepIngredient) *mealplanningsvc.RecipeStepIngredient {
	var validIngredient *mealplanningsvc.ValidIngredient
	if input.Ingredient != nil {
		validIngredient = ConvertValidIngredientToGRPCValidIngredient(input.Ingredient)
	}

	ingredient := &mealplanningsvc.RecipeStepIngredient{
		MeasurementUnit:           ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(&input.MeasurementUnit),
		CreatedAt:                 grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		MinQuantity:               input.MinQuantity,
		MaxQuantity:               input.MaxQuantity,
		ArchivedAt:                grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		Ingredient:                validIngredient,
		LastUpdatedAt:             grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		IngredientNotes:           input.IngredientNotes,
		QuantityNotes:             input.QuantityNotes,
		Id:                        input.ID,
		Name:                      input.Name,
		Index:                     uint32(input.Index),
		OptionIndex:               uint32(input.OptionIndex),
		Optional:                  input.Optional,
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		RecipeStepProductRecipeId: input.RecipeStepProductRecipeID,
		RecipeStepProductId:       input.RecipeStepProductID,
		VesselIndex:               grpcconverters.ConvertUint16PointerToUint32Pointer(input.VesselIndex),
		ScaleFactor:               input.ScaleFactor,
	}

	return ingredient
}

func ConvertGRPCRecipeStepIngredientToRecipeStepIngredient(input *mealplanningsvc.RecipeStepIngredient) *mealplanning.RecipeStepIngredient {
	var validIngredient *mealplanning.ValidIngredient
	if input.Ingredient != nil {
		validIngredient = ConvertGRPCValidIngredientToValidIngredient(input.Ingredient)
	}

	scaleFactor := input.ScaleFactor
	if scaleFactor <= 0 {
		scaleFactor = 1.0
	}
	ingredient := &mealplanning.RecipeStepIngredient{
		MeasurementUnit: *ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(input.MeasurementUnit),
		CreatedAt:       grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		MinQuantity:     input.MinQuantity,

		MaxQuantity:               input.MaxQuantity,
		ArchivedAt:                grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		Ingredient:                validIngredient,
		LastUpdatedAt:             grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		IngredientNotes:           input.IngredientNotes,
		QuantityNotes:             input.QuantityNotes,
		ID:                        input.Id,
		Name:                      input.Name,
		Index:                     uint16(input.Index),
		OptionIndex:               uint16(input.OptionIndex),
		Optional:                  input.Optional,
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		RecipeStepProductRecipeID: input.RecipeStepProductRecipeId,
		RecipeStepProductID:       input.RecipeStepProductId,
		VesselIndex:               grpcconverters.ConvertUint32PointerToUint16Pointer(input.VesselIndex),
		ScaleFactor:               scaleFactor,
	}

	return ingredient
}

func ConvertRecipeStepProductToGRPCRecipeStepProduct(input *mealplanning.RecipeStepProduct) *mealplanningsvc.RecipeStepProduct {
	var validMeasurementUnit *mealplanningsvc.ValidMeasurementUnit
	if input.MeasurementUnit != nil {
		validMeasurementUnit = ConvertValidMeasurementUnitToGRPCValidMeasurementUnit(input.MeasurementUnit)
	}

	rsp := &mealplanningsvc.RecipeStepProduct{
		CreatedAt:                      grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		ArchivedAt:                     grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:                  grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		MeasurementUnit:                validMeasurementUnit,
		BelongsToRecipeStep:            input.BelongsToRecipeStep,
		Name:                           input.Name,
		Type:                           ConvertStringToRecipeStepProductType(input.Type),
		Id:                             input.ID,
		StorageInstructions:            input.StorageInstructions,
		QuantityNotes:                  input.QuantityNotes,
		Index:                          uint32(input.Index),
		IsWaste:                        input.IsWaste,
		IsLiquid:                       input.IsLiquid,
		Compostable:                    input.Compostable,
	}

	if input.ContainedInVesselIndex != nil {
		rsp.ContainedInVesselIndex = grpcconverters.ConvertUint16PointerToUint32Pointer(input.ContainedInVesselIndex)
	}

	return rsp
}

func ConvertGRPCRecipeStepProductToRecipeStepProduct(input *mealplanningsvc.RecipeStepProduct) *mealplanning.RecipeStepProduct {
	rsp := &mealplanning.RecipeStepProduct{
		CreatedAt:                      grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		ArchivedAt:                     grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:                  grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		MeasurementUnit:                ConvertGRPCValidMeasurementUnitToValidMeasurementUnit(input.MeasurementUnit),
		BelongsToRecipeStep:            input.BelongsToRecipeStep,
		Name:                           input.Name,
		Type:                           ConvertRecipeStepProductTypeToString(input.Type),
		ID:                             input.Id,
		StorageInstructions:            input.StorageInstructions,
		QuantityNotes:                  input.QuantityNotes,
		Index:                          uint16(input.Index),
		IsWaste:                        input.IsWaste,
		IsLiquid:                       input.IsLiquid,
		Compostable:                    input.Compostable,
		ContainedInVesselIndex:         grpcconverters.ConvertUint32PointerToUint16Pointer(input.ContainedInVesselIndex),
	}

	return rsp
}

func ConvertRecipeMediaToGRPCRecipeMedia(input *mealplanning.RecipeMedia) *mealplanningsvc.RecipeMedia {
	x := &mealplanningsvc.RecipeMedia{
		CreatedAt:       grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		ArchivedAt:      grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:   grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		Id:              input.ID,
		MimeType:        input.MimeType,
		InternalPath:    input.InternalPath,
		ExternalPath:    input.ExternalPath,
		BelongsToRecipe: input.BelongsToRecipe,
		Index:           uint32(input.Index),
	}

	return x
}

func ConvertGRPCRecipeMediaToRecipeMedia(input *mealplanningsvc.RecipeMedia) *mealplanning.RecipeMedia {
	x := &mealplanning.RecipeMedia{
		CreatedAt:       grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		ArchivedAt:      grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:   grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ID:              input.Id,
		MimeType:        input.MimeType,
		InternalPath:    input.InternalPath,
		ExternalPath:    input.ExternalPath,
		BelongsToRecipe: input.BelongsToRecipe,
		Index:           uint16(input.Index),
	}

	return x
}

func ConvertGRPCRecipeMediaCreationRequestInputToRecipeMediaCreationRequestInput(input *mealplanningsvc.RecipeMediaCreationRequestInput) *mealplanning.RecipeMediaCreationRequestInput {
	x := &mealplanning.RecipeMediaCreationRequestInput{
		MimeType:            input.MimeType,
		InternalPath:        input.InternalPath,
		ExternalPath:        input.ExternalPath,
		BelongsToRecipe:     new(input.BelongsToRecipe),
		BelongsToRecipeStep: new(input.BelongsToRecipeStep),
		Index:               uint16(input.Index),
	}

	return x
}

func ConvertRecipeMediaCreationRequestInputToGRPCRecipeMediaCreationRequestInput(input *mealplanning.RecipeMediaCreationRequestInput) *mealplanningsvc.RecipeMediaCreationRequestInput {
	x := &mealplanningsvc.RecipeMediaCreationRequestInput{
		MimeType:            input.MimeType,
		InternalPath:        input.InternalPath,
		ExternalPath:        input.ExternalPath,
		BelongsToRecipe:     pointer.Dereference(input.BelongsToRecipe),
		BelongsToRecipeStep: pointer.Dereference(input.BelongsToRecipeStep),
		Index:               uint32(input.Index),
	}

	return x
}

func ConvertRecipePrepTaskToGRPCRecipePrepTask(input *mealplanning.RecipePrepTask) *mealplanningsvc.RecipePrepTask {
	recipePrepTask := &mealplanningsvc.RecipePrepTask{
		CreatedAt:                          grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		ArchivedAt:                         grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		LastUpdatedAt:                      grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		BelongsToRecipe:                    input.BelongsToRecipe,
		StorageType:                        input.StorageType,
		Id:                                 input.ID,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		Optional:                           input.Optional,
	}

	for _, taskStep := range input.TaskSteps {
		recipePrepTask.TaskSteps = append(recipePrepTask.TaskSteps, ConvertRecipePrepTaskStepToGRPCRecipePrepTaskStep(taskStep))
	}

	return recipePrepTask
}

func ConvertGRPCRecipePrepTaskToRecipePrepTask(input *mealplanningsvc.RecipePrepTask) *mealplanning.RecipePrepTask {
	recipePrepTask := &mealplanning.RecipePrepTask{
		CreatedAt:                          grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		ArchivedAt:                         grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		LastUpdatedAt:                      grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		BelongsToRecipe:                    input.BelongsToRecipe,
		StorageType:                        input.StorageType,
		ID:                                 input.Id,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		Notes:                              input.Notes,
		Name:                               input.Name,
		Description:                        input.Description,
		Optional:                           input.Optional,
	}

	for _, taskStep := range input.TaskSteps {
		recipePrepTask.TaskSteps = append(recipePrepTask.TaskSteps, ConvertGRPCRecipePrepTaskStepToRecipePrepTaskStep(taskStep))
	}

	return recipePrepTask
}

func ConvertRecipePrepTaskStepToGRPCRecipePrepTaskStep(input *mealplanning.RecipePrepTaskStep) *mealplanningsvc.RecipePrepTaskStep {
	return &mealplanningsvc.RecipePrepTaskStep{
		Id:                      input.ID,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
	}
}

func ConvertGRPCRecipePrepTaskStepToRecipePrepTaskStep(input *mealplanningsvc.RecipePrepTaskStep) *mealplanning.RecipePrepTaskStep {
	return &mealplanning.RecipePrepTaskStep{
		ID:                      input.Id,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
	}
}

func ConvertRecipeRatingToGRPCRecipeRating(input *mealplanning.RecipeRating) *mealplanningsvc.RecipeRating {
	return &mealplanningsvc.RecipeRating{
		CreatedAt:     grpcconverters.ConvertTimeToPBTimestamp(input.CreatedAt),
		LastUpdatedAt: grpcconverters.ConvertTimePointerToPBTimestamp(input.LastUpdatedAt),
		ArchivedAt:    grpcconverters.ConvertTimePointerToPBTimestamp(input.ArchivedAt),
		RecipeId:      input.BelongsToRecipe,
		Id:            input.ID,
		Notes:         input.Notes,
		ByUser:        input.CreatedByUser,
		Taste:         input.Taste,
		Instructions:  input.Instructions,
		Overall:       input.Overall,
		Cleanup:       input.Cleanup,
		Difficulty:    input.Difficulty,
	}
}

func ConvertGRPCRecipeRatingToRecipeRating(input *mealplanningsvc.RecipeRating) *mealplanning.RecipeRating {
	return &mealplanning.RecipeRating{
		CreatedAt:       grpcconverters.ConvertPBTimestampToTime(input.CreatedAt),
		LastUpdatedAt:   grpcconverters.ConvertPBTimestampToTimePointer(input.LastUpdatedAt),
		ArchivedAt:      grpcconverters.ConvertPBTimestampToTimePointer(input.ArchivedAt),
		BelongsToRecipe: input.RecipeId,
		ID:              input.Id,
		Notes:           input.Notes,
		CreatedByUser:   input.ByUser,
		Taste:           input.Taste,
		Instructions:    input.Instructions,
		Overall:         input.Overall,
		Cleanup:         input.Cleanup,
		Difficulty:      input.Difficulty,
	}
}

func ConvertGRPCRecipeUpdateRequestInputToRecipeUpdateRequestInput(input *mealplanningsvc.RecipeUpdateRequestInput) *mealplanning.RecipeUpdateRequestInput {
	var componentType *string
	if input.YieldsComponentType != nil {
		componentType = new(ConvertMealComponentTypeToString(*input.YieldsComponentType))
	}

	return &mealplanning.RecipeUpdateRequestInput{
		Name:                 input.Name,
		Slug:                 input.Slug,
		Source:               input.Source,
		SourceISBN:           input.SourceIsbn,
		Description:          input.Description,
		InspiredByRecipeID:   input.InspiredByRecipeId,
		PortionName:          input.PortionName,
		PluralPortionName:    input.PluralPortionName,
		EligibleForMeals:     input.EligibleForMeals,
		YieldsComponentType:  componentType,
		MinEstimatedPortions: input.MinEstimatedPortions,

		MaxEstimatedPortions: input.MaxEstimatedPortions,
	}
}

func ConvertRecipeUpdateRequestInputToGRPCRecipeUpdateRequestInput(input *mealplanning.RecipeUpdateRequestInput) *mealplanningsvc.RecipeUpdateRequestInput {
	var componentType *mealplanningsvc.MealComponentType
	if input.YieldsComponentType != nil {
		componentType = new(ConvertStringToMealComponentType(*input.YieldsComponentType))
	}

	return &mealplanningsvc.RecipeUpdateRequestInput{
		Name:                 input.Name,
		Slug:                 input.Slug,
		Source:               input.Source,
		SourceIsbn:           input.SourceISBN,
		Description:          input.Description,
		InspiredByRecipeId:   input.InspiredByRecipeID,
		PortionName:          input.PortionName,
		PluralPortionName:    input.PluralPortionName,
		EligibleForMeals:     input.EligibleForMeals,
		YieldsComponentType:  componentType,
		MinEstimatedPortions: input.MinEstimatedPortions,
		MaxEstimatedPortions: input.MaxEstimatedPortions,
	}
}

func ConvertGRPCRecipePrepTaskUpdateRequestInputToRecipePrepTaskUpdateRequestInput(input *mealplanningsvc.RecipePrepTaskUpdateRequestInput) *mealplanning.RecipePrepTaskUpdateRequestInput {
	var taskSteps []*mealplanning.RecipePrepTaskStepUpdateRequestInput
	for _, step := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertGRPCRecipePrepTaskStepUpdateRequestInputToRecipePrepTaskStepUpdateRequestInput(step))
	}

	return &mealplanning.RecipePrepTaskUpdateRequestInput{
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		StorageType:                        input.StorageType,
		Name:                               input.Name,
		Optional:                           input.Optional,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
	}
}

func ConvertRecipePrepTaskUpdateRequestInputToGRPCRecipePrepTaskUpdateRequestInput(input *mealplanning.RecipePrepTaskUpdateRequestInput) *mealplanningsvc.RecipePrepTaskUpdateRequestInput {
	var taskSteps []*mealplanningsvc.RecipePrepTaskStepUpdateRequestInput
	for _, step := range input.TaskSteps {
		taskSteps = append(taskSteps, ConvertRecipePrepTaskStepUpdateRequestInputToGRPCRecipePrepTaskStepUpdateRequestInput(step))
	}

	return &mealplanningsvc.RecipePrepTaskUpdateRequestInput{
		Notes:                              input.Notes,
		ExplicitStorageInstructions:        input.ExplicitStorageInstructions,
		StorageType:                        input.StorageType,
		Name:                               input.Name,
		Optional:                           input.Optional,
		Description:                        input.Description,
		BelongsToRecipe:                    input.BelongsToRecipe,
		TaskSteps:                          taskSteps,
		MinTimeBufferBeforeRecipeInSeconds: input.MinTimeBufferBeforeRecipeInSeconds,
		MaxTimeBufferBeforeRecipeInSeconds: input.MaxTimeBufferBeforeRecipeInSeconds,
		MinStorageTemperatureInCelsius:     input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius:     input.MaxStorageTemperatureInCelsius,
	}
}

func ConvertGRPCRecipePrepTaskStepUpdateRequestInputToRecipePrepTaskStepUpdateRequestInput(input *mealplanningsvc.RecipePrepTaskStepUpdateRequestInput) *mealplanning.RecipePrepTaskStepUpdateRequestInput {
	return &mealplanning.RecipePrepTaskStepUpdateRequestInput{
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
	}
}

func ConvertRecipePrepTaskStepUpdateRequestInputToGRPCRecipePrepTaskStepUpdateRequestInput(input *mealplanning.RecipePrepTaskStepUpdateRequestInput) *mealplanningsvc.RecipePrepTaskStepUpdateRequestInput {
	return &mealplanningsvc.RecipePrepTaskStepUpdateRequestInput{
		SatisfiesRecipeStep:     input.SatisfiesRecipeStep,
		BelongsToRecipeStep:     input.BelongsToRecipeStep,
		BelongsToRecipePrepTask: input.BelongsToRecipePrepTask,
	}
}

func ConvertGRPCRecipeRatingUpdateRequestInputToRecipeRatingUpdateRequestInput(input *mealplanningsvc.RecipeRatingUpdateRequestInput) *mealplanning.RecipeRatingUpdateRequestInput {
	return &mealplanning.RecipeRatingUpdateRequestInput{
		BelongsToRecipe: input.RecipeId,
		Taste:           input.Taste,
		Difficulty:      input.Difficulty,
		Cleanup:         input.Cleanup,
		Instructions:    input.Instructions,
		Overall:         input.Overall,
		Notes:           input.Notes,
	}
}

func ConvertRecipeRatingUpdateRequestInputToGRPCRecipeRatingUpdateRequestInput(input *mealplanning.RecipeRatingUpdateRequestInput) *mealplanningsvc.RecipeRatingUpdateRequestInput {
	return &mealplanningsvc.RecipeRatingUpdateRequestInput{
		RecipeId:     input.BelongsToRecipe,
		Taste:        input.Taste,
		Difficulty:   input.Difficulty,
		Cleanup:      input.Cleanup,
		Instructions: input.Instructions,
		Overall:      input.Overall,
		Notes:        input.Notes,
	}
}

func ConvertGRPCRecipeStepUpdateRequestInputToRecipeStepUpdateRequestInput(input *mealplanningsvc.RecipeStepUpdateRequestInput) *mealplanning.RecipeStepUpdateRequestInput {
	return &mealplanning.RecipeStepUpdateRequestInput{
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		Notes:                     input.Notes,
		Preparation:               ConvertGRPCValidPreparationToValidPreparation(input.Preparation),
		Index:                     input.Index,
		Optional:                  input.Optional,
		ExplicitInstructions:      input.ExplicitInstructions,
		ConditionExpression:       input.ConditionExpression,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		BelongsToRecipe:           input.BelongsToRecipe,
	}
}

func ConvertRecipeStepUpdateRequestInputToGRPCRecipeStepUpdateRequestInput(input *mealplanning.RecipeStepUpdateRequestInput) *mealplanningsvc.RecipeStepUpdateRequestInput {
	return &mealplanningsvc.RecipeStepUpdateRequestInput{
		MinEstimatedTimeInSeconds: input.MinEstimatedTimeInSeconds,
		MaxEstimatedTimeInSeconds: input.MaxEstimatedTimeInSeconds,
		MinTemperatureInCelsius:   input.MinTemperatureInCelsius,
		MaxTemperatureInCelsius:   input.MaxTemperatureInCelsius,
		Notes:                     input.Notes,
		Preparation:               ConvertValidPreparationToGRPCValidPreparation(input.Preparation),
		Index:                     input.Index,
		Optional:                  input.Optional,
		ExplicitInstructions:      input.ExplicitInstructions,
		ConditionExpression:       input.ConditionExpression,
		StartTimerAutomatically:   input.StartTimerAutomatically,
		BelongsToRecipe:           input.BelongsToRecipe,
	}
}

func ConvertGRPCRecipeStepCompletionConditionUpdateRequestInputToRecipeStepCompletionConditionUpdateRequestInput(input *mealplanningsvc.RecipeStepCompletionConditionUpdateRequestInput) *mealplanning.RecipeStepCompletionConditionUpdateRequestInput {
	return &mealplanning.RecipeStepCompletionConditionUpdateRequestInput{
		IngredientStateID:   input.IngredientStateId,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Optional:            input.Optional,
	}
}

func ConvertRecipeStepCompletionConditionUpdateRequestInputToGRPCRecipeStepCompletionConditionUpdateRequestInput(input *mealplanning.RecipeStepCompletionConditionUpdateRequestInput) *mealplanningsvc.RecipeStepCompletionConditionUpdateRequestInput {
	return &mealplanningsvc.RecipeStepCompletionConditionUpdateRequestInput{
		IngredientStateId:   input.IngredientStateID,
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Notes:               input.Notes,
		Optional:            input.Optional,
	}
}

func ConvertGRPCRecipeStepIngredientUpdateRequestInputToRecipeStepIngredientUpdateRequestInput(input *mealplanningsvc.RecipeStepIngredientUpdateRequestInput) *mealplanning.RecipeStepIngredientUpdateRequestInput {
	scaleFactor := new(float32(1.0))
	if input.ScaleFactor != nil {
		sf := input.GetScaleFactor()
		if sf > 0 {
			scaleFactor = &sf
		}
	}

	out := &mealplanning.RecipeStepIngredientUpdateRequestInput{
		IngredientID:              input.IngredientId,
		RecipeStepProductID:       input.RecipeStepProductId,
		Name:                      input.Name,
		Optional:                  input.Optional,
		MeasurementUnitID:         input.MeasurementUnitId,
		QuantityNotes:             input.QuantityNotes,
		IngredientNotes:           input.IngredientNotes,
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		Index:                     grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:               grpcconverters.ConvertUint32PointerToUint16Pointer(input.OptionIndex),
		VesselIndex:               grpcconverters.ConvertUint32PointerToUint16Pointer(input.VesselIndex),
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		RecipeStepProductRecipeID: input.RecipeStepProductRecipeId,
		MinQuantity:               input.MinQuantity,

		MaxQuantity: input.MaxQuantity,
		ScaleFactor: scaleFactor,
	}
	return out
}

func ConvertRecipeStepIngredientUpdateRequestInputToGRPCRecipeStepIngredientUpdateRequestInput(input *mealplanning.RecipeStepIngredientUpdateRequestInput) *mealplanningsvc.RecipeStepIngredientUpdateRequestInput {
	return &mealplanningsvc.RecipeStepIngredientUpdateRequestInput{
		IngredientId:              input.IngredientID,
		RecipeStepProductId:       input.RecipeStepProductID,
		Name:                      input.Name,
		Optional:                  input.Optional,
		MeasurementUnitId:         input.MeasurementUnitID,
		QuantityNotes:             input.QuantityNotes,
		IngredientNotes:           input.IngredientNotes,
		BelongsToRecipeStep:       input.BelongsToRecipeStep,
		Index:                     grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:               grpcconverters.ConvertUint16PointerToUint32Pointer(input.OptionIndex),
		VesselIndex:               grpcconverters.ConvertUint16PointerToUint32Pointer(input.VesselIndex),
		ToTaste:                   input.ToTaste,
		ProductPercentageToUse:    input.ProductPercentageToUse,
		RecipeStepProductRecipeId: input.RecipeStepProductRecipeID,
		MinQuantity:               input.MinQuantity,
		MaxQuantity:               input.MaxQuantity,
		ScaleFactor:               input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepInstrumentUpdateRequestInputToRecipeStepInstrumentUpdateRequestInput(input *mealplanningsvc.RecipeStepInstrumentUpdateRequestInput) *mealplanning.RecipeStepInstrumentUpdateRequestInput {
	scaleFactor := new(float32(1.0))
	if input.ScaleFactor != nil {
		sf := input.GetScaleFactor()
		if sf > 0 {
			scaleFactor = &sf
		}
	}

	out := &mealplanning.RecipeStepInstrumentUpdateRequestInput{
		InstrumentID:        input.InstrumentId,
		RecipeStepProductID: input.RecipeStepProductId,
		Notes:               input.Notes,
		PreferenceRank:      grpcconverters.ConvertUint32PointerToUint8Pointer(input.PreferenceRank),
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Name:                input.Name,
		Optional:            input.Optional,
		Index:               grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:         grpcconverters.ConvertUint32PointerToUint16Pointer(input.OptionIndex),
		MinQuantity:         input.MinQuantity,

		MaxQuantity: input.MaxQuantity,
		ScaleFactor: scaleFactor,
	}

	return out
}

func ConvertRecipeStepInstrumentUpdateRequestInputToGRPCRecipeStepInstrumentUpdateRequestInput(input *mealplanning.RecipeStepInstrumentUpdateRequestInput) *mealplanningsvc.RecipeStepInstrumentUpdateRequestInput {
	return &mealplanningsvc.RecipeStepInstrumentUpdateRequestInput{
		InstrumentId:        input.InstrumentID,
		RecipeStepProductId: input.RecipeStepProductID,
		Notes:               input.Notes,
		PreferenceRank:      grpcconverters.ConvertUint8PointerToUint32Pointer(input.PreferenceRank),
		BelongsToRecipeStep: input.BelongsToRecipeStep,
		Name:                input.Name,
		Optional:            input.Optional,
		Index:               grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:         grpcconverters.ConvertUint16PointerToUint32Pointer(input.OptionIndex),
		MinQuantity:         input.MinQuantity,
		MaxQuantity:         input.MaxQuantity,
		ScaleFactor:         input.ScaleFactor,
	}
}

func ConvertGRPCRecipeStepProductUpdateRequestInputToRecipeStepProductUpdateRequestInput(input *mealplanningsvc.RecipeStepProductUpdateRequestInput) *mealplanning.RecipeStepProductUpdateRequestInput {
	var newType *string
	if input.Type != nil {
		newType = new(ConvertRecipeStepProductTypeToString(*input.Type))
	}

	return &mealplanning.RecipeStepProductUpdateRequestInput{
		Name:                           input.Name,
		Type:                           newType,
		MeasurementUnitID:              input.MeasurementUnitId,
		QuantityNotes:                  input.QuantityNotes,
		BelongsToRecipeStep:            input.BelongsToRecipeStep,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		Compostable:                    input.Compostable,
		StorageInstructions:            input.StorageInstructions,
		IsLiquid:                       input.IsLiquid,
		IsWaste:                        input.IsWaste,
		Index:                          grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		ContainedInVesselIndex:         grpcconverters.ConvertUint32PointerToUint16Pointer(input.ContainedInVesselIndex),
	}
}

func ConvertRecipeStepProductUpdateRequestInputToGRPCRecipeStepProductUpdateRequestInput(input *mealplanning.RecipeStepProductUpdateRequestInput) *mealplanningsvc.RecipeStepProductUpdateRequestInput {
	var newType *mealplanningsvc.RecipeStepProductType
	if input.Type != nil {
		newType = new(ConvertStringToRecipeStepProductType(*input.Type))
	}

	return &mealplanningsvc.RecipeStepProductUpdateRequestInput{
		Name:                           input.Name,
		Type:                           newType,
		MeasurementUnitId:              input.MeasurementUnitID,
		QuantityNotes:                  input.QuantityNotes,
		BelongsToRecipeStep:            input.BelongsToRecipeStep,
		MinStorageTemperatureInCelsius: input.MinStorageTemperatureInCelsius,
		MaxStorageTemperatureInCelsius: input.MaxStorageTemperatureInCelsius,
		MinStorageDurationInSeconds:    input.MinStorageDurationInSeconds,
		MaxStorageDurationInSeconds:    input.MaxStorageDurationInSeconds,
		MinMeasurementQuantity:         input.MinMeasurementQuantity,
		MaxMeasurementQuantity:         input.MaxMeasurementQuantity,
		MinItemQuantity:                input.MinItemQuantity,
		MaxItemQuantity:                input.MaxItemQuantity,
		Compostable:                    input.Compostable,
		StorageInstructions:            input.StorageInstructions,
		IsLiquid:                       input.IsLiquid,
		IsWaste:                        input.IsWaste,
		Index:                          grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		ContainedInVesselIndex:         grpcconverters.ConvertUint16PointerToUint32Pointer(input.ContainedInVesselIndex),
	}
}

func ConvertGRPCRecipeStepVesselUpdateRequestInputToRecipeStepVesselUpdateRequestInput(input *mealplanningsvc.RecipeStepVesselUpdateRequestInput) *mealplanning.RecipeStepVesselUpdateRequestInput {
	scaleFactor := new(float32(1.0))
	if input.ScaleFactor != nil {
		sf := input.GetScaleFactor()
		if sf > 0 {
			scaleFactor = &sf
		}
	}

	out := &mealplanning.RecipeStepVesselUpdateRequestInput{
		RecipeStepProductID:  input.RecipeStepProductId,
		Name:                 input.Name,
		Notes:                input.Notes,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		VesselID:             input.VesselId,
		Index:                grpcconverters.ConvertUint32PointerToUint16Pointer(input.Index),
		OptionIndex:          grpcconverters.ConvertUint32PointerToUint16Pointer(input.OptionIndex),
		VesselPreposition:    input.VesselPreposition,
		UnavailableAfterStep: input.UnavailableAfterStep,
		MinQuantity:          grpcconverters.ConvertUint32PointerToUint16Pointer(input.MinQuantity),

		MaxQuantity: grpcconverters.ConvertUint32PointerToUint16Pointer(input.MaxQuantity),
		ScaleFactor: scaleFactor,
	}

	return out
}

func ConvertRecipeStepVesselUpdateRequestInputToGRPCRecipeStepVesselUpdateRequestInput(input *mealplanning.RecipeStepVesselUpdateRequestInput) *mealplanningsvc.RecipeStepVesselUpdateRequestInput {
	return &mealplanningsvc.RecipeStepVesselUpdateRequestInput{
		RecipeStepProductId:  input.RecipeStepProductID,
		Name:                 input.Name,
		Notes:                input.Notes,
		BelongsToRecipeStep:  input.BelongsToRecipeStep,
		VesselId:             input.VesselID,
		Index:                grpcconverters.ConvertUint16PointerToUint32Pointer(input.Index),
		OptionIndex:          grpcconverters.ConvertUint16PointerToUint32Pointer(input.OptionIndex),
		VesselPreposition:    input.VesselPreposition,
		UnavailableAfterStep: input.UnavailableAfterStep,
		MinQuantity:          grpcconverters.ConvertUint16PointerToUint32Pointer(input.MinQuantity),
		MaxQuantity:          grpcconverters.ConvertUint16PointerToUint32Pointer(input.MaxQuantity),
		ScaleFactor:          input.ScaleFactor,
	}
}
