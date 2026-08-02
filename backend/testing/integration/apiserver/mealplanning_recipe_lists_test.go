package integration

import (
	"testing"

	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRecipeLists_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		createRes, err := userClient.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{
			Input: &mealplanninggrpc.RecipeListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createRes)

		listID := createRes.Created.Id

		listsRes, err := userClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.NotNil(t, listsRes)
		assert.NotEmpty(t, listsRes.Results)

		newName := t.Name() + "_updated"
		newDesc := "new desc"
		_, err = userClient.UpdateRecipeList(ctx, &mealplanninggrpc.UpdateRecipeListRequest{
			RecipeListId: listID,
			Input: &mealplanninggrpc.RecipeListUpdateRequestInput{
				Name:        &newName,
				Description: &newDesc,
			},
		})
		assert.NoError(t, err)

		_, err = userClient.ArchiveRecipeList(ctx, &mealplanninggrpc.ArchiveRecipeListRequest{RecipeListId: listID})
		assert.NoError(t, err)
	})
}

func TestRecipeListItems_CompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("should CRUD items", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, userClient := createUserAndClientForTest(t)

		// Create a recipe list
		createListRes, err := userClient.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{
			Input: &mealplanninggrpc.RecipeListCreationRequestInput{
				Name:        t.Name(),
				Description: "desc",
			},
		})
		require.NoError(t, err)
		listID := createListRes.Created.Id

		// Create a recipe to reference
		_, _, createdRecipe := createRecipeForTest(t, nil)

		// Create item
		createItemRes, err := userClient.CreateRecipeListItem(ctx, &mealplanninggrpc.CreateRecipeListItemRequest{
			Input: &mealplanninggrpc.RecipeListItemCreationRequestInput{
				BelongsToRecipeList: listID,
				RecipeId:            createdRecipe.ID,
				Notes:               "notes",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, createItemRes)
		itemID := createItemRes.Created.Id

		// List lists and confirm item present
		listsRes, err := userClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.NotNil(t, listsRes)
		require.NotEmpty(t, listsRes.Results)

		found := false
		for _, l := range listsRes.Results {
			if l.Id == listID && len(l.Items) > 0 {
				found = true
				assert.Equal(t, listID, l.Items[0].BelongsToRecipeList)
				break
			}
		}
		require.True(t, found)

		// Update item
		newNotes := new("new notes")
		_, err = userClient.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input: &mealplanninggrpc.RecipeListItemUpdateRequestInput{
				BelongsToRecipeList: &listID,
				RecipeId:            &createdRecipe.ID,
				Notes:               newNotes,
			},
		})
		assert.NoError(t, err)

		// Archive item
		_, err = userClient.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		assert.NoError(t, err)
	})
}
