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

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	"github.com/primandproper/primitives-go/v2/identifiers"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"

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
func buildDatabaseClientForTest(t *testing.T) (mediaregistry.Store, audit.Repository, database.Client) {
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

	return c, auditLogEntryRepo, pgc
}

// ownedBy builds the RecordObject input for an object belonging to userID.
func ownedBy(userID string) *mediaregistry.ObjectInput {
	input := fakes.BuildFakeUploadedMediaInput()
	input.OwnerID = userID

	return &input
}

// recordT registers one object on a transaction of its own.
//
// As of platform-go v14 a store write takes the caller's database.Tx, so a test that wants one
// row written supplies the transaction the production caller would. These helpers answer with
// the error rather than asserting on it, because two of the writes here are supposed to fail.
func recordT(ctx context.Context, db database.Client, dbc mediaregistry.Store, input *mediaregistry.ObjectInput) (*mediaregistry.Object, error) {
	return writeT(ctx, db, func(tx database.Tx) (*mediaregistry.Object, error) {
		return dbc.RecordObject(ctx, tx, ddbuploadedmedia.Scope(), *input)
	})
}

// archiveT retires one object on a transaction of its own.
func archiveT(ctx context.Context, db database.Client, dbc mediaregistry.Store, objectID string) (*mediaregistry.Object, error) {
	return writeT(ctx, db, func(tx database.Tx) (*mediaregistry.Object, error) {
		return dbc.ArchiveObject(ctx, tx, ddbuploadedmedia.Scope(), objectID)
	})
}

func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}

func TestRepository_Integration_UploadedMedia(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())
	input := ownedBy(user.ID)

	// record
	object, err := recordT(ctx, db, dbc, input)
	require.NoError(t, err)
	assert.NotZero(t, object.CreatedAt)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
	})

	fetched, err := dbc.GetObject(ctx, db.Reader(), ddbuploadedmedia.Scope(), object.ID)
	require.NoError(t, err)
	assert.Equal(t, object.Key, fetched.Key)
	assert.Equal(t, object.ContentType, fetched.ContentType)
	assert.Equal(t, user.ID, fetched.OwnerID)

	// the key is how a request holding a URL path rather than a row id finds the row
	byKey, err := dbc.GetObjectByKey(ctx, db.Reader(), ddbuploadedmedia.Scope(), object.Key)
	require.NoError(t, err)
	assert.Equal(t, object.ID, byKey.ID)

	// the owner's page
	page, err := dbc.ListObjectsByOwner(ctx, db.Reader(), ddbuploadedmedia.Scope(), user.ID, nil)
	require.NoError(t, err)
	require.Len(t, page.Data, 1)
	assert.Equal(t, object.ID, page.Data[0].ID)

	// archive
	_, err = archiveT(ctx, db, dbc, object.ID)
	require.NoError(t, err)
	pgtesting.AssertAuditLogContainsForUser(t, ctx, auditRepo, user.ID, []*audit.AuditLogEntry{
		{EventType: audit.AuditLogEventTypeCreated, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
		{EventType: audit.AuditLogEventTypeArchived, ResourceType: resourceTypeUploadedMedia, RelevantID: object.ID},
	})

	fetchedAfterArchive, err := dbc.GetObject(ctx, db.Reader(), ddbuploadedmedia.Scope(), object.ID)
	require.Error(t, err)
	assert.Nil(t, fetchedAfterArchive)
	assert.ErrorIs(t, err, mediaregistry.ErrObjectNotFound)
}

// TestRepository_Integration_ArchiveRecordsTheOwner pins the one thing this package's
// ArchiveObject does that the platform's does not: it reads the object first so the
// audit entry can name whose it was.
func TestRepository_Integration_ArchiveRecordsTheOwner(t *testing.T) {
	ctx := t.Context()
	dbc, auditRepo, db := buildDatabaseClientForTest(t)

	owner := pgtesting.CreateUserForTest(t, nil, db.Writer())
	archiver := pgtesting.CreateUserForTest(t, nil, db.Writer())

	object, err := recordT(ctx, db, dbc, ownedBy(owner.ID))
	require.NoError(t, err)

	_, err = archiveT(ctx, db, dbc, object.ID)
	require.NoError(t, err)

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
	dbc, _, db := buildDatabaseClientForTest(t)

	_, err := archiveT(ctx, db, dbc, identifiers.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, mediaregistry.ErrObjectNotFound)
}

// TestRepository_Integration_KeyIsUniqueAcrossArchival pins that archiving does not
// free the key. Archival here is metadata-only — the bytes are still in the bucket —
// so a second row claiming the key would be a second row for one object, which is
// exactly the drift this table exists to prevent.
func TestRepository_Integration_KeyIsUniqueAcrossArchival(t *testing.T) {
	ctx := t.Context()
	dbc, _, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())

	first := ownedBy(user.ID)

	object, err := recordT(ctx, db, dbc, first)
	require.NoError(t, err)

	_, err = archiveT(ctx, db, dbc, object.ID)
	require.NoError(t, err)

	second := ownedBy(user.ID)
	second.Key = first.Key

	_, err = recordT(ctx, db, dbc, second)
	require.Error(t, err)
	assert.ErrorIs(t, err, mediaregistry.ErrObjectKeyTaken)
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
	dbc, _, db := buildDatabaseClientForTest(t)

	user := pgtesting.CreateUserForTest(t, nil, db.Writer())

	object, err := recordT(ctx, db, dbc, ownedBy(user.ID))
	require.NoError(t, err)

	_, err = db.Writer().ExecContext(ctx, "DELETE FROM ddb_identity_users WHERE id = $1", user.ID)
	require.NoError(t, err)

	fetched, err := dbc.GetObject(ctx, db.Reader(), ddbuploadedmedia.Scope(), object.ID)
	require.Error(t, err)
	assert.Nil(t, fetched)
	assert.ErrorIs(t, err, mediaregistry.ErrObjectNotFound)
}
