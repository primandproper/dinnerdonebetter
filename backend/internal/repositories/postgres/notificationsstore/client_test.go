package notificationsstore

import (
	"context"
	"database/sql"
	"os"
	"testing"

	ddbnotifications "github.com/primandproper/dinnerdonebetter/backend/internal/domain/notifications"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v14/audit"
	platformnotifications "github.com/primandproper/platform-go/v14/notifications"
	"github.com/primandproper/primitives-go/v2/database"
	"github.com/primandproper/primitives-go/v2/database/postgres"
	loggingnoop "github.com/primandproper/primitives-go/v2/observability/logging/noop"
	metricsnoop "github.com/primandproper/primitives-go/v2/observability/metrics/noop"
	tracingnoop "github.com/primandproper/primitives-go/v2/observability/tracing/noop"
	"github.com/primandproper/primitives-go/v2/tenancy"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMain starts the one postgres container this package's tests share and migrates
// the template database each of them is cloned from. See
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

// buildStoresForTest builds the hooked inbox and registry over a real database.
func buildStoresForTest(t *testing.T) (platformnotifications.Inbox, platformnotifications.Registry, database.Client) {
	t.Helper()

	ctx := t.Context()

	// Already migrated: the template this was cloned from was migrated once in TestMain.
	_, config := pgtesting.NewIsolatedDatabaseForTest(t)

	pgc, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
	require.NotNil(t, pgc)
	require.NoError(t, err)

	auditLogEntryRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), pgc)
	require.NoError(t, err)

	inbox, registry, err := ProvideStores(
		ctx,
		loggingnoop.NewLogger(),
		tracingnoop.NewTracerProvider(),
		metricsnoop.NewMetricsProvider(),
		auditLogEntryRepo,
		nil,
		pgc,
	)
	require.NoError(t, err)

	return inbox, registry, pgc
}

// TestRepository_Integration_InboxWritesAreRecorded pins which inbox writes land an
// entry under the person they are about, and which deliberately do not.
func TestRepository_Integration_InboxWritesAreRecorded(t *testing.T) {
	ctx := t.Context()
	inbox, _, db := buildStoresForTest(t)
	scope := tenancy.Global()

	userID := pgtesting.CreateUserForTest(t, nil, db.Writer()).ID

	created, err := writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Notification, error) {
		return inbox.CreateNotification(ctx, tx, scope, &platformnotifications.Notification{
			Principal: userID,
			Topic:     ddbnotifications.DefaultTopic,
			Title:     "Your meal plan is ready",
		})
	})
	require.NoError(t, err)

	// Reading is not recorded: the row's read_at is the record.
	_, err = writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Notification, error) {
		return inbox.MarkNotificationRead(ctx, tx, scope, userID, created.ID)
	})
	require.NoError(t, err)

	_, err = writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Notification, error) {
		return inbox.ArchiveNotification(ctx, tx, scope, userID, created.ID)
	})
	require.NoError(t, err)

	// Nor is the erasure, which dataprivacy records for itself.
	erased, err := writeT(ctx, db, func(tx database.Tx) (int64, error) {
		return inbox.DeleteNotificationsForPrincipal(ctx, tx, scope, userID)
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), erased)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, userID, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: resourceTypeUserNotifications, ResourceID: created.ID},
		{EventType: platformaudit.EventArchived, ResourceType: resourceTypeUserNotifications, ResourceID: created.ID},
	})
	assert.Len(t, pgtesting.AuditEntriesForActor(t, ctx, db, userID), 2)
}

// TestRepository_Integration_RegistryWritesAreRecorded pins the registry's entries,
// including the one a handset changing hands writes under its new owner.
func TestRepository_Integration_RegistryWritesAreRecorded(t *testing.T) {
	ctx := t.Context()
	_, registry, db := buildStoresForTest(t)
	scope := tenancy.Global()

	firstOwner := pgtesting.CreateUserForTest(t, nil, db.Writer()).ID
	secondOwner := pgtesting.CreateUserForTest(t, nil, db.Writer()).ID

	registered, err := writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Device, error) {
		return registry.RegisterDevice(ctx, tx, scope, &platformnotifications.Device{
			Principal: firstOwner,
			Platform:  platformnotifications.PlatformIOS,
			Token:     "token-a",
		})
	})
	require.NoError(t, err)

	moved, err := writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Device, error) {
		return registry.RegisterDevice(ctx, tx, scope, &platformnotifications.Device{
			Principal: secondOwner,
			Platform:  platformnotifications.PlatformIOS,
			Token:     "token-a",
		})
	})
	require.NoError(t, err)
	require.Equal(t, registered.ID, moved.ID)

	_, err = writeT(ctx, db, func(tx database.Tx) (*platformnotifications.Device, error) {
		return registry.RevokeDevice(ctx, tx, scope, secondOwner, moved.ID)
	})
	require.NoError(t, err)

	_, err = writeT(ctx, db, func(tx database.Tx) (int64, error) {
		return registry.DeleteDevicesForPrincipal(ctx, tx, scope, firstOwner)
	})
	require.NoError(t, err)

	firstEntries := pgtesting.AuditEntriesForActor(t, ctx, db, firstOwner)
	require.Len(t, firstEntries, 1)
	assert.Equal(t, platformaudit.EventCreated, firstEntries[0].EventType)
	assert.Equal(t, resourceTypeUserDeviceTokens, firstEntries[0].ResourceType)
	assert.Equal(t, registered.ID, firstEntries[0].ResourceID)
	assert.Empty(t, firstEntries[0].Changes)

	pgtesting.AssertAuditLogContainsForUser(t, ctx, db, secondOwner, []pgtesting.ExpectedAuditEntry{
		{EventType: platformaudit.EventCreated, ResourceType: resourceTypeUserDeviceTokens, ResourceID: moved.ID},
		{EventType: platformaudit.EventArchived, ResourceType: resourceTypeUserDeviceTokens, ResourceID: moved.ID},
	})

	// The re-registration says whose handset it was.
	var reregistration *platformaudit.Entry
	for _, e := range pgtesting.AuditEntriesForActor(t, ctx, db, secondOwner) {
		if e.EventType == platformaudit.EventCreated {
			reregistration = e
		}
	}
	require.NotNil(t, reregistration)
	require.Contains(t, reregistration.Changes, "principal")
	assert.Equal(t, firstOwner, reregistration.Changes["principal"].Old)
	assert.Equal(t, secondOwner, reregistration.Changes["principal"].New)
}

// writeT runs one store write on a transaction of its own, answering with the error
// rather than asserting on it.
func writeT[T any](ctx context.Context, db database.Client, write func(tx database.Tx) (T, error)) (T, error) {
	var out T

	err := db.WithTransaction(ctx, func(tx database.Tx) error {
		var writeErr error
		out, writeErr = write(tx)

		return writeErr
	})

	return out, err
}
