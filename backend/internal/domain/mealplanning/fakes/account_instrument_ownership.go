package fakes

import (
	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"

	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
)

// BuildFakeAccountInstrumentOwnership builds a faked account instrument ownership.
func BuildFakeAccountInstrumentOwnership() *types.AccountInstrumentOwnership {
	ownership := fake.BuildFakeRecord[types.AccountInstrumentOwnership]()

	// The instrument owned, built by its own builder.
	ownership.Instrument = *BuildFakeValidInstrument()

	return ownership
}

// BuildFakeAccountInstrumentOwnershipsList builds a faked AccountInstrumentOwnershipList.
func BuildFakeAccountInstrumentOwnershipsList() *filtering.QueryFilteredResult[types.AccountInstrumentOwnership] {
	return fake.BuildFakePage(BuildFakeAccountInstrumentOwnership)
}

// BuildFakeAccountInstrumentOwnershipUpdateRequestInput builds a faked AccountInstrumentOwnershipUpdateRequestInput from an ownership.
func BuildFakeAccountInstrumentOwnershipUpdateRequestInput() *types.AccountInstrumentOwnershipUpdateRequestInput {
	ownership := BuildFakeAccountInstrumentOwnership()

	return converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipUpdateRequestInput(ownership)
}

// BuildFakeAccountInstrumentOwnershipCreationRequestInput builds a faked AccountInstrumentOwnershipCreationRequestInput.
func BuildFakeAccountInstrumentOwnershipCreationRequestInput() *types.AccountInstrumentOwnershipCreationRequestInput {
	ownership := BuildFakeAccountInstrumentOwnership()

	return converters.ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput(ownership)
}
