package dbcleaner

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers the one config sub-field platform has no block for in this process:
// the authorization server's table prefix — see config.DBCleanerConfig.
func RegisterConfigs(i do.Injector) {
	do.Provide[*oauth2servercfg.Config](i, func(i do.Injector) (*oauth2servercfg.Config, error) {
		return &do.MustInvoke[*config.DBCleanerConfig](i).OAuth2, nil
	})
}
