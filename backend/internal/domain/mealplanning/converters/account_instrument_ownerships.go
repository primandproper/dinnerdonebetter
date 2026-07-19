package converters

import (
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v5/identifiers"
)

// ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipUpdateRequestInput creates a DatabaseCreationInput from a CreationInput.
func ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipUpdateRequestInput(x *types.AccountInstrumentOwnership) *types.AccountInstrumentOwnershipUpdateRequestInput {
	out := &types.AccountInstrumentOwnershipUpdateRequestInput{
		Notes:             &x.Notes,
		Quantity:          &x.Quantity,
		ValidInstrumentID: &x.Instrument.ID,
	}

	return out
}

// ConvertAccountInstrumentOwnershipCreationRequestInputToAccountInstrumentOwnershipDatabaseCreationInput creates a DatabaseCreationInput from a CreationInput.
func ConvertAccountInstrumentOwnershipCreationRequestInputToAccountInstrumentOwnershipDatabaseCreationInput(x *types.AccountInstrumentOwnershipCreationRequestInput) *types.AccountInstrumentOwnershipDatabaseCreationInput {
	out := &types.AccountInstrumentOwnershipDatabaseCreationInput{
		ID:                identifiers.New(),
		Notes:             x.Notes,
		Quantity:          x.Quantity,
		ValidInstrumentID: x.ValidInstrumentID,
	}

	return out
}

// ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput builds a AccountInstrumentOwnershipCreationRequestInput from a Ingredient.
func ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipCreationRequestInput(x *types.AccountInstrumentOwnership) *types.AccountInstrumentOwnershipCreationRequestInput {
	return &types.AccountInstrumentOwnershipCreationRequestInput{
		Notes:             x.Notes,
		Quantity:          x.Quantity,
		ValidInstrumentID: x.Instrument.ID,
	}
}

// ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipDatabaseCreationInput builds a AccountInstrumentOwnershipDatabaseCreationInput from a AccountInstrumentOwnership.
func ConvertAccountInstrumentOwnershipToAccountInstrumentOwnershipDatabaseCreationInput(x *types.AccountInstrumentOwnership) *types.AccountInstrumentOwnershipDatabaseCreationInput {
	return &types.AccountInstrumentOwnershipDatabaseCreationInput{
		ID:                x.ID,
		Notes:             x.Notes,
		Quantity:          x.Quantity,
		ValidInstrumentID: x.Instrument.ID,
		BelongsToAccount:  x.BelongsToAccount,
	}
}
