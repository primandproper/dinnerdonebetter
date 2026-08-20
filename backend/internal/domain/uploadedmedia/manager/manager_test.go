package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediamock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"

	platformerrors "github.com/primandproper/platform-go/v12/errors"
	"github.com/primandproper/platform-go/v12/filtering"
	loggingnoop "github.com/primandproper/platform-go/v12/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v12/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// buildUploadedMediaManagerForTest builds a manager backed by the given repository mock. A nil repo
// gets an unconfigured mock, which panics if any of its methods are called.
func buildUploadedMediaManagerForTest(t *testing.T, repo *uploadedmediamock.RepositoryMock) *uploadedMediaManager {
	t.Helper()

	if repo == nil {
		repo = &uploadedmediamock.RepositoryMock{}
	}

	m := NewUploadedMediaDataManager(tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)

	return m.(*uploadedMediaManager)
}

func TestUploadedMediaDataManager_GetUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		expected := fakes.BuildFakeUploadedMedia()

		repo := &uploadedmediamock.RepositoryMock{
			GetUploadedMediaFunc: func(_ context.Context, uploadedMediaID string) (*uploadedmedia.UploadedMedia, error) {
				assert.Equal(t, expected.ID, uploadedMediaID)

				return expected, nil
			},
		}
		manager := buildUploadedMediaManagerForTest(t, repo)

		result, err := manager.GetUploadedMedia(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetUploadedMediaCalls(), 1)
	})
}

func TestUploadedMediaDataManager_GetUploadedMediaForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		userID := fakes.BuildFakeUploadedMedia().CreatedByUser
		filter := filtering.DefaultQueryFilter()
		media := fakes.BuildFakeUploadedMedia()
		expected := &filtering.QueryFilteredResult[uploadedmedia.UploadedMedia]{
			Data: []*uploadedmedia.UploadedMedia{media},
		}

		repo := &uploadedmediamock.RepositoryMock{
			GetUploadedMediaForUserFunc: func(_ context.Context, actualUserID string, actualFilter *filtering.QueryFilter) (*filtering.QueryFilteredResult[uploadedmedia.UploadedMedia], error) {
				assert.Equal(t, userID, actualUserID)
				assert.Equal(t, filter, actualFilter)

				return expected, nil
			},
		}
		manager := buildUploadedMediaManagerForTest(t, repo)

		result, err := manager.GetUploadedMediaForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		assert.Len(t, repo.GetUploadedMediaForUserCalls(), 1)
	})
}

func TestUploadedMediaDataManager_CreateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		dbInput := fakes.BuildFakeUploadedMediaDatabaseCreationInput()
		created := fakes.BuildFakeUploadedMedia()
		created.ID = dbInput.ID

		repo := &uploadedmediamock.RepositoryMock{
			CreateUploadedMediaFunc: func(_ context.Context, input *uploadedmedia.UploadedMediaDatabaseCreationInput) (*uploadedmedia.UploadedMedia, error) {
				assert.NotNil(t, input)

				return created, nil
			},
		}
		manager := buildUploadedMediaManagerForTest(t, repo)

		result, err := manager.CreateUploadedMedia(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, dbInput.ID, result.ID)
		assert.Len(t, repo.CreateUploadedMediaCalls(), 1)
	})

	t.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		repo := &uploadedmediamock.RepositoryMock{}
		manager := buildUploadedMediaManagerForTest(t, repo)

		result, err := manager.CreateUploadedMedia(ctx, nil)

		require.ErrorIs(t, err, platformerrors.ErrNilInputParameter)
		assert.Nil(t, result)
		assert.Empty(t, repo.CreateUploadedMediaCalls())
	})
}

func TestUploadedMediaDataManager_ArchiveUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		uploadedMediaID := fakes.BuildFakeUploadedMedia().ID

		repo := &uploadedmediamock.RepositoryMock{
			ArchiveUploadedMediaFunc: func(_ context.Context, actualID string) error {
				assert.Equal(t, uploadedMediaID, actualID)

				return nil
			},
		}
		manager := buildUploadedMediaManagerForTest(t, repo)

		err := manager.ArchiveUploadedMedia(ctx, uploadedMediaID)

		require.NoError(t, err)
		assert.Len(t, repo.ArchiveUploadedMediaCalls(), 1)
	})
}
