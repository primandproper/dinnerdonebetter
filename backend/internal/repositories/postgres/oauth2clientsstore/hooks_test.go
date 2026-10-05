package oauth2clientsstore

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformoauth2clients "github.com/primandproper/platform-go/v15/authentication/oauth2clients"
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

func TestHooks_Integration(T *testing.T) {
	T.Parallel()

	T.Run("every write through the service lands its audit entry", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		_, config := pgtesting.NewIsolatedDatabaseForTest(t)

		db, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
		require.NoError(t, err)

		auditRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), db)
		require.NoError(t, err)

		store, err := platformoauth2clients.NewSQLStore(db, platformoauth2clients.WithTablePrefix(branding.TablePrefix))
		require.NoError(t, err)

		svc, err := platformoauth2clients.NewService(db, store, platformoauth2clients.WithHooks(
			ProvideHooks(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditRepo, nil),
		))
		require.NoError(t, err)

		// The entries are filed under the registration's owner, and the audit chain
		// requires that owner to exist.
		owner := pgtesting.CreateUserForTest(t, nil, db.Writer())
		account := pgtesting.CreateAccountForTest(t, nil, owner.ID, db.Writer())
		scope := tenancy.Of(account.ID)

		issued, err := svc.CreateClient(ctx, scope, owner.ID, &platformoauth2clients.CreationInput{
			Name:         "a client",
			RedirectURIs: []string{"https://example.com/callback"},
		})
		require.NoError(t, err)

		_, err = svc.UpdateClient(ctx, scope, issued.Client.ID, &platformoauth2clients.UpdateInput{
			Name:         "renamed",
			RedirectURIs: []string{"https://example.com/callback"},
		})
		require.NoError(t, err)

		require.NoError(t, svc.ArchiveClient(ctx, scope, issued.Client.ID))

		pgtesting.AssertAuditLogContainsForUser(t, ctx, db, owner.ID, []pgtesting.ExpectedAuditEntry{
			{EventType: platformaudit.EventCreated, ResourceType: resourceTypeOAuth2Clients, ResourceID: issued.Client.ID},
			{EventType: platformaudit.EventUpdated, ResourceType: resourceTypeOAuth2Clients, ResourceID: issued.Client.ID},
			{EventType: platformaudit.EventArchived, ResourceType: resourceTypeOAuth2Clients, ResourceID: issued.Client.ID},
		})

		// The revision's entry says what moved, from the row platform read before
		// overwriting it.
		var revision *platformaudit.Entry
		for _, entry := range pgtesting.AuditEntriesForActor(t, ctx, db, owner.ID) {
			if entry.EventType == platformaudit.EventUpdated && entry.ResourceID == issued.Client.ID {
				revision = entry
			}
		}
		require.NotNil(t, revision)
		require.Contains(t, revision.Changes, "name")
		assert.Equal(t, "a client", revision.Changes["name"].Old)
		assert.Equal(t, "renamed", revision.Changes["name"].New)
	})
}
