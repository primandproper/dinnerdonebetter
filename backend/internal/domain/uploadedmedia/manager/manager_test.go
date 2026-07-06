package manager

import (
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	uploadedmediamock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v3/database/filtering"
	loggingnoop "github.com/primandproper/platform-go/v3/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v3/observability/tracing/noop"
	"github.com/primandproper/platform-go/v3/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildUploadedMediaManagerForTest(t *testing.T) (*uploadedMediaManager, *uploadedmediamock.Repository) {
	t.Helper()

	repo := &uploadedmediamock.Repository{}
	m := NewUploadedMediaDataManager(tracingnoop.NewTracerProvider(), loggingnoop.NewLogger(), repo)
	return m.(*uploadedMediaManager), repo
}

func TestUploadedMediaDataManager_GetUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildUploadedMediaManagerForTest(t)

		expected := fakes.BuildFakeUploadedMedia()
		repo.On(reflection.GetMethodName(repo.GetUploadedMedia), testutils.ContextMatcher, expected.ID).Return(expected, nil)

		result, err := manager.GetUploadedMedia(ctx, expected.ID)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestUploadedMediaDataManager_GetUploadedMediaForUser(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildUploadedMediaManagerForTest(t)

		userID := fakes.BuildFakeUploadedMedia().CreatedByUser
		filter := filtering.DefaultQueryFilter()
		media := fakes.BuildFakeUploadedMedia()
		expected := &filtering.QueryFilteredResult[uploadedmedia.UploadedMedia]{
			Data: []*uploadedmedia.UploadedMedia{media},
		}
		repo.On(reflection.GetMethodName(repo.GetUploadedMediaForUser), testutils.ContextMatcher, userID, filter).Return(expected, nil)

		result, err := manager.GetUploadedMediaForUser(ctx, userID, filter)

		require.NoError(t, err)
		assert.Equal(t, expected, result)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestUploadedMediaDataManager_CreateUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildUploadedMediaManagerForTest(t)

		dbInput := fakes.BuildFakeUploadedMediaDatabaseCreationInput()
		created := fakes.BuildFakeUploadedMedia()
		created.ID = dbInput.ID

		repo.On(reflection.GetMethodName(repo.CreateUploadedMedia), testutils.ContextMatcher, mock.Anything).Return(created, nil)

		result, err := manager.CreateUploadedMedia(ctx, dbInput)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, dbInput.ID, result.ID)
		mock.AssertExpectationsForObjects(t, repo)
	})
}

func TestUploadedMediaDataManager_ArchiveUploadedMedia(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		manager, repo := buildUploadedMediaManagerForTest(t)

		uploadedMediaID := fakes.BuildFakeUploadedMedia().ID
		repo.On(reflection.GetMethodName(repo.ArchiveUploadedMedia), testutils.ContextMatcher, uploadedMediaID).Return(nil)

		err := manager.ArchiveUploadedMedia(ctx, uploadedMediaID)

		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, repo)
	})
}
