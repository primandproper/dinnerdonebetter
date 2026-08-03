package integration

import (
	"fmt"
	"testing"
	"time"

	types "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/filtering"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func checkMealEquality(t *testing.T, expected, actual *types.Meal) {
	t.Helper()

	assert.NotEmpty(t, actual.ID)

	assert.Equal(t, expected.Name, actual.Name, "expected Name for meal %s to be %v, but it was %v", expected.ID, expected.Name, actual.Name)
	assert.Equal(t, expected.Description, actual.Description, "expected Description for meal %s to be %v, but it was %v", expected.ID, expected.Description, actual.Description)
	assert.Equal(t, expected.MinEstimatedPortions, actual.MinEstimatedPortions, "expected MinEstimatedPortions for meal %s to be %v, but it was %v", expected.ID, expected.MinEstimatedPortions, actual.MinEstimatedPortions)
	assert.Equal(t, expected.MaxEstimatedPortions, actual.MaxEstimatedPortions, "expected MaxEstimatedPortions for meal %s to be %v, but it was %v", expected.ID, expected.MaxEstimatedPortions, actual.MaxEstimatedPortions)
	assert.Equal(t, expected.EligibleForMealPlans, actual.EligibleForMealPlans, "expected EligibleForMealPlans for meal %s to be %v, but it was %v", expected.ID, expected.EligibleForMealPlans, actual.EligibleForMealPlans)

	assert.NotZero(t, actual.CreatedAt)
}

func createMealForTest(t *testing.T, clientToUse client.Client, mealInput *types.Meal) *types.Meal {
	t.Helper()

	ctx := t.Context()

	createdComponents := []*types.MealComponentCreationRequestInput{}
	createdRecipeIDs := []string{}
	for range 3 {
		_, _, recipe := createRecipeForTest(t, nil)
		createdRecipeIDs = append(createdRecipeIDs, recipe.ID)
		createdComponents = append(createdComponents, &types.MealComponentCreationRequestInput{
			RecipeID:      recipe.ID,
			ComponentType: types.MealComponentTypesMain,
			RecipeScale:   1.0,
		})
	}

	exampleMeal := mealInput
	if exampleMeal == nil {
		exampleMeal = fakes.BuildFakeMeal()
	}

	exampleMealInput := mpconverters.ConvertMealToMealCreationRequestInput(exampleMeal)
	exampleMealInput.Components = createdComponents

	createdMealRes, err := clientToUse.CreateMeal(ctx, &mealplanninggrpc.CreateMealRequest{Input: converters.ConvertMealCreationRequestInputToGRPCMealCreationRequestInput(exampleMealInput)})
	require.NoError(t, err)

	fetchedMealRes, err := clientToUse.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: createdMealRes.Created.Id})
	require.NoError(t, err)

	createdMeal := converters.ConvertGRPCMealToMeal(fetchedMealRes.Result)
	checkMealEquality(t, exampleMeal, createdMeal)

	componentRecipeIDs := []string{}
	for _, component := range createdMeal.Components {
		componentRecipeIDs = append(componentRecipeIDs, component.Recipe.ID)
	}

	assert.ElementsMatch(t, createdRecipeIDs, componentRecipeIDs)

	return createdMeal
}

func TestMeals_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createdMeal := createMealForTest(t, userClient, nil)

		_, err := userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
		assert.NoError(t, err)
	})

	T.Run("requires auth for creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		exampleMeal := fakes.BuildFakeMeal()
		exampleMealInput := mpconverters.ConvertMealToMealCreationRequestInput(exampleMeal)
		convertedInput := converters.ConvertMealCreationRequestInputToGRPCMealCreationRequestInput(exampleMealInput)

		c := buildUnauthenticatedGRPCClientForTest(t)
		created, err := c.CreateMeal(ctx, &mealplanninggrpc.CreateMealRequest{
			Input: convertedInput,
		})
		assert.Error(t, err)
		assert.Nil(t, created)
	})
}

func TestMeals_Listing(T *testing.T) {
	T.Parallel()

	T.Run("should be readable in paginated form", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		// Capture timestamp before creating meals so we can filter to only our meals.
		// GetMeals returns all meals globally; without this filter, other tests' meals
		// can fill the first page and ours may be excluded when running the full suite.
		// Use 5 minutes to tolerate clock skew between test runner and API/DB.
		createdAfter := time.Now().UTC().Add(-5 * time.Minute)
		createdAfterProto := timestamppb.New(createdAfter)

		var expected []*types.Meal
		for range 5 {
			createdMeal := createMealForTest(t, userClient, nil)
			expected = append(expected, createdMeal)
		}

		// assert meal list equality - filter to our meals only
		actual, err := userClient.GetMeals(ctx, &mealplanninggrpc.GetMealsRequest{
			Filter: &filtering.QueryFilter{CreatedAfter: createdAfterProto},
		})
		require.NoError(t, err)
		assert.LessOrEqual(
			t,
			len(expected), len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdMeal := range expected {
			_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
			assert.NoError(t, err)
		}
	})

	T.Run("requires auth for listing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		meals, err := c.GetMeals(ctx, &mealplanninggrpc.GetMealsRequest{})
		assert.Error(t, err)
		assert.Nil(t, meals)
	})
}

func TestMeals_Searching(T *testing.T) {
	T.Parallel()

	T.Run("should be searchable", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		// Use a unique prefix based on test name to avoid collisions with other tests
		// that might also create meals/recipes with "example" in the name
		uniquePrefix := fmt.Sprintf("mealsearch-%s", t.Name())
		exampleMeal := fakes.BuildFakeMeal()
		var expected []*types.Meal
		for i := range 5 {
			exampleMeal.Name = fmt.Sprintf("%s-example%d", uniquePrefix, i)
			createdMeal := createMealForTest(t, userClient, exampleMeal)

			expected = append(expected, createdMeal)
		}

		// assert meal list equality - search for the unique prefix to avoid finding meals from other tests
		actual, err := userClient.SearchForMeals(ctx, &mealplanninggrpc.SearchForMealsRequest{
			Query:            uniquePrefix,
			UseSearchService: false,
		})
		require.NoError(t, err)
		assert.LessOrEqual(
			t,
			len(expected), len(actual.Results),
			"expected %d to be <= %d",
			len(expected),
			len(actual.Results),
		)

		for _, createdMeal := range expected {
			_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
			assert.NoError(t, err)
		}
	})

	T.Run("requires auth for searching", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		results, err := c.SearchForMeals(ctx, &mealplanninggrpc.SearchForMealsRequest{
			Query:            "test",
			UseSearchService: false,
		})
		assert.Error(t, err)
		assert.Nil(t, results)
	})
}

func TestMeals_Reading(T *testing.T) {
	T.Parallel()

	T.Run("requires auth for reading", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMeal := createMealForTest(t, userClient, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		meal, err := c.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: createdMeal.ID})
		assert.Error(t, err)
		assert.Nil(t, meal)

		// Clean up
		_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
		assert.NoError(t, err)
	})

	T.Run("nonexistent meal", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		meal, err := userClient.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: nonexistentID})
		assert.Error(t, err)
		assert.Nil(t, meal)
	})
}

func TestMeals_GetMermaidDiagramForMeal(T *testing.T) {
	T.Parallel()

	T.Run("returns non-empty mermaid diagram", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMeal := createMealForTest(t, userClient, nil)

		res, err := userClient.GetMermaidDiagramForMeal(ctx, &mealplanninggrpc.GetMermaidDiagramForMealRequest{MealId: createdMeal.ID})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.NotEmpty(t, res.Response, "mermaid diagram should not be empty")
	})
}

func createMealWithInputForTest(t *testing.T, clientToUse client.Client, name string, components []*types.MealComponentCreationRequestInput) (*types.Meal, *types.MealCreationRequestInput) {
	t.Helper()

	ctx := t.Context()

	exampleMeal := fakes.BuildFakeMeal()
	exampleMeal.Name = name
	exampleMealInput := mpconverters.ConvertMealToMealCreationRequestInput(exampleMeal)
	exampleMealInput.Components = components

	createdMealRes, err := clientToUse.CreateMeal(ctx, &mealplanninggrpc.CreateMealRequest{Input: converters.ConvertMealCreationRequestInputToGRPCMealCreationRequestInput(exampleMealInput)})
	require.NoError(t, err)

	fetchedMealRes, err := clientToUse.GetMeal(ctx, &mealplanninggrpc.GetMealRequest{MealId: createdMealRes.Created.Id})
	require.NoError(t, err)

	return converters.ConvertGRPCMealToMeal(fetchedMealRes.Result), exampleMealInput
}

func TestMeals_DuplicatePrevention(T *testing.T) {
	T.Parallel()

	T.Run("rejects duplicate meal (same name and components)", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		_, _, r1 := createRecipeForTest(t, nil)
		_, _, r2 := createRecipeForTest(t, nil)
		components := []*types.MealComponentCreationRequestInput{
			{RecipeID: r1.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
			{RecipeID: r2.ID, ComponentType: types.MealComponentTypesSide, RecipeScale: 1.0},
		}

		createdMeal, input := createMealWithInputForTest(t, userClient, "Dup Meal", components)
		require.NotNil(t, createdMeal)

		_, err := userClient.CreateMeal(ctx, &mealplanninggrpc.CreateMealRequest{Input: converters.ConvertMealCreationRequestInputToGRPCMealCreationRequestInput(input)})
		assert.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))

		_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
		assert.NoError(t, err)
	})

	T.Run("allows meal with same name but different components", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		_, _, r1 := createRecipeForTest(t, nil)
		_, _, r2 := createRecipeForTest(t, nil)
		m1, _ := createMealWithInputForTest(t, userClient, "Shared Name", []*types.MealComponentCreationRequestInput{
			{RecipeID: r1.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
		})
		m2, _ := createMealWithInputForTest(t, userClient, "Shared Name", []*types.MealComponentCreationRequestInput{
			{RecipeID: r2.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
		})

		require.NotEqual(t, m1.ID, m2.ID)

		_, err := userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m1.ID})
		assert.NoError(t, err)
		_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m2.ID})
		assert.NoError(t, err)
	})

	T.Run("allows meal with same components but different name", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		_, _, r1 := createRecipeForTest(t, nil)
		_, _, r2 := createRecipeForTest(t, nil)
		components := []*types.MealComponentCreationRequestInput{
			{RecipeID: r1.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
			{RecipeID: r2.ID, ComponentType: types.MealComponentTypesSide, RecipeScale: 1.0},
		}

		m1, _ := createMealWithInputForTest(t, userClient, "Meal A", components)
		m2, _ := createMealWithInputForTest(t, userClient, "Meal B", components)

		require.NotEqual(t, m1.ID, m2.ID)

		_, err := userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m1.ID})
		assert.NoError(t, err)
		_, err = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m2.ID})
		assert.NoError(t, err)
	})

	T.Run("allows same meal structure for different user", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, user1Client := createUserAndClientForTest(t)
		_, user2Client := createUserAndClientForTest(t)

		_, _, r1 := createRecipeForTest(t, nil)
		components := []*types.MealComponentCreationRequestInput{
			{RecipeID: r1.ID, ComponentType: types.MealComponentTypesMain, RecipeScale: 1.0},
		}

		m1, _ := createMealWithInputForTest(t, user1Client, "Shared", components)
		m2, _ := createMealWithInputForTest(t, user2Client, "Shared", components)

		require.NotEqual(t, m1.ID, m2.ID)

		_, err := user1Client.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m1.ID})
		assert.NoError(t, err)
		_, err = user2Client.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: m2.ID})
		assert.NoError(t, err)
	})
}

func TestMeals_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("requires auth for archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMeal := createMealForTest(t, userClient, nil)

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
		assert.Error(t, err)
	})

	T.Run("nonexistent meal for archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		_, err := userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: nonexistentID})
		assert.Error(t, err)
	})
}
