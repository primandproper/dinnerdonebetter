package repositories

import (
	postgresmigrations "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/repositories/postgres/migrations"

	"github.com/primandproper/platform-go/v4/database"
	databasecfg "github.com/primandproper/platform-go/v4/database/config"
	"github.com/primandproper/platform-go/v4/observability/logging"
)

// ProvideMigrator returns a Migrator appropriate for the configured database provider.
func ProvideMigrator(
	cfg *databasecfg.Config,
	logger logging.Logger,
) database.Migrator {
	switch cfg.Provider {
	case databasecfg.ProviderPostgres:
		return postgresmigrations.NewMigrator(logger)
	default:
		return nil
	}
}
