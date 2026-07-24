package grpc

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/authentication/sessions"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	commentsfakes "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/fakes"
	commentsmanager "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager"
	commentsmanagermock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/manager/mock"
	commentssvc "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/grpc/generated/services/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/filtering"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	"github.com/primandproper/platform-go/v6/observability/tracing"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// noopCommentsManager is a stub implementation for tests that only need service construction.
type noopCommentsManager struct{}

func (n *noopCommentsManager) CreateComment(_ context.Context, _ *comments.CommentCreationRequestInput) (*comments.Comment, error) {
	return nil, nil
}
func (n *noopCommentsManager) GetComment(_ context.Context, _ string) (*comments.Comment, error) {
	return nil, nil
}
func (n *noopCommentsManager) GetCommentsForReference(_ context.Context, _, _ string, _ *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
	return nil, nil
}
func (n *noopCommentsManager) UpdateComment(_ context.Context, _, _ string, _ *comments.CommentUpdateRequestInput) error {
	return nil
}
func (n *noopCommentsManager) ArchiveComment(_ context.Context, _ string) error {
	return nil
}
func (n *noopCommentsManager) ArchiveCommentsForReference(_ context.Context, _, _ string) error {
	return nil
}

var _ commentsmanager.CommentsDataManager = (*noopCommentsManager)(nil)

func buildCommentsServiceImplForTest(t *testing.T) *serviceImpl {
	t.Helper()

	return &serviceImpl{
		tracer:          tracing.NewTracerForTest(t.Name()),
		logger:          loggingnoop.NewLogger(),
		commentsManager: &noopCommentsManager{},
		sessionContextDataFetcher: func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{
					UserID: commentsfakes.BuildFakeID(),
				},
			}, nil
		},
	}
}

func TestServiceImpl_CreateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		userID := commentsfakes.BuildFakeID()
		recipeID := commentsfakes.BuildFakeID()

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.TargetType = "recipes"
		fakeComment.ReferencedID = recipeID

		mcm.On(reflection.GetMethodName(mcm.CreateComment), testutils.ContextMatcher, mock.Anything).Return(fakeComment, nil)
		s.commentsManager = mcm

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Content:      "test comment",
				TargetType:   "recipes",
				ReferencedId: recipeID,
			},
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, fakeComment.ID, res.Comment.Id)

		mock.AssertExpectationsForObjects(t, mcm)
	})

	T.Run("missing input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{})
		assert.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("missing target_type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Content:      "test",
				ReferencedId: commentsfakes.BuildFakeID(),
			},
		})
		assert.Error(t, err)
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

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		recipeID := commentsfakes.BuildFakeID()
		expected := commentsfakes.BuildFakeCommentList("recipes", recipeID)

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		mcm.On(reflection.GetMethodName(mcm.GetCommentsForReference), testutils.ContextMatcher, "recipes", recipeID, testutils.QueryFilterMatcher).Return(expected, nil)
		s.commentsManager = mcm

		res, err := s.GetCommentsForReference(ctx, &commentssvc.GetCommentsForReferenceRequest{
			TargetType:   "recipes",
			ReferencedId: recipeID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Len(t, res.Data, len(expected.Data))

		mock.AssertExpectationsForObjects(t, mcm)
	})
}

func TestServiceImpl_UpdateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		commentID := commentsfakes.BuildFakeID()
		userID := commentsfakes.BuildFakeID()
		newContent := "updated content"

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = userID

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		mcm.On(reflection.GetMethodName(mcm.GetComment), testutils.ContextMatcher, commentID).Return(fakeComment, nil).Twice()
		mcm.On(reflection.GetMethodName(mcm.UpdateComment), testutils.ContextMatcher, commentID, userID, mock.Anything).Return(nil)
		s.commentsManager = mcm

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: commentID,
			Input:     &commentssvc.CommentUpdateRequestInput{Content: newContent},
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, commentID, res.Comment.Id)

		mock.AssertExpectationsForObjects(t, mcm)
	})

	T.Run("permission_denied_when_different_user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		commentID := commentsfakes.BuildFakeID()
		ownerID := commentsfakes.BuildFakeID()
		requestingUserID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = ownerID

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		mcm.On(reflection.GetMethodName(mcm.GetComment), testutils.ContextMatcher, commentID).Return(fakeComment, nil)
		s.commentsManager = mcm

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: requestingUserID},
			}, nil
		}

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: commentID,
			Input:     &commentssvc.CommentUpdateRequestInput{Content: "updated"},
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		mock.AssertExpectationsForObjects(t, mcm)
	})
}

func TestServiceImpl_ArchiveComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		commentID := commentsfakes.BuildFakeID()
		userID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = userID

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		mcm.On(reflection.GetMethodName(mcm.GetComment), testutils.ContextMatcher, commentID).Return(fakeComment, nil)
		mcm.On(reflection.GetMethodName(mcm.ArchiveComment), testutils.ContextMatcher, commentID).Return(nil)
		s.commentsManager = mcm

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: userID},
			}, nil
		}

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{
			CommentId: commentID,
		})
		assert.NoError(t, err)
		assert.NotNil(t, res)

		mock.AssertExpectationsForObjects(t, mcm)
	})

	T.Run("permission_denied_when_different_user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		s := buildCommentsServiceImplForTest(t)

		commentID := commentsfakes.BuildFakeID()
		ownerID := commentsfakes.BuildFakeID()
		requestingUserID := commentsfakes.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.ID = commentID
		fakeComment.BelongsToUser = ownerID

		mcm := &commentsmanagermock.MockCommentsDataManager{}
		mcm.On(reflection.GetMethodName(mcm.GetComment), testutils.ContextMatcher, commentID).Return(fakeComment, nil)
		s.commentsManager = mcm

		s.sessionContextDataFetcher = func(ctx context.Context) (*sessions.ContextData, error) {
			return &sessions.ContextData{
				Requester: sessions.RequesterInfo{UserID: requestingUserID},
			}, nil
		}

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{
			CommentId: commentID,
		})
		assert.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		mock.AssertExpectationsForObjects(t, mcm)
	})
}
