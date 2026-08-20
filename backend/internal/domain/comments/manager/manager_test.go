package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	commentsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/mock"

	"github.com/primandproper/platform-go/v12/fake"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildCommentsManagerForTest(t *testing.T) *commentsManager {
	t.Helper()

	ctx := t.Context()

	m, err := NewCommentsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&commentsmock.RepositoryMock{},
	)
	require.NoError(t, err)

	return m.(*commentsManager)
}

// attachRepositoryToCommentsManager wires a configured repository mock into the manager under test.
func attachRepositoryToCommentsManager(manager *commentsManager, repo *commentsmock.RepositoryMock) {
	manager.repo = repo
}

func TestCommentsManager_CreateComment(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cm := buildCommentsManagerForTest(t)

		input := fakes.BuildFakeCommentCreationRequestInput()
		expected := fakes.BuildFakeComment()

		repo := &commentsmock.RepositoryMock{
			CreateCommentFunc: func(_ context.Context, _ *comments.CommentDatabaseCreationInput) (*comments.Comment, error) {
				return expected, nil
			},
		}
		attachRepositoryToCommentsManager(cm, repo)

		actual, err := cm.CreateComment(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateCommentCalls(), 1)
	})
}

func TestCommentsManager_UpdateComment(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cm := buildCommentsManagerForTest(t)

		comment := fakes.BuildFakeComment()
		commentID := comment.ID
		belongsToUser := comment.BelongsToUser
		input := &comments.CommentUpdateRequestInput{Content: "Updated content"}

		repo := &commentsmock.RepositoryMock{
			UpdateCommentFunc: func(_ context.Context, id, updatedBy, content string) error {
				assert.Equal(t, commentID, id)
				assert.Equal(t, belongsToUser, updatedBy)
				assert.Equal(t, input.Content, content)

				return nil
			},
		}
		attachRepositoryToCommentsManager(cm, repo)

		err := cm.UpdateComment(ctx, commentID, belongsToUser, input)
		require.NoError(t, err)

		assert.Len(t, repo.UpdateCommentCalls(), 1)
	})
}

func TestCommentsManager_ArchiveComment(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cm := buildCommentsManagerForTest(t)

		commentID := fake.BuildFakeID()

		repo := &commentsmock.RepositoryMock{
			ArchiveCommentFunc: func(_ context.Context, id string) error {
				assert.Equal(t, commentID, id)

				return nil
			},
		}
		attachRepositoryToCommentsManager(cm, repo)

		err := cm.ArchiveComment(ctx, commentID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveCommentCalls(), 1)
	})
}
