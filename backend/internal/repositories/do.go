package repositories

import (
	"github.com/primandproper/platform-go/v9/database"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
	"github.com/primandproper/platform-go/v9/observability/logging"

	"github.com/samber/do/v2"
)

// RegisterMigrator registers the database migrator with the injector.
func RegisterMigrator(i do.Injector) {
	do.Provide[database.Migrator](i, func(i do.Injector) (database.Migrator, error) {
		return ProvideMigrator(
			do.MustInvoke[*databasecfg.Config](i),
			do.MustInvoke[logging.Logger](i),
		)
	})
}
