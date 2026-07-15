package mockmanagers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"

	"github.com/primandproper/platform-go/v4/filtering"
)

// Valid-enumeration mock methods for MockMealPlanningManager. Struct is defined in meal_planning_manager.go.

func (m *MockMealPlanningManager) SearchValidIngredientGroups(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredientGroup(ctx context.Context, input *mealplanning.ValidIngredientGroupCreationRequestInput) (*mealplanning.ValidIngredientGroup, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientGroup), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredientGroup(ctx context.Context, validIngredientGroupID string) (*mealplanning.ValidIngredientGroup, error) {
	returnValues := m.Called(ctx, validIngredientGroupID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientGroup), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredientGroup(ctx context.Context, validIngredientGroupID string, input *mealplanning.ValidIngredientGroupUpdateRequestInput) (*mealplanning.ValidIngredientGroup, error) {
	returnValues := m.Called(ctx, validIngredientGroupID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientGroup), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredientGroup(ctx context.Context, validIngredientGroupID string) error {
	returnValues := m.Called(ctx, validIngredientGroupID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredientMeasurementUnit(ctx context.Context, input *mealplanning.ValidIngredientMeasurementUnitCreationRequestInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, validIngredientMeasurementUnitID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string, input *mealplanning.ValidIngredientMeasurementUnitUpdateRequestInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, validIngredientMeasurementUnitID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error {
	returnValues := m.Called(ctx, validIngredientMeasurementUnitID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidIngredientMeasurementUnitsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, validIngredientID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidIngredientMeasurementUnitsByMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, validMeasurementUnitID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredientPreparation(ctx context.Context, input *mealplanning.ValidIngredientPreparationCreationRequestInput) (*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, validIngredientPreparationID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string, input *mealplanning.ValidIngredientPreparationUpdateRequestInput) (*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, validIngredientPreparationID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error {
	returnValues := m.Called(ctx, validIngredientPreparationID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidIngredientPreparationsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, ingredientID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidIngredientPreparationsByPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, preparationID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidPrepTaskConfig(ctx context.Context, input *mealplanning.ValidPrepTaskConfigCreationRequestInput) (*mealplanning.ValidPrepTaskConfig, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidPrepTaskConfig), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) (*mealplanning.ValidPrepTaskConfig, error) {
	returnValues := m.Called(ctx, validPrepTaskConfigID)

	return returnValues.Get(0).(*mealplanning.ValidPrepTaskConfig), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string, input *mealplanning.ValidPrepTaskConfigUpdateRequestInput) (*mealplanning.ValidPrepTaskConfig, error) {
	returnValues := m.Called(ctx, validPrepTaskConfigID, input)

	return returnValues.Get(0).(*mealplanning.ValidPrepTaskConfig), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidPrepTaskConfig(ctx context.Context, validPrepTaskConfigID string) error {
	returnValues := m.Called(ctx, validPrepTaskConfigID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidPrepTaskConfigsByIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, ingredientID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidPrepTaskConfigsByPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, preparationID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidPrepTaskConfigsByIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, ingredientID, preparationID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidIngredients(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredient(ctx context.Context, input *mealplanning.ValidIngredientCreationRequestInput) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredient(ctx context.Context, validIngredientID string) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, validIngredientID)

	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RandomValidIngredient(ctx context.Context) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredient(ctx context.Context, validIngredientID string, input *mealplanning.ValidIngredientUpdateRequestInput) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, validIngredientID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredient(ctx context.Context, validIngredientID string) error {
	returnValues := m.Called(ctx, validIngredientID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error {
	returnValues := m.Called(ctx, validIngredientID, uploadedMediaID, index)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidIngredientsByPreparationAndIngredientName(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, preparationID, query, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredientStateIngredient(ctx context.Context, input *mealplanning.ValidIngredientStateIngredientCreationRequestInput) (*mealplanning.ValidIngredientStateIngredient, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientStateIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) (*mealplanning.ValidIngredientStateIngredient, error) {
	returnValues := m.Called(ctx, validIngredientStateIngredientID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientStateIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string, input *mealplanning.ValidIngredientStateIngredientUpdateRequestInput) (*mealplanning.ValidIngredientStateIngredient, error) {
	returnValues := m.Called(ctx, validIngredientStateIngredientID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientStateIngredient), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) error {
	returnValues := m.Called(ctx, validIngredientStateIngredientID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidIngredientStateIngredientsByIngredient(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, validIngredientID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidIngredientStateIngredientsByIngredientState(ctx context.Context, validIngredientStateID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, validIngredientStateID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidIngredientStates(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientState]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidIngredientStates(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientState]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidIngredientState(ctx context.Context, input *mealplanning.ValidIngredientStateCreationRequestInput) (*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientState), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidIngredientState(ctx context.Context, validIngredientStateID string) (*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, validIngredientStateID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientState), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidIngredientState(ctx context.Context, validIngredientStateID string, input *mealplanning.ValidIngredientStateUpdateRequestInput) (*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, validIngredientStateID, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientState), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidIngredientState(ctx context.Context, validIngredientStateID string) error {
	returnValues := m.Called(ctx, validIngredientStateID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidMeasurementUnits(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidMeasurementUnitsByIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, validIngredientID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidMeasurementUnit(ctx context.Context, input *mealplanning.ValidMeasurementUnitCreationRequestInput) (*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string, input *mealplanning.ValidMeasurementUnitUpdateRequestInput) (*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID, input)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error {
	returnValues := m.Called(ctx, validMeasurementUnitID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidInstruments(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidInstrument(ctx context.Context, input *mealplanning.ValidInstrumentCreationRequestInput) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidInstrument(ctx context.Context, validInstrumentID string) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, validInstrumentID)

	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RandomValidInstrument(ctx context.Context) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidInstrument(ctx context.Context, validInstrumentID string, input *mealplanning.ValidInstrumentUpdateRequestInput) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, validInstrumentID, input)

	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error {
	returnValues := m.Called(ctx, validInstrumentID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ValidMeasurementUnitConversionsForMeasurementUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
	returnValues := m.Called(ctx, validMeasurementUnitID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validIngredientIDs)

	return returnValues.Get(0).([]*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*mealplanning.MeasurementUnitConversionMismatch, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).([]*mealplanning.MeasurementUnitConversionMismatch), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ValidMeasurementUnitConversionsFromMeasurementUnit(ctx context.Context, validMeasurementUnitID string) ([]*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID)

	return returnValues.Get(0).([]*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ValidMeasurementUnitConversionsToMeasurementUnit(ctx context.Context, validMeasurementUnitID string) ([]*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID)

	return returnValues.Get(0).([]*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidMeasurementUnitConversion(ctx context.Context, input *mealplanning.ValidMeasurementUnitConversionCreationRequestInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) (*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validMeasurementUnitConversionID)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string, input *mealplanning.ValidMeasurementUnitConversionUpdateRequestInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validMeasurementUnitConversionID, input)

	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidMeasurementUnitConversion(ctx context.Context, validMeasurementUnitConversionID string) error {
	returnValues := m.Called(ctx, validMeasurementUnitConversionID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidPreparationInstrument(ctx context.Context, input *mealplanning.ValidPreparationInstrumentCreationRequestInput) (*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, validPreparationInstrumentID)

	return returnValues.Get(0).(*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string, input *mealplanning.ValidPreparationInstrumentUpdateRequestInput) (*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, validPreparationInstrumentID, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error {
	returnValues := m.Called(ctx, validPreparationInstrumentID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidPreparationInstrumentsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, validPreparationID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidPreparationInstrumentsByInstrument(ctx context.Context, validInstrumentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, validInstrumentID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidPreparations(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparation]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparation]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidPreparation(ctx context.Context, input *mealplanning.ValidPreparationCreationRequestInput) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidPreparation(ctx context.Context, validPreparationID string) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, validPreparationID)

	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RandomValidPreparation(ctx context.Context) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidPreparation(ctx context.Context, validPreparationID string, input *mealplanning.ValidPreparationUpdateRequestInput) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, validPreparationID, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidPreparation(ctx context.Context, validPreparationID string) error {
	returnValues := m.Called(ctx, validPreparationID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error {
	returnValues := m.Called(ctx, validPreparationID, forIngredientID, uploadedMediaID, index)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListValidPreparationVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidPreparationVessel(ctx context.Context, input *mealplanning.ValidPreparationVesselCreationRequestInput) (*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidPreparationVessel(ctx context.Context, validPreparationVesselID string) (*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, validPreparationVesselID)

	return returnValues.Get(0).(*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidPreparationVessel(ctx context.Context, validPreparationVesselID string, input *mealplanning.ValidPreparationVesselUpdateRequestInput) (*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, validPreparationVesselID, input)

	return returnValues.Get(0).(*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidPreparationVessel(ctx context.Context, validPreparationVesselID string) error {
	returnValues := m.Called(ctx, validPreparationVesselID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) SearchValidPreparationVesselsByPreparation(ctx context.Context, validPreparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, validPreparationID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidPreparationVesselsByVessel(ctx context.Context, validVesselID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, validVesselID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) SearchValidVessels(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ListValidVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidVessel]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateValidVessel(ctx context.Context, input *mealplanning.ValidVesselCreationRequestInput) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ReadValidVessel(ctx context.Context, validVesselID string) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, validVesselID)

	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) RandomValidVessel(ctx context.Context) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx)

	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateValidVessel(ctx context.Context, validVesselID string, input *mealplanning.ValidVesselUpdateRequestInput) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, validVesselID, input)

	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

func (m *MockMealPlanningManager) ArchiveValidVessel(ctx context.Context, validVesselID string) error {
	returnValues := m.Called(ctx, validVesselID)

	return returnValues.Error(0)
}
