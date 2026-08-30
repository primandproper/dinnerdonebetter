package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"

	gofakeit "github.com/brianvoe/gofakeit/v7"
)

// BuildFakeValidPrepTaskConfig builds a faked valid prep task config.
func BuildFakeValidPrepTaskConfig() *types.ValidPrepTaskConfig {
	cfg := fake.BuildFakeRecord[types.ValidPrepTaskConfig]()

	cfg.MinStorageDurationInSeconds, cfg.MaxStorageDurationInSeconds = BuildFakeUint32WithOptionalMax()
	cfg.MinStorageTemperatureInCelsius, cfg.MaxStorageTemperatureInCelsius = BuildFakeOptionalFloat32MinMax()

	// One of the four ways this domain knows to store something between steps.
	cfg.StorageType = gofakeit.RandomString([]string{
		types.RecipePrepTaskStorageTypeUncovered,
		types.RecipePrepTaskStorageTypeCovered,
		types.RecipePrepTaskStorageTypeAirtightContainer,
		types.RecipePrepTaskStorageTypeWireRack,
	})

	cfg.Preparation = *BuildFakeValidPreparation()
	cfg.Ingredient = *BuildFakeValidIngredient()

	return cfg
}

// BuildFakeValidPrepTaskConfigsList builds a faked ValidPrepTaskConfigList.
func BuildFakeValidPrepTaskConfigsList() *filtering.QueryFilteredResult[types.ValidPrepTaskConfig] {
	return fake.BuildFakePage(BuildFakeValidPrepTaskConfig)
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
