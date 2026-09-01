package comments

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbcomments "github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformcomments "github.com/primandproper/platform-go/v13/comments"
	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"

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

// buildDatabaseClientForTest builds the store over a real database.
//
// The target catalog is the read-only one, with no existence checks: nothing here
// creates the recipes and meals the fakes point at, and a checked catalog would
// make every write in this file a test of the meal planning repository.
func buildDatabaseClientForTest(t *testing.T) (platformcomments.Store, audit.Repository, database.SQLQueryExecutor) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvideCommentsRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		pgc,
		nil,
		comments.Catalog(),
	)
	require.NoError(t, err)

	return c, auditLogEntryRepo, pgc.Writer()
}

func TestRepository_Integration_Comments(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)
	target := platformcomments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: identifiers.New()}

	comment := fakes.BuildFakeComment()
	comment.Author = user.ID
	comment.Target = target

	// create
	require.NoError(t, dbc.CreateComment(ctx, comment))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
	})

	fetched, err := dbc.GetComment(ctx, ddbcomments.Scope(), comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.Body, fetched.Body)
	assert.Equal(t, target, fetched.Target)
	assert.Equal(t, user.ID, fetched.Author)

	// read as the target's root list
	roots, err := dbc.ListRootComments(ctx, ddbcomments.Scope(), target, nil)
	require.NoError(t, err)
	require.Len(t, roots.Data, 1)
	assert.Equal(t, comment.ID, roots.Data[0].ID)

	// update
	fetched.Body = "updated body"
	require.NoError(t, dbc.UpdateComment(ctx, fetched))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
	})

	updated, err := dbc.GetComment(ctx, ddbcomments.Scope(), comment.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated body", updated.Body)
	assert.NotNil(t, updated.LastUpdatedAt)

	// archive
	require.NoError(t, dbc.ArchiveComment(ctx, ddbcomments.Scope(), comment.ID))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
		{EventType: audit.AuditLogEventTypeUpdated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeComments, RelevantID: comment.ID},
	})

	fetchedAfterArchive, err := dbc.GetComment(ctx, ddbcomments.Scope(), comment.ID)
	require.Error(t, err)
	assert.Nil(t, fetchedAfterArchive)
	assert.ErrorIs(t, err, platformcomments.ErrCommentNotFound)
}

// TestRepository_Integration_ArchiveRecordsTheAuthor pins the one thing this
// package's ArchiveComment does that the platform's does not: it reads the comment
// first so the audit entry can name whose it was.
func TestRepository_Integration_ArchiveRecordsTheAuthor(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)

	author := pgtesting.CreateUserForTest(t, nil, writer)
	archiver := pgtesting.CreateUserForTest(t, nil, writer)

	comment := fakes.BuildFakeComment()
	comment.Author = author.ID
	require.NoError(t, dbc.CreateComment(ctx, comment))

	require.NoError(t, dbc.ArchiveComment(ctx, ddbcomments.Scope(), comment.ID))

	// The entry belongs to whoever wrote the comment, not to whoever happened to be
	// signed in when it was archived.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, author.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeComments, RelevantID: comment.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeComments, RelevantID: comment.ID},
	})

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, archiver.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, entries.Data)
}

// TestRepository_Integration_ArchiveMissingRecordsNothing pins that a failed
// archive records nothing. The read that finds the author is also what makes an
// absent comment an error before anything is written down about it.
func TestRepository_Integration_ArchiveMissingRecordsNothing(t *testing.T) {
	ctx := t.Context()
	dbc, _, _ := buildDatabaseClientForTest(t)

	err := dbc.ArchiveComment(ctx, ddbcomments.Scope(), identifiers.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, platformcomments.ErrCommentNotFound)
}

// TestRepository_Integration_UnknownTargetType pins that the catalog gates writes.
// A misspelled target type is refused rather than stored under a name nothing
// lists.
func TestRepository_Integration_UnknownTargetType(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)

	comment := fakes.BuildFakeComment()
	comment.Author = user.ID
	comment.Target.Type = "recipies"

	err := dbc.CreateComment(ctx, comment)
	require.Error(t, err)
	assert.ErrorIs(t, err, platformcomments.ErrUnknownTargetType)
}

// TestRepository_Integration_Replies pins the two-read thread shape: the target's
// roots, then one root's replies. The reply is not in the root list, which is what
// makes the root list's count the count a client renders beside the discussion.
func TestRepository_Integration_Replies(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)
	target := platformcomments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: identifiers.New()}

	root := fakes.BuildFakeComment()
	root.Author = user.ID
	root.Target = target
	require.NoError(t, dbc.CreateComment(ctx, root))

	reply := fakes.BuildFakeCommentReply(root)
	reply.Author = user.ID
	require.NoError(t, dbc.CreateComment(ctx, reply))

	roots, err := dbc.ListRootComments(ctx, ddbcomments.Scope(), target, nil)
	require.NoError(t, err)
	require.Len(t, roots.Data, 1)
	assert.Equal(t, root.ID, roots.Data[0].ID)

	replies, err := dbc.ListReplies(ctx, ddbcomments.Scope(), target, root.ID, nil)
	require.NoError(t, err)
	require.Len(t, replies.Data, 1)
	assert.Equal(t, reply.ID, replies.Data[0].ID)
	assert.Equal(t, root.ID, replies.Data[0].ParentID)
}
