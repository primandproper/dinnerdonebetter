package migrations

import (
	"context"
	"database/sql"
	"embed"
	"fmt"

	"github.com/primandproper/platform-go/v5/database"
	"github.com/primandproper/platform-go/v5/observability/logging"

	"github.com/GuiaBolso/darwin"
)

var (
	//go:embed migration_files/*.sql
	rawMigrations embed.FS
)

func fetchMigration(name string) string {
	file, err := rawMigrations.ReadFile(fmt.Sprintf("migration_files/%s.sql", name))
	if err != nil {
		panic(err)
	}

	return string(file)
}

// Migrator implements database.Migrator for postgres databases.
type Migrator struct {
	logger logging.Logger
}

var _ database.Migrator = (*Migrator)(nil)

// NewMigrator creates a new postgres Migrator.
func NewMigrator(logger logging.Logger) *Migrator {
	return &Migrator{
		logger: logging.EnsureLogger(logger),
	}
}

// Migrate runs all postgres migrations on the given database connection.
func (m *Migrator) Migrate(ctx context.Context, db *sql.DB) error {
	migrations := []darwin.Migration{
		{Version: 1, Description: "identity tables", Script: fetchMigration("00001_identity")},
		{Version: 2, Description: "audit log entries tables", Script: fetchMigration("00002_auditlogentries")},
		{Version: 3, Description: "auth tables", Script: fetchMigration("00003_auth")},
		{Version: 4, Description: "oauth tables", Script: fetchMigration("00004_oauth")},
		{Version: 5, Description: "settings tables", Script: fetchMigration("00005_settings")},
		{Version: 6, Description: "user notifications table", Script: fetchMigration("00006_notifications")},
		{Version: 7, Description: "webhooks tables", Script: fetchMigration("00007_webhooks")},
		{Version: 8, Description: "waitlist tables", Script: fetchMigration("00008_waitlists")},
		{Version: 9, Description: "issue reports table", Script: fetchMigration("00009_issue_reports")},
		{Version: 10, Description: "uploaded media table", Script: fetchMigration("00010_uploaded_media")},
		{Version: 11, Description: "payments tables", Script: fetchMigration("00011_payments")},
		{Version: 12, Description: "comments table", Script: fetchMigration("00012_comments")},
		{Version: 13, Description: "data privacy tables", Script: fetchMigration("00013_dataprivacy")},
		{Version: 14, Description: "queue test messages tables", Script: fetchMigration("00014_internalops")},
		{Version: 15, Description: "user device tokens table", Script: fetchMigration("00015_user_device_tokens")},
		{Version: 16, Description: "webauthn credentials table", Script: fetchMigration("00016_webauthn_credentials")},
		{Version: 17, Description: "webauthn sessions table", Script: fetchMigration("00017_webauthn_sessions")},
		{Version: 18, Description: "user sessions table", Script: fetchMigration("00018_user_sessions")},
		{Version: 19, Description: "rbac tables", Script: fetchMigration("00019_rbac")},
		{Version: 20, Description: "per-service database users", Script: fetchMigration("00020_service_users")},
		{Version: 21, Description: "meal planning tables", Script: fetchMigration("00021_mealplanning")},
		{Version: 22, Description: "run meal plan workers permission", Script: fetchMigration("00022_run_meal_plan_workers_permission")},
		{Version: 23, Description: "nullable prep task reference for thaw tasks", Script: fetchMigration("00023_nullable_prep_task_for_thaw_tasks")},
		{Version: 24, Description: "clone recipes permission", Script: fetchMigration("00024_clone_recipes_permission")},
		{Version: 25, Description: "hash stored oauth2 client secrets", Script: fetchMigration("00025_hash_oauth2_client_secrets")},
	}

	if err := darwin.New(darwin.NewGenericDriver(db, darwin.PostgresDialect{}), migrations, nil).Migrate(); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	m.logger.Info("migrations completed successfully")
	return nil
}
