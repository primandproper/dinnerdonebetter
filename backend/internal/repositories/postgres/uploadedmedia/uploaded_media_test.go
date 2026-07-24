package uploadedmedia

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	types "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	pgtesting "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v6/filtering"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createUploadedMediaForTest(t *testing.T, ctx context.Context, exampleUploadedMedia *types.UploadedMedia, dbc *repository) *types.UploadedMedia {
	t.Helper()

	// create
	if exampleUploadedMedia == nil {
		exampleUploadedMedia = fakes.BuildFakeUploadedMedia()
	}
	dbInput := &types.UploadedMediaDatabaseCreationInput{
		ID:            exampleUploadedMedia.ID,
		StoragePath:   exampleUploadedMedia.StoragePath,
		MimeType:      exampleUploadedMedia.MimeType,
		CreatedByUser: exampleUploadedMedia.CreatedByUser,
	}

	created, err := dbc.CreateUploadedMedia(ctx, dbInput)
	assert.NoError(t, err)
	require.NotNil(t, created)

	exampleUploadedMedia.CreatedAt = created.CreatedAt
	assert.Equal(t, exampleUploadedMedia, created)

	uploadedMedia, err := dbc.GetUploadedMedia(ctx, created.ID)
	exampleUploadedMedia.CreatedAt = uploadedMedia.CreatedAt

	assert.NoError(t, err)
	assert.Equal(t, uploadedMedia, exampleUploadedMedia)

	return created
}

func TestQuerier_Integration_UploadedMedia(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, auditRepo, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	exampleUploadedMedia := fakes.BuildFakeUploadedMedia()
	exampleUploadedMedia.CreatedByUser = user.ID
	createdUploadedMedia := []*types.UploadedMedia{}

	// create
	createdUploadedMedia = append(createdUploadedMedia, createUploadedMediaForTest(t, ctx, exampleUploadedMedia, dbc))

	// create more
	for i := range exampleQuantity {
		input := fakes.BuildFakeUploadedMedia()
		input.StoragePath = fmt.Sprintf("/storage/path/%d.png", i)
		input.CreatedByUser = user.ID
		createdUploadedMedia = append(createdUploadedMedia, createUploadedMediaForTest(t, ctx, input, dbc))
	}

	// Assert audit log entries for creates (uploaded media is user-scoped)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: createdUploadedMedia[0].ID},
	})

	// fetch as list for user
	uploadedMediaList, err := dbc.GetUploadedMediaForUser(ctx, user.ID, nil)
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadedMediaList.Data)
	assert.Equal(t, len(createdUploadedMedia), len(uploadedMediaList.Data))

	// fetch with IDs
	ids := []string{createdUploadedMedia[0].ID, createdUploadedMedia[1].ID}
	uploadedMediaWithIDs, err := dbc.GetUploadedMediaWithIDs(ctx, ids)
	assert.NoError(t, err)
	assert.Len(t, uploadedMediaWithIDs, 2)
	assert.Contains(t, ids, uploadedMediaWithIDs[0].ID)
	assert.Contains(t, ids, uploadedMediaWithIDs[1].ID)

	// update
	createdUploadedMedia[0].StoragePath = "/new/storage/path.png"
	assert.NoError(t, dbc.UpdateUploadedMedia(ctx, createdUploadedMedia[0]))

	// Assert audit log entry for update
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeUploadedMedia, RelevantID: createdUploadedMedia[0].ID},
	})

	// fetch again to verify update
	updated, err := dbc.GetUploadedMedia(ctx, createdUploadedMedia[0].ID)
	assert.NoError(t, err)
	assert.Equal(t, "/new/storage/path.png", updated.StoragePath)

	// delete
	for _, uploadedMedia := range createdUploadedMedia {
		assert.NoError(t, dbc.ArchiveUploadedMedia(ctx, uploadedMedia.ID))
	}

	for _, uploadedMedia := range createdUploadedMedia {
		var y *types.UploadedMedia
		y, err = dbc.GetUploadedMedia(ctx, uploadedMedia.ID)
		assert.Nil(t, y)
		assert.Error(t, err)
		assert.ErrorIs(t, err, sql.ErrNoRows)
	}
}

func TestQuerier_GetUploadedMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid uploaded media MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUploadedMedia(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUploadedMediaWithIDs(T *testing.T) {
	T.Parallel()

	T.Run("with empty MealPlanTaskID list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUploadedMediaWithIDs(ctx, []string{})
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil MealPlanTaskID list", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetUploadedMediaWithIDs(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetUploadedMediaForUser(T *testing.T) {
	T.Parallel()

	T.Run("with invalid user MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		filter := filtering.DefaultQueryFilter()
		c := buildInertClientForTest(t)

		actual, err := c.GetUploadedMediaForUser(ctx, "", filter)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with nil filter", func(t *testing.T) {
		t.Parallel()

		if !pgtesting.RunContainerTests {
			t.SkipNow()
		}

		ctx := t.Context()
		dbc, _, container := buildDatabaseClientForTest(t)

		databaseURI, err := container.ConnectionString(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, databaseURI)

		defer func(t *testing.T) {
			t.Helper()
			assert.NoError(t, container.Terminate(ctx))
		}(t)

		user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

		exampleUploadedMedia := fakes.BuildFakeUploadedMedia()
		exampleUploadedMedia.CreatedByUser = user.ID

		created := createUploadedMediaForTest(t, ctx, exampleUploadedMedia, dbc)

		// Should work with nil filter (uses default)
		actual, err := dbc.GetUploadedMediaForUser(ctx, user.ID, nil)
		assert.NoError(t, err)
		assert.NotNil(t, actual)
		assert.NotEmpty(t, actual.Data)

		// Cleanup
		assert.NoError(t, dbc.ArchiveUploadedMedia(ctx, created.ID))
	})
}

func TestQuerier_CreateUploadedMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateUploadedMedia(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateUploadedMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateUploadedMedia(ctx, nil)
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveUploadedMedia(T *testing.T) {
	T.Parallel()

	T.Run("with invalid uploaded media MealPlanTaskID", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		assert.Error(t, c.ArchiveUploadedMedia(ctx, ""))
	})
}

func TestQuerier_Integration_CursorBasedPagination(t *testing.T) {
	if !pgtesting.RunContainerTests {
		t.SkipNow()
	}

	ctx := t.Context()
	dbc, _, container := buildDatabaseClientForTest(t)

	databaseURI, err := container.ConnectionString(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, databaseURI)

	defer func(t *testing.T) {
		t.Helper()
		assert.NoError(t, container.Terminate(ctx))
	}(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)

	// Use the generic pagination test helper
	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[types.UploadedMedia]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "uploaded media",
		CreateItem: func(ctx context.Context, i int) *types.UploadedMedia {
			uploadedMedia := fakes.BuildFakeUploadedMedia()
			uploadedMedia.StoragePath = fmt.Sprintf("/storage/path/%02d.png", i) // Use zero-padded numbers for consistent sorting
			uploadedMedia.CreatedByUser = user.ID
			return createUploadedMediaForTest(t, ctx, uploadedMedia, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[types.UploadedMedia], error) {
			return dbc.GetUploadedMediaForUser(ctx, user.ID, filter)
		},
		GetID: func(uploadedMedia *types.UploadedMedia) string {
			return uploadedMedia.ID
		},
		CleanupItem: func(ctx context.Context, uploadedMedia *types.UploadedMedia) error {
			return dbc.ArchiveUploadedMedia(ctx, uploadedMedia.ID)
		},
	})
}
