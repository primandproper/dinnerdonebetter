package mcpbuild

import (
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/config"

	databasecfg "github.com/primandproper/platform-go/v8/database/config"
	"github.com/primandproper/platform-go/v8/observability"
	routingcfg "github.com/primandproper/platform-go/v8/routing/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs extracts sub-fields from MCPServiceConfig into the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*databasecfg.Config](i, func(i do.Injector) (*databasecfg.Config, error) {
		cfg := do.MustInvoke[*config.MCPServiceConfig](i)
		return &cfg.Database, nil
	})
	do.Provide[*observability.Config](i, func(i do.Injector) (*observability.Config, error) {
		cfg := do.MustInvoke[*config.MCPServiceConfig](i)
		return &cfg.Observability, nil
	})
	do.Provide[config.MetaSettings](i, func(i do.Injector) (config.MetaSettings, error) {
		cfg := do.MustInvoke[*config.MCPServiceConfig](i)
		return cfg.Meta, nil
	})
	do.Provide[*routingcfg.Config](i, func(i do.Injector) (*routingcfg.Config, error) {
		cfg := do.MustInvoke[*config.MCPServiceConfig](i)
		return &cfg.Routing, nil
	})
}
