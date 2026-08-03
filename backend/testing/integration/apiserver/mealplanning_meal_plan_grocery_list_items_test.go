package integration

import (
	"testing"
	"time"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	mpconverters "github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning/converters"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	converters "github.com/primandproper/dinnerdonebetter/backend/internal/services/mealplanning/grpc/converters"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func checkMealPlanGroceryListItemEquality(t *testing.T, expected, actual *mealplanning.MealPlanGroceryListItem) {
	t.Helper()

	assert.NotEmpty(t, actual.ID)

	assert.Equal(t, expected.QuantityPurchased, actual.QuantityPurchased, "expected QuantityPurchased for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.QuantityPurchased, actual.QuantityPurchased)
	assert.Equal(t, expected.PurchasePrice, actual.PurchasePrice, "expected PurchasePrice for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.PurchasePrice, actual.PurchasePrice)
	assert.Equal(t, expected.PurchasedUPC, actual.PurchasedUPC, "expected PurchasedUPC for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.PurchasedUPC, actual.PurchasedUPC)
	assert.Equal(t, expected.PurchasedMeasurementUnit, actual.PurchasedMeasurementUnit, "expected PurchasedMeasurementUnit for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.PurchasedMeasurementUnit, actual.PurchasedMeasurementUnit)
	assert.Equal(t, expected.StatusExplanation, actual.StatusExplanation, "expected StatusExplanation for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.StatusExplanation, actual.StatusExplanation)
	assert.Equal(t, expected.Status, actual.Status, "expected Status for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.Status, actual.Status)
	assert.Equal(t, expected.MeasurementUnit.ID, actual.MeasurementUnit.ID, "expected MeasurementUnitID for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.MeasurementUnit.ID, actual.MeasurementUnit.ID)
	assert.Equal(t, expected.BelongsToMealPlan, actual.BelongsToMealPlan, "expected BelongsToMealPlan for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.BelongsToMealPlan, actual.BelongsToMealPlan)
	assert.Equal(t, expected.Ingredient.ID, actual.Ingredient.ID, "expected ValidIngredientID for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.Ingredient.ID, actual.Ingredient.ID)
	assert.Equal(t, expected.MinQuantityNeeded, actual.MinQuantityNeeded, "expected MinQuantityNeeded for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.MinQuantityNeeded, actual.MinQuantityNeeded)
	assert.Equal(t, expected.MaxQuantityNeeded, actual.MaxQuantityNeeded, "expected MaxQuantityNeeded for meal plan grocery list item %s to be %v, but it was %v", expected.ID, expected.MaxQuantityNeeded, actual.MaxQuantityNeeded)

	assert.NotZero(t, actual.CreatedAt)
}

func TestMealPlanGroceryListItems_Getting(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		meal := createMealForTest(t, userClient, nil)

		now := time.Now().Truncate(time.Second).UTC()
		inTenMinutes := now.Add(10 * time.Minute)
		inOneWeek := now.Add(7 * 24 * time.Hour)

		mealPlan := &mealplanning.MealPlan{
			Notes: "get grocery list item test",
			// voting deadline must be before every event's start time; see MealPlanCreationRequestInput.ValidateWithContext.
			VotingDeadline: now.Add(5 * time.Minute),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Status:         string(mealplanning.MealPlanStatusFinalized),
			TasksCreated:   true,
			Events: []*mealplanning.MealPlanEvent{
				{
					Notes:    "dinner",
					StartsAt: inTenMinutes,
					EndsAt:   inOneWeek,
					MealName: mealplanning.DinnerMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal: mealplanning.Meal{ID: meal.ID},
						},
					},
				},
			},
		}

		exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
		createdRes, err := userClient.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{
			Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput),
		})
		require.NoError(t, err)
		mealPlanID := createdRes.Created.Id

		// The plan was created finalized, so its finalization saga is already running. Wait for
		// the step that builds the grocery list rather than reading straight back.
		firstItemID := awaitMealPlanGroceryListItems(t, ctx, userClient, mealPlanID)[0].Id

		// Get the individual grocery list item
		getRes, err := userClient.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                mealPlanID,
			MealPlanGroceryListItemId: firstItemID,
		})
		require.NoError(t, err)
		require.NotNil(t, getRes)
		require.NotNil(t, getRes.Result)
		assert.Equal(t, firstItemID, getRes.Result.Id)

		// Clean up
		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: mealPlanID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan grocery list item", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		res, err := userClient.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestMealPlanGroceryListItems_Updating(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		meal := createMealForTest(t, userClient, nil)

		now := time.Now().Truncate(time.Second).UTC()
		inTenMinutes := now.Add(10 * time.Minute)
		inOneWeek := now.Add(7 * 24 * time.Hour)

		mealPlan := &mealplanning.MealPlan{
			Notes: "update grocery list item test",
			// voting deadline must be before every event's start time; see MealPlanCreationRequestInput.ValidateWithContext.
			VotingDeadline: now.Add(5 * time.Minute),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Status:         string(mealplanning.MealPlanStatusFinalized),
			TasksCreated:   true,
			Events: []*mealplanning.MealPlanEvent{
				{
					Notes:    "dinner",
					StartsAt: inTenMinutes,
					EndsAt:   inOneWeek,
					MealName: mealplanning.DinnerMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal: mealplanning.Meal{ID: meal.ID},
						},
					},
				},
			},
		}

		exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
		createdRes, err := userClient.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{
			Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput),
		})
		require.NoError(t, err)
		mealPlanID := createdRes.Created.Id

		// The plan was created finalized, so its finalization saga is already running. Wait for
		// the step that builds the grocery list rather than reading straight back.
		firstItemID := awaitMealPlanGroceryListItems(t, ctx, userClient, mealPlanID)[0].Id

		// Update the grocery list item: mark as acquired with purchase details
		newStatus := mealplanninggrpc.MealPlanGroceryListItemStatus_MEAL_PLAN_GROCERY_LIST_ITEM_STATUS_ACQUIRED
		newStatusExplanation := new("purchased from store")
		newQuantityPurchased := new(float32(2.5))
		newPurchasePrice := new(float32(4.99))
		newPurchasedUPC := new("012345678901")

		updateRes, err := userClient.UpdateMealPlanGroceryListItem(ctx, &mealplanninggrpc.UpdateMealPlanGroceryListItemRequest{
			MealPlanId:                mealPlanID,
			MealPlanGroceryListItemId: firstItemID,
			Input: &mealplanninggrpc.MealPlanGroceryListItemUpdateRequestInput{
				Status:            &newStatus,
				StatusExplanation: newStatusExplanation,
				QuantityPurchased: newQuantityPurchased,
				PurchasePrice:     newPurchasePrice,
				PurchasedUpc:      newPurchasedUPC,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, updateRes)
		require.NotNil(t, updateRes.Updated)

		updated := converters.ConvertGRPCMealPlanGroceryListItemToMealPlanGroceryListItem(updateRes.Updated)
		assert.Equal(t, firstItemID, updated.ID)
		assert.Equal(t, string(mealplanning.MealPlanGroceryListItemStatusAcquired), updated.Status)
		assert.Equal(t, "purchased from store", updated.StatusExplanation)
		assert.NotNil(t, updated.QuantityPurchased)
		assert.InDelta(t, float32(2.5), *updated.QuantityPurchased, 0.01)
		assert.NotNil(t, updated.PurchasePrice)
		assert.InDelta(t, float32(4.99), *updated.PurchasePrice, 0.01)
		assert.NotNil(t, updated.PurchasedUPC)
		assert.Equal(t, "012345678901", *updated.PurchasedUPC)

		// Verify the update persisted by re-fetching
		getRes, err := userClient.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                mealPlanID,
			MealPlanGroceryListItemId: firstItemID,
		})
		require.NoError(t, err)
		require.NotNil(t, getRes)

		refetched := converters.ConvertGRPCMealPlanGroceryListItemToMealPlanGroceryListItem(getRes.Result)
		assert.Equal(t, string(mealplanning.MealPlanGroceryListItemStatusAcquired), refetched.Status)
		assert.Equal(t, "purchased from store", refetched.StatusExplanation)

		// Clean up
		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: mealPlanID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		newStatus := mealplanninggrpc.MealPlanGroceryListItemStatus_MEAL_PLAN_GROCERY_LIST_ITEM_STATUS_ACQUIRED

		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.UpdateMealPlanGroceryListItem(ctx, &mealplanninggrpc.UpdateMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanGroceryListItemUpdateRequestInput{
				Status: &newStatus,
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan grocery list item", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		newStatus := mealplanninggrpc.MealPlanGroceryListItemStatus_MEAL_PLAN_GROCERY_LIST_ITEM_STATUS_ACQUIRED

		res, err := userClient.UpdateMealPlanGroceryListItem(ctx, &mealplanninggrpc.UpdateMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
			Input: &mealplanninggrpc.MealPlanGroceryListItemUpdateRequestInput{
				Status: &newStatus,
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}

func TestMealPlanGroceryListItems_Archiving(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		meal := createMealForTest(t, userClient, nil)

		now := time.Now().Truncate(time.Second).UTC()
		inTenMinutes := now.Add(10 * time.Minute)
		inOneWeek := now.Add(7 * 24 * time.Hour)

		mealPlan := &mealplanning.MealPlan{
			Notes: "archive grocery list item test",
			// voting deadline must be before every event's start time; see MealPlanCreationRequestInput.ValidateWithContext.
			VotingDeadline: now.Add(5 * time.Minute),
			ElectionMethod: mealplanning.MealPlanElectionMethodSchulze,
			Status:         string(mealplanning.MealPlanStatusFinalized),
			TasksCreated:   true,
			Events: []*mealplanning.MealPlanEvent{
				{
					Notes:    "dinner",
					StartsAt: inTenMinutes,
					EndsAt:   inOneWeek,
					MealName: mealplanning.DinnerMealName,
					Options: []*mealplanning.MealPlanOption{
						{
							Meal: mealplanning.Meal{ID: meal.ID},
						},
					},
				},
			},
		}

		exampleMealPlanInput := mpconverters.ConvertMealPlanToMealPlanCreationRequestInput(mealPlan)
		createdRes, err := userClient.CreateMealPlan(ctx, &mealplanninggrpc.CreateMealPlanRequest{
			Input: converters.ConvertMealPlanCreationRequestInputToGRPCMealPlanCreationRequestInput(exampleMealPlanInput),
		})
		require.NoError(t, err)
		mealPlanID := createdRes.Created.Id

		// The plan was created finalized, so its finalization saga is already running. Wait for
		// the step that builds the grocery list rather than reading straight back.
		firstItemID := awaitMealPlanGroceryListItems(t, ctx, userClient, mealPlanID)[0].Id

		// Archive the grocery list item
		archiveRes, err := userClient.ArchiveMealPlanGroceryListItem(ctx, &mealplanninggrpc.ArchiveMealPlanGroceryListItemRequest{
			MealPlanId:                mealPlanID,
			MealPlanGroceryListItemId: firstItemID,
		})
		require.NoError(t, err)
		require.NotNil(t, archiveRes)

		// Verify the item is no longer fetchable
		getRes, err := userClient.GetMealPlanGroceryListItem(ctx, &mealplanninggrpc.GetMealPlanGroceryListItemRequest{
			MealPlanId:                mealPlanID,
			MealPlanGroceryListItemId: firstItemID,
		})
		require.Error(t, err)
		assert.Nil(t, getRes)

		// Clean up
		_, err = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: mealPlanID})
		assert.NoError(t, err)
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		c := buildUnauthenticatedGRPCClientForTest(t)
		res, err := c.ArchiveMealPlanGroceryListItem(ctx, &mealplanninggrpc.ArchiveMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})

	T.Run("nonexistent meal plan grocery list item", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		res, err := userClient.ArchiveMealPlanGroceryListItem(ctx, &mealplanninggrpc.ArchiveMealPlanGroceryListItemRequest{
			MealPlanId:                nonexistentID,
			MealPlanGroceryListItemId: nonexistentID,
		})
		require.Error(t, err)
		assert.Nil(t, res)
	})
}
