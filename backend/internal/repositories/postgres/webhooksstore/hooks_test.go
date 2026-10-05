package webhooksstore

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/auditlogentries"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"
	pgtesting "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/testing"

	platformaudit "github.com/primandproper/platform-go/v15/audit"
	platformwebhooks "github.com/primandproper/platform-go/v15/webhooks"
	webhookscfg "github.com/primandproper/platform-go/v15/webhooks/config"
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

// inTx runs one store write in a transaction of its own, the way a handler's would.
func inTx[T any](t *testing.T, db database.Client, write func(tx database.Tx) (T, error)) T {
	t.Helper()

	var out T

	require.NoError(t, db.WithTransaction(t.Context(), func(tx database.Tx) error {
		var err error
		out, err = write(tx)

		return err
	}))

	return out
}

func TestHooks_Integration(T *testing.T) {
	T.Parallel()

	T.Run("every endpoint and subscription write lands its audit entry", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		// Already migrated: the template this was cloned from was migrated once in TestMain.
		_, config := pgtesting.NewIsolatedDatabaseForTest(t)

		db, err := postgres.NewDatabaseClient(ctx, config, postgres.WithLogger(loggingnoop.NewLogger()), postgres.WithTracerProvider(tracingnoop.NewTracerProvider()))
		require.NoError(t, err)

		auditRepo, err := auditlogentries.ProvideAuditLogRepository(loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), metricsnoop.NewMetricsProvider(), db)
		require.NoError(t, err)

		store, err := ProvideStore(ctx, &webhookscfg.Config{}, db, loggingnoop.NewLogger(), tracingnoop.NewTracerProvider(), auditRepo, nil)
		require.NoError(t, err)

		// The entries are filed under the endpoint's account, and the audit chain
		// requires that account to exist.
		owner := pgtesting.CreateUserForTest(t, nil, db.Writer())
		account := pgtesting.CreateAccountForTest(t, nil, owner.ID, db.Writer())
		scope := tenancy.Of(account.ID)

		endpoint := &platformwebhooks.Endpoint{
			ID:            "endpoint-1",
			Name:          "orders",
			URL:           "https://93.184.216.34/hooks",
			ContentType:   platformwebhooks.DefaultContentType,
			Secret:        platformwebhooks.Secret{Current: []byte("current")},
			Subscriptions: platformwebhooks.SubscribeTo("webhook_created"),
		}

		created := inTx(t, db, func(tx database.Tx) (*platformwebhooks.Endpoint, error) {
			return store.SaveEndpoint(ctx, tx, scope, endpoint)
		})
		require.True(t, created.Created)

		endpoint.Name = "renamed"
		inTx(t, db, func(tx database.Tx) (*platformwebhooks.Endpoint, error) {
			return store.SaveEndpoint(ctx, tx, scope, endpoint)
		})

		inTx(t, db, func(tx database.Tx) (struct{}, error) {
			return struct{}{}, store.RotateSecret(ctx, tx, scope, endpoint.ID, []byte("next"))
		})

		subscription := inTx(t, db, func(tx database.Tx) (*platformwebhooks.Subscription, error) {
			return store.AddSubscription(ctx, tx, scope, endpoint.ID, "webhook_archived")
		})

		inTx(t, db, func(tx database.Tx) (*platformwebhooks.Subscription, error) {
			return store.ArchiveSubscription(ctx, tx, scope, subscription.ID)
		})

		inTx(t, db, func(tx database.Tx) (*platformwebhooks.Endpoint, error) {
			return store.ArchiveEndpoint(ctx, tx, scope, endpoint.ID)
		})

		pgtesting.AssertAuditLogContains(t, ctx, db, account.ID, []pgtesting.ExpectedAuditEntry{
			{EventType: platformaudit.EventCreated, ResourceType: resourceTypeWebhooks, ResourceID: endpoint.ID},
			{EventType: platformaudit.EventUpdated, ResourceType: resourceTypeWebhooks, ResourceID: endpoint.ID},
			{EventType: platformaudit.EventCreated, ResourceType: resourceTypeWebhookTriggerConfigs, ResourceID: subscription.ID},
			{EventType: platformaudit.EventArchived, ResourceType: resourceTypeWebhookTriggerConfigs, ResourceID: subscription.ID},
			{EventType: platformaudit.EventArchived, ResourceType: resourceTypeWebhooks, ResourceID: endpoint.ID},
		})

		// The save and the rotation are both updates to the endpoint; the save is the
		// one that carries what changed, and no entry carries a key.
		var updates []*platformaudit.Entry
		for _, entry := range pgtesting.AuditEntriesForAccount(t, ctx, db, account.ID) {
			if entry.ResourceType == resourceTypeWebhooks && entry.EventType == platformaudit.EventUpdated {
				updates = append(updates, entry)
			}
		}
		require.Len(t, updates, 2)

		var diffed *platformaudit.Entry
		for _, entry := range updates {
			if len(entry.Changes) > 0 {
				diffed = entry
			}
		}
		require.NotNil(t, diffed)
		assert.Contains(t, diffed.Changes, "name")
		assert.NotContains(t, diffed.Changes, "secret")
		assert.NotContains(t, diffed.Changes, "Secret")
	})
}
