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

func commentsServiceCreateCommentOnRecipe(t *testing.T, recipeID string, c client.Client, content string) *commentsgrpc.Comment {
	t.Helper()
	ctx := t.Context()

	if content == "" {
		content = "test comment via CommentsService"
	}

	res, err := c.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
		Input: &commentsgrpc.CommentCreationRequestInput{
			Content:      content,
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: recipeID,
		},
	})
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, res.Comment)

	return res.Comment
}

func TestCommentsService_CreateComment(T *testing.T) {
	T.Parallel()

	T.Run("recipe target", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		user, testClient := createUserAndClientForTest(t)

		res, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Content:      "created via CreateComment",
				TargetType:   mealplanning.CommentTargetTypeRecipes,
				ReferencedId: createdRecipe.ID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeRecipes, res.Comment.TargetType)
		assert.Equal(t, createdRecipe.ID, res.Comment.ReferencedId)
		assert.Equal(t, "created via CreateComment", res.Comment.Content)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: res.Comment.Id},
		})

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("meal plan target", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		user, userClient := createUserAndClientForTest(t)
		createdMealPlan := createMealPlanForTest(t, userClient, nil)

		res, err := userClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Content:      "CreateComment on meal plan",
				TargetType:   mealplanning.CommentTargetTypeMealPlans,
				ReferencedId: createdMealPlan.ID,
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeMealPlans, res.Comment.TargetType)
		assert.Equal(t, createdMealPlan.ID, res.Comment.ReferencedId)

		AssertAuditLogContainsFuzzyForUser(t, ctx, userClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: res.Comment.Id},
		})

		_, _ = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		c := buildUnauthenticatedGRPCClientForTest(t)

		res, err := c.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Content:      "test",
				TargetType:   mealplanning.CommentTargetTypeRecipes,
				ReferencedId: createdRecipe.ID,
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

func TestCommentsService_GetCommentsForReference(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		_ = commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "")

		listRes, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		require.NotNil(t, listRes)
		assert.GreaterOrEqual(t, len(listRes.Data), 1)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		_ = commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "")

		c := buildUnauthenticatedGRPCClientForTest(t)
		listRes, err := c.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.Error(t, err)
		assert.Nil(t, listRes)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

func TestCommentsService_UpdateComment(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		user, testClient := createUserAndClientForTest(t)
		createdComment := commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "original")

		_, err := testClient.CommentsService().UpdateComment(ctx, &commentsgrpc.UpdateCommentRequest{
			CommentId: createdComment.Id,
			Input:     &commentsgrpc.CommentUpdateRequestInput{Content: "updated via CommentsService"},
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: createdComment.Id},
			{EventType: "updated", ResourceType: "comments", RelevantID: createdComment.Id},
		})

		listRes, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		for _, c := range listRes.Data {
			if c.Id == createdComment.Id {
				assert.Equal(t, "updated via CommentsService", c.Content)
				break
			}
		}

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

func TestCommentsService_ArchiveComment(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		user, testClient := createUserAndClientForTest(t)
		createdComment := commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "")

		_, err := testClient.CommentsService().ArchiveComment(ctx, &commentsgrpc.ArchiveCommentRequest{
			CommentId: createdComment.Id,
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: createdComment.Id},
			{EventType: "archived", ResourceType: "comments", RelevantID: createdComment.Id},
		})

		listRes, err := testClient.CommentsService().GetCommentsForReference(ctx, &commentsgrpc.GetCommentsForReferenceRequest{
			TargetType:   mealplanning.CommentTargetTypeRecipes,
			ReferencedId: createdRecipe.ID,
		})
		require.NoError(t, err)
		for _, c := range listRes.Data {
			assert.NotEqual(t, createdComment.Id, c.Id, "archived comment should not appear")
		}

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}
