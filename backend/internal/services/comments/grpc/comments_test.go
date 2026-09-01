package grpc

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/authentication/sessions"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	commentsfakes "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	commentssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/comments"

	comments "github.com/primandproper/platform-go/v13/comments"
	commentsmock "github.com/primandproper/platform-go/v13/comments/mock"
	"github.com/primandproper/platform-go/v13/fake"
	"github.com/primandproper/platform-go/v13/filtering"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// buildCommentsServiceImplForTest builds a service backed by the given store mock. A nil store
// gets an unconfigured mock, which panics if any of its methods are called.
func buildCommentsServiceImplForTest(t *testing.T, store *commentsmock.StoreMock) *serviceImpl {
	t.Helper()

	if store == nil {
		store = &commentsmock.StoreMock{}
	}

	return &serviceImpl{
		tracer:   tracing.NewTracerForTest(t.Name()),
		logger:   loggingnoop.NewLogger(),
		comments: store,
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

	return sessionContextForUser(t, fake.BuildFakeID())
}

// protoTarget is the request half of a target the fakes build.
func protoTarget(target comments.Target) *commentssvc.CommentTarget {
	return &commentssvc.CommentTarget{Type: target.Type.String(), Id: target.ID}
}

func TestServiceImpl_CreateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		target := comments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: fake.BuildFakeID()}
		body := "halved the sugar and it was still too sweet"

		var written *comments.Comment
		mcm := &commentsmock.StoreMock{
			CreateCommentFunc: func(_ context.Context, comment *comments.Comment) error {
				comment.ID = fake.BuildFakeID()
				written = comment

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Body:   body,
				Target: protoTarget(target),
			},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, written.ID, res.Comment.Id)

		// The author comes from the session rather than the request, and the scope is
		// the one this deployment files every comment under.
		assert.Equal(t, userID, written.Author)
		assert.Equal(t, body, written.Body)
		assert.Equal(t, target, written.Target)
		assert.Equal(t, ddbcomments.Scope(), written.Scope)
		assert.Equal(t, comments.RootParentID, written.ParentID)

		assert.Len(t, mcm.CreateCommentCalls(), 1)
	})

	T.Run("with parent", func(t *testing.T) {
		t.Parallel()

		parentID := fake.BuildFakeID()
		target := comments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: fake.BuildFakeID()}

		var written *comments.Comment
		mcm := &commentsmock.StoreMock{
			CreateCommentFunc: func(_ context.Context, comment *comments.Comment) error {
				written = comment

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)

		_, err := s.CreateComment(buildSessionContextForTest(t), &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Body:     "agreed",
				ParentId: parentID,
				Target:   protoTarget(target),
			},
		})
		require.NoError(t, err)
		assert.Equal(t, parentID, written.ParentID)
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

	T.Run("missing target type", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Body:   "test",
				Target: &commentssvc.CommentTarget{Id: fake.BuildFakeID()},
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})

	T.Run("missing target id", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)
		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.CreateComment(ctx, &commentssvc.CreateCommentRequest{
			Input: &commentssvc.CommentCreationRequestInput{
				Body:   "test",
				Target: &commentssvc.CommentTarget{Type: mealplanning.CommentTargetTypeRecipes.String()},
			},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServiceImpl_GetRootComments(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		target := comments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: fake.BuildFakeID()}
		expected := commentsfakes.BuildFakeCommentList(target)

		mcm := &commentsmock.StoreMock{
			ListRootCommentsFunc: func(_ context.Context, scope tenancy.Scope, actualTarget comments.Target, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
				assert.Equal(t, ddbcomments.Scope(), scope)
				assert.Equal(t, target, actualTarget)
				assert.NotNil(t, filter)

				return expected, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)

		res, err := s.GetRootComments(ctx, &commentssvc.GetRootCommentsRequest{Target: protoTarget(target)})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Len(t, res.Data, len(expected.Data))

		assert.Len(t, mcm.ListRootCommentsCalls(), 1)
	})

	T.Run("with malformed target", func(t *testing.T) {
		t.Parallel()

		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.GetRootComments(buildSessionContextForTest(t), &commentssvc.GetRootCommentsRequest{
			Target: &commentssvc.CommentTarget{Type: mealplanning.CommentTargetTypeRecipes.String()},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServiceImpl_GetCommentReplies(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := buildSessionContextForTest(t)

		target := comments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: fake.BuildFakeID()}
		parentID := fake.BuildFakeID()
		expected := commentsfakes.BuildFakeCommentList(target)

		mcm := &commentsmock.StoreMock{
			ListRepliesFunc: func(_ context.Context, scope tenancy.Scope, actualTarget comments.Target, actualParentID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
				assert.Equal(t, ddbcomments.Scope(), scope)
				assert.Equal(t, target, actualTarget)
				assert.Equal(t, parentID, actualParentID)
				assert.NotNil(t, filter)

				return expected, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)

		res, err := s.GetCommentReplies(ctx, &commentssvc.GetCommentRepliesRequest{
			Target:   protoTarget(target),
			ParentId: parentID,
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Len(t, res.Data, len(expected.Data))

		assert.Len(t, mcm.ListRepliesCalls(), 1)
	})
}

func TestServiceImpl_UpdateComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()
		newBody := "updated body"

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.Author = userID

		var written *comments.Comment
		mcm := &commentsmock.StoreMock{
			GetCommentFunc: func(_ context.Context, _ tenancy.Scope, id string) (*comments.Comment, error) {
				assert.Equal(t, fakeComment.ID, id)

				return fakeComment, nil
			},
			UpdateCommentFunc: func(_ context.Context, comment *comments.Comment) error {
				written = comment

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: fakeComment.ID,
			Input:     &commentssvc.CommentUpdateRequestInput{Body: newBody},
		})
		require.NoError(t, err)
		require.NotNil(t, res)
		assert.Equal(t, fakeComment.ID, res.Comment.Id)
		assert.Equal(t, newBody, written.Body)

		assert.Len(t, mcm.GetCommentCalls(), 2)
		assert.Len(t, mcm.UpdateCommentCalls(), 1)
	})

	T.Run("permission denied when different user", func(t *testing.T) {
		t.Parallel()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.Author = fake.BuildFakeID()

		mcm := &commentsmock.StoreMock{
			GetCommentFunc: func(_ context.Context, _ tenancy.Scope, id string) (*comments.Comment, error) {
				assert.Equal(t, fakeComment.ID, id)

				return fakeComment, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, fake.BuildFakeID())

		res, err := s.UpdateComment(ctx, &commentssvc.UpdateCommentRequest{
			CommentId: fakeComment.ID,
			Input:     &commentssvc.CommentUpdateRequestInput{Body: "updated"},
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Empty(t, mcm.UpdateCommentCalls())
	})

	T.Run("missing input", func(t *testing.T) {
		t.Parallel()

		s := buildCommentsServiceImplForTest(t, nil)

		res, err := s.UpdateComment(buildSessionContextForTest(t), &commentssvc.UpdateCommentRequest{
			CommentId: fake.BuildFakeID(),
		})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
	})
}

func TestServiceImpl_ArchiveComment(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		userID := fake.BuildFakeID()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.Author = userID

		mcm := &commentsmock.StoreMock{
			GetCommentFunc: func(_ context.Context, _ tenancy.Scope, id string) (*comments.Comment, error) {
				assert.Equal(t, fakeComment.ID, id)

				return fakeComment, nil
			},
			ArchiveCommentFunc: func(_ context.Context, scope tenancy.Scope, id string) error {
				assert.Equal(t, ddbcomments.Scope(), scope)
				assert.Equal(t, fakeComment.ID, id)

				return nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, userID)

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{CommentId: fakeComment.ID})
		require.NoError(t, err)
		assert.NotNil(t, res)

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Len(t, mcm.ArchiveCommentCalls(), 1)
	})

	T.Run("permission denied when different user", func(t *testing.T) {
		t.Parallel()

		fakeComment := commentsfakes.BuildFakeComment()
		fakeComment.Author = fake.BuildFakeID()

		mcm := &commentsmock.StoreMock{
			GetCommentFunc: func(_ context.Context, _ tenancy.Scope, id string) (*comments.Comment, error) {
				assert.Equal(t, fakeComment.ID, id)

				return fakeComment, nil
			},
		}
		s := buildCommentsServiceImplForTest(t, mcm)
		ctx := sessionContextForUser(t, fake.BuildFakeID())

		res, err := s.ArchiveComment(ctx, &commentssvc.ArchiveCommentRequest{CommentId: fakeComment.ID})
		require.Error(t, err)
		assert.Nil(t, res)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.PermissionDenied, st.Code())

		assert.Len(t, mcm.GetCommentCalls(), 1)
		assert.Empty(t, mcm.ArchiveCommentCalls())
	})
}
