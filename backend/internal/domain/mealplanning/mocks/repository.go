package mocks

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ mealplanning.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// ValidInstrumentExists is a mock function.
func (m *Repository) ValidInstrumentExists(ctx context.Context, validInstrumentID string) (bool, error) {
	returnValues := m.Called(ctx, validInstrumentID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidInstrument is a mock function.
func (m *Repository) GetValidInstrument(ctx context.Context, validInstrumentID string) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, validInstrumentID)
	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

// GetRandomValidInstrument is a mock function.
func (m *Repository) GetRandomValidInstrument(ctx context.Context) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

// SearchForValidInstruments is a mock function.
func (m *Repository) SearchForValidInstruments(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

// SearchForValidInstrumentsNotOwnedByAccount is a mock function.
func (m *Repository) SearchForValidInstrumentsNotOwnedByAccount(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, accountID, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

// GetValidInstruments is a mock function.
func (m *Repository) GetValidInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidInstrument], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidInstrument]), returnValues.Error(1)
}

// CreateValidInstrument is a mock function.
func (m *Repository) CreateValidInstrument(ctx context.Context, input *mealplanning.ValidInstrumentDatabaseCreationInput) (*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidInstrument), returnValues.Error(1)
}

// UpdateValidInstrument is a mock function.
func (m *Repository) UpdateValidInstrument(ctx context.Context, updated *mealplanning.ValidInstrument) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidInstrument is a mock function.
func (m *Repository) ArchiveValidInstrument(ctx context.Context, validInstrumentID string) error {
	return m.Called(ctx, validInstrumentID).Error(0)
}

// MarkValidInstrumentAsIndexed is a mock function.
func (m *Repository) MarkValidInstrumentAsIndexed(ctx context.Context, validInstrumentID string) error {
	return m.Called(ctx, validInstrumentID).Error(0)
}

// GetValidInstrumentIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidInstrumentIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetValidInstrumentsWithIDs is a mock function.
func (m *Repository) GetValidInstrumentsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidInstrument, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidInstrument), returnValues.Error(1)
}

// ValidIngredientExists is a mock function.
func (m *Repository) ValidIngredientExists(ctx context.Context, validIngredientID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredient is a mock function.
func (m *Repository) GetValidIngredient(ctx context.Context, validIngredientID string) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, validIngredientID)
	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

// GetRandomValidIngredient is a mock function.
func (m *Repository) GetRandomValidIngredient(ctx context.Context) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

// SearchForValidIngredients is a mock function.
func (m *Repository) SearchForValidIngredients(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

// SearchForValidIngredientsForPreparation is a mock function.
func (m *Repository) SearchForValidIngredientsForPreparation(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, preparationID, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

// GetValidIngredients is a mock function.
func (m *Repository) GetValidIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredient], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredient]), returnValues.Error(1)
}

// CreateValidIngredient is a mock function.
func (m *Repository) CreateValidIngredient(ctx context.Context, input *mealplanning.ValidIngredientDatabaseCreationInput) (*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidIngredient), returnValues.Error(1)
}

// UpdateValidIngredient is a mock function.
func (m *Repository) UpdateValidIngredient(ctx context.Context, updated *mealplanning.ValidIngredient) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredient is a mock function.
func (m *Repository) ArchiveValidIngredient(ctx context.Context, validIngredientID string) error {
	return m.Called(ctx, validIngredientID).Error(0)
}

// MarkValidIngredientAsIndexed is a mock function.
func (m *Repository) MarkValidIngredientAsIndexed(ctx context.Context, validIngredientID string) error {
	return m.Called(ctx, validIngredientID).Error(0)
}

// GetValidIngredientIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidIngredientIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetValidIngredientsWithIDs is a mock function.
func (m *Repository) GetValidIngredientsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidIngredient, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidIngredient), returnValues.Error(1)
}

// ValidIngredientGroupExists is a mock method.
func (m *Repository) ValidIngredientGroupExists(ctx context.Context, validIngredientID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredientGroup is a mock method.
func (m *Repository) GetValidIngredientGroup(ctx context.Context, validIngredientID string) (*mealplanning.ValidIngredientGroup, error) {
	returnValues := m.Called(ctx, validIngredientID)

	return returnValues.Get(0).(*mealplanning.ValidIngredientGroup), returnValues.Error(1)
}

// GetValidIngredientGroups is a mock method.
func (m *Repository) GetValidIngredientGroups(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
	returnValues := m.Called(ctx, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup]), returnValues.Error(1)
}

// SearchForValidIngredientGroups is a mock method.
func (m *Repository) SearchForValidIngredientGroups(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup], error) {
	returnValues := m.Called(ctx, query, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientGroup]), returnValues.Error(1)
}

// CreateValidIngredientGroup is a mock method.
func (m *Repository) CreateValidIngredientGroup(ctx context.Context, input *mealplanning.ValidIngredientGroupDatabaseCreationInput) (*mealplanning.ValidIngredientGroup, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.ValidIngredientGroup), returnValues.Error(1)
}

// UpdateValidIngredientGroup is a mock method.
func (m *Repository) UpdateValidIngredientGroup(ctx context.Context, updated *mealplanning.ValidIngredientGroup) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredientGroup is a mock method.
func (m *Repository) ArchiveValidIngredientGroup(ctx context.Context, validIngredientID string) error {
	return m.Called(ctx, validIngredientID).Error(0)
}

// ValidPreparationExists is a mock function.
func (m *Repository) ValidPreparationExists(ctx context.Context, validPreparationID string) (bool, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidPreparation is a mock function.
func (m *Repository) GetValidPreparation(ctx context.Context, validPreparationID string) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

// GetRandomValidPreparation is a mock function.
func (m *Repository) GetRandomValidPreparation(ctx context.Context) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

// SearchForValidPreparations is a mock function.
func (m *Repository) SearchForValidPreparations(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparation]), returnValues.Error(1)
}

// GetValidPreparations is a mock function.
func (m *Repository) GetValidPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparation], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparation]), returnValues.Error(1)
}

// CreateValidPreparation is a mock function.
func (m *Repository) CreateValidPreparation(ctx context.Context, input *mealplanning.ValidPreparationDatabaseCreationInput) (*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidPreparation), returnValues.Error(1)
}

// UpdateValidPreparation is a mock function.
func (m *Repository) UpdateValidPreparation(ctx context.Context, updated *mealplanning.ValidPreparation) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidPreparation is a mock function.
func (m *Repository) ArchiveValidPreparation(ctx context.Context, validPreparationID string) error {
	return m.Called(ctx, validPreparationID).Error(0)
}

// MarkValidPreparationAsIndexed is a mock function.
func (m *Repository) MarkValidPreparationAsIndexed(ctx context.Context, validPreparationID string) error {
	return m.Called(ctx, validPreparationID).Error(0)
}

// GetValidPreparationIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidPreparationIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetValidPreparationsWithIDs is a mock function.
func (m *Repository) GetValidPreparationsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidPreparation, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidPreparation), returnValues.Error(1)
}

// ValidIngredientPreparationExists is a mock function.
func (m *Repository) ValidIngredientPreparationExists(ctx context.Context, validIngredientPreparationID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientPreparationID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredientPreparation is a mock function.
func (m *Repository) GetValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) (*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, validIngredientPreparationID)
	return returnValues.Get(0).(*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

// GetValidIngredientPreparations is a mock function.
func (m *Repository) GetValidIngredientPreparations(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

// GetValidIngredientPreparationsForIngredient is a mock function.
func (m *Repository) GetValidIngredientPreparationsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, ingredientID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

// GetValidIngredientPreparationsForPreparation is a mock function.
func (m *Repository) GetValidIngredientPreparationsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, preparationID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

// GetValidIngredientPreparationsForIngredientNameQuery is a mock function.
func (m *Repository) GetValidIngredientPreparationsForIngredientNameQuery(ctx context.Context, preparationID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation], error) {
	returnValues := m.Called(ctx, preparationID, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientPreparation]), returnValues.Error(1)
}

// CreateValidIngredientPreparation is a mock function.
func (m *Repository) CreateValidIngredientPreparation(ctx context.Context, input *mealplanning.ValidIngredientPreparationDatabaseCreationInput) (*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

// UpdateValidIngredientPreparation is a mock function.
func (m *Repository) UpdateValidIngredientPreparation(ctx context.Context, updated *mealplanning.ValidIngredientPreparation) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredientPreparation is a mock function.
func (m *Repository) ArchiveValidIngredientPreparation(ctx context.Context, validIngredientPreparationID string) error {
	return m.Called(ctx, validIngredientPreparationID).Error(0)
}

// GetValidIngredientPreparationsByIDs is a mock function.
func (m *Repository) GetValidIngredientPreparationsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidIngredientPreparation, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).(map[string]*mealplanning.ValidIngredientPreparation), returnValues.Error(1)
}

// ValidPrepTaskConfigExists is a mock function.
func (m *Repository) ValidPrepTaskConfigExists(ctx context.Context, validIngredientPreparationStorageConfigID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientPreparationStorageConfigID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidPrepTaskConfig is a mock function.
func (m *Repository) GetValidPrepTaskConfig(ctx context.Context, validIngredientPreparationStorageConfigID string) (*mealplanning.ValidPrepTaskConfig, error) {
	returnValues := m.Called(ctx, validIngredientPreparationStorageConfigID)
	return returnValues.Get(0).(*mealplanning.ValidPrepTaskConfig), returnValues.Error(1)
}

// GetValidPrepTaskConfigs is a mock function.
func (m *Repository) GetValidPrepTaskConfigs(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

// GetValidPrepTaskConfigsForIngredient is a mock function.
func (m *Repository) GetValidPrepTaskConfigsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, ingredientID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

// GetValidPrepTaskConfigsForPreparation is a mock function.
func (m *Repository) GetValidPrepTaskConfigsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, preparationID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

// GetValidPrepTaskConfigsForIngredientAndPreparation is a mock function.
func (m *Repository) GetValidPrepTaskConfigsForIngredientAndPreparation(ctx context.Context, ingredientID, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig], error) {
	returnValues := m.Called(ctx, ingredientID, preparationID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPrepTaskConfig]), returnValues.Error(1)
}

// CreateValidPrepTaskConfig is a mock function.
func (m *Repository) CreateValidPrepTaskConfig(ctx context.Context, input *mealplanning.ValidPrepTaskConfigDatabaseCreationInput) (*mealplanning.ValidPrepTaskConfig, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidPrepTaskConfig), returnValues.Error(1)
}

// UpdateValidPrepTaskConfig is a mock function.
func (m *Repository) UpdateValidPrepTaskConfig(ctx context.Context, updated *mealplanning.ValidPrepTaskConfig) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidPrepTaskConfig is a mock function.
func (m *Repository) ArchiveValidPrepTaskConfig(ctx context.Context, validIngredientPreparationStorageConfigID string) error {
	return m.Called(ctx, validIngredientPreparationStorageConfigID).Error(0)
}

// MealExists is a mock function.
func (m *Repository) MealExists(ctx context.Context, recipeID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMeal is a mock function.
func (m *Repository) GetMeal(ctx context.Context, recipeID string) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, recipeID)
	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// GetMealByIDAndUser is a mock function.
func (m *Repository) GetMealByIDAndUser(ctx context.Context, recipeID, userID string) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, recipeID, userID)
	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// GetMeals is a mock function.
func (m *Repository) GetMeals(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Meal]), returnValues.Error(1)
}

func (m *Repository) GetMealsCreatedByUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (x *filtering.QueryFilteredResult[mealplanning.Meal], err error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Meal]), returnValues.Error(1)
}

// SearchForMeals is a mock function.
func (m *Repository) SearchForMeals(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Meal], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Meal]), returnValues.Error(1)
}

// FindMealWithSameComponents is a mock function.
func (m *Repository) FindMealWithSameComponents(ctx context.Context, creatorID string, input *mealplanning.MealCreationRequestInput) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, creatorID, input)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// CreateMeal is a mock function.
func (m *Repository) CreateMeal(ctx context.Context, input *mealplanning.MealDatabaseCreationInput) (*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.Meal), returnValues.Error(1)
}

// UpdateMeal is a mock function.
func (m *Repository) UpdateMeal(ctx context.Context, updated *mealplanning.Meal) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMeal is a mock function.
func (m *Repository) ArchiveMeal(ctx context.Context, recipeID, accountID string) error {
	return m.Called(ctx, recipeID, accountID).Error(0)
}

// AddMealImage is a mock function.
func (m *Repository) AddMealImage(ctx context.Context, mealID, uploadedMediaID, uploadedByUser string) error {
	return m.Called(ctx, mealID, uploadedMediaID, uploadedByUser).Error(0)
}

// MarkMealAsIndexed is a mock function.
func (m *Repository) MarkMealAsIndexed(ctx context.Context, mealID string) error {
	return m.Called(ctx, mealID).Error(0)
}

// GetMealIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetMealIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetMealsWithIDs is a mock function.
func (m *Repository) GetMealsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.Meal, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.Meal), returnValues.Error(1)
}

// RecipeExists is a mock function.
func (m *Repository) RecipeExists(ctx context.Context, recipeID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipe is a mock function.
func (m *Repository) GetRecipe(ctx context.Context, recipeID string) (*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, recipeID)
	return returnValues.Get(0).(*mealplanning.Recipe), returnValues.Error(1)
}

// SearchForRecipes is a mock function.
func (m *Repository) SearchForRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

// SearchForMealEligibleRecipes is a mock function.
func (m *Repository) SearchForMealEligibleRecipes(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

// SearchForRecipesWithInstrumentOwnership is a mock function.
func (m *Repository) SearchForRecipesWithInstrumentOwnership(ctx context.Context, accountID, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, accountID, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

// GetRecipes is a mock function.
func (m *Repository) GetRecipes(ctx context.Context, status string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, status, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

// GetRecipesCreatedByUser is a mock function.
func (m *Repository) GetRecipesCreatedByUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.Recipe], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.Recipe]), returnValues.Error(1)
}

// CreateRecipe is a mock function.
func (m *Repository) CreateRecipe(ctx context.Context, input *mealplanning.RecipeDatabaseCreationInput) (*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.Recipe), returnValues.Error(1)
}

// UpdateRecipe is a mock function.
func (m *Repository) UpdateRecipe(ctx context.Context, updated *mealplanning.Recipe) error {
	return m.Called(ctx, updated).Error(0)
}

func (m *Repository) UpdateRecipeStatus(ctx context.Context, recipeID, newStatus string) error {
	return m.Called(ctx, recipeID, newStatus).Error(0)
}

// ArchiveRecipe is a mock function.
func (m *Repository) ArchiveRecipe(ctx context.Context, recipeID, accountID string) error {
	return m.Called(ctx, recipeID, accountID).Error(0)
}

// AddRecipeImage is a mock function.
func (m *Repository) AddRecipeImage(ctx context.Context, recipeID, uploadedMediaID, uploadedByUser string) error {
	return m.Called(ctx, recipeID, uploadedMediaID, uploadedByUser).Error(0)
}

// MarkRecipeAsIndexed is a mock function.
func (m *Repository) MarkRecipeAsIndexed(ctx context.Context, recipeID string) error {
	return m.Called(ctx, recipeID).Error(0)
}

// GetRecipeIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetRecipeIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetRecipesWithIDs is a mock function.
func (m *Repository) GetRecipesWithIDs(ctx context.Context, ids []string) ([]*mealplanning.Recipe, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.Recipe), returnValues.Error(1)
}

// RecipeStepExists is a mock function.
func (m *Repository) RecipeStepExists(ctx context.Context, recipeID, recipeStepID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStep is a mock function.
func (m *Repository) GetRecipeStep(ctx context.Context, recipeID, recipeStepID string) (*mealplanning.RecipeStep, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID)
	return returnValues.Get(0).(*mealplanning.RecipeStep), returnValues.Error(1)
}

// GetRecipeSteps is a mock function.
func (m *Repository) GetRecipeSteps(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStep], error) {
	returnValues := m.Called(ctx, recipeID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStep]), returnValues.Error(1)
}

// CreateRecipeStep is a mock function.
func (m *Repository) CreateRecipeStep(ctx context.Context, input *mealplanning.RecipeStepDatabaseCreationInput) (*mealplanning.RecipeStep, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStep), returnValues.Error(1)
}

// UpdateRecipeStep is a mock function.
func (m *Repository) UpdateRecipeStep(ctx context.Context, updated *mealplanning.RecipeStep) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStep is a mock function.
func (m *Repository) ArchiveRecipeStep(ctx context.Context, recipeID, recipeStepID string) error {
	return m.Called(ctx, recipeID, recipeStepID).Error(0)
}

// RecipeStepProductExists is a mock function.
func (m *Repository) RecipeStepProductExists(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepProductID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStepProduct is a mock function.
func (m *Repository) GetRecipeStepProduct(ctx context.Context, recipeID, recipeStepID, recipeStepProductID string) (*mealplanning.RecipeStepProduct, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepProductID)
	return returnValues.Get(0).(*mealplanning.RecipeStepProduct), returnValues.Error(1)
}

// GetRecipeStepProducts is a mock function.
func (m *Repository) GetRecipeStepProducts(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepProduct], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepProduct]), returnValues.Error(1)
}

// CreateRecipeStepProduct is a mock function.
func (m *Repository) CreateRecipeStepProduct(ctx context.Context, input *mealplanning.RecipeStepProductDatabaseCreationInput) (*mealplanning.RecipeStepProduct, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStepProduct), returnValues.Error(1)
}

// UpdateRecipeStepProduct is a mock function.
func (m *Repository) UpdateRecipeStepProduct(ctx context.Context, updated *mealplanning.RecipeStepProduct) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStepProduct is a mock function.
func (m *Repository) ArchiveRecipeStepProduct(ctx context.Context, recipeStepID, recipeStepProductID string) error {
	return m.Called(ctx, recipeStepID, recipeStepProductID).Error(0)
}

// RecipeStepInstrumentExists is a mock function.
func (m *Repository) RecipeStepInstrumentExists(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepInstrumentID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStepInstrument is a mock function.
func (m *Repository) GetRecipeStepInstrument(ctx context.Context, recipeID, recipeStepID, recipeStepInstrumentID string) (*mealplanning.RecipeStepInstrument, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepInstrumentID)
	return returnValues.Get(0).(*mealplanning.RecipeStepInstrument), returnValues.Error(1)
}

// GetRecipeStepInstruments is a mock function.
func (m *Repository) GetRecipeStepInstruments(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepInstrument]), returnValues.Error(1)
}

// CreateRecipeStepInstrument is a mock function.
func (m *Repository) CreateRecipeStepInstrument(ctx context.Context, input *mealplanning.RecipeStepInstrumentDatabaseCreationInput) (*mealplanning.RecipeStepInstrument, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStepInstrument), returnValues.Error(1)
}

// UpdateRecipeStepInstrument is a mock function.
func (m *Repository) UpdateRecipeStepInstrument(ctx context.Context, updated *mealplanning.RecipeStepInstrument) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStepInstrument is a mock function.
func (m *Repository) ArchiveRecipeStepInstrument(ctx context.Context, recipeStepID, recipeStepInstrumentID string) error {
	return m.Called(ctx, recipeStepID, recipeStepInstrumentID).Error(0)
}

// RecipeStepIngredientExists is a mock function.
func (m *Repository) RecipeStepIngredientExists(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStepIngredient is a mock function.
func (m *Repository) GetRecipeStepIngredient(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*mealplanning.RecipeStepIngredient, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	return returnValues.Get(0).(*mealplanning.RecipeStepIngredient), returnValues.Error(1)
}

// GetRecipeStepIngredients is a mock function.
func (m *Repository) GetRecipeStepIngredients(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepIngredient]), returnValues.Error(1)
}

// CreateRecipeStepIngredient is a mock function.
func (m *Repository) CreateRecipeStepIngredient(ctx context.Context, input *mealplanning.RecipeStepIngredientDatabaseCreationInput) (*mealplanning.RecipeStepIngredient, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStepIngredient), returnValues.Error(1)
}

// UpdateRecipeStepIngredient is a mock function.
func (m *Repository) UpdateRecipeStepIngredient(ctx context.Context, updated *mealplanning.RecipeStepIngredient) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStepIngredient is a mock function.
func (m *Repository) ArchiveRecipeStepIngredient(ctx context.Context, recipeStepID, recipeStepIngredientID string) error {
	return m.Called(ctx, recipeStepID, recipeStepIngredientID).Error(0)
}

// MealPlanExists is a mock function.
func (m *Repository) MealPlanExists(ctx context.Context, mealPlanID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, accountID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlan is a mock function.
func (m *Repository) GetMealPlan(ctx context.Context, mealPlanID, accountID string) (*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx, mealPlanID, accountID)
	return returnValues.Get(0).(*mealplanning.MealPlan), returnValues.Error(1)
}

// GetMealPlansForAccount is a mock function.
func (m *Repository) GetMealPlansForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlan], error) {
	returnValues := m.Called(ctx, accountID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlan]), returnValues.Error(1)
}

// CreateMealPlan is a mock function.
func (m *Repository) CreateMealPlan(ctx context.Context, input *mealplanning.MealPlanDatabaseCreationInput) (*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlan), returnValues.Error(1)
}

// UpdateMealPlan is a mock function.
func (m *Repository) UpdateMealPlan(ctx context.Context, updated *mealplanning.MealPlan) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMealPlan is a mock function.
func (m *Repository) ArchiveMealPlan(ctx context.Context, mealPlanID, accountID string) error {
	return m.Called(ctx, mealPlanID, accountID).Error(0)
}

// AttemptToFinalizeMealPlan is a mock function.
func (m *Repository) AttemptToFinalizeMealPlan(ctx context.Context, mealPlanID, accountID string) (changed bool, err error) {
	returnValues := m.Called(ctx, mealPlanID, accountID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// MarkMealPlanAsGroceryListInitialized is a mock function.
func (m *Repository) MarkMealPlanAsGroceryListInitialized(ctx context.Context, mealPlanID string) error {
	return m.Called(ctx, mealPlanID).Error(0)
}

// GetUnfinalizedMealPlansWithExpiredVotingPeriods is a mock function.
func (m *Repository) GetUnfinalizedMealPlansWithExpiredVotingPeriods(ctx context.Context) ([]*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]*mealplanning.MealPlan), returnValues.Error(1)
}

// GetFinalizedMealPlanIDsForTheNextWeek is a mock function.
func (m *Repository) GetFinalizedMealPlanIDsForTheNextWeek(ctx context.Context) ([]*mealplanning.FinalizedMealPlanDatabaseResult, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]*mealplanning.FinalizedMealPlanDatabaseResult), returnValues.Error(1)
}

// GetFinalizedMealPlansWithUninitializedGroceryLists is a mock function.
func (m *Repository) GetFinalizedMealPlansWithUninitializedGroceryLists(ctx context.Context) ([]*mealplanning.MealPlan, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]*mealplanning.MealPlan), returnValues.Error(1)
}

// MealPlanOptionExists is a mock function.
func (m *Repository) MealPlanOptionExists(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// MealPlanOptionBelongsToAccount is a mock function.
func (m *Repository) MealPlanOptionBelongsToAccount(ctx context.Context, mealPlanOptionID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanOptionID, accountID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlanOption is a mock function.
func (m *Repository) GetMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) (*mealplanning.MealPlanOption, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	return returnValues.Get(0).(*mealplanning.MealPlanOption), returnValues.Error(1)
}

// GetMealPlanOptions is a mock function.
func (m *Repository) GetMealPlanOptions(ctx context.Context, mealPlanID, mealPlanEventID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOption], error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanOption]), returnValues.Error(1)
}

// MealExistsAsOptionInEvent is a mock function.
func (m *Repository) MealExistsAsOptionInEvent(ctx context.Context, mealPlanEventID, mealID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanEventID, mealID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// CreateMealPlanOption is a mock function.
func (m *Repository) CreateMealPlanOption(ctx context.Context, input *mealplanning.MealPlanOptionDatabaseCreationInput) (*mealplanning.MealPlanOption, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlanOption), returnValues.Error(1)
}

// UpdateMealPlanOption is a mock function.
func (m *Repository) UpdateMealPlanOption(ctx context.Context, updated *mealplanning.MealPlanOption) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMealPlanOption is a mock function.
func (m *Repository) ArchiveMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) error {
	return m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID).Error(0)
}

// FinalizeMealPlanOption is a mock function.
func (m *Repository) FinalizeMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, accountID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// MealPlanOptionVoteExists is a mock function.
func (m *Repository) MealPlanOptionVoteExists(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlanOptionVote is a mock function.
func (m *Repository) GetMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) (*mealplanning.MealPlanOptionVote, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID)
	return returnValues.Get(0).(*mealplanning.MealPlanOptionVote), returnValues.Error(1)
}

// GetMealPlanOptionVotes is a mock function.
func (m *Repository) GetMealPlanOptionVotes(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanOptionVote], error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanOptionVote]), returnValues.Error(1)
}

// CreateMealPlanOptionVote is a mock function.
func (m *Repository) CreateMealPlanOptionVote(ctx context.Context, input *mealplanning.MealPlanOptionVotesDatabaseCreationInput) ([]*mealplanning.MealPlanOptionVote, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).([]*mealplanning.MealPlanOptionVote), returnValues.Error(1)
}

// UpdateMealPlanOptionVote is a mock function.
func (m *Repository) UpdateMealPlanOptionVote(ctx context.Context, updated *mealplanning.MealPlanOptionVote) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMealPlanOptionVote is a mock function.
func (m *Repository) ArchiveMealPlanOptionVote(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID string) error {
	return m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID, mealPlanOptionVoteID).Error(0)
}

// GetMealPlanOptionVotesForMealPlanOption is a mock function.
func (m *Repository) GetMealPlanOptionVotesForMealPlanOption(ctx context.Context, mealPlanID, mealPlanEventID, mealPlanOptionID string) ([]*mealplanning.MealPlanOptionVote, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID, mealPlanOptionID)
	return returnValues.Get(0).([]*mealplanning.MealPlanOptionVote), returnValues.Error(1)
}

// ValidMeasurementUnitsForIngredientID is a mock function.
func (m *Repository) ValidMeasurementUnitsForIngredientID(ctx context.Context, validIngredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, validIngredientID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

// ValidMeasurementUnitExists is a mock function.
func (m *Repository) ValidMeasurementUnitExists(ctx context.Context, validMeasurementUnitID string) (bool, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidMeasurementUnit is a mock function.
func (m *Repository) GetValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) (*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, validMeasurementUnitID)
	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

// SearchForValidMeasurementUnits is a mock function.
func (m *Repository) SearchForValidMeasurementUnits(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

// GetValidMeasurementUnits is a mock function.
func (m *Repository) GetValidMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnit]), returnValues.Error(1)
}

// CreateValidMeasurementUnit is a mock function.
func (m *Repository) CreateValidMeasurementUnit(ctx context.Context, input *mealplanning.ValidMeasurementUnitDatabaseCreationInput) (*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

// UpdateValidMeasurementUnit is a mock function.
func (m *Repository) UpdateValidMeasurementUnit(ctx context.Context, updated *mealplanning.ValidMeasurementUnit) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidMeasurementUnit is a mock function.
func (m *Repository) ArchiveValidMeasurementUnit(ctx context.Context, validMeasurementUnitID string) error {
	return m.Called(ctx, validMeasurementUnitID).Error(0)
}

// MarkValidMeasurementUnitAsIndexed is a mock function.
func (m *Repository) MarkValidMeasurementUnitAsIndexed(ctx context.Context, validMeasurementUnitID string) error {
	return m.Called(ctx, validMeasurementUnitID).Error(0)
}

// GetValidMeasurementUnitIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidMeasurementUnitIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetValidMeasurementUnitsWithIDs is a mock function.
func (m *Repository) GetValidMeasurementUnitsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidMeasurementUnit, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidMeasurementUnit), returnValues.Error(1)
}

// ValidPreparationInstrumentExists is a mock function.
func (m *Repository) ValidPreparationInstrumentExists(ctx context.Context, validPreparationInstrumentID string) (bool, error) {
	returnValues := m.Called(ctx, validPreparationInstrumentID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidPreparationInstrument is a mock function.
func (m *Repository) GetValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) (*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, validPreparationInstrumentID)
	return returnValues.Get(0).(*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

// GetValidPreparationInstruments is a mock function.
func (m *Repository) GetValidPreparationInstruments(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

// GetValidPreparationInstrumentsForPreparation is a mock function.
func (m *Repository) GetValidPreparationInstrumentsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, preparationID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

// GetValidPreparationInstrumentsForInstrument is a mock function.
func (m *Repository) GetValidPreparationInstrumentsForInstrument(ctx context.Context, instrumentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument], error) {
	returnValues := m.Called(ctx, instrumentID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationInstrument]), returnValues.Error(1)
}

// CreateValidPreparationInstrument is a mock function.
func (m *Repository) CreateValidPreparationInstrument(ctx context.Context, input *mealplanning.ValidPreparationInstrumentDatabaseCreationInput) (*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

// UpdateValidPreparationInstrument is a mock function.
func (m *Repository) UpdateValidPreparationInstrument(ctx context.Context, updated *mealplanning.ValidPreparationInstrument) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidPreparationInstrument is a mock function.
func (m *Repository) ArchiveValidPreparationInstrument(ctx context.Context, validPreparationInstrumentID string) error {
	return m.Called(ctx, validPreparationInstrumentID).Error(0)
}

// GetValidPreparationInstrumentsByIDs is a mock function.
func (m *Repository) GetValidPreparationInstrumentsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidPreparationInstrument, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).(map[string]*mealplanning.ValidPreparationInstrument), returnValues.Error(1)
}

// ValidIngredientMeasurementUnitExists is a mock function.
func (m *Repository) ValidIngredientMeasurementUnitExists(ctx context.Context, validIngredientMeasurementUnitID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientMeasurementUnitID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredientMeasurementUnit is a mock function.
func (m *Repository) GetValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, validIngredientMeasurementUnitID)
	return returnValues.Get(0).(*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

// GetValidIngredientMeasurementUnits is a mock function.
func (m *Repository) GetValidIngredientMeasurementUnits(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

// GetValidIngredientMeasurementUnitsForIngredient is a mock function.
func (m *Repository) GetValidIngredientMeasurementUnitsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, ingredientID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

// GetValidIngredientMeasurementUnitsForMeasurementUnit is a mock function.
func (m *Repository) GetValidIngredientMeasurementUnitsForMeasurementUnit(ctx context.Context, measurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit], error) {
	returnValues := m.Called(ctx, measurementUnitID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientMeasurementUnit]), returnValues.Error(1)
}

// CreateValidIngredientMeasurementUnit is a mock function.
func (m *Repository) CreateValidIngredientMeasurementUnit(ctx context.Context, input *mealplanning.ValidIngredientMeasurementUnitDatabaseCreationInput) (*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

// UpdateValidIngredientMeasurementUnit is a mock function.
func (m *Repository) UpdateValidIngredientMeasurementUnit(ctx context.Context, updated *mealplanning.ValidIngredientMeasurementUnit) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredientMeasurementUnit is a mock function.
func (m *Repository) ArchiveValidIngredientMeasurementUnit(ctx context.Context, validIngredientMeasurementUnitID string) error {
	return m.Called(ctx, validIngredientMeasurementUnitID).Error(0)
}

// GetValidIngredientMeasurementUnitsByIDs is a mock function.
func (m *Repository) GetValidIngredientMeasurementUnitsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidIngredientMeasurementUnit, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).(map[string]*mealplanning.ValidIngredientMeasurementUnit), returnValues.Error(1)
}

// MealPlanEventIsEligibleForVoting is a mock function.
func (m *Repository) MealPlanEventIsEligibleForVoting(ctx context.Context, mealPlanID, mealPlanEventID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// MealPlanEventExists is a mock function.
func (m *Repository) MealPlanEventExists(ctx context.Context, mealPlanID, mealPlanEventID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlanEvent is a mock function.
func (m *Repository) GetMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) (*mealplanning.MealPlanEvent, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanEventID)
	return returnValues.Get(0).(*mealplanning.MealPlanEvent), returnValues.Error(1)
}

// GetMealPlanEvents is a mock function.
func (m *Repository) GetMealPlanEvents(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanEvent], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanEvent]), returnValues.Error(1)
}

// CreateMealPlanEvent is a mock function.
func (m *Repository) CreateMealPlanEvent(ctx context.Context, input *mealplanning.MealPlanEventDatabaseCreationInput) (*mealplanning.MealPlanEvent, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlanEvent), returnValues.Error(1)
}

// UpdateMealPlanEvent is a mock function.
func (m *Repository) UpdateMealPlanEvent(ctx context.Context, updated *mealplanning.MealPlanEvent) error {
	return m.Called(ctx, updated).Error(0)
}

// SwapMealPlanEvents is a mock function.
func (m *Repository) SwapMealPlanEvents(ctx context.Context, mealPlanID, mealPlanEventIDA, mealPlanEventIDB string) error {
	return m.Called(ctx, mealPlanID, mealPlanEventIDA, mealPlanEventIDB).Error(0)
}

// ArchiveMealPlanEvent is a mock function.
func (m *Repository) ArchiveMealPlanEvent(ctx context.Context, mealPlanID, mealPlanEventID string) error {
	return m.Called(ctx, mealPlanID, mealPlanEventID).Error(0)
}

// AttemptToFinalizeMealPlanEvent is a mock function.
func (m *Repository) AttemptToFinalizeMealPlanEvent(ctx context.Context, mealPlanEventID string) (changed bool, err error) {
	returnValues := m.Called(ctx, mealPlanEventID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetUnfinalizedMealPlanEventsWithExpiredVotingPeriods is a mock function.
func (m *Repository) GetUnfinalizedMealPlanEventsWithExpiredVotingPeriods(ctx context.Context) ([]*mealplanning.MealPlanEvent, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]*mealplanning.MealPlanEvent), returnValues.Error(1)
}

// GetFinalizedMealPlanEventIDsForTheNextWeek is a mock function.
func (m *Repository) GetFinalizedMealPlanEventIDsForTheNextWeek(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// MarkMealPlanAsHavingTasksCreated is a mock function.
func (m *Repository) MarkMealPlanAsHavingTasksCreated(ctx context.Context, mealPlanID string) error {
	return m.Called(ctx, mealPlanID).Error(0)
}

// MealPlanTaskExists is a mock function.
func (m *Repository) MealPlanTaskExists(ctx context.Context, mealPlanID, mealPlanTaskID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanTaskID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlanTask is a mock function.
func (m *Repository) GetMealPlanTask(ctx context.Context, mealPlanTaskID string) (*mealplanning.MealPlanTask, error) {
	returnValues := m.Called(ctx, mealPlanTaskID)
	return returnValues.Get(0).(*mealplanning.MealPlanTask), returnValues.Error(1)
}

// CreateMealPlanTask is a mock function.
func (m *Repository) CreateMealPlanTask(ctx context.Context, input *mealplanning.MealPlanTaskDatabaseCreationInput) (*mealplanning.MealPlanTask, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlanTask), returnValues.Error(1)
}

// GetMealPlanTasksForMealPlan is a mock function.
func (m *Repository) GetMealPlanTasksForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanTask], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanTask]), returnValues.Error(1)
}

// CreateMealPlanTasksForMealPlanOption is a mock function.
func (m *Repository) CreateMealPlanTasksForMealPlanOption(ctx context.Context, inputs []*mealplanning.MealPlanTaskDatabaseCreationInput) ([]*mealplanning.MealPlanTask, error) {
	returnValues := m.Called(ctx, inputs)
	return returnValues.Get(0).([]*mealplanning.MealPlanTask), returnValues.Error(1)
}

// ChangeMealPlanTaskStatus is a mock function.
func (m *Repository) ChangeMealPlanTaskStatus(ctx context.Context, input *mealplanning.MealPlanTaskStatusChangeRequestInput) error {
	return m.Called(ctx, input).Error(0)
}

// MealPlanTaskNotificationHasBeenSent is a mock function.
func (m *Repository) MealPlanTaskNotificationHasBeenSent(ctx context.Context, mealPlanTaskID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanTaskID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// MarkMealPlanTaskNotificationSent is a mock function.
func (m *Repository) MarkMealPlanTaskNotificationSent(ctx context.Context, mealPlanTaskID string) error {
	return m.Called(ctx, mealPlanTaskID).Error(0)
}

// ClearMealPlanTaskNotificationSentForEvent is a mock function.
func (m *Repository) ClearMealPlanTaskNotificationSentForEvent(ctx context.Context, mealPlanEventID string) error {
	return m.Called(ctx, mealPlanEventID).Error(0)
}

// GetMealPlanTaskIDsThatNeedNotification is a mock function.
func (m *Repository) GetMealPlanTaskIDsThatNeedNotification(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetMealPlanTaskAccountID is a mock function.
func (m *Repository) GetMealPlanTaskAccountID(ctx context.Context, mealPlanTaskID string) (string, error) {
	returnValues := m.Called(ctx, mealPlanTaskID)
	return returnValues.String(0), returnValues.Error(1)
}

// GetMealPlanTaskNotificationContext is a mock function.
func (m *Repository) GetMealPlanTaskNotificationContext(ctx context.Context, mealPlanTaskID string) (*mealplanning.MealPlanTaskNotificationContext, error) {
	returnValues := m.Called(ctx, mealPlanTaskID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*mealplanning.MealPlanTaskNotificationContext), returnValues.Error(1)
}

// RecipePrepTaskExists implements the requisite interface.
func (m *Repository) RecipePrepTaskExists(ctx context.Context, recipeID, recipePrepTaskID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID)

	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipePrepTask implements the requisite interface.
func (m *Repository) GetRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) (*mealplanning.RecipePrepTask, error) {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID)

	return returnValues.Get(0).(*mealplanning.RecipePrepTask), returnValues.Error(1)
}

// GetRecipePrepTasks implements the requisite interface.
func (m *Repository) GetRecipePrepTasks(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipePrepTask], error) {
	returnValues := m.Called(ctx, recipeID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipePrepTask]), returnValues.Error(1)
}

// GetRecipePrepTasksForRecipe implements the requisite interface.
func (m *Repository) GetRecipePrepTasksForRecipe(ctx context.Context, recipeID string) ([]*mealplanning.RecipePrepTask, error) {
	returnValues := m.Called(ctx, recipeID)

	return returnValues.Get(0).([]*mealplanning.RecipePrepTask), returnValues.Error(1)
}

// CreateRecipePrepTask implements the requisite interface.
func (m *Repository) CreateRecipePrepTask(ctx context.Context, input *mealplanning.RecipePrepTaskDatabaseCreationInput) (*mealplanning.RecipePrepTask, error) {
	returnValues := m.Called(ctx, input)

	return returnValues.Get(0).(*mealplanning.RecipePrepTask), returnValues.Error(1)
}

// UpdateRecipePrepTask implements the requisite interface.
func (m *Repository) UpdateRecipePrepTask(ctx context.Context, updated *mealplanning.RecipePrepTask) error {
	returnValues := m.Called(ctx, updated)

	return returnValues.Error(0)
}

// ArchiveRecipePrepTask implements the requisite interface.
func (m *Repository) ArchiveRecipePrepTask(ctx context.Context, recipeID, recipePrepTaskID string) error {
	returnValues := m.Called(ctx, recipeID, recipePrepTaskID)

	return returnValues.Error(0)
}

// MealPlanGroceryListItemExists is a mock function.
func (m *Repository) MealPlanGroceryListItemExists(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (bool, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanGroceryListItemID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealPlanGroceryListItem is a mock function.
func (m *Repository) GetMealPlanGroceryListItem(ctx context.Context, mealPlanID, mealPlanGroceryListItemID string) (*mealplanning.MealPlanGroceryListItem, error) {
	returnValues := m.Called(ctx, mealPlanID, mealPlanGroceryListItemID)
	return returnValues.Get(0).(*mealplanning.MealPlanGroceryListItem), returnValues.Error(1)
}

// GetMealPlanGroceryListItemsForMealPlan is a mock function.
func (m *Repository) GetMealPlanGroceryListItemsForMealPlan(ctx context.Context, mealPlanID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem], error) {
	returnValues := m.Called(ctx, mealPlanID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanGroceryListItem]), returnValues.Error(1)
}

// CreateMealPlanGroceryListItem is a mock function.
func (m *Repository) CreateMealPlanGroceryListItem(ctx context.Context, input *mealplanning.MealPlanGroceryListItemDatabaseCreationInput) (*mealplanning.MealPlanGroceryListItem, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlanGroceryListItem), returnValues.Error(1)
}

// UpdateMealPlanGroceryListItem is a mock function.
func (m *Repository) UpdateMealPlanGroceryListItem(ctx context.Context, updated *mealplanning.MealPlanGroceryListItem) error {
	returnValues := m.Called(ctx, updated)
	return returnValues.Error(0)
}

// ArchiveMealPlanGroceryListItem is a mock function.
func (m *Repository) ArchiveMealPlanGroceryListItem(ctx context.Context, mealPlanGroceryListItemID string) error {
	returnValues := m.Called(ctx, mealPlanGroceryListItemID)
	return returnValues.Error(0)
}

// GetValidMeasurementUnitConversionsForUnit is a mock function.
func (m *Repository) GetValidMeasurementUnitConversionsForUnit(ctx context.Context, validMeasurementUnitID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion], error) {
	returnValues := m.Called(ctx, validMeasurementUnitID, filter)

	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidMeasurementUnitConversion]), returnValues.Error(1)
}

// GetValidMeasurementUnitConversionsForIngredients is a mock function.
func (m *Repository) GetValidMeasurementUnitConversionsForIngredients(ctx context.Context, validIngredientIDs []string) ([]*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validIngredientIDs)

	return returnValues.Get(0).([]*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

// ValidMeasurementUnitConversionExists is a mock function.
func (m *Repository) ValidMeasurementUnitConversionExists(ctx context.Context, validPreparationID string) (bool, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidMeasurementUnitConversion is a mock function.
func (m *Repository) GetValidMeasurementUnitConversion(ctx context.Context, validPreparationID string) (*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

// CreateValidMeasurementUnitConversion is a mock function.
func (m *Repository) CreateValidMeasurementUnitConversion(ctx context.Context, input *mealplanning.ValidMeasurementUnitConversionDatabaseCreationInput) (*mealplanning.ValidMeasurementUnitConversion, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidMeasurementUnitConversion), returnValues.Error(1)
}

// UpdateValidMeasurementUnitConversion is a mock function.
func (m *Repository) UpdateValidMeasurementUnitConversion(ctx context.Context, updated *mealplanning.ValidMeasurementUnitConversion) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidMeasurementUnitConversion is a mock function.
func (m *Repository) ArchiveValidMeasurementUnitConversion(ctx context.Context, validPreparationID string) error {
	return m.Called(ctx, validPreparationID).Error(0)
}

// GetMeasurementUnitConversionMismatches is a mock function.
func (m *Repository) GetMeasurementUnitConversionMismatches(ctx context.Context) ([]*mealplanning.MeasurementUnitConversionMismatch, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]*mealplanning.MeasurementUnitConversionMismatch), returnValues.Error(1)
}

// RecipeMediaExists is a mock function.
func (m *Repository) RecipeMediaExists(ctx context.Context, validPreparationID string) (bool, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeMedia is a mock function.
func (m *Repository) GetRecipeMedia(ctx context.Context, validPreparationID string) (*mealplanning.RecipeMedia, error) {
	returnValues := m.Called(ctx, validPreparationID)
	return returnValues.Get(0).(*mealplanning.RecipeMedia), returnValues.Error(1)
}

// CreateRecipeMedia is a mock function.
func (m *Repository) CreateRecipeMedia(ctx context.Context, input *mealplanning.RecipeMediaDatabaseCreationInput) (*mealplanning.RecipeMedia, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeMedia), returnValues.Error(1)
}

// UpdateRecipeMedia is a mock function.
func (m *Repository) UpdateRecipeMedia(ctx context.Context, updated *mealplanning.RecipeMedia) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeMedia is a mock function.
func (m *Repository) ArchiveRecipeMedia(ctx context.Context, validPreparationID string) error {
	return m.Called(ctx, validPreparationID).Error(0)
}

// ValidIngredientStateExists is a mock function.
func (m *Repository) ValidIngredientStateExists(ctx context.Context, validIngredientStateID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientStateID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredientState is a mock function.
func (m *Repository) GetValidIngredientState(ctx context.Context, validIngredientStateID string) (*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, validIngredientStateID)
	return returnValues.Get(0).(*mealplanning.ValidIngredientState), returnValues.Error(1)
}

// SearchForValidIngredientStates is a mock function.
func (m *Repository) SearchForValidIngredientStates(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientState]), returnValues.Error(1)
}

// GetValidIngredientStates is a mock function.
func (m *Repository) GetValidIngredientStates(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientState], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientState]), returnValues.Error(1)
}

// CreateValidIngredientState is a mock function.
func (m *Repository) CreateValidIngredientState(ctx context.Context, input *mealplanning.ValidIngredientStateDatabaseCreationInput) (*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidIngredientState), returnValues.Error(1)
}

// UpdateValidIngredientState is a mock function.
func (m *Repository) UpdateValidIngredientState(ctx context.Context, updated *mealplanning.ValidIngredientState) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredientState is a mock function.
func (m *Repository) ArchiveValidIngredientState(ctx context.Context, validIngredientStateID string) error {
	return m.Called(ctx, validIngredientStateID).Error(0)
}

// MarkValidIngredientStateAsIndexed is a mock function.
func (m *Repository) MarkValidIngredientStateAsIndexed(ctx context.Context, validIngredientStateID string) error {
	return m.Called(ctx, validIngredientStateID).Error(0)
}

// GetValidIngredientStatesWithIDs is a mock function.
func (m *Repository) GetValidIngredientStatesWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidIngredientState, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidIngredientState), returnValues.Error(1)
}

// ValidIngredientStateIngredientExists is a mock function.
func (m *Repository) ValidIngredientStateIngredientExists(ctx context.Context, validIngredientStateIngredientID string) (bool, error) {
	returnValues := m.Called(ctx, validIngredientStateIngredientID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidIngredientStateIngredient is a mock function.
func (m *Repository) GetValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) (*mealplanning.ValidIngredientStateIngredient, error) {
	returnValues := m.Called(ctx, validIngredientStateIngredientID)
	return returnValues.Get(0).(*mealplanning.ValidIngredientStateIngredient), returnValues.Error(1)
}

// GetValidIngredientStateIngredients is a mock function.
func (m *Repository) GetValidIngredientStateIngredients(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

// GetValidIngredientStateIngredientsForIngredient is a mock function.
func (m *Repository) GetValidIngredientStateIngredientsForIngredient(ctx context.Context, ingredientID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, ingredientID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

// GetValidIngredientStateIngredientsForIngredientState is a mock function.
func (m *Repository) GetValidIngredientStateIngredientsForIngredientState(ctx context.Context, ingredientStateID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient], error) {
	returnValues := m.Called(ctx, ingredientStateID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidIngredientStateIngredient]), returnValues.Error(1)
}

// CreateValidIngredientStateIngredient is a mock function.
func (m *Repository) CreateValidIngredientStateIngredient(ctx context.Context, input *mealplanning.ValidIngredientStateIngredientDatabaseCreationInput) (*mealplanning.ValidIngredientStateIngredient, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidIngredientStateIngredient), returnValues.Error(1)
}

// UpdateValidIngredientStateIngredient is a mock function.
func (m *Repository) UpdateValidIngredientStateIngredient(ctx context.Context, updated *mealplanning.ValidIngredientStateIngredient) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidIngredientStateIngredient is a mock function.
func (m *Repository) ArchiveValidIngredientStateIngredient(ctx context.Context, validIngredientStateIngredientID string) error {
	return m.Called(ctx, validIngredientStateIngredientID).Error(0)
}

// GetValidIngredientStateIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidIngredientStateIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// RecipeStepCompletionConditionExists is a mock function.
func (m *Repository) RecipeStepCompletionConditionExists(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStepCompletionCondition is a mock function.
func (m *Repository) GetRecipeStepCompletionCondition(ctx context.Context, recipeID, recipeStepID, recipeStepIngredientID string) (*mealplanning.RecipeStepCompletionCondition, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepIngredientID)
	return returnValues.Get(0).(*mealplanning.RecipeStepCompletionCondition), returnValues.Error(1)
}

// GetRecipeStepCompletionConditions is a mock function.
func (m *Repository) GetRecipeStepCompletionConditions(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepCompletionCondition]), returnValues.Error(1)
}

// CreateRecipeStepCompletionCondition is a mock function.
func (m *Repository) CreateRecipeStepCompletionCondition(ctx context.Context, input *mealplanning.RecipeStepCompletionConditionDatabaseCreationInput) (*mealplanning.RecipeStepCompletionCondition, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStepCompletionCondition), returnValues.Error(1)
}

// UpdateRecipeStepCompletionCondition is a mock function.
func (m *Repository) UpdateRecipeStepCompletionCondition(ctx context.Context, updated *mealplanning.RecipeStepCompletionCondition) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStepCompletionCondition is a mock function.
func (m *Repository) ArchiveRecipeStepCompletionCondition(ctx context.Context, recipeStepID, recipeStepIngredientID string) error {
	return m.Called(ctx, recipeStepID, recipeStepIngredientID).Error(0)
}

// RecipeStepVesselExists is a mock function.
func (m *Repository) RecipeStepVesselExists(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepVesselID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeStepVessel is a mock function.
func (m *Repository) GetRecipeStepVessel(ctx context.Context, recipeID, recipeStepID, recipeStepVesselID string) (*mealplanning.RecipeStepVessel, error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, recipeStepVesselID)
	return returnValues.Get(0).(*mealplanning.RecipeStepVessel), returnValues.Error(1)
}

// GetRecipeStepVessels is a mock function.
func (m *Repository) GetRecipeStepVessels(ctx context.Context, recipeID, recipeStepID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeStepVessel], error) {
	returnValues := m.Called(ctx, recipeID, recipeStepID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeStepVessel]), returnValues.Error(1)
}

// CreateRecipeStepVessel is a mock function.
func (m *Repository) CreateRecipeStepVessel(ctx context.Context, input *mealplanning.RecipeStepVesselDatabaseCreationInput) (*mealplanning.RecipeStepVessel, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeStepVessel), returnValues.Error(1)
}

// UpdateRecipeStepVessel is a mock function.
func (m *Repository) UpdateRecipeStepVessel(ctx context.Context, updated *mealplanning.RecipeStepVessel) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeStepVessel is a mock function.
func (m *Repository) ArchiveRecipeStepVessel(ctx context.Context, recipeStepID, recipeStepVesselID string) error {
	return m.Called(ctx, recipeStepID, recipeStepVesselID).Error(0)
}

// UserIngredientPreferenceExists is a mock function.
func (m *Repository) UserIngredientPreferenceExists(ctx context.Context, userUserIngredientPreferenceID, userID string) (bool, error) {
	returnValues := m.Called(ctx, userUserIngredientPreferenceID, userID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetUserIngredientPreference is a mock function.
func (m *Repository) GetUserIngredientPreference(ctx context.Context, userUserIngredientPreferenceID, userID string) (*mealplanning.UserIngredientPreference, error) {
	returnValues := m.Called(ctx, userUserIngredientPreferenceID, userID)
	return returnValues.Get(0).(*mealplanning.UserIngredientPreference), returnValues.Error(1)
}

// GetUserIngredientPreferences is a mock function.
func (m *Repository) GetUserIngredientPreferences(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.UserIngredientPreference]), returnValues.Error(1)
}

// CreateUserIngredientPreference is a mock function.
func (m *Repository) CreateUserIngredientPreference(ctx context.Context, input *mealplanning.UserIngredientPreferenceDatabaseCreationInput) ([]*mealplanning.UserIngredientPreference, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).([]*mealplanning.UserIngredientPreference), returnValues.Error(1)
}

// UpdateUserIngredientPreference is a mock function.
func (m *Repository) UpdateUserIngredientPreference(ctx context.Context, updated *mealplanning.UserIngredientPreference) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveUserIngredientPreference is a mock function.
func (m *Repository) ArchiveUserIngredientPreference(ctx context.Context, userUserIngredientPreferenceID, userID string) error {
	return m.Called(ctx, userUserIngredientPreferenceID, userID).Error(0)
}

// AccountInstrumentOwnershipExists is a mock function.
func (m *Repository) AccountInstrumentOwnershipExists(ctx context.Context, accountInstrumentOwnershipID, accountID string) (bool, error) {
	returnValues := m.Called(ctx, accountInstrumentOwnershipID, accountID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetAccountInstrumentOwnership is a mock function.
func (m *Repository) GetAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) (*mealplanning.AccountInstrumentOwnership, error) {
	returnValues := m.Called(ctx, accountInstrumentOwnershipID, accountID)
	return returnValues.Get(0).(*mealplanning.AccountInstrumentOwnership), returnValues.Error(1)
}

// GetAccountInstrumentOwnerships is a mock function.
func (m *Repository) GetAccountInstrumentOwnerships(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership], error) {
	returnValues := m.Called(ctx, accountID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.AccountInstrumentOwnership]), returnValues.Error(1)
}

// CreateAccountInstrumentOwnership is a mock function.
func (m *Repository) CreateAccountInstrumentOwnership(ctx context.Context, input *mealplanning.AccountInstrumentOwnershipDatabaseCreationInput) (*mealplanning.AccountInstrumentOwnership, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.AccountInstrumentOwnership), returnValues.Error(1)
}

// UpdateAccountInstrumentOwnership is a mock function.
func (m *Repository) UpdateAccountInstrumentOwnership(ctx context.Context, updated *mealplanning.AccountInstrumentOwnership) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveAccountInstrumentOwnership is a mock function.
func (m *Repository) ArchiveAccountInstrumentOwnership(ctx context.Context, accountInstrumentOwnershipID, accountID string) error {
	return m.Called(ctx, accountInstrumentOwnershipID, accountID).Error(0)
}

// RecipeRatingExists is a mock function.
func (m *Repository) RecipeRatingExists(ctx context.Context, recipeID, recipeRatingID string) (bool, error) {
	returnValues := m.Called(ctx, recipeID, recipeRatingID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetRecipeRating is a mock function.
func (m *Repository) GetRecipeRating(ctx context.Context, recipeID, recipeRatingID string) (*mealplanning.RecipeRating, error) {
	returnValues := m.Called(ctx, recipeID, recipeRatingID)
	return returnValues.Get(0).(*mealplanning.RecipeRating), returnValues.Error(1)
}

// GetRecipeRatingsForRecipe is a mock function.
func (m *Repository) GetRecipeRatingsForRecipe(ctx context.Context, recipeID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
	returnValues := m.Called(ctx, recipeID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeRating]), returnValues.Error(1)
}

// GetRecipeRatingsForUser is a mock function.
func (m *Repository) GetRecipeRatingsForUser(ctx context.Context, user string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeRating], error) {
	returnValues := m.Called(ctx, user, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeRating]), returnValues.Error(1)
}

// CreateRecipeRating is a mock function.
func (m *Repository) CreateRecipeRating(ctx context.Context, input *mealplanning.RecipeRatingDatabaseCreationInput) (*mealplanning.RecipeRating, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeRating), returnValues.Error(1)
}

// UpdateRecipeRating is a mock function.
func (m *Repository) UpdateRecipeRating(ctx context.Context, updated *mealplanning.RecipeRating) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeRating is a mock function.
func (m *Repository) ArchiveRecipeRating(ctx context.Context, recipeID, recipeRatingID string) error {
	return m.Called(ctx, recipeID, recipeRatingID).Error(0)
}

// ValidVesselExists is a mock function.
func (m *Repository) ValidVesselExists(ctx context.Context, validVesselID string) (bool, error) {
	returnValues := m.Called(ctx, validVesselID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidVessel is a mock function.
func (m *Repository) GetValidVessel(ctx context.Context, validVesselID string) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, validVesselID)
	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

// GetRandomValidVessel is a mock function.
func (m *Repository) GetRandomValidVessel(ctx context.Context) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

// SearchForValidVessels is a mock function.
func (m *Repository) SearchForValidVessels(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
	returnValues := m.Called(ctx, query, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidVessel]), returnValues.Error(1)
}

// GetValidVessels is a mock function.
func (m *Repository) GetValidVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidVessel], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidVessel]), returnValues.Error(1)
}

// CreateValidVessel is a mock function.
func (m *Repository) CreateValidVessel(ctx context.Context, input *mealplanning.ValidVesselDatabaseCreationInput) (*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidVessel), returnValues.Error(1)
}

// UpdateValidVessel is a mock function.
func (m *Repository) UpdateValidVessel(ctx context.Context, updated *mealplanning.ValidVessel) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidVessel is a mock function.
func (m *Repository) ArchiveValidVessel(ctx context.Context, validVesselID string) error {
	return m.Called(ctx, validVesselID).Error(0)
}

// MarkValidVesselAsIndexed is a mock function.
func (m *Repository) MarkValidVesselAsIndexed(ctx context.Context, validVesselID string) error {
	return m.Called(ctx, validVesselID).Error(0)
}

// GetValidVesselIDsThatNeedSearchIndexing is a mock function.
func (m *Repository) GetValidVesselIDsThatNeedSearchIndexing(ctx context.Context) ([]string, error) {
	returnValues := m.Called(ctx)
	return returnValues.Get(0).([]string), returnValues.Error(1)
}

// GetValidVesselsWithIDs is a mock function.
func (m *Repository) GetValidVesselsWithIDs(ctx context.Context, ids []string) ([]*mealplanning.ValidVessel, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).([]*mealplanning.ValidVessel), returnValues.Error(1)
}

// ValidPreparationVesselExists is a mock function.
func (m *Repository) ValidPreparationVesselExists(ctx context.Context, validPreparationVesselID string) (bool, error) {
	returnValues := m.Called(ctx, validPreparationVesselID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetValidPreparationVessel is a mock function.
func (m *Repository) GetValidPreparationVessel(ctx context.Context, validPreparationVesselID string) (*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, validPreparationVesselID)
	return returnValues.Get(0).(*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

// GetValidPreparationVessels is a mock function.
func (m *Repository) GetValidPreparationVessels(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

// GetValidPreparationVesselsForPreparation is a mock function.
func (m *Repository) GetValidPreparationVesselsForPreparation(ctx context.Context, preparationID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, preparationID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

// GetValidPreparationVesselsForVessel is a mock function.
func (m *Repository) GetValidPreparationVesselsForVessel(ctx context.Context, vesselID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel], error) {
	returnValues := m.Called(ctx, vesselID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.ValidPreparationVessel]), returnValues.Error(1)
}

// CreateValidPreparationVessel is a mock function.
func (m *Repository) CreateValidPreparationVessel(ctx context.Context, input *mealplanning.ValidPreparationVesselDatabaseCreationInput) (*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

// UpdateValidPreparationVessel is a mock function.
func (m *Repository) UpdateValidPreparationVessel(ctx context.Context, updated *mealplanning.ValidPreparationVessel) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveValidPreparationVessel is a mock function.
func (m *Repository) ArchiveValidPreparationVessel(ctx context.Context, validPreparationVesselID string) error {
	return m.Called(ctx, validPreparationVesselID).Error(0)
}

// GetValidPreparationVesselsByIDs is a mock function.
func (m *Repository) GetValidPreparationVesselsByIDs(ctx context.Context, ids []string) (map[string]*mealplanning.ValidPreparationVessel, error) {
	returnValues := m.Called(ctx, ids)
	return returnValues.Get(0).(map[string]*mealplanning.ValidPreparationVessel), returnValues.Error(1)
}

// GetRecipeLists is a mock function.
func (m *Repository) GetRecipeLists(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeList], error) {
	returnValues := m.Called(ctx, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeList]), returnValues.Error(1)
}

// CreateRecipeList is a mock function.
func (m *Repository) CreateRecipeList(ctx context.Context, input *mealplanning.RecipeListDatabaseCreationInput) (*mealplanning.RecipeList, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeList), returnValues.Error(1)
}

// UpdateRecipeList is a mock function.
func (m *Repository) UpdateRecipeList(ctx context.Context, updated *mealplanning.RecipeList) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeList is a mock function.
func (m *Repository) ArchiveRecipeList(ctx context.Context, recipeListID, userID string) error {
	return m.Called(ctx, recipeListID, userID).Error(0)
}

// GetMealLists is a mock function.
func (m *Repository) GetMealLists(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealList], error) {
	returnValues := m.Called(ctx, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealList]), returnValues.Error(1)
}

// CreateMealList is a mock function.
func (m *Repository) CreateMealList(ctx context.Context, input *mealplanning.MealListDatabaseCreationInput) (*mealplanning.MealList, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealList), returnValues.Error(1)
}

// UpdateMealList is a mock function.
func (m *Repository) UpdateMealList(ctx context.Context, updated *mealplanning.MealList) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMealList is a mock function.
func (m *Repository) ArchiveMealList(ctx context.Context, mealListID, userID string) error {
	return m.Called(ctx, mealListID, userID).Error(0)
}

// GetRecipeListItems is a mock function.
func (m *Repository) GetRecipeListItems(ctx context.Context, recipeListID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.RecipeListItem], error) {
	returnValues := m.Called(ctx, recipeListID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.RecipeListItem]), returnValues.Error(1)
}

// CreateRecipeListItem is a mock function.
func (m *Repository) CreateRecipeListItem(ctx context.Context, input *mealplanning.RecipeListItemDatabaseCreationInput) (*mealplanning.RecipeListItem, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.RecipeListItem), returnValues.Error(1)
}

// UpdateRecipeListItem is a mock function.
func (m *Repository) UpdateRecipeListItem(ctx context.Context, updated *mealplanning.RecipeListItem) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveRecipeListItem is a mock function.
func (m *Repository) ArchiveRecipeListItem(ctx context.Context, recipeListItemID, recipeListID string) error {
	return m.Called(ctx, recipeListItemID, recipeListID).Error(0)
}

// MealExistsInMealList is a mock function.
func (m *Repository) MealExistsInMealList(ctx context.Context, mealListID, mealID string) (bool, error) {
	returnValues := m.Called(ctx, mealListID, mealID)
	return returnValues.Bool(0), returnValues.Error(1)
}

// GetMealListItems is a mock function.
func (m *Repository) GetMealListItems(ctx context.Context, mealListID, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealListItem], error) {
	returnValues := m.Called(ctx, mealListID, userID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealListItem]), returnValues.Error(1)
}

// CreateMealListItem is a mock function.
func (m *Repository) CreateMealListItem(ctx context.Context, input *mealplanning.MealListItemDatabaseCreationInput) (*mealplanning.MealListItem, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealListItem), returnValues.Error(1)
}

// UpdateMealListItem is a mock function.
func (m *Repository) UpdateMealListItem(ctx context.Context, updated *mealplanning.MealListItem) error {
	return m.Called(ctx, updated).Error(0)
}

// ArchiveMealListItem is a mock function.
func (m *Repository) ArchiveMealListItem(ctx context.Context, mealListItemID, mealListID string) error {
	return m.Called(ctx, mealListItemID, mealListID).Error(0)
}

// GetSelection is a mock function.
func (m *Repository) GetMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) (*mealplanning.MealPlanRecipeOptionSelection, error) {
	returnValues := m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).(*mealplanning.MealPlanRecipeOptionSelection), returnValues.Error(1)
}

// GetSelectionsForMealPlanOption is a mock function.
func (m *Repository) GetSelectionsForMealPlanOption(ctx context.Context, mealPlanOptionID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[mealplanning.MealPlanRecipeOptionSelection], error) {
	returnValues := m.Called(ctx, mealPlanOptionID, filter)
	return returnValues.Get(0).(*filtering.QueryFilteredResult[mealplanning.MealPlanRecipeOptionSelection]), returnValues.Error(1)
}

// CreateSelection is a mock function.
func (m *Repository) CreateMealPlanRecipeOptionSelection(ctx context.Context, input *mealplanning.MealPlanRecipeOptionSelectionDatabaseCreationInput) (*mealplanning.MealPlanRecipeOptionSelection, error) {
	returnValues := m.Called(ctx, input)
	return returnValues.Get(0).(*mealplanning.MealPlanRecipeOptionSelection), returnValues.Error(1)
}

// UpdateSelection is a mock function.
func (m *Repository) UpdateMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string, input *mealplanning.MealPlanRecipeOptionSelectionUpdateRequestInput) error {
	return m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType, input).Error(0)
}

// ArchiveSelection is a mock function.
func (m *Repository) ArchiveMealPlanRecipeOptionSelection(ctx context.Context, mealPlanOptionID, recipeStepID string, ingredientIndex uint16, selectionType string) error {
	return m.Called(ctx, mealPlanOptionID, recipeStepID, ingredientIndex, selectionType).Error(0)
}

// AddPreparationMedia is a mock function.
func (m *Repository) AddPreparationMedia(ctx context.Context, validPreparationID string, forIngredientID *string, uploadedMediaID string, index int32) error {
	return m.Called(ctx, validPreparationID, forIngredientID, uploadedMediaID, index).Error(0)
}

// GetPreparationMediaByPreparation is a mock function.
func (m *Repository) GetPreparationMediaByPreparation(ctx context.Context, validPreparationID string) ([]*mealplanning.PreparationMediaRow, error) {
	returnValues := m.Called(ctx, validPreparationID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*mealplanning.PreparationMediaRow), returnValues.Error(1)
}

// GetPreparationMediaByPreparationAndIngredient is a mock function.
func (m *Repository) GetPreparationMediaByPreparationAndIngredient(ctx context.Context, validPreparationID string, forIngredientID *string) ([]*mealplanning.PreparationMediaRow, error) {
	returnValues := m.Called(ctx, validPreparationID, forIngredientID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*mealplanning.PreparationMediaRow), returnValues.Error(1)
}

// AddIngredientMedia is a mock function.
func (m *Repository) AddIngredientMedia(ctx context.Context, validIngredientID, uploadedMediaID string, index int32) error {
	return m.Called(ctx, validIngredientID, uploadedMediaID, index).Error(0)
}

// GetIngredientMediaByIngredient is a mock function.
func (m *Repository) GetIngredientMediaByIngredient(ctx context.Context, validIngredientID string) ([]*mealplanning.IngredientMediaRow, error) {
	returnValues := m.Called(ctx, validIngredientID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*mealplanning.IngredientMediaRow), returnValues.Error(1)
}

// GetUploadedMediaWithIDs is a mock function.
func (m *Repository) GetUploadedMediaWithIDs(ctx context.Context, ids []string) ([]*uploadedmedia.UploadedMedia, error) {
	returnValues := m.Called(ctx, ids)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*uploadedmedia.UploadedMedia), returnValues.Error(1)
}

// AddRecipeStepImage is a mock function.
func (m *Repository) AddRecipeStepImage(ctx context.Context, recipeStepID, uploadedMediaID, uploadedByUser string) error {
	return m.Called(ctx, recipeStepID, uploadedMediaID, uploadedByUser).Error(0)
}

// GetRecipeStepImagesByStep is a mock function.
func (m *Repository) GetRecipeStepImagesByStep(ctx context.Context, recipeStepID string) ([]*mealplanning.RecipeStepImageRow, error) {
	returnValues := m.Called(ctx, recipeStepID)
	if returnValues.Get(0) == nil {
		return nil, returnValues.Error(1)
	}
	return returnValues.Get(0).([]*mealplanning.RecipeStepImageRow), returnValues.Error(1)
}
