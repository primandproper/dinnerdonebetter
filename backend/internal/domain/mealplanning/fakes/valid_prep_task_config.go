package fakes

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v5/filtering"

	fake "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeValidPrepTaskConfig builds a faked valid prep task config.
func BuildFakeValidPrepTaskConfig() *types.ValidPrepTaskConfig {
	minSD, maxSD := BuildFakeUint32WithOptionalMax()
	minST, maxST := BuildFakeOptionalFloat32MinMax()
	return &types.ValidPrepTaskConfig{
		ID:                             BuildFakeID(),
		MinStorageDurationInSeconds:    minSD,
		MaxStorageDurationInSeconds:    maxSD,
		MinStorageTemperatureInCelsius: minST,
		MaxStorageTemperatureInCelsius: maxST,
		StorageType: fake.RandomString([]string{
			types.RecipePrepTaskStorageTypeUncovered,
			types.RecipePrepTaskStorageTypeCovered,
			types.RecipePrepTaskStorageTypeAirtightContainer,
			types.RecipePrepTaskStorageTypeWireRack,
		}),
		StorageInstructions: buildUniqueString(),
		Notes:               buildUniqueString(),
		Source:              buildUniqueString(),
		Preparation:         *BuildFakeValidPreparation(),
		Ingredient:          *BuildFakeValidIngredient(),
		CreatedAt:           BuildFakeTime(),
	}
}

// BuildFakeValidPrepTaskConfigsList builds a faked ValidPrepTaskConfigList.
func BuildFakeValidPrepTaskConfigsList() *filtering.QueryFilteredResult[types.ValidPrepTaskConfig] {
	var examples []*types.ValidPrepTaskConfig
	for range exampleQuantity {
		examples = append(examples, BuildFakeValidPrepTaskConfig())
	}

	return &filtering.QueryFilteredResult[types.ValidPrepTaskConfig]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeValidPrepTaskConfigUpdateRequestInput builds a faked ValidPrepTaskConfigUpdateRequestInput from a valid prep task config.
func BuildFakeValidPrepTaskConfigUpdateRequestInput() *types.ValidPrepTaskConfigUpdateRequestInput {
	validPrepTaskConfig := BuildFakeValidPrepTaskConfig()
	return converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigUpdateRequestInput(validPrepTaskConfig)
}

// BuildFakeValidPrepTaskConfigCreationRequestInput builds a faked ValidPrepTaskConfigCreationRequestInput.
func BuildFakeValidPrepTaskConfigCreationRequestInput() *types.ValidPrepTaskConfigCreationRequestInput {
	validPrepTaskConfig := BuildFakeValidPrepTaskConfig()
	return converters.ConvertValidPrepTaskConfigToValidPrepTaskConfigCreationRequestInput(validPrepTaskConfig)
}
