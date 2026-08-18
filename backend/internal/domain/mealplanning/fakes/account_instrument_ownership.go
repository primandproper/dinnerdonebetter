package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v11/filtering"
	"github.com/primandproper/platform-go/v11/identifiers"
)

// BuildFakeAccountInstrumentOwnership builds a faked valid ingredient.
func BuildFakeAccountInstrumentOwnership() *types.AccountInstrumentOwnership {
	return &types.AccountInstrumentOwnership{
		CreatedAt:        BuildFakeTime(),
		ID:               identifiers.New(),
		Notes:            buildUniqueString(),
		BelongsToAccount: buildUniqueString(),
		Instrument:       *BuildFakeValidInstrument(),
		Quantity:         uint16(buildFakeNumber()),
	}
}

// BuildFakeAccountInstrumentOwnershipsList builds a faked AccountInstrumentOwnershipList.
func BuildFakeAccountInstrumentOwnershipsList() *filtering.QueryFilteredResult[types.AccountInstrumentOwnership] {
	var examples []*types.AccountInstrumentOwnership
	for range exampleQuantity {
		examples = append(examples, BuildFakeAccountInstrumentOwnership())
	}

	return &filtering.QueryFilteredResult[types.AccountInstrumentOwnership]{
		Pagination: filtering.Pagination{
			Cursor:          BuildFakeID(),
			MaxResponseSize: 50,
			FilteredCount:   exampleQuantity / 2,
			TotalCount:      exampleQuantity,
		},
		Data: examples,
	}
}

// BuildFakeAccountInstrumentOwnershipUpdateRequestInput builds a faked AccountInstrumentOwnershipUpdateRequestInput from a valid ingredient.
func BuildFakeAccountInstrumentOwnershipUpdateRequestInput() *types.AccountInstrumentOwnershipUpdateRequestInput {
	validIngredient := BuildFakeAccountInstrumentOwnership()
	return converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipUpdateRequestInput(validIngredient)
}

// BuildFakeAccountInstrumentOwnershipCreationRequestInput builds a faked AccountInstrumentOwnershipCreationRequestInput.
func BuildFakeAccountInstrumentOwnershipCreationRequestInput() *types.AccountInstrumentOwnershipCreationRequestInput {
	validIngredient := BuildFakeAccountInstrumentOwnership()
	return converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput(validIngredient)
}
