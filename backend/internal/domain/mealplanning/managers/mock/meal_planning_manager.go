package mockmanagers

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	mealplanningmgr "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning/managers"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ mealplanningmgr.MealPlanningManager = (*MockMealPlanningManager)(nil)

type MockMealPlanningManager struct {
	mock.Mock
}

// ListMeals is a mock method.
func (m *MockMealPlanningManager) ListMeals(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
	returnValues := m.Called(ctx, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Meal]), returnValues.Error(1)
}

// CreateMeal is a mock method.
func (m *MockMealPlanningManager) CreateMeal(ctx context.Context, creatorID string, input *mealplanning.MealCreationRequestInput) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, creatorID, input)

	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// ReadMeal is a mock method.
func (m *MockMealPlanningManager) ReadMeal(ctx context.Context, mealID string) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, mealID)

	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// SearchMeals is a mock method.
func (m *MockMealPlanningManager) SearchMeals(ctx context.Context, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
	returnValues := m.Called(ctx, query, useSearchService, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Meal]), returnValues.Error(1)
}

// ArchiveMeal is a mock method.
func (m *MockMealPlanningManager) ArchiveMeal(ctx context.Context, mealID, ownerID string) error {
	returnValues := m.Called(ctx, mealID, ownerID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddMealImage(ctx context.Context, mealID, uploadedMediaID, uploadedByUser string) error {
	returnValues := m.Called(ctx, mealID, uploadedMediaID, uploadedByUser)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListMealLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealList], error) {
	returnValues := m.Called(ctx, userID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealList]), returnValues.Error(1)
}

func (m *MockMealPlanningManager) CreateMealList(ctx context.Context, userID string, input *mealplanning.MealListCreationRequestInput) (*mealplanning.MealList, error) {
	returnValues := m.Called(ctx, userID, input)

	return returnValues.Get(0).(*mealplanning.MealList), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateMealList(ctx context.Context, mealListID, userID string, input *mealplanning.MealListUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealListID, userID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ArchiveMealList(ctx context.Context, mealListID, userID string) error {
	returnValues := m.Called(ctx, mealListID, userID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) AddMealToMealList(ctx context.Context, mealListID, mealID, notes string) (*mealplanning.MealListItem, error) {
	returnValues := m.Called(ctx, mealListID, mealID, notes)

	return returnValues.Get(0).(*mealplanning.MealListItem), returnValues.Error(1)
}

func (m *MockMealPlanningManager) UpdateMealListItem(ctx context.Context, mealListItemID, mealListID, mealID string, input *mealplanning.MealListItemUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealListItemID, mealListID, mealID, input)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) RemoveMealFromMealList(ctx context.Context, mealListID, mealListItemID string) error {
	returnValues := m.Called(ctx, mealListID, mealListItemID)

	return returnValues.Error(0)
}

func (m *MockMealPlanningManager) ListMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealListItem], error) {
	returnValues := m.Called(ctx, mealListID, userID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealListItem]), returnValues.Error(1)
}

// ListMealPlans is a mock method.
func (m *MockMealPlanningManager) ListMealPlans(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
	returnValues := m.Called(ctx, ownerID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlan]), returnValues.Error(1)
}

// CreateMealPlan is a mock method.
func (m *MockMealPlanningManager) CreateMealPlan(ctx context.Context, ownerID, creatorID string, input *mealplanning.MealPlanCreationRequestInput) (*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx, ownerID, creatorID, input)

	return returnValues.Get(0).(*mealplanning.MealPlan), returnValues.Error(1)
}

// ReadMealPlan is a mock method.
func (m *MockMealPlanningManager) ReadMealPlan(ctx context.Context, mealPlanID, ownerID string) (*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx, mealPlanID, ownerID)

	return returnValues.Get(0).(*mealplanning.MealPlan), returnValues.Error(1)
}

// UpdateMealPlan is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlan(ctx context.Context, mealPlanID, ownerID string, input *mealplanning.MealPlanUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanID, ownerID, input)

	return returnValues.Error(0)
}

// ArchiveMealPlan is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlan(ctx context.Context, mealPlanID, ownerID string) error {
	returnValues := m.Called(ctx, mealPlanID, ownerID)

	return returnValues.Error(0)
}

// FinalizeMealPlan is a mock method.
func (m *MockMealPlanningManager) FinalizeMealPlan(ctx context.Context, mealPlanID, ownerID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, ownerID)

	return returnValues.Get(0).(bool), returnValues.Error(1)
}

// ListMealPlanEvents is a mock method.
func (m *MockMealPlanningManager) ListMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanEvent], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanEvent]), returnValues.Error(1)
}

// CreateMealPlanEvent is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanEvent(ctx context.Context, mealPlanID string, input *mealplanning.MealPlanEventCreationRequestInput) (*mealplanning.MealPlanEvent, error) {
	returnValues := m.Called(ctx, mealPlanID, input)

	return returnValues.Get(0).(*mealplanning.MealPlanEvent), returnValues.Error(1)
}

// ReadMealPlanEvent is a mock method.
func (m *MockMealPlanningManager) ReadMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*mealplanning.MealPlanEvent, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID)

	return returnValues.Get(0).(*mealplanning.MealPlanEvent), returnValues.Error(1)
}

// UpdateMealPlanEvent is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string, input *mealplanning.MealPlanEventUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, input)

	return returnValues.Error(0)
}

// SwapMealPlanEvents is a mock method.
func (m *MockMealPlanningManager) SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventIDA, mealPlanEventIDB)

	return returnValues.Error(0)
}

// ArchiveMealPlanEvent is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID)

	return returnValues.Error(0)
}

// ListMealPlanOptions is a mock method.
func (m *MockMealPlanningManager) ListMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOption], error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanOption]), returnValues.Error(1)
}

// CreateMealPlanOption is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanOption(ctx context.Context, input *mealplanning.MealPlanOptionCreationRequestInput) (*mealplanning.MealPlanOption, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.MealPlanOption), returnValues.Error(1)
}

// CreateMealPlanOptionWithEventID is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanOptionWithEventID(ctx context.Context, mealPlanEventID string, input *mealplanning.MealPlanOptionCreationRequestInput) (*mealplanning.MealPlanOption, error) {
	returnValues := m.Called(ctx, mealPlanEventID, input)

	return returnValues.Get(0).(*mealplanning.MealPlanOption), returnValues.Error(1)
}

// MealPlanOptionBelongsToAccount is a mock method.
func (m *MockMealPlanningManager) MealPlanOptionBelongsToAccount(ctx context.Context, mealPlanOptionID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanOptionID, accountID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// ReadMealPlanOption is a mock method.
func (m *MockMealPlanningManager) ReadMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)

	return returnValues.Get(0).(*mealplanning.MealPlanOption), returnValues.Error(1)
}

// UpdateMealPlanOption is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, input *mealplanning.MealPlanOptionUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, input)

	return returnValues.Error(0)
}

// ArchiveMealPlanOption is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)

	return returnValues.Error(0)
}

// ListMealPlanOptionVotes is a mock method.
func (m *MockMealPlanningManager) ListMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOptionVote], error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanOptionVote]), returnValues.Error(1)
}

// CreateMealPlanOptionVotes is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, creatorID string, input *mealplanning.MealPlanOptionVoteCreationRequestInput) ([]*mealplanning.MealPlanOptionVote, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, creatorID, input)

	return returnValues.Get(0).([]*mealplanning.MealPlanOptionVote), returnValues.Error(1)
}

// ReadMealPlanOptionVote is a mock method.
func (m *MockMealPlanningManager) ReadMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*mealplanning.MealPlanOptionVote, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)

	return returnValues.Get(0).(*mealplanning.MealPlanOptionVote), returnValues.Error(1)
}

// UpdateMealPlanOptionVote is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string, input *mealplanning.MealPlanOptionVoteUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID, input)

	return returnValues.Error(0)
}

// ArchiveMealPlanOptionVote is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)

	return returnValues.Error(0)
}

// ListMealPlanTasksByMealPlan is a mock method.
func (m *MockMealPlanningManager) ListMealPlanTasksByMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanTask], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanTask]), returnValues.Error(1)
}

// ReadMealPlanTask is a mock method.
func (m *MockMealPlanningManager) ReadMealPlanTask(ctx context.Context, mealPlanID, mealPlanTaskID string) (*mealplanning.MealPlanTask, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanTaskID)

	return returnValues.Get(0).(*mealplanning.MealPlanTask), returnValues.Error(1)
}

// CreateMealPlanTask is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanTask(ctx context.Context, input *mealplanning.MealPlanTaskCreationRequestInput) (*mealplanning.MealPlanTask, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.MealPlanTask), returnValues.Error(1)
}

// MealPlanTaskStatusChange is a mock method.
func (m *MockMealPlanningManager) MealPlanTaskStatusChange(ctx context.Context, input *mealplanning.MealPlanTaskStatusChangeRequestInput) error {
	returnValues := m.Called(ctx, input)

	return returnValues.Error(0)
}

// ListMealPlanGroceryListItemsByMealPlan is a mock method.
func (m *MockMealPlanningManager) ListMealPlanGroceryListItemsByMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem]), returnValues.Error(1)
}

// CreateMealPlanGroceryListItem is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanGroceryListItem(ctx context.Context, input *mealplanning.MealPlanGroceryListItemCreationRequestInput) (*mealplanning.MealPlanGroceryListItem, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.MealPlanGroceryListItem), returnValues.Error(1)
}

// ReadMealPlanGroceryListItem is a mock method.
func (m *MockMealPlanningManager) ReadMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (*mealplanning.MealPlanGroceryListItem, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanGroceryListItemID)

	return returnValues.Get(0).(*mealplanning.MealPlanGroceryListItem), returnValues.Error(1)
}

// UpdateMealPlanGroceryListItem is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string, input *mealplanning.MealPlanGroceryListItemUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanGroceryListItemID, input)

	return returnValues.Error(0)
}

// ArchiveMealPlanGroceryListItem is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) error {
	returnValues := m.Called(ctx, mealPlanID, mealPlanGroceryListItemID)

	return returnValues.Error(0)
}

// GetMealPlanRecipeOptionSelection is a mock method.
func (m *MockMealPlanningManager) GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*mealplanning.MealPlanRecipeOptionSelection, error) {
	returnValues := m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*mealplanning.MealPlanRecipeOptionSelection), returnValues.Error(1)
}

// GetMealPlanRecipeOptionSelectionsForMealPlanOption is a mock method.
func (m *MockMealPlanningManager) GetMealPlanRecipeOptionSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanRecipeOptionSelection], error) {
	returnValues := m.Called(ctx, mealPlanOptionID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanRecipeOptionSelection]), returnValues.Error(1)
}

// CreateMealPlanRecipeOptionSelection is a mock method.
func (m *MockMealPlanningManager) CreateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID string, input *mealplanning.MealPlanRecipeOptionSelectionCreationRequestInput) (*mealplanning.MealPlanRecipeOptionSelection, error) {
	returnValues := m.Called(ctx, mealPlanOptionID, input)

	return returnValues.Get(0).(*mealplanning.MealPlanRecipeOptionSelection), returnValues.Error(1)
}

// UpdateMealPlanRecipeOptionSelection is a mock method.
func (m *MockMealPlanningManager) UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *mealplanning.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
	returnValues := m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, input)

	return returnValues.Error(0)
}

// ArchiveMealPlanRecipeOptionSelection is a mock method.
func (m *MockMealPlanningManager) ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error {
	returnValues := m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)

	return returnValues.Error(0)
}

// ListUserIngredientPreferences is a mock method.
func (m *MockMealPlanningManager) ListUserIngredientPreferences(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
	returnValues := m.Called(ctx, ownerID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference]), returnValues.Error(1)
}

// ReadUserIngredientPreference is a mock method.
func (m *MockMealPlanningManager) ReadUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) (*mealplanning.UserIngredientPreference, error) {
	returnValues := m.Called(ctx, ownerID, ingredientPreferenceID)

	return returnValues.Get(0).(*mealplanning.UserIngredientPreference), returnValues.Error(1)
}

// CreateUserIngredientPreference is a mock method.
func (m *MockMealPlanningManager) CreateUserIngredientPreference(ctx context.Context, ownerID string, input *mealplanning.UserIngredientPreferenceCreationRequestInput) ([]*mealplanning.UserIngredientPreference, error) {
	returnValues := m.Called(ctx, ownerID, input)

	return returnValues.Get(0).([]*mealplanning.UserIngredientPreference), returnValues.Error(1)
}

// UpdateUserIngredientPreference is a mock method.
func (m *MockMealPlanningManager) UpdateUserIngredientPreference(ctx context.Context, ingredientPreferenceID, ownerID string, input *mealplanning.UserIngredientPreferenceUpdateRequestInput) error {
	returnValues := m.Called(ctx, ingredientPreferenceID, ownerID, input)

	return returnValues.Error(0)
}

// ArchiveUserIngredientPreference is a mock method.
func (m *MockMealPlanningManager) ArchiveUserIngredientPreference(ctx context.Context, ownerID, ingredientPreferenceID string) error {
	returnValues := m.Called(ctx, ownerID, ingredientPreferenceID)

	return returnValues.Error(0)
}

// ListAccountInstrumentOwnerships is a mock method.
func (m *MockMealPlanningManager) ListAccountInstrumentOwnerships(ctx context.Context, ownerID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
	returnValues := m.Called(ctx, ownerID, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership]), returnValues.Error(1)
}

// SearchValidInstrumentsNotOwnedByAccount is a mock method.
func (m *MockMealPlanningManager) SearchValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, useSearchService bool, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, accountID, query, useSearchService, filter)

	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

// CreateAccountInstrumentOwnership is a mock method.
func (m *MockMealPlanningManager) CreateAccountInstrumentOwnership(ctx context.Context, ownerID string, input *mealplanning.AccountInstrumentOwnershipCreationRequestInput) (*mealplanning.AccountInstrumentOwnership, error) {
	returnValues := m.Called(ctx, ownerID, input)

	return returnValues.Get(0).(*mealplanning.AccountInstrumentOwnership), returnValues.Error(1)
}

// ReadAccountInstrumentOwnership is a mock method.
func (m *MockMealPlanningManager) ReadAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) (*mealplanning.AccountInstrumentOwnership, error) {
	returnValues := m.Called(ctx, ownerID, instrumentOwnershipID)

	return returnValues.Get(0).(*mealplanning.AccountInstrumentOwnership), returnValues.Error(1)
}

// UpdateAccountInstrumentOwnership is a mock method.
func (m *MockMealPlanningManager) UpdateAccountInstrumentOwnership(ctx context.Context, instrumentOwnershipID, ownerID string, input *mealplanning.AccountInstrumentOwnershipUpdateRequestInput) error {
	returnValues := m.Called(ctx, instrumentOwnershipID, ownerID, input)

	return returnValues.Error(0)
}

// ArchiveAccountInstrumentOwnership is a mock method.
func (m *MockMealPlanningManager) ArchiveAccountInstrumentOwnership(ctx context.Context, ownerID, instrumentOwnershipID string) error {
	returnValues := m.Called(ctx, ownerID, instrumentOwnershipID)

	return returnValues.Error(0)
}
