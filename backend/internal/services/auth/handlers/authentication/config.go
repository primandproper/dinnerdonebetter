package authentication

import (
	"context"
	"time"

	tokenscfg "github.com/primandproper/platform-go/v7/authentication/tokens/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config is our configuration.
	Config struct {
		_ struct{} `json:"-"`

		Tokens                tokenscfg.Config `envPrefix:"TOKENS_"           json:"tokens"`
		OAuth2                OAuth2Config     `envPrefix:"OAUTH2"            json:"oauth2"`
		TokenLifetime         time.Duration    `env:"JWT_LIFETIME"            json:"jwtLifetime"`
		Debug                 bool             `env:"DEBUG"                   json:"debug,omitempty"`
		EnableUserSignup      bool             `env:"ENABLE_USER_SIGNUP"      json:"enableUserSignup,omitempty"`
		MinimumUsernameLength uint8            `env:"MINIMUM_USERNAME_LENGTH" json:"minimumUsernameLength,omitempty"`
		MinimumPasswordLength uint8            `env:"MINIMUM_PASSWORD_LENGTH" json:"minimumPasswordLength,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
		validation.Field(&cfg.Tokens, validation.Required),
	)
}

// OAuth2Config represents our database configuration.
type OAuth2Config struct {
	_ struct{} `json:"-"`

	Domain               string        `env:"DOMAIN"                 json:"domain"`
	AccessTokenLifespan  time.Duration `env:"ACCESS_TOKEN_LIFESPAN"  json:"accessTokenLifespan"`
	RefreshTokenLifespan time.Duration `env:"REFRESH_TOKEN_LIFESPAN" json:"refreshTokenLifespan"`
	Debug                bool          `env:"DEBUG"                  json:"debug"`
}

var _ validation.ValidatableWithContext = (*OAuth2Config)(nil)

// ValidateWithContext validates a OAuth2Config struct.
func (cfg OAuth2Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, &cfg,
		validation.Field(&cfg.AccessTokenLifespan, validation.Required),
		validation.Field(&cfg.RefreshTokenLifespan, validation.Required),
		validation.Field(&cfg.Domain, validation.Required),
	)
}
