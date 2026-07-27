package authcfg

import (
	tokenscfg "github.com/primandproper/platform-go/v7/authentication/tokens/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers auth config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*tokenscfg.Config](i, func(i do.Injector) (*tokenscfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Tokens, nil
	})
}
