package manager

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/fakes"
	commentskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/keys"
	commentsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/comments/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v5/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v5/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v5/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v5/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v5/observability/tracing/noop"
	"github.com/primandproper/platform-go/v5/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildCommentsManagerForTest(t *testing.T) *commentsManager {
	t.Helper()

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewCommentsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&commentsmock.Repository{},
		queueCfg,
		mpp,
	)
	require.NoError(t, err)

	return m.(*commentsManager)
}

func setupExpectationsForCommentsManager(
	manager *commentsManager,
	repoSetupFunc func(repo *commentsmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &commentsmock.Repository{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.repo = repo

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{repo}
}

func TestCommentsManager_CreateComment(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cm := buildCommentsManagerForTest(t)

		input := fakes.BuildFakeCommentCreationRequestInput()
		expected := fakes.BuildFakeComment()

		expectations := setupExpectationsForCommentsManager(
			cm,
			func(repo *commentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateComment), testutils.ContextMatcher, testutils.MatchType[*comments.CommentDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				comments.CommentCreatedServiceEventType: {commentskeys.CommentIDKey},
			},
		)

		actual, err := cm.CreateComment(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForCommentsManager(
			cm,
			func(repo *commentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateComment), testutils.ContextMatcher, commentID, belongsToUser, input.Content).Return(nil)
			},
			map[string][]string{
				comments.CommentUpdatedServiceEventType: {commentskeys.CommentIDKey},
			},
		)

		err := cm.UpdateComment(ctx, commentID, belongsToUser, input)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestCommentsManager_ArchiveComment(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		cm := buildCommentsManagerForTest(t)

		commentID := fakes.BuildFakeID()

		expectations := setupExpectationsForCommentsManager(
			cm,
			func(repo *commentsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveComment), testutils.ContextMatcher, commentID).Return(nil)
			},
			map[string][]string{
				comments.CommentArchivedServiceEventType: {commentskeys.CommentIDKey},
			},
		)

		err := cm.ArchiveComment(ctx, commentID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
