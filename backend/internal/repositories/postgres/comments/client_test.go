package comments

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v9/database"
	mockdatabase "github.com/primandproper/platform-go/v9/database/mock"
	"github.com/primandproper/platform-go/v9/database/postgres"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v9/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v9/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from, so that a test costs a database
// clone rather than a container start plus a migration replay. See
// pgtesting.RunTestsWithSharedDatabase.
func TestMain(m *testing.M) {
	os.Exit(pgtesting.RunTestsWithSharedDatabase(m, func(ctx context.Context, db *sql.DB) error {
		migrator, err := migrations.NewMigrator(loggingnoop.NewLogger())
		if err != nil {
			return err
		}

		return migrator.Migrate(ctx, db)
	}))
}

func buildDatabaseClientForTest(t *testing.T) (*repository, audit.Repository) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, pgc)
	require.NoError(t, err)

	c := ProvideCommentsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditLogEntryRepo, pgc, nil)

	return c.(*repository), auditLogEntryRepo
}

func buildInertClientForTest(t *testing.T) *repository {
	t.Helper()

	c := ProvideCommentsRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), nil, &mockdatabase.ClientMock{ReaderFunc: func() database.SQLQueryExecutor { return nil }, WriterFunc: func() database.SQLQueryExecutor { return nil }}, nil)

	return c.(*repository)
}

func createCommentForTest(t *testing.T, ctx context.Context, input *comments.CommentDatabaseCreationInput, dbc *repository) *comments.Comment {
	t.Helper()

	if input == nil {
		input = fakes.BuildFakeCommentDatabaseCreationInput()
	}

	created, err := dbc.CreateComment(ctx, input)
	assert.NoError(t, err)
	require.NotNil(t, created)

	fetched, err := dbc.GetComment(ctx, created.ID)
	assert.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, created.Content, fetched.Content)
	assert.Equal(t, created.TargetType, fetched.TargetType)
	assert.Equal(t, created.ReferencedID, fetched.ReferencedID)
	assert.Equal(t, created.BelongsToUser, fetched.BelongsToUser)

	return created
}

func TestQuerier_Integration_Comments(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	referencedID := identifiers.New()
	targetType := mealplanning.CommentTargetTypeRecipes

	input := fakes.BuildFakeCommentDatabaseCreationInput()
	input.BelongsToUser = user.ID
	input.ReferencedID = referencedID
	input.TargetType = targetType

	// create
	created := createCommentForTest(t, ctx, input, dbc)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: created.ID},
	})

	// fetch as list
	result, err := dbc.GetCommentsForReference(ctx, targetType, referencedID, nil)
	assert.NoError(t, err)
	require.NotEmpty(t, result.Data)
	assert.Len(t, result.Data, 1)
	assert.Equal(t, created.ID, result.Data[0].ID)

	// update
	newContent := "updated content"
	err = dbc.UpdateComment(ctx, created.ID, user.ID, newContent)
	assert.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeComments, RelevantID: created.ID},
	})

	updated, err := dbc.GetComment(ctx, created.ID)
	assert.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, newContent, updated.Content)
	assert.NotNil(t, updated.LastUpdatedAt)

	// archive
	err = dbc.ArchiveComment(ctx, created.ID)
	assert.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeComments, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeComments, RelevantID: created.ID},
	})

	fetchedAfterArchive, err := dbc.GetComment(ctx, created.ID)
	assert.Error(t, err)
	assert.Nil(t, fetchedAfterArchive)
	assert.ErrorIs(t, err, sql.ErrNoRows)
}

func TestQuerier_Integration_Comments_WithReplies(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	referencedID := identifiers.New()
	targetType := mealplanning.CommentTargetTypeRecipes

	// create parent comment
	parentInput := fakes.BuildFakeCommentDatabaseCreationInput()
	parentInput.BelongsToUser = user.ID
	parentInput.ReferencedID = referencedID
	parentInput.TargetType = targetType
	parent := createCommentForTest(t, ctx, parentInput, dbc)

	// create reply
	replyInput := fakes.BuildFakeCommentDatabaseCreationInput()
	replyInput.BelongsToUser = user.ID
	replyInput.ReferencedID = referencedID
	replyInput.TargetType = targetType
	replyInput.ParentCommentID = &parent.ID
	reply := createCommentForTest(t, ctx, replyInput, dbc)

	// fetch all for reference - should get both parent and reply
	result, err := dbc.GetCommentsForReference(ctx, targetType, referencedID, nil)
	assert.NoError(t, err)
	require.Len(t, result.Data, 2)
	var foundParent, foundReply bool
	for _, c := range result.Data {
		if c.ID == parent.ID {
			foundParent = true
			assert.Nil(t, c.ParentCommentID)
		}
		if c.ID == reply.ID {
			foundReply = true
			require.NotNil(t, c.ParentCommentID)
			assert.Equal(t, parent.ID, *c.ParentCommentID)
		}
	}
	assert.True(t, foundParent)
	assert.True(t, foundReply)
}

func TestQuerier_Integration_ArchiveCommentsForReference(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	referencedID := identifiers.New()
	targetType := mealplanning.CommentTargetTypeMealPlans

	input := fakes.BuildFakeCommentDatabaseCreationInput()
	input.BelongsToUser = user.ID
	input.ReferencedID = referencedID
	input.TargetType = targetType

	created := createCommentForTest(t, ctx, input, dbc)

	// archive all for reference
	err := dbc.ArchiveCommentsForReference(ctx, targetType, referencedID)
	assert.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: created.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeComments, RelevantID: created.ID},
	})

	// comment should no longer be fetchable (GetComment returns archived)
	fetched, err := dbc.GetComment(ctx, created.ID)
	assert.Error(t, err)
	assert.Nil(t, fetched)
	assert.ErrorIs(t, err, sql.ErrNoRows)

	// list should be empty when not including archived
	result, err := dbc.GetCommentsForReference(ctx, targetType, referencedID, nil)
	assert.NoError(t, err)
	assert.Empty(t, result.Data)
}

func TestQuerier_CreateComment(T *testing.T) {
	T.Parallel()

	T.Run("with nil input", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.CreateComment(ctx, nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetComment(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetComment(ctx, "")
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_GetCommentsForReference(T *testing.T) {
	T.Parallel()

	T.Run("with empty target type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetCommentsForReference(ctx, "", "ref-id", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})

	T.Run("with empty referenced id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		actual, err := c.GetCommentsForReference(ctx, mealplanning.CommentTargetTypeRecipes, "", nil)
		assert.Error(t, err)
		assert.Nil(t, actual)
	})
}

func TestQuerier_UpdateComment(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateComment(ctx, "", "user-id", "content")
		assert.Error(t, err)
	})

	T.Run("with empty belongs to user", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.UpdateComment(ctx, "comment-id", "", "content")
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveComment(T *testing.T) {
	T.Parallel()

	T.Run("with empty id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveComment(ctx, "")
		assert.Error(t, err)
	})
}

func TestQuerier_ArchiveCommentsForReference(T *testing.T) {
	T.Parallel()

	T.Run("with empty target type", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveCommentsForReference(ctx, "", "ref-id")
		assert.Error(t, err)
	})

	T.Run("with empty referenced id", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		c := buildInertClientForTest(t)

		err := c.ArchiveCommentsForReference(ctx, mealplanning.CommentTargetTypeRecipes, "")
		assert.Error(t, err)
	})
}

func TestQuerier_Integration_Comments_CursorPagination(t *testing.T) {
	ctx := t.Context()
	dbc, _ := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, dbc.writeDB)
	referencedID := identifiers.New()
	targetType := mealplanning.CommentTargetTypeMeals

	pgtesting.TestCursorBasedPagination(t, ctx, pgtesting.PaginationTestConfig[comments.Comment]{
		TotalItems: 9,
		PageSize:   3,
		ItemName:   "comment",
		CreateItem: func(ctx context.Context, i int) *comments.Comment {
			input := fakes.BuildFakeCommentDatabaseCreationInput()
			input.BelongsToUser = user.ID
			input.ReferencedID = referencedID
			input.TargetType = targetType
			return createCommentForTest(t, ctx, input, dbc)
		},
		FetchPage: func(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[comments.Comment], error) {
			return dbc.GetCommentsForReference(ctx, targetType, referencedID, filter)
		},
		GetID: func(c *comments.Comment) string {
			return c.ID
		},
		CleanupItem: func(ctx context.Context, c *comments.Comment) error {
			return dbc.ArchiveComment(ctx, c.ID)
		},
	})
}
