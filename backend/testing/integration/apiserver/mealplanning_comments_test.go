package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	commentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createCommentForRecipeForTest(t *testing.T, recipeID string, clientToUse client.Client, content string) *commentsgrpc.Comment {
	t.Helper()
	ctx := t.Context()

	if content == "" {
		content = "test comment on recipe"
	}

	res, err := clientToUse.AddCommentToRecipe(ctx, &mealplanninggrpc.AddCommentToRecipeRequest{
		RecipeId: recipeID,
		Input:    &commentsgrpc.CommentCreationRequestInput{Content: content},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Comment)

	return res.Comment
}

func TestComments_RecipeCompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		user, testClient := createUserAndClientForTest(t)
		createdComment := createCommentForRecipeForTest(t, createdRecipe.ID, testClient, "initial content")

		assert.NotEmpty(t, createdComment.Id)
		assert.Equal(t, mealplanning.CommentTargetTypeRecipes, createdComment.TargetType)
		assert.Equal(t, createdRecipe.ID, createdComment.ReferencedId)
		assert.Equal(t, "initial content", createdComment.Content)
		assert.NotNil(t, createdComment.CreatedAt)

		// List comments
		listRes, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, listRes)
		assert.GreaterOrEqual(t, len(listRes.Data), 1)
		found := false
		for _, c := range listRes.Data {
			if c.Id == createdComment.Id {
				found = true
				assert.Equal(t, "initial content", c.Content)
				break
			}
		}
		assert.True(t, found, "created comment should appear in list")

		// Update comment
		updatedContent := "updated content"
		_, err = testClient.CommentsService().UpdateComment(ctx, &commentsgrpc.UpdateCommentRequest{
			CommentId: createdComment.Id,
			Input:     &commentsgrpc.CommentUpdateRequestInput{Content: updatedContent},
		})
		assert.NoError(t, err)

		// List again and verify update
		listRes2, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		for _, c := range listRes2.Data {
			if c.Id == createdComment.Id {
				assert.Equal(t, updatedContent, c.Content)
				break
			}
		}

		// Archive comment
		_, err = testClient.CommentsService().ArchiveComment(ctx, &commentsgrpc.ArchiveCommentRequest{
			CommentId: createdComment.Id,
		})
		assert.NoError(t, err)

		// List again - archived comment may or may not appear depending on implementation
		listRes3, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		for _, c := range listRes3.Data {
			assert.NotEqual(t, createdComment.Id, c.Id, "archived comment should not appear")
		}

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: createdComment.Id},
			{EventType: "updated", ResourceType: "comments", RelevantID: createdComment.Id},
			{EventType: "archived", ResourceType: "comments", RelevantID: createdComment.Id},
		})

		// Cleanup
		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("requires auth for creating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		c := buildUnauthenticatedGRPCClientForTest(t)

		res, err := c.AddCommentToRecipe(ctx, &mealplanninggrpc.AddCommentToRecipeRequest{
			RecipeId: createdRecipe.ID,
			Input:    &commentsgrpc.CommentCreationRequestInput{Content: "test"},
		})
		assert.Error(t, err)
		assert.Nil(t, res)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("requires auth for listing", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		_ = createCommentForRecipeForTest(t, createdRecipe.ID, testClient, "")

		c := buildUnauthenticatedGRPCClientForTest(t)
		listRes, err := c.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		assert.Error(t, err)
		assert.Nil(t, listRes)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("requires auth for updating", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		createdComment := createCommentForRecipeForTest(t, createdRecipe.ID, testClient, "")

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.CommentsService().UpdateComment(ctx, &commentsgrpc.UpdateCommentRequest{
			CommentId: createdComment.Id,
			Input:     &commentsgrpc.CommentUpdateRequestInput{Content: "updated"},
		})
		assert.Error(t, err)

		_, _ = testClient.CommentsService().ArchiveComment(ctx, &commentsgrpc.ArchiveCommentRequest{CommentId: createdComment.Id})
		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("requires auth for archiving", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		createdComment := createCommentForRecipeForTest(t, createdRecipe.ID, testClient, "")

		c := buildUnauthenticatedGRPCClientForTest(t)
		_, err := c.CommentsService().ArchiveComment(ctx, &commentsgrpc.ArchiveCommentRequest{
			CommentId: createdComment.Id,
		})
		assert.Error(t, err)

		_, _ = testClient.CommentsService().ArchiveComment(ctx, &commentsgrpc.ArchiveCommentRequest{CommentId: createdComment.Id})
		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

func TestComments_MealCompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)
		createdMeal := createMealForTest(t, userClient, nil)

		res, err := userClient.AddCommentToMeal(ctx, &mealplanninggrpc.AddCommentToMealRequest{
			MealId: createdMeal.ID,
			Input:  &commentsgrpc.CommentCreationRequestInput{Content: "comment on meal"},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeMeals, res.Comment.TargetType)
		assert.Equal(t, createdMeal.ID, res.Comment.ReferencedId)

		listRes, err := userClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeMeals,
			ReferencedId: createdMeal.ID,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listRes.Data), 1)

		AssertAuditLogContainsFuzzyForUser(t, ctx, userClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments"},
		})

		_, _ = userClient.ArchiveMeal(ctx, &mealplanninggrpc.ArchiveMealRequest{MealId: createdMeal.ID})
	})
}

func TestComments_MealPlanCompleteLifecycle(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		res, err := userClient.AddCommentToMealPlan(ctx, &mealplanninggrpc.AddCommentToMealPlanRequest{
			MealPlanId: createdMealPlan.ID,
			Input:      &commentsgrpc.CommentCreationRequestInput{Content: "comment on meal plan"},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeMealPlans, res.Comment.TargetType)
		assert.Equal(t, createdMealPlan.ID, res.Comment.ReferencedId)

		listRes, err := userClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeMealPlans,
			ReferencedId: createdMealPlan.ID,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(listRes.Data), 1)

		AssertAuditLogContainsFuzzyForUser(t, ctx, userClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments"},
		})

		_, _ = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
	})
}
