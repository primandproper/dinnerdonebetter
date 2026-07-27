package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v7/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeValidPreparation builds a faked valid preparation.
func BuildFakeValidPreparation() *types.ValidPreparation {
	minIngredientCount, maxIngredientCount := BuildFakeUint16WithOptionalMax()
	minInstrumentCount, maxInstrumentCount := BuildFakeUint16WithOptionalMax()
	minVesselCount, maxVesselCount := BuildFakeUint16WithOptionalMax()
	return &types.ValidPreparation{
		ID:                          BuildFakeID(),
		Name:                        buildUniqueString(),
		Description:                 buildUniqueString(),
		IconPath:                    buildUniqueString(),
		YieldsNothing:               fake.Bool(),
		RestrictToIngredients:       fake.Bool(),
		Slug:                        buildUniqueString(),
		PastTense:                   buildUniqueString(),
		CreatedAt:                   BuildFakeTime(),
		MinIngredientCount:          minIngredientCount,
		MaxIngredientCount:          maxIngredientCount,
		MinInstrumentCount:          minInstrumentCount,
		MaxInstrumentCount:          maxInstrumentCount,
		MinVesselCount:              minVesselCount,
		MaxVesselCount:              maxVesselCount,
		TemperatureRequired:         fake.Bool(),
		TimeEstimateRequired:        fake.Bool(),
		ConditionExpressionRequired: fake.Bool(),
		ConsumesVessel:              fake.Bool(),
		OnlyForVessels:              fake.Bool(),
	}
}

// BuildFakeValidPreparationsList builds a faked ValidPreparationList.
func BuildFakeValidPreparationsList() *filtering.QueryFilteredResult[types.ValidPreparation] {
	var examples []*types.ValidPreparation
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidPreparation())
	}

	return &filtering.QueryFilteredResult[types.ValidPreparation]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidPreparationUpdateRequestInput builds a faked ValidPreparationUpdateRequestInput from a valid preparation.
func BuildFakeValidPreparationUpdateRequestInput() *types.ValidPreparationUpdateRequestInput {
	validPreparation := BuildFakeValidPreparation()
	return converters.ConvertValidPreparationToValidPreparationUpdateRequestInput(validPreparation)
}

// BuildFakeValidPreparationCreationRequestInput builds a faked ValidPreparationCreationRequestInput.
func BuildFakeValidPreparationCreationRequestInput() *types.ValidPreparationCreationRequestInput {
	validPreparation := BuildFakeValidPreparation()
	return converters.ConvertValidPreparationToValidPreparationCreationRequestInput(validPreparation)
}
