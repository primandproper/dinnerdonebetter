package integration

import (
	"testing"

	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		require.NoError(t, err)

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
		require.NoError(t, err)

		// Archive item
		_, err = userClient.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		assert.NoError(t, err)
	})
}

func TestRecipeLists_ConfinedToOwner(T *testing.T) {
	T.Parallel()

	T.Run("another household's user can neither see nor change a user's recipe list", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		// Each registered user gets a household of their own, so A and B share no account.
		_, ownerClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		_, _, createdRecipe := createRecipeForTest(t, nil)

		createListRes, err := ownerClient.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{
			Input: &mealplanninggrpc.RecipeListCreationRequestInput{
				Name:        t.Name(),
				Description: t.Name(),
			},
		})
		require.NoError(t, err)
		listID := createListRes.Created.Id

		otherListRes, err := otherClient.CreateRecipeList(ctx, &mealplanninggrpc.CreateRecipeListRequest{
			Input: &mealplanninggrpc.RecipeListCreationRequestInput{
				Name:        t.Name(),
				Description: t.Name(),
			},
		})
		require.NoError(t, err)
		otherListID := otherListRes.Created.Id

		// Reading: B sees only B's own list.
		otherLists, err := otherClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		for _, l := range otherLists.Results {
			assert.NotEqual(t, listID, l.Id, "B was handed A's recipe list")
		}
		require.Len(t, otherLists.Results, 1)
		assert.Equal(t, otherListID, otherLists.Results[0].Id)

		ownerLists, err := ownerClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.Len(t, ownerLists.Results, 1)
		assert.Equal(t, listID, ownerLists.Results[0].Id)

		// Creating: B is refused, A is not.
		itemInput := &mealplanninggrpc.RecipeListItemCreationRequestInput{
			BelongsToRecipeList: listID,
			RecipeId:            createdRecipe.ID,
			Notes:               t.Name(),
		}

		_, err = otherClient.CreateRecipeListItem(ctx, &mealplanninggrpc.CreateRecipeListItemRequest{Input: itemInput})
		assert.Equal(t, codes.NotFound, status.Code(err))

		createItemRes, err := ownerClient.CreateRecipeListItem(ctx, &mealplanninggrpc.CreateRecipeListItemRequest{Input: itemInput})
		require.NoError(t, err)
		itemID := createItemRes.Created.Id

		// Updating: B is refused naming A's list, and naming B's own list with A's item; A is not.
		otherNotes := new(t.Name() + " by B")
		_, err = otherClient.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input: &mealplanninggrpc.RecipeListItemUpdateRequestInput{
				BelongsToRecipeList: &listID,
				RecipeId:            &createdRecipe.ID,
				Notes:               otherNotes,
			},
		})
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = otherClient.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input: &mealplanninggrpc.RecipeListItemUpdateRequestInput{
				BelongsToRecipeList: &otherListID,
				RecipeId:            &createdRecipe.ID,
				Notes:               otherNotes,
			},
		})
		assert.Equal(t, codes.NotFound, status.Code(err))

		ownerLists, err = ownerClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.Len(t, ownerLists.Results, 1)
		require.Len(t, ownerLists.Results[0].Items, 1)
		assert.Equal(t, itemInput.Notes, ownerLists.Results[0].Items[0].Notes, "B's refused update still landed")

		ownerNotes := new(t.Name() + " by A")
		_, err = ownerClient.UpdateRecipeListItem(ctx, &mealplanninggrpc.UpdateRecipeListItemRequest{
			RecipeListItemId: itemID,
			Input: &mealplanninggrpc.RecipeListItemUpdateRequestInput{
				BelongsToRecipeList: &listID,
				RecipeId:            &createdRecipe.ID,
				Notes:               ownerNotes,
			},
		})
		require.NoError(t, err)

		// Archiving: B is refused both ways; A is not.
		_, err = otherClient.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		assert.Equal(t, codes.NotFound, status.Code(err))

		_, err = otherClient.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     otherListID,
		})
		assert.Equal(t, codes.NotFound, status.Code(err))

		ownerLists, err = ownerClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.Len(t, ownerLists.Results, 1)
		require.Len(t, ownerLists.Results[0].Items, 1, "B's refused archive still landed")
		assert.Equal(t, *ownerNotes, ownerLists.Results[0].Items[0].Notes)

		_, err = ownerClient.ArchiveRecipeListItem(ctx, &mealplanninggrpc.ArchiveRecipeListItemRequest{
			RecipeListItemId: itemID,
			RecipeListId:     listID,
		})
		require.NoError(t, err)

		ownerLists, err = ownerClient.GetRecipeLists(ctx, &mealplanninggrpc.GetRecipeListsRequest{})
		require.NoError(t, err)
		require.Len(t, ownerLists.Results, 1)
		assert.Empty(t, ownerLists.Results[0].Items)
	})
}
