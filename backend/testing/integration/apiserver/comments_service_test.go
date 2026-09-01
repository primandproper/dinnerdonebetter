package integration

import (
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	commentsgrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	mealplanninggrpc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/pkg/client"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// recipeTarget names one recipe as a comment target.
func recipeTarget(recipeID string) *commentsgrpc.CommentTarget {
	return &commentsgrpc.CommentTarget{
		Type: mealplanning.CommentTargetTypeRecipes.String(),
		Id:   recipeID,
	}
}

func commentsServiceCreateCommentOnRecipe(t *testing.T, recipeID string, c client.Client, body string) *commentsgrpc.Comment {
	t.Helper()
	ctx := t.Context()

	if body == "" {
		body = "test comment via CommentsService"
	}

	res, err := c.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
		Input: &commentsgrpc.CommentCreationRequestInput{
			Body:   body,
			Target: recipeTarget(recipeID),
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
				Body:   "created via CreateComment",
				Target: recipeTarget(createdRecipe.ID),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeRecipes.String(), res.Comment.Target.Type)
		assert.Equal(t, createdRecipe.ID, res.Comment.Target.Id)
		assert.Equal(t, "created via CreateComment", res.Comment.Body)
		assert.Equal(t, user.ID, res.Comment.Author)

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
				Body: "CreateComment on meal plan",
				Target: &commentsgrpc.CommentTarget{
					Type: mealplanning.CommentTargetTypeMealPlans.String(),
					Id:   createdMealPlan.ID,
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res.Comment)
		assert.Equal(t, mealplanning.CommentTargetTypeMealPlans.String(), res.Comment.Target.Type)
		assert.Equal(t, createdMealPlan.ID, res.Comment.Target.Id)

		AssertAuditLogContainsFuzzyForUser(t, ctx, userClient, user.ID, 10, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: res.Comment.Id},
		})

		_, _ = userClient.ArchiveMealPlan(ctx, &mealplanninggrpc.ArchiveMealPlanRequest{MealPlanId: createdMealPlan.ID})
	})

	// The target catalog is what refuses this, and it is the reason the catalog
	// exists: a target type is a string underneath, and a comment stored under a
	// misspelled one is counted and shown nowhere.
	T.Run("rejects an unknown target type", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:   "test",
				Target: &commentsgrpc.CommentTarget{Type: "recipies", Id: createdRecipe.ID},
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.InvalidArgument, status.Code(err))

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	// The existence check the catalog registers for recipes. Without it this write
	// would land, and the comment would be about nothing.
	T.Run("rejects a target that is not there", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, testClient := createUserAndClientForTest(t)

		res, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:   "test",
				Target: recipeTarget("nonexistent_recipe"),
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.NotFound, status.Code(err))
	})

	T.Run("requires auth", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		c := buildUnauthenticatedGRPCClientForTest(t)

		res, err := c.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:   "test",
				Target: recipeTarget(createdRecipe.ID),
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

func TestCommentsService_GetRootComments(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)
		_ = commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "")

		listRes, err := testClient.CommentsService().GetRootComments(ctx, &commentsgrpc.GetRootCommentsRequest{
			Target: recipeTarget(createdRecipe.ID),
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
		listRes, err := c.CommentsService().GetRootComments(ctx, &commentsgrpc.GetRootCommentsRequest{
			Target: recipeTarget(createdRecipe.ID),
		})
		require.Error(t, err)
		assert.Nil(t, listRes)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}

// TestCommentsService_GetCommentReplies pins the two-read thread shape: a reply is
// not in the target's root list, and is found by asking its parent for its replies.
func TestCommentsService_GetCommentReplies(T *testing.T) {
	T.Parallel()

	T.Run("happy path", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		root := commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "the root")

		replyRes, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:     "the reply",
				ParentId: root.Id,
				Target:   recipeTarget(createdRecipe.ID),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, replyRes.Comment)
		assert.Equal(t, root.Id, replyRes.Comment.ParentId)

		roots, err := testClient.CommentsService().GetRootComments(ctx, &commentsgrpc.GetRootCommentsRequest{
			Target: recipeTarget(createdRecipe.ID),
		})
		require.NoError(t, err)
		for _, c := range roots.Data {
			assert.NotEqual(t, replyRes.Comment.Id, c.Id, "a reply should not be in the root list")
		}

		replies, err := testClient.CommentsService().GetCommentReplies(ctx, &commentsgrpc.GetCommentRepliesRequest{
			Target:   recipeTarget(createdRecipe.ID),
			ParentId: root.Id,
		})
		require.NoError(t, err)
		require.Len(t, replies.Data, 1)
		assert.Equal(t, replyRes.Comment.Id, replies.Data[0].Id)

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	// Threads are one level deep, and this is where that is enforced.
	T.Run("refuses a reply to a reply", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, testClient := createUserAndClientForTest(t)

		root := commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, testClient, "the root")

		replyRes, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:     "the reply",
				ParentId: root.Id,
				Target:   recipeTarget(createdRecipe.ID),
			},
		})
		require.NoError(t, err)

		res, err := testClient.CommentsService().CreateComment(ctx, &commentsgrpc.CreateCommentRequest{
			Input: &commentsgrpc.CommentCreationRequestInput{
				Body:     "a reply to a reply",
				ParentId: replyRes.Comment.Id,
				Target:   recipeTarget(createdRecipe.ID),
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.FailedPrecondition, status.Code(err))

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
			Input:     &commentsgrpc.CommentUpdateRequestInput{Body: "updated via CommentsService"},
		})
		require.NoError(t, err)

		AssertAuditLogContainsFuzzyForUser(t, ctx, testClient, user.ID, 15, []*ExpectedAuditEntry{
			{EventType: "created", ResourceType: "comments", RelevantID: createdComment.Id},
			{EventType: "updated", ResourceType: "comments", RelevantID: createdComment.Id},
		})

		listRes, err := testClient.CommentsService().GetRootComments(ctx, &commentsgrpc.GetRootCommentsRequest{
			Target: recipeTarget(createdRecipe.ID),
		})
		require.NoError(t, err)
		for _, c := range listRes.Data {
			if c.Id == createdComment.Id {
				assert.Equal(t, "updated via CommentsService", c.Body)
				break
			}
		}

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})

	T.Run("refuses somebody else's comment", func(t *testing.T) {
		t.Parallel()
		ctx := t.Context()

		_, _, createdRecipe := createRecipeForTest(t, nil)
		_, authorClient := createUserAndClientForTest(t)
		_, otherClient := createUserAndClientForTest(t)

		createdComment := commentsServiceCreateCommentOnRecipe(t, createdRecipe.ID, authorClient, "original")

		res, err := otherClient.CommentsService().UpdateComment(ctx, &commentsgrpc.UpdateCommentRequest{
			CommentId: createdComment.Id,
			Input:     &commentsgrpc.CommentUpdateRequestInput{Body: "not mine to edit"},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, codes.PermissionDenied, status.Code(err))

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

		listRes, err := testClient.CommentsService().GetRootComments(ctx, &commentsgrpc.GetRootCommentsRequest{
			Target: recipeTarget(createdRecipe.ID),
		})
		require.NoError(t, err)
		for _, c := range listRes.Data {
			assert.NotEqual(t, createdComment.Id, c.Id, "archived comment should not appear")
		}

		_, _ = adminClient.ArchiveRecipe(ctx, &mealplanninggrpc.ArchiveRecipeRequest{RecipeId: createdRecipe.ID})
	})
}
