package converters

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v6/identifiers"
)

// ConvertValidPreparationToValidPreparationUpdateRequestInput creates a ValidPreparationUpdateRequestInput from a ValidPreparation.
func ConvertValidPreparationToValidPreparationUpdateRequestInput(input *mealplanning.ValidPreparation) *mealplanning.ValidPreparationUpdateRequestInput {
	x := &mealplanning.ValidPreparationUpdateRequestInput{
		Name:                        &input.Name,
		Description:                 &input.Description,
		IconPath:                    &input.IconPath,
		YieldsNothing:               &input.YieldsNothing,
		RestrictToIngredients:       &input.RestrictToIngredients,
		Slug:                        &input.Slug,
		PastTense:                   &input.PastTense,
		MinIngredientCount:          &input.MinIngredientCount,
		MaxIngredientCount:          input.MaxIngredientCount,
		MinInstrumentCount:          &input.MinInstrumentCount,
		MaxInstrumentCount:          input.MaxInstrumentCount,
		MinVesselCount:              &input.MinVesselCount,
		MaxVesselCount:              input.MaxVesselCount,
		TemperatureRequired:         &input.TemperatureRequired,
		TimeEstimateRequired:        &input.TimeEstimateRequired,
		ConditionExpressionRequired: &input.ConditionExpressionRequired,
		ConsumesVessel:              &input.ConsumesVessel,
		OnlyForVessels:              &input.OnlyForVessels,
	}

	return x
}

// ConvertValidPreparationCreationRequestInputToValidPreparationDatabaseCreationInput creates a ValidPreparationDatabaseCreationInput from a ValidPreparationCreationRequestInput.
func ConvertValidPreparationCreationRequestInputToValidPreparationDatabaseCreationInput(input *mealplanning.ValidPreparationCreationRequestInput) *mealplanning.ValidPreparationDatabaseCreationInput {
	x := &mealplanning.ValidPreparationDatabaseCreationInput{
		ID:                          identifiers.New(),
		Name:                        input.Name,
		Description:                 input.Description,
		IconPath:                    input.IconPath,
		YieldsNothing:               input.YieldsNothing,
		RestrictToIngredients:       input.RestrictToIngredients,
		Slug:                        input.Slug,
		PastTense:                   input.PastTense,
		MinIngredientCount:          input.MinIngredientCount,
		MaxIngredientCount:          input.MaxIngredientCount,
		MinInstrumentCount:          input.MinInstrumentCount,
		MaxInstrumentCount:          input.MaxInstrumentCount,
		MinVesselCount:              input.MinVesselCount,
		MaxVesselCount:              input.MaxVesselCount,
		TemperatureRequired:         input.TemperatureRequired,
		TimeEstimateRequired:        input.TimeEstimateRequired,
		ConditionExpressionRequired: input.ConditionExpressionRequired,
		ConsumesVessel:              input.ConsumesVessel,
		OnlyForVessels:              input.OnlyForVessels,
	}

	return x
}

// ConvertValidPreparationToValidPreparationCreationRequestInput builds a ValidPreparationCreationRequestInput from a ValidPreparation.
func ConvertValidPreparationToValidPreparationCreationRequestInput(validPreparation *mealplanning.ValidPreparation) *mealplanning.ValidPreparationCreationRequestInput {
	return &mealplanning.ValidPreparationCreationRequestInput{
		Name:                        validPreparation.Name,
		Description:                 validPreparation.Description,
		IconPath:                    validPreparation.IconPath,
		YieldsNothing:               validPreparation.YieldsNothing,
		RestrictToIngredients:       validPreparation.RestrictToIngredients,
		Slug:                        validPreparation.Slug,
		PastTense:                   validPreparation.PastTense,
		MinIngredientCount:          validPreparation.MinIngredientCount,
		MaxIngredientCount:          validPreparation.MaxIngredientCount,
		MinInstrumentCount:          validPreparation.MinInstrumentCount,
		MaxInstrumentCount:          validPreparation.MaxInstrumentCount,
		MinVesselCount:              validPreparation.MinVesselCount,
		MaxVesselCount:              validPreparation.MaxVesselCount,
		TemperatureRequired:         validPreparation.TemperatureRequired,
		TimeEstimateRequired:        validPreparation.TimeEstimateRequired,
		ConditionExpressionRequired: validPreparation.ConditionExpressionRequired,
		ConsumesVessel:              validPreparation.ConsumesVessel,
		OnlyForVessels:              validPreparation.OnlyForVessels,
	}
}

// ConvertValidPreparationToValidPreparationDatabaseCreationInput builds a ValidPreparationDatabaseCreationInput from a ValidPreparation.
func ConvertValidPreparationToValidPreparationDatabaseCreationInput(validPreparation *mealplanning.ValidPreparation) *mealplanning.ValidPreparationDatabaseCreationInput {
	return &mealplanning.ValidPreparationDatabaseCreationInput{
		ID:                          validPreparation.ID,
		Name:                        validPreparation.Name,
		Description:                 validPreparation.Description,
		IconPath:                    validPreparation.IconPath,
		YieldsNothing:               validPreparation.YieldsNothing,
		RestrictToIngredients:       validPreparation.RestrictToIngredients,
		Slug:                        validPreparation.Slug,
		PastTense:                   validPreparation.PastTense,
		MinIngredientCount:          validPreparation.MinIngredientCount,
		MaxIngredientCount:          validPreparation.MaxIngredientCount,
		MinInstrumentCount:          validPreparation.MinInstrumentCount,
		MaxInstrumentCount:          validPreparation.MaxInstrumentCount,
		MinVesselCount:              validPreparation.MinVesselCount,
		MaxVesselCount:              validPreparation.MaxVesselCount,
		TemperatureRequired:         validPreparation.TemperatureRequired,
		TimeEstimateRequired:        validPreparation.TimeEstimateRequired,
		ConditionExpressionRequired: validPreparation.ConditionExpressionRequired,
		ConsumesVessel:              validPreparation.ConsumesVessel,
		OnlyForVessels:              validPreparation.OnlyForVessels,
	}
}
