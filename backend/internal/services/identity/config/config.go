package config

import (
	"context"

	uploadscfg "github.com/primandproper/platform-go/v5/uploads/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures the service.
type Config struct {
	_ struct{} `json:"-"`

	PublicMediaURLPrefix string            `env:"PUBLIC_MEDIA_URL_PREFIX" json:"publicMediaURLPrefix"`
	Uploads              uploadscfg.Config `envPrefix:"UPLOADS_"          json:"uploads"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Uploads, validation.Required),
	)
}
