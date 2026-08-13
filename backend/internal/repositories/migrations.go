package repositories

import (
	postgresmigrations "github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/migrations"

	"github.com/primandproper/platform-go/v10/database"
	databasecfg "github.com/primandproper/platform-go/v10/database/config"
	"github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/observability/logging"
)

// ErrUnsupportedDatabaseProvider indicates the configured provider has no migrator.
// The migrations are postgres-specific SQL, so any other provider is a
// misconfiguration rather than a database that merely needs no schema — better to
// fail at startup than to serve traffic against an unmigrated database.
var ErrUnsupportedDatabaseProvider = errors.New("unsupported database provider for migrations")

// ProvideMigrator returns a Migrator appropriate for the configured database provider.
func ProvideMigrator(
	cfg *databasecfg.Config,
	logger logging.Logger,
) (database.Migrator, error) {
	switch cfg.Provider {
	case databasecfg.ProviderPostgres:
		return postgresmigrations.NewMigrator(logger)
	default:
		return nil, errors.Wrapf(ErrUnsupportedDatabaseProvider, "provider %q", cfg.Provider)
	}
}
