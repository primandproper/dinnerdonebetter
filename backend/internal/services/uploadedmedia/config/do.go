package config

import (
	uploadscfg "github.com/primandproper/platform-go/v12/uploads/config"

	"github.com/samber/do/v2"
)

// RegisterUploadedMediaConfig registers the uploaded media config fields with the injector.
func RegisterUploadedMediaConfig(i do.Injector) {
	do.Provide[*uploadscfg.Config](i, func(i do.Injector) (*uploadscfg.Config, error) {
		cfg := do.MustInvoke[*Config](i)
		return &cfg.Uploads, nil
	})
}
