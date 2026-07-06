package dbcleaner

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	databasecfg "github.com/primandproper/platform-go/v3/database/config"
	"github.com/primandproper/platform-go/v3/observability"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers all config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.DBCleanerConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.DBCleanerConfig](i)
		return &cfg.Database, nil
	})
}
