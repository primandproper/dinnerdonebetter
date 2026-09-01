package uploadedmedia

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia/fakes"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	"github.com/primandproper/platform-go/v13/database"
	"github.com/primandproper/platform-go/v13/database/postgres"
	"github.com/primandproper/platform-go/v13/identifiers"
	loggingnoop "github.com/primandproper/platform-go/v13/observability/logging/noop"
	metricsnoop "github.com/primandproper/platform-go/v13/observability/metrics/noop"
	tracingnoop "github.com/primandproper/platform-go/v13/observability/tracing/noop"
	"github.com/primandproper/platform-go/v13/uploads/registry"

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
func buildDatabaseClientForTest(t *testing.T) (registry.Store, audit.Repository, database.SQLQueryExecutor) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	c, err := ProvideUploadedMediaRepository(
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		pgc,
		nil,
	)
	require.NoError(t, err)

	return c, auditLogEntryRepo, pgc.Writer()
}

// ownedBy builds a fake object registered to userID.
func ownedBy(userID string) *registry.Object {
	object := fakes.BuildFakeUploadedMedia()
	object.OwnerID = userID

	return object
}

func TestRepository_Integration_UploadedMedia(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)
	object := ownedBy(user.ID)

	// record
	require.NoError(t, dbc.RecordObject(ctx, object))
	assert.NotZero(t, object.CreatedAt)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
	})

	fetched, err := dbc.GetObject(ctx, ddbuploadedmedia.Scope(), object.ID)
	require.NoError(t, err)
	assert.Equal(t, object.Key, fetched.Key)
	assert.Equal(t, object.ContentType, fetched.ContentType)
	assert.Equal(t, user.ID, fetched.OwnerID)

	// the key is how a request holding a URL path rather than a row id finds the row
	byKey, err := dbc.GetObjectByKey(ctx, ddbuploadedmedia.Scope(), object.Key)
	require.NoError(t, err)
	assert.Equal(t, object.ID, byKey.ID)

	// the owner's page
	page, err := dbc.ListObjectsByOwner(ctx, ddbuploadedmedia.Scope(), user.ID, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	assert.Equal(t, object.ID, page.Data[0].ID)

	// archive
	require.NoError(t, dbc.ArchiveObject(ctx, ddbuploadedmedia.Scope(), object.ID))
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
	})

	fetchedAfterArchive, err := dbc.GetObject(ctx, ddbuploadedmedia.Scope(), object.ID)
	require.Error(t, err)
	assert.Nil(t, fetchedAfterArchive)
	assert.ErrorIs(t, err, registry.ErrObjectNotFound)
}

// TestRepository_Integration_ArchiveRecordsTheOwner pins the one thing this package's
// ArchiveObject does that the platform's does not: it reads the object first so the
// audit entry can name whose it was.
func TestRepository_Integration_ArchiveRecordsTheOwner(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, writer := buildDatabaseClientForTest(t)

	owner := pgtesting.CreateUserForTest(t, nil, writer)
	archiver := pgtesting.CreateUserForTest(t, nil, writer)

	object := ownedBy(owner.ID)
	require.NoError(t, dbc.RecordObject(ctx, object))
	require.NoError(t, dbc.ArchiveObject(ctx, ddbuploadedmedia.Scope(), object.ID))

	// The entry belongs to whoever uploaded the object, not to whoever happened to be
	// signed in when it was archived.
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, owner.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
	})

	entries, err := auditRepo.GetAuditLogEntriesForUser(ctx, archiver.ID, nil)
	require.NoError(t, err)
	assert.Empty(t, entries.Data)
}

// TestRepository_Integration_ArchiveMissingRecordsNothing pins that a failed archive
// records nothing. The read that finds the owner is also what makes an absent object
// an error before anything is written down about it.
func TestRepository_Integration_ArchiveMissingRecordsNothing(t *testing.T) {
	ctx := t.Context()
	dbc, _, _ := buildDatabaseClientForTest(t)

	err := dbc.ArchiveObject(ctx, ddbuploadedmedia.Scope(), identifiers.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, registry.ErrObjectNotFound)
}

// TestRepository_Integration_KeyIsUniqueAcrossArchival pins that archiving does not
// free the key. Archival here is metadata-only — the bytes are still in the bucket —
// so a second row claiming the key would be a second row for one object, which is
// exactly the drift this table exists to prevent.
func TestRepository_Integration_KeyIsUniqueAcrossArchival(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)

	object := ownedBy(user.ID)
	require.NoError(t, dbc.RecordObject(ctx, object))
	require.NoError(t, dbc.ArchiveObject(ctx, ddbuploadedmedia.Scope(), object.ID))

	second := ownedBy(user.ID)
	second.Key = object.Key

	err := dbc.RecordObject(ctx, second)
	require.Error(t, err)
	assert.ErrorIs(t, err, registry.ErrObjectKeyTaken)
}

// TestRepository_Integration_ErasingTheOwnerRemovesTheRow pins the foreign key this
// repository's migration adds to the platform's table.
//
// The registry ships no key on owner_id — it cannot, because it does not know which
// of a consumer's tables holds a principal — so without the one added in
// internal/repositories/postgres/migrations, a deleted user would leave rows nobody
// can name and nothing erases. The single identity eraser in internal/build/dataprivacy
// covers uploads only for as long as this holds.
func TestRepository_Integration_ErasingTheOwnerRemovesTheRow(t *testing.T) {
	ctx := t.Context()
	dbc, _, writer := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, writer)

	object := ownedBy(user.ID)
	require.NoError(t, dbc.RecordObject(ctx, object))

	_, err := writer.ExecContext(ctx, "DELETE FROM users WHERE id = $1", user.ID)
	require.NoError(t, err)

	fetched, err := dbc.GetObject(ctx, ddbuploadedmedia.Scope(), object.ID)
	require.Error(t, err)
	assert.Nil(t, fetched)
	assert.ErrorIs(t, err, registry.ErrObjectNotFound)
}
