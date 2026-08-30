package authcfg

import (
	tokenscfg "github.com/primandproper/platform-go/v13/authentication/tokens/config"

	"github.com/samber/do/v2"
)

// RegisterConfigs registers auth config sub-fields with the injector.
func RegisterConfigs(i do.Injector) {
	do.Provide[*TokensConfig](i, func(i do.Injector) (*TokensConfig, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Tokens, nil
	})

	do.Provide[*tokenscfg.Config](i, func(i do.Injector) (*tokenscfg.Config, error) {
		return &do.MustInvoke[*TokensConfig](i).Config, nil
	})

	do.Provide[*SessionsConfig](i, func(i do.Injector) (*SessionsConfig, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Sessions, nil
	})
}
