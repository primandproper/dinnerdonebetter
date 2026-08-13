package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	commentsmanagermock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/manager/mock"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	"github.com/primandproper/platform-go/v10/filtering"
	loggingnoop "github.com/primandproper/platform-go/v10/observability/logging/noop"
	"github.com/primandproper/platform-go/v10/observability/tracing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildCommentsServiceImplForTest builds a service backed by the given manager mock. A nil manager
// gets an unconfigured mock, which panics if any of its methods are called.
func buildCommentsServiceImplForTest(t *testing.T, commentsManager *commentsmanagermock.CommentsDataManagerMock) *serviceImpl {
	t.Helper()

	if commentsManager == nil {
		commentsManager = &commentsmanagermock.CommentsDataManagerMock{}
	}

	return &serviceImpl{
		tracer:          tracing.NewTracerForTest(t.Name()),
		logger:          loggingnoop.NewLogger(),
		commentsManager: commentsManager,
	}
}

// sessionContextForUser returns a context carrying session data that reports the given user.
func sessionContextForUser(t *testing.T, userID string) context.Context {
	t.Helper()

	return sessions.AttachToContext(t.Context(), &sessions.ContextData{
		Requester: sessions.RequesterInfo{UserID: userID},
	})
}

// buildSessionContextForTest returns a context carrying session data for an arbitrary user.
func buildSessionContextForTest(t *testing.T) context.Context {
	t.Helper()

	return sessionContextForUser(t, commentsfakes.BuildFakeID())
}

func TestServiceImpl_CreateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := commentsfakes.BuildFakeID()
		recipeID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.TargetType = "recipes"
		fakeComment.ReferencedID = recipeID

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			CreateCommentFunc: func(_ context.Context, _ *comments.CommentCreationRequestInput) (*comments.Comment, error) {
				return fakeComment, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Content:      "test comment",
				TargetType:   "recipes",
				ReferencedId: recipeID,
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, fakeComment.ID, res.Comment.Id)

		assert.Len(t, mcm.CreateCommentCalls(), 1)
	})

	T.Run("missing input", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("missing target_type", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Content:      "test",
				ReferencedId: commentsfakes.BuildFakeID(),
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServiceImpl_GetCommentsForReference(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		recipeID := commentsfakes.BuildFakeID()
		expected := commentsfakes.BuildFakeCommentList("recipes", recipeID)

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			GetCommentsForReferenceFunc: func(_ context.Context, targetType, referencedID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
				assert.Equal(t, "recipes", targetType)
				assert.Equal(t, recipeID, referencedID)
				assert.NotNil(t, filter)

				return expected, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)

		res, err := s.GetCommentsForReference(ctx, &commentssvc.GetCommentsForReferenceRequest{
			TargetType:   "recipes",
			ReferencedId: recipeID,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Data, len(expected.Data))

		assert.Len(t, mcm.GetCommentsForReferenceCalls(), 1)
	})
}

func TestServiceImpl_UpdateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		commentID := commentsfakes.BuildFakeID()
		userID := commentsfakes.BuildFakeID()
		newContent := "updated content"

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = userID

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			GetCommentFunc: func(_ context.Context, id string) (*comments.Comment, error) {
				assert.Equal(t, commentID, id)

				return fakeComment, nil
			},
			UpdateCommentFunc: func(_ context.Context, id, belongsToUser string, _ *comments.CommentUpdateRequestInput) error {
				assert.Equal(t, commentID, id)
				assert.Equal(t, userID, belongsToUser)

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: commentID,
			Input:     &commentssvc.CommentUpdateRequestInput{Content: newContent},
		})
		require.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, commentID, res.Comment.Id)

		assert.Len(t, mcm.GetCommentCalls(), 2)
		assert.Len(t, mcm.UpdateCommentCalls(), 1)
	})

	T.Run("permission_denied_when_different_user", func(t *testing.T) {
		t.Parallel()

		commentID := commentsfakes.BuildFakeID()
		ownerID := commentsfakes.BuildFakeID()
		requestingUserID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = ownerID

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			GetCommentFunc: func(_ context.Context, id string) (*comments.Comment, error) {
				assert.Equal(t, commentID, id)

				return fakeComment, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, requestingUserID)

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: commentID,
			Input:     &commentssvc.CommentUpdateRequestInput{Content: "updated"},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Empty(t, mcm.UpdateCommentCalls())
	})
}

func TestServiceImpl_ArchiveComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		commentID := commentsfakes.BuildFakeID()
		userID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = userID

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			GetCommentFunc: func(_ context.Context, id string) (*comments.Comment, error) {
				assert.Equal(t, commentID, id)

				return fakeComment, nil
			},
			ArchiveCommentFunc: func(_ context.Context, id string) error {
				assert.Equal(t, commentID, id)

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{
			CommentId: commentID,
		})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Len(t, mcm.ArchiveCommentCalls(), 1)
	})

	T.Run("permission_denied_when_different_user", func(t *testing.T) {
		t.Parallel()

		commentID := commentsfakes.BuildFakeID()
		ownerID := commentsfakes.BuildFakeID()
		requestingUserID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = ownerID

		mcm := &commentsmanagermock.CommentsDataManagerMock{
			GetCommentFunc: func(_ context.Context, id string) (*comments.Comment, error) {
				assert.Equal(t, commentID, id)

				return fakeComment, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, requestingUserID)

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{
			CommentId: commentID,
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Empty(t, mcm.ArchiveCommentCalls())
	})
}
