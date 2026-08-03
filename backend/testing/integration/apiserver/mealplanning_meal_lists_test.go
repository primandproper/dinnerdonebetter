package integration

import (
	"testing"

	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestMealLists_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createRes, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createRes)

		listID := createRes.Created.Id

		listRes, err := userClient.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		require.NoError(t, err)
		require.NotNil(t, listRes)
		assert.NotEmpty(t, listRes.Results)

		newName := t.Name() + "_updated"
		newDesc := "new desc"
		_, err = userClient.UpdateMealList(ctx, &mealplanninggrpc.UpdateMealListRequest{
			MealListId: listID,
			Input: &mealplanninggrpc.MealListUpdateRequestInput{
				Name:        &newName,
				Description: &newDesc,
			},
		})
		require.NoError(t, err)

		_, err = userClient.ArchiveMealList(ctx, &mealplanninggrpc.ArchiveMealListRequest{MealListId: listID})
		assert.NoError(t, err)
	})
}

func TestMealListItems_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD items", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		// Create a meal list
		createListRes, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		listID := createListRes.Created.Id

		// Create a meal to reference
		createdMeal := createMealForTest(t, userClient, nil)

		// Create item
		createItemRes, err := userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: listID,
				MealId:            createdMeal.ID,
				Notes:             "notes",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createItemRes)

		// List lists and confirm item present
		listsRes, err := userClient.GetMealLists(ctx, &mealplanninggrpc.GetMealListsRequest{})
		require.NoError(t, err)
		require.NotNil(t, listsRes)
		require.NotEmpty(t, listsRes.Results)

		var itemID string
		for _, l := range listsRes.Results {
			if l.Id == listID && len(l.Items) > 0 {
				itemID = l.Items[0].Id
				assert.Equal(t, listID, l.Items[0].BelongsToMealList)
				break
			}
		}
		require.NotEmpty(t, itemID)

		// Update item
		newNotes := new("new notes")
		_, err = userClient.UpdateMealListItem(ctx, &mealplanninggrpc.UpdateMealListItemRequest{
			MealListItemId: itemID,
			Input: &mealplanninggrpc.MealListItemUpdateRequestInput{
				BelongsToMealList: &listID,
				MealId:            &createdMeal.ID,
				Notes:             newNotes,
			},
		})
		require.NoError(t, err)

		// Archive item
		_, err = userClient.ArchiveMealListItem(ctx, &mealplanninggrpc.ArchiveMealListItemRequest{
			MealListItemId: itemID,
			MealListId:     listID,
		})
		assert.NoError(t, err)
	})
}

func TestMealListItems_DuplicatePrevention(T *testing.T) {
	T.Parallel()

	T.Run("rejects adding same meal twice to list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createListRes, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		listID := createListRes.Created.Id

		createdMeal := createMealForTest(t, userClient, nil)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: listID,
				MealId:            createdMeal.ID,
				Notes:             "notes",
			},
		})
		require.NoError(t, err)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: listID,
				MealId:            createdMeal.ID,
				Notes:             "notes2",
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})

	T.Run("allows adding different meals to same list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createListRes, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		listID := createListRes.Created.Id

		m1 := createMealForTest(t, userClient, nil)
		m2 := createMealForTest(t, userClient, nil)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: listID,
				MealId:            m1.ID,
				Notes:             "notes1",
			},
		})
		require.NoError(t, err)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: listID,
				MealId:            m2.ID,
				Notes:             "notes2",
			},
		})
		require.NoError(t, err)
	})

	T.Run("allows same meal in different lists", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createList1Res, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name() + "_1",
				Description: "desc",
			},
		})
		require.NoError(t, err)
		createList2Res, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name() + "_2",
				Description: "desc",
			},
		})
		require.NoError(t, err)

		createdMeal := createMealForTest(t, userClient, nil)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: createList1Res.Created.Id,
				MealId:            createdMeal.ID,
				Notes:             "notes",
			},
		})
		require.NoError(t, err)

		_, err = userClient.CreateMealListItem(ctx, &mealplanninggrpc.CreateMealListItemRequest{
			Input: &mealplanninggrpc.MealListItemCreationRequestInput{
				BelongsToMealList: createList2Res.Created.Id,
				MealId:            createdMeal.ID,
				Notes:             "notes",
			},
		})
		require.NoError(t, err)
	})
}

func TestMealLists_DuplicatePrevention(T *testing.T) {
	T.Parallel()

	T.Run("rejects CreateMealList with duplicate meal_ids in items", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)
		createdMeal := createMealForTest(t, userClient, nil)

		_, err := userClient.CreateMealList(ctx, &mealplanninggrpc.CreateMealListRequest{
			Input: &mealplanninggrpc.MealListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
				Items: []*mealplanninggrpc.MealListItemCreationRequestInput{
					{MealId: createdMeal.ID, Notes: "notes1"},
					{MealId: createdMeal.ID, Notes: "notes2"},
				},
			},
		})
		require.Error(t, err)
		assert.Equal(t, codes.AlreadyExists, status.Code(err))
	})
}
