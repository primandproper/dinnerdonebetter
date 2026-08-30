package dbcleaner

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"
	dbcfg "github.com/primandproper/dinnerdonebetter/backend/internal/database/config"

	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
	"github.com/primandproper/platform-go/v13/observability"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.DBCleanerConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*dbcfg.Config](i, func(i do.Injector) (*dbcfg.Config, error) {
		cfg := do.MustInvoke[*config.DBCleanerConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		return &do.MustInvoke[*dbcfg.Config](i).Config, nil
	})
	do.Provide[*oauth2servercfg.Config](i, func(i do.Injector) (*oauth2servercfg.Config, error) {
		cfg := do.MustInvoke[*config.DBCleanerConfig](i)
		return &cfg.OAuth2, nil
	})
}
