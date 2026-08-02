package grpc

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/authorization"
	mealplanningsvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
)

// MealPlanningMethodPermissions is a named type for Wire dependency injection.
type MealPlanningMethodPermissions map[string][]authorization.Permission

// ProvideMethodPermissions returns a Wire provider for the meal planning service's method permissions.
func ProvideMethodPermissions() MealPlanningMethodPermissions {
	return MealPlanningMethodPermissions{
		// ValidIngredients
		mealplanningsvc.MealPlanningService_CreateValidIngredient_FullMethodName:     {authorization.CreateValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredient_FullMethodName:        {authorization.ReadValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredients_FullMethodName:       {authorization.ReadValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidIngredients_FullMethodName: {authorization.ReadValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredient_FullMethodName:     {authorization.UpdateValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_UploadIngredientMedia_FullMethodName:     {authorization.UpdateValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredient_FullMethodName:    {authorization.ArchiveValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetRandomValidIngredient_FullMethodName:  {authorization.ReadValidIngredientsPermission},

		// ValidIngredientGroups
		mealplanningsvc.MealPlanningService_CreateValidIngredientGroup_FullMethodName:     {authorization.CreateValidIngredientGroupsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientGroup_FullMethodName:        {authorization.ReadValidIngredientGroupsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientGroups_FullMethodName:       {authorization.ReadValidIngredientGroupsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidIngredientGroups_FullMethodName: {authorization.ReadValidIngredientGroupsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredientGroup_FullMethodName:     {authorization.UpdateValidIngredientGroupsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredientGroup_FullMethodName:    {authorization.ArchiveValidIngredientGroupsPermission},

		// ValidIngredientStates
		mealplanningsvc.MealPlanningService_CreateValidIngredientState_FullMethodName:     {authorization.CreateValidIngredientStatesPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientState_FullMethodName:        {authorization.ReadValidIngredientStatesPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientStates_FullMethodName:       {authorization.ReadValidIngredientStatesPermission},
		mealplanningsvc.MealPlanningService_SearchForValidIngredientStates_FullMethodName: {authorization.ReadValidIngredientStatesPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredientState_FullMethodName:     {authorization.UpdateValidIngredientStatesPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredientState_FullMethodName:    {authorization.ArchiveValidIngredientStatesPermission},

		// ValidIngredientStateIngredients
		mealplanningsvc.MealPlanningService_CreateValidIngredientStateIngredient_FullMethodName:                {authorization.CreateValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientStateIngredient_FullMethodName:                   {authorization.ReadValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientStateIngredients_FullMethodName:                  {authorization.ReadValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredientStateIngredient_FullMethodName:                {authorization.UpdateValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredientStateIngredient_FullMethodName:               {authorization.ArchiveValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientStateIngredientsByIngredient_FullMethodName:      {authorization.ReadValidIngredientStateIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientStateIngredientsByIngredientState_FullMethodName: {authorization.ReadValidIngredientStateIngredientsPermission},

		// ValidPreparations
		mealplanningsvc.MealPlanningService_CreateValidPreparation_FullMethodName:     {authorization.CreateValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparation_FullMethodName:        {authorization.ReadValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparations_FullMethodName:       {authorization.ReadValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidPreparations_FullMethodName: {authorization.ReadValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidPreparation_FullMethodName:     {authorization.UpdateValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_UploadPreparationMedia_FullMethodName:     {authorization.UpdateValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidPreparation_FullMethodName:    {authorization.ArchiveValidPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetRandomValidPreparation_FullMethodName:  {authorization.ReadValidPreparationsPermission},

		// ValidMeasurementUnits
		mealplanningsvc.MealPlanningService_CreateValidMeasurementUnit_FullMethodName:     {authorization.CreateValidMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_GetValidMeasurementUnit_FullMethodName:        {authorization.ReadValidMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_GetValidMeasurementUnits_FullMethodName:       {authorization.ReadValidMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidMeasurementUnits_FullMethodName: {authorization.ReadValidMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidMeasurementUnit_FullMethodName:     {authorization.UpdateValidMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidMeasurementUnit_FullMethodName:    {authorization.ArchiveValidMeasurementUnitsPermission},

		// ValidVessels
		mealplanningsvc.MealPlanningService_CreateValidVessel_FullMethodName:     {authorization.CreateValidVesselsPermission},
		mealplanningsvc.MealPlanningService_GetValidVessel_FullMethodName:        {authorization.ReadValidVesselsPermission},
		mealplanningsvc.MealPlanningService_GetValidVessels_FullMethodName:       {authorization.ReadValidVesselsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidVessels_FullMethodName: {authorization.ReadValidVesselsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidVessel_FullMethodName:     {authorization.UpdateValidVesselsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidVessel_FullMethodName:    {authorization.ArchiveValidVesselsPermission},
		mealplanningsvc.MealPlanningService_GetRandomValidVessel_FullMethodName:  {authorization.ReadValidVesselsPermission},

		// ValidInstruments
		mealplanningsvc.MealPlanningService_CreateValidInstrument_FullMethodName:     {authorization.CreateValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetValidInstrument_FullMethodName:        {authorization.ReadValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetValidInstruments_FullMethodName:       {authorization.ReadValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidInstruments_FullMethodName: {authorization.ReadValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidInstrument_FullMethodName:     {authorization.UpdateValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidInstrument_FullMethodName:    {authorization.ArchiveValidInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetRandomValidInstrument_FullMethodName:  {authorization.ReadValidInstrumentsPermission},

		// ValidPreparationVessels
		mealplanningsvc.MealPlanningService_GetValidPreparationVessel_FullMethodName:               {authorization.ReadValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_CreateValidPreparationVessel_FullMethodName:            {authorization.CreateValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationVessels_FullMethodName:              {authorization.ReadValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationVesselsByVessel_FullMethodName:      {authorization.ReadValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationVesselsByPreparation_FullMethodName: {authorization.ReadValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidPreparationVessel_FullMethodName:            {authorization.UpdateValidPreparationVesselsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidPreparationVessel_FullMethodName:           {authorization.ArchiveValidPreparationVesselsPermission},

		// ValidIngredientPreparations
		mealplanningsvc.MealPlanningService_GetValidIngredientPreparation_FullMethodName:               {authorization.ReadValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_CreateValidIngredientPreparation_FullMethodName:            {authorization.CreateValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientPreparations_FullMethodName:              {authorization.ReadValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientPreparationsByPreparation_FullMethodName: {authorization.ReadValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientPreparationsByIngredient_FullMethodName:  {authorization.ReadValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredientPreparation_FullMethodName:            {authorization.UpdateValidIngredientPreparationsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredientPreparation_FullMethodName:           {authorization.ArchiveValidIngredientPreparationsPermission},

		// ValidPrepTaskConfigs
		mealplanningsvc.MealPlanningService_GetValidPrepTaskConfig_FullMethodName:                            {authorization.ReadValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_CreateValidPrepTaskConfig_FullMethodName:                         {authorization.CreateValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_GetValidPrepTaskConfigs_FullMethodName:                           {authorization.ReadValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_GetValidPrepTaskConfigsByPreparation_FullMethodName:              {authorization.ReadValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_GetValidPrepTaskConfigsByIngredient_FullMethodName:               {authorization.ReadValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_GetValidPrepTaskConfigsByIngredientAndPreparation_FullMethodName: {authorization.ReadValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidPrepTaskConfig_FullMethodName:                         {authorization.UpdateValidPrepTaskConfigsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidPrepTaskConfig_FullMethodName:                        {authorization.ArchiveValidPrepTaskConfigsPermission},

		// ValidIngredientMeasurementUnits
		mealplanningsvc.MealPlanningService_GetValidIngredientMeasurementUnit_FullMethodName:                   {authorization.ReadValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_CreateValidIngredientMeasurementUnit_FullMethodName:                {authorization.CreateValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientMeasurementUnits_FullMethodName:                  {authorization.ReadValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientMeasurementUnitsByMeasurementUnit_FullMethodName: {authorization.ReadValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_GetValidIngredientMeasurementUnitsByIngredient_FullMethodName:      {authorization.ReadValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidIngredientMeasurementUnit_FullMethodName:                {authorization.UpdateValidIngredientMeasurementUnitsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidIngredientMeasurementUnit_FullMethodName:               {authorization.ArchiveValidIngredientMeasurementUnitsPermission},

		// ValidPreparationInstruments
		mealplanningsvc.MealPlanningService_GetValidPreparationInstrument_FullMethodName:               {authorization.ReadValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_CreateValidPreparationInstrument_FullMethodName:            {authorization.CreateValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationInstruments_FullMethodName:              {authorization.ReadValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationInstrumentsByInstrument_FullMethodName:  {authorization.ReadValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetValidPreparationInstrumentsByPreparation_FullMethodName: {authorization.ReadValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidPreparationInstrument_FullMethodName:            {authorization.UpdateValidPreparationInstrumentsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidPreparationInstrument_FullMethodName:           {authorization.ArchiveValidPreparationInstrumentsPermission},

		// ValidMeasurementUnitConversions
		mealplanningsvc.MealPlanningService_CreateValidMeasurementUnitConversion_FullMethodName:             {authorization.CreateValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_GetValidMeasurementUnitConversion_FullMethodName:                {authorization.ReadValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_GetValidMeasurementUnitConversionsForUnit_FullMethodName:        {authorization.ReadValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_GetValidMeasurementUnitConversionsForIngredients_FullMethodName: {authorization.ReadValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_GetMeasurementUnitConversionMismatches_FullMethodName:           {authorization.ReadValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_UpdateValidMeasurementUnitConversion_FullMethodName:             {authorization.UpdateValidMeasurementUnitConversionsPermission},
		mealplanningsvc.MealPlanningService_ArchiveValidMeasurementUnitConversion_FullMethodName:            {authorization.ArchiveValidMeasurementUnitConversionsPermission},

		// UserIngredientPreferences
		mealplanningsvc.MealPlanningService_CreateUserIngredientPreference_FullMethodName:  {authorization.CreateUserIngredientPreferencesPermission},
		mealplanningsvc.MealPlanningService_GetUserIngredientPreference_FullMethodName:     {authorization.ReadUserIngredientPreferencesPermission},
		mealplanningsvc.MealPlanningService_GetUserIngredientPreferences_FullMethodName:    {authorization.ReadUserIngredientPreferencesPermission},
		mealplanningsvc.MealPlanningService_UpdateUserIngredientPreference_FullMethodName:  {authorization.UpdateUserIngredientPreferencesPermission},
		mealplanningsvc.MealPlanningService_ArchiveUserIngredientPreference_FullMethodName: {authorization.ArchiveUserIngredientPreferencesPermission},

		// AccountInstrumentOwnerships
		mealplanningsvc.MealPlanningService_CreateAccountInstrumentOwnership_FullMethodName:           {authorization.CreateAccountInstrumentOwnershipsPermission},
		mealplanningsvc.MealPlanningService_GetAccountInstrumentOwnership_FullMethodName:              {authorization.ReadAccountInstrumentOwnershipsPermission},
		mealplanningsvc.MealPlanningService_GetAccountInstrumentOwnerships_FullMethodName:             {authorization.ReadAccountInstrumentOwnershipsPermission},
		mealplanningsvc.MealPlanningService_SearchForValidInstrumentsNotOwnedByAccount_FullMethodName: {authorization.ReadAccountInstrumentOwnershipsPermission},
		mealplanningsvc.MealPlanningService_UpdateAccountInstrumentOwnership_FullMethodName:           {authorization.UpdateAccountInstrumentOwnershipsPermission},
		mealplanningsvc.MealPlanningService_ArchiveAccountInstrumentOwnership_FullMethodName:          {authorization.ArchiveAccountInstrumentOwnershipsPermission},

		// Recipes
		mealplanningsvc.MealPlanningService_CreateRecipe_FullMethodName:                            {authorization.CreateRecipesPermission},
		mealplanningsvc.MealPlanningService_GetRecipe_FullMethodName:                               {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_GetRecipes_FullMethodName:                              {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipe_FullMethodName:                            {authorization.UpdateRecipesPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStatus_FullMethodName:                      {authorization.UpdateRecipesStatusPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipe_FullMethodName:                           {authorization.ArchiveRecipesPermission},
		mealplanningsvc.MealPlanningService_SearchForRecipes_FullMethodName:                        {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_CloneRecipe_FullMethodName:                             {authorization.CloneRecipesPermission},
		mealplanningsvc.MealPlanningService_UploadRecipeImage_FullMethodName:                       {authorization.UpdateRecipesPermission},
		mealplanningsvc.MealPlanningService_SearchForMealEligibleRecipes_FullMethodName:            {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_SearchForRecipesWithInstrumentOwnership_FullMethodName: {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_GetMermaidDiagramForMeal_FullMethodName:                {authorization.ReadMealsPermission},
		mealplanningsvc.MealPlanningService_GetMermaidDiagramForRecipe_FullMethodName:              {authorization.ReadRecipesPermission},
		mealplanningsvc.MealPlanningService_EstimateRecipePrepTasks_FullMethodName:                 {authorization.ReadRecipesPermission},

		// Meals
		mealplanningsvc.MealPlanningService_CreateMeal_FullMethodName:      {authorization.CreateMealsPermission},
		mealplanningsvc.MealPlanningService_GetMeal_FullMethodName:         {authorization.ReadMealsPermission},
		mealplanningsvc.MealPlanningService_GetMeals_FullMethodName:        {authorization.ReadMealsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMeal_FullMethodName:     {authorization.ArchiveMealsPermission},
		mealplanningsvc.MealPlanningService_SearchForMeals_FullMethodName:  {authorization.ReadMealsPermission},
		mealplanningsvc.MealPlanningService_UploadMealImage_FullMethodName: {authorization.UpdateMealsPermission},

		// MealPlans
		mealplanningsvc.MealPlanningService_CreateMealPlan_FullMethodName:         {authorization.CreateMealPlansPermission},
		mealplanningsvc.MealPlanningService_GetMealPlan_FullMethodName:            {authorization.ReadMealPlansPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlan_FullMethodName:         {authorization.UpdateMealPlansPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlan_FullMethodName:        {authorization.ArchiveMealPlansPermission},
		mealplanningsvc.MealPlanningService_GetMealPlansForAccount_FullMethodName: {authorization.ReadMealPlansPermission},
		mealplanningsvc.MealPlanningService_FinalizeMealPlan_FullMethodName:       {authorization.UpdateMealPlansPermission},

		// MealPlanOptions
		mealplanningsvc.MealPlanningService_CreateMealPlanOption_FullMethodName:  {authorization.CreateMealPlanOptionsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanOption_FullMethodName:     {authorization.ReadMealPlanOptionsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanOptions_FullMethodName:    {authorization.ReadMealPlanOptionsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanOption_FullMethodName:  {authorization.UpdateMealPlanOptionsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlanOption_FullMethodName: {authorization.ArchiveMealPlanOptionsPermission},

		// MealPlanEvents
		mealplanningsvc.MealPlanningService_CreateMealPlanEvent_FullMethodName:  {authorization.CreateMealPlanEventsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanEvent_FullMethodName:     {authorization.ReadMealPlanEventsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanEvents_FullMethodName:    {authorization.ReadMealPlanEventsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanEvent_FullMethodName:  {authorization.UpdateMealPlanEventsPermission},
		mealplanningsvc.MealPlanningService_SwapMealPlanEvents_FullMethodName:   {authorization.UpdateMealPlanEventsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlanEvent_FullMethodName: {authorization.ArchiveMealPlanEventsPermission},

		// MealPlanTasks
		mealplanningsvc.MealPlanningService_CreateMealPlanTask_FullMethodName:       {authorization.CreateMealPlanTasksPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanTask_FullMethodName:          {authorization.ReadMealPlanTasksPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanTasks_FullMethodName:         {authorization.ReadMealPlanTasksPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanTaskStatus_FullMethodName: {authorization.UpdateMealPlanTasksPermission},

		// MealPlanGroceryListItems
		mealplanningsvc.MealPlanningService_GetMealPlanGroceryListItem_FullMethodName:             {authorization.ReadMealPlanGroceryListItemsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanGroceryListItemsForMealPlan_FullMethodName: {authorization.ReadMealPlanGroceryListItemsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanGroceryListItem_FullMethodName:          {authorization.UpdateMealPlanGroceryListItemsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlanGroceryListItem_FullMethodName:         {authorization.ArchiveMealPlanGroceryListItemsPermission},

		// MealLists
		mealplanningsvc.MealPlanningService_ArchiveMealList_FullMethodName:     {authorization.ArchiveMealListsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealListItem_FullMethodName: {authorization.ArchiveMealListsPermission},
		mealplanningsvc.MealPlanningService_CreateMealList_FullMethodName:      {authorization.CreateMealListsPermission},
		mealplanningsvc.MealPlanningService_CreateMealListItem_FullMethodName:  {authorization.CreateMealListsPermission},
		mealplanningsvc.MealPlanningService_GetMealLists_FullMethodName:        {authorization.ReadMealListsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealList_FullMethodName:      {authorization.UpdateMealListsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealListItem_FullMethodName:  {authorization.UpdateMealListsPermission},

		// RecipeLists
		mealplanningsvc.MealPlanningService_ArchiveRecipeList_FullMethodName:     {authorization.ArchiveRecipeListsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeListItem_FullMethodName: {authorization.ArchiveRecipeListsPermission},
		mealplanningsvc.MealPlanningService_CreateRecipeList_FullMethodName:      {authorization.CreateRecipeListsPermission},
		mealplanningsvc.MealPlanningService_CreateRecipeListItem_FullMethodName:  {authorization.CreateRecipeListsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeLists_FullMethodName:        {authorization.ReadMealListsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeList_FullMethodName:      {authorization.UpdateRecipeListsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeListItem_FullMethodName:  {authorization.UpdateRecipeListsPermission},

		// RecipeSteps
		mealplanningsvc.MealPlanningService_CreateRecipeStep_FullMethodName:      {authorization.CreateRecipeStepsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeSteps_FullMethodName:        {authorization.ReadRecipeStepsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStep_FullMethodName:         {authorization.ReadRecipeStepsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStep_FullMethodName:      {authorization.UpdateRecipeStepsPermission},
		mealplanningsvc.MealPlanningService_UploadRecipeStepImage_FullMethodName: {authorization.UpdateRecipeStepsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStep_FullMethodName:     {authorization.ArchiveRecipeStepsPermission},

		// RecipeStepVessels
		mealplanningsvc.MealPlanningService_CreateRecipeStepVessel_FullMethodName:  {authorization.CreateRecipeStepVesselsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepVessels_FullMethodName:    {authorization.ReadRecipeStepVesselsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepVessel_FullMethodName:     {authorization.ReadRecipeStepVesselsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStepVessel_FullMethodName:  {authorization.UpdateRecipeStepVesselsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStepVessel_FullMethodName: {authorization.ArchiveRecipeStepVesselsPermission},

		// RecipeStepProducts
		mealplanningsvc.MealPlanningService_CreateRecipeStepProduct_FullMethodName:  {authorization.CreateRecipeStepProductsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepProducts_FullMethodName:    {authorization.ReadRecipeStepProductsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepProduct_FullMethodName:     {authorization.ReadRecipeStepProductsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStepProduct_FullMethodName:  {authorization.UpdateRecipeStepProductsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStepProduct_FullMethodName: {authorization.ArchiveRecipeStepProductsPermission},

		// RecipePrepTasks
		mealplanningsvc.MealPlanningService_CreateRecipePrepTask_FullMethodName:  {authorization.CreateRecipePrepTasksPermission},
		mealplanningsvc.MealPlanningService_GetRecipePrepTasks_FullMethodName:    {authorization.ReadRecipePrepTasksPermission},
		mealplanningsvc.MealPlanningService_GetRecipePrepTask_FullMethodName:     {authorization.ReadRecipePrepTasksPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipePrepTask_FullMethodName:  {authorization.UpdateRecipePrepTasksPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipePrepTask_FullMethodName: {authorization.ArchiveRecipePrepTasksPermission},

		// RecipeStepInstruments
		mealplanningsvc.MealPlanningService_CreateRecipeStepInstrument_FullMethodName:  {authorization.CreateRecipeStepInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepInstruments_FullMethodName:    {authorization.ReadRecipeStepInstrumentsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepInstrument_FullMethodName:     {authorization.ReadRecipeStepInstrumentsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStepInstrument_FullMethodName:  {authorization.UpdateRecipeStepInstrumentsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStepInstrument_FullMethodName: {authorization.ArchiveRecipeStepInstrumentsPermission},

		// RecipeStepIngredients
		mealplanningsvc.MealPlanningService_CreateRecipeStepIngredient_FullMethodName:  {authorization.CreateRecipeStepIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepIngredients_FullMethodName:    {authorization.ReadRecipeStepIngredientsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepIngredient_FullMethodName:     {authorization.ReadRecipeStepIngredientsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStepIngredient_FullMethodName:  {authorization.UpdateRecipeStepIngredientsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStepIngredient_FullMethodName: {authorization.ArchiveRecipeStepIngredientsPermission},

		// RecipeStepCompletionConditions
		mealplanningsvc.MealPlanningService_CreateRecipeStepCompletionCondition_FullMethodName:  {authorization.CreateRecipeStepCompletionConditionsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepCompletionConditions_FullMethodName:    {authorization.ReadRecipeStepCompletionConditionsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeStepCompletionCondition_FullMethodName:     {authorization.ReadRecipeStepCompletionConditionsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeStepCompletionCondition_FullMethodName:  {authorization.UpdateRecipeStepCompletionConditionsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeStepCompletionCondition_FullMethodName: {authorization.ArchiveRecipeStepCompletionConditionsPermission},

		// RecipeRatings
		mealplanningsvc.MealPlanningService_CreateRecipeRating_FullMethodName:        {authorization.CreateRecipeRatingsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeRating_FullMethodName:           {authorization.ReadRecipeRatingsPermission},
		mealplanningsvc.MealPlanningService_UpdateRecipeRating_FullMethodName:        {authorization.UpdateRecipeRatingsPermission},
		mealplanningsvc.MealPlanningService_ArchiveRecipeRating_FullMethodName:       {authorization.ArchiveRecipeRatingsPermission},
		mealplanningsvc.MealPlanningService_GetRecipeRatingsForRecipe_FullMethodName: {authorization.ReadRecipeRatingsPermission},

		// Comments (typed entrypoints; generic CRUD lives on CommentsService)
		mealplanningsvc.MealPlanningService_AddCommentToRecipe_FullMethodName:   {authorization.CreateCommentsPermission},
		mealplanningsvc.MealPlanningService_AddCommentToMeal_FullMethodName:     {authorization.CreateCommentsPermission},
		mealplanningsvc.MealPlanningService_AddCommentToMealPlan_FullMethodName: {authorization.CreateCommentsPermission},

		// MealPlanOptionVotes
		mealplanningsvc.MealPlanningService_CreateMealPlanOptionVote_FullMethodName:  {authorization.CreateMealPlanOptionVotesPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanOptionVote_FullMethodName:  {authorization.UpdateMealPlanOptionVotesPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanOptionVote_FullMethodName:     {authorization.ReadMealPlanOptionVotesPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanOptionVotes_FullMethodName:    {authorization.ReadMealPlanOptionVotesPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlanOptionVote_FullMethodName: {authorization.ArchiveMealPlanOptionVotesPermission},

		// MealPlanRecipeOptionSelections
		mealplanningsvc.MealPlanningService_CreateMealPlanRecipeOptionSelection_FullMethodName:                {authorization.CreateMealPlanRecipeOptionSelectionsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanRecipeOptionSelection_FullMethodName:                   {authorization.ReadMealPlanRecipeOptionSelectionsPermission},
		mealplanningsvc.MealPlanningService_GetMealPlanRecipeOptionSelectionsForMealPlanOption_FullMethodName: {authorization.ReadMealPlanRecipeOptionSelectionsPermission},
		mealplanningsvc.MealPlanningService_UpdateMealPlanRecipeOptionSelection_FullMethodName:                {authorization.UpdateMealPlanRecipeOptionSelectionsPermission},
		mealplanningsvc.MealPlanningService_ArchiveMealPlanRecipeOptionSelection_FullMethodName:               {authorization.ArchiveMealPlanRecipeOptionSelectionsPermission},

		// Workers
		mealplanningsvc.MealPlanningService_RunFinalizeMealPlanWorker_FullMethodName:               {authorization.RunMealPlanWorkersPermission},
		mealplanningsvc.MealPlanningService_RunMealPlanGroceryListInitializerWorker_FullMethodName: {authorization.RunMealPlanWorkersPermission},
		mealplanningsvc.MealPlanningService_RunMealPlanTaskCreatorWorker_FullMethodName:            {authorization.RunMealPlanWorkersPermission},

		// Search helpers
		mealplanningsvc.MealPlanningService_SearchValidIngredientsByPreparation_FullMethodName:     {authorization.ReadValidIngredientsPermission},
		mealplanningsvc.MealPlanningService_SearchValidMeasurementUnitsByIngredient_FullMethodName: {authorization.ReadValidMeasurementUnitsPermission},
	}
}
