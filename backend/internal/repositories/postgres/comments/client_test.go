package comments

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/build/comments"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/comments/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/mealplanning"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformcomments "github.com/primandproper/platform-go/v15/comments"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

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
func buildDatabaseClientForTest(t *testing.T) (platformcomments.Store, database.Client) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLog(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	auditRecorder := auditLogEntryRepo.Recorder()

	catalog, err := comments.Catalog()
	require.NoError(t, err)

	c, err := ProvideCommentsRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		pgc,
		pgtesting.NewRecorderForTest(t, ctx, pgc, auditRecorder),
		catalog,
	)
	require.NoError(t, err)

	return c, pgc
}

// createComment writes one comment on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test
// that wants one row written supplies the transaction the production caller
// would. See TestRepository_Integration_RecordingJoinsTheCallersTransaction for
// what that buys.
func createComment(t *testing.T, ctx context.Context, db database.Client, store platformcomments.Store, comment *platformcomments.Comment) (*platformcomments.Comment, error) {
	t.Helper()

	var created *platformcomments.Comment

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var createErr error
		created, createErr = store.CreateComment(ctx, tx, tenancy.Global(), comment)

		return createErr
	})

	return created, err
}

func TestRepository_Integration_Comments(t *testing.T) {
	ctx := t.Context()
	dbc, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	target := platformcomments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: identifiers.New()}

	// The author is the one making the requests, which is who every entry below names.
	ctx = pgtesting.AsRequester(ctx, user.ID)

	comment := fakes.BuildFakeComment(mealplanning.CommentTargetTypeRecipes)
	comment.Author = user.ID
	comment.Target = target

	// create
	_, err := createComment(t, ctx, db, dbc, comment)
	require.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})

	fetched, err := dbc.GetComment(ctx, db.Reader(), tenancy.Global(), comment.ID)
	require.NoError(t, err)
	assert.Equal(t, comment.Body, fetched.Body)
	assert.Equal(t, target, fetched.Target)
	assert.Equal(t, user.ID, fetched.Author)

	// read as the target's root list
	roots, err := dbc.ListRootComments(ctx, db.Reader(), tenancy.Global(), target, nil)
	require.NoError(t, err)
	require.Len(t, roots.Data, 1)
	assert.Equal(t, comment.ID, roots.Data[0].ID)

	// update
	fetched.Body = "updated body"
	require.NoError(t, db.WithTransaction(ctx, func(tx database.Tx) error {
		_, updateErr := dbc.UpdateComment(ctx, tx, tenancy.Global(), fetched)

		return updateErr
	}))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
		{EventType: platformaudit.EventUpdated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})

	updated, err := dbc.GetComment(ctx, db.Reader(), tenancy.Global(), comment.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated body", updated.Body)
	assert.NotNil(t, updated.LastUpdatedAt)

	// archive
	require.NoError(t, db.WithTransaction(ctx, func(tx database.Tx) error {
		_, archiveErr := dbc.ArchiveComment(ctx, tx, tenancy.Global(), comment.ID)

		return archiveErr
	}))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, user.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
		{EventType: platformaudit.EventUpdated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
		{EventType: platformaudit.EventArchived, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})

	fetchedAfterArchive, err := dbc.GetComment(ctx, db.Reader(), tenancy.Global(), comment.ID)
	require.Error(t, err)
	assert.Nil(t, fetchedAfterArchive)
	assert.ErrorIs(t, err, platformcomments.ErrCommentNotFound)
}

// TestRepository_Integration_ArchiveRecordsTheAuthorAndTheArchiver pins the two halves of an
// entry about somebody else's comment: it is filed on the author's chain, from the row platform
// hands AfterArchiveComment, and it names the archiver as the one who did it.
//
// The hooks this replaced filed the author as the actor, so an administrator archiving a comment
// was recorded as the author archiving their own. Platform reads the actor off the request and
// the subject off the row, which is the distinction an audit log exists to make.
func TestRepository_Integration_ArchiveRecordsTheAuthorAndTheArchiver(t *testing.T) {
	ctx := t.Context()
	dbc, db := buildDatabaseClientForTest(t)

	author := pgtesting.CreateUserForTest(t, nil, db.Writer())
	archiver := pgtesting.CreateUserForTest(t, nil, db.Writer())

	comment := fakes.BuildFakeComment(mealplanning.CommentTargetTypeRecipes)
	comment.Author = author.ID
	_, err := createComment(t, pgtesting.AsRequester(ctx, author.ID), db, dbc, comment)
	require.NoError(t, err)

	archiving := pgtesting.AsRequester(ctx, archiver.ID)
	require.NoError(t, db.WithTransaction(archiving, func(tx database.Tx) error {
		_, archiveErr := dbc.ArchiveComment(archiving, tx, tenancy.Global(), comment.ID)

		return archiveErr
	}))

	// Both entries are on the author's chain: it is their comment.
	pgtesting.AssertAuditLogContains(t, ctx, db, author.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
		{EventType: platformaudit.EventArchived, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})

	// The author did the writing, and the archiver did the archiving.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, author.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})
	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, archiver.ID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventArchived, ResourceType: platformcomments.ResourceTypeComment, ResourceID: comment.ID},
	})
	assert.Len(t, pgtesting.AuditEntriesForActor(t, ctx, db, archiver.ID), 1)
}

// TestRepository_Integration_ArchiveMissingRecordsNothing pins that a failed
// archive records nothing. Platform's guard makes an absent comment an error
// before the hook that would write anything down is called.
func TestRepository_Integration_ArchiveMissingRecordsNothing(t *testing.T) {
	ctx := t.Context()
	dbc, db := buildDatabaseClientForTest(t)

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		_, archiveErr := dbc.ArchiveComment(ctx, tx, tenancy.Global(), identifiers.New())

		return archiveErr
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, platformcomments.ErrCommentNotFound)
}

// TestRepository_Integration_UnknownTargetType pins that the catalog gates writes.
// A misspelled target type is refused rather than stored under a name nothing
// lists.
func TestRepository_Integration_UnknownTargetType(t *testing.T) {
	ctx := t.Context()
	dbc, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())

	comment := fakes.BuildFakeComment(mealplanning.CommentTargetTypeRecipes)
	comment.Author = user.ID
	comment.Target.Type = "recipies"

	_, err := createComment(t, ctx, db, dbc, comment)
	require.Error(t, err)
	assert.ErrorIs(t, err, platformcomments.ErrUnknownTargetType)
}

// TestRepository_Integration_Replies pins the two-read thread shape: the target's
// roots, then one root's replies. The reply is not in the root list, which is what
// makes the root list's count the count a client renders beside the discussion.
func TestRepository_Integration_Replies(t *testing.T) {
	ctx := t.Context()
	dbc, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	target := platformcomments.Target{Type: mealplanning.CommentTargetTypeRecipes, ID: identifiers.New()}

	root := fakes.BuildFakeComment(mealplanning.CommentTargetTypeRecipes)
	root.Author = user.ID
	root.Target = target
	_, err := createComment(t, ctx, db, dbc, root)
	require.NoError(t, err)

	reply := fakes.BuildFakeCommentReply(root)
	reply.Author = user.ID
	_, err = createComment(t, ctx, db, dbc, reply)
	require.NoError(t, err)

	roots, err := dbc.ListRootComments(ctx, db.Reader(), tenancy.Global(), target, nil)
	require.NoError(t, err)
	require.Len(t, roots.Data, 1)
	assert.Equal(t, root.ID, roots.Data[0].ID)

	replies, err := dbc.ListReplies(ctx, db.Reader(), tenancy.Global(), target, root.ID, nil)
	require.NoError(t, err)
	require.Len(t, replies.Data, 1)
	assert.Equal(t, reply.ID, replies.Data[0].ID)
	assert.Equal(t, root.ID, replies.Data[0].ParentID)
}
