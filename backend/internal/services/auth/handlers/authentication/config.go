package authentication

import (
	"context"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"

	oauth2servercfg "github.com/primandproper/platform-go/v13/authentication/oauth2server/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config is our configuration.
	Config struct {
		_ struct{} `json:"-"`

		Tokens                authcfg.TokensConfig   `envPrefix:"TOKENS_"           json:"tokens,omitzero"`
		OAuth2                oauth2servercfg.Config `envPrefix:"OAUTH2_"           json:"oauth2,omitzero"`
		TokenLifetime         time.Duration          `env:"JWT_LIFETIME"            json:"jwtLifetime,omitempty"`
		Debug                 bool                   `env:"DEBUG"                   json:"debug,omitempty"`
		EnableUserSignup      bool                   `env:"ENABLE_USER_SIGNUP"      json:"enableUserSignup,omitempty"`
		MinimumUsernameLength uint8                  `env:"MINIMUM_USERNAME_LENGTH" json:"minimumUsernameLength,omitempty"`
		MinimumPasswordLength uint8                  `env:"MINIMUM_PASSWORD_LENGTH" json:"minimumPasswordLength,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
		validation.Field(&cfg.Tokens, validation.Required),
		// Called explicitly rather than left to ozzo's nested-struct handling: the platform's
		// Config implements ValidatableWithContext on its pointer receiver, and ozzo hands
		// the dereferenced value to that check — so the nested rules would be silently
		// skipped, and an unusable authorization server config would validate clean.
		validation.Field(&cfg.OAuth2, validation.By(func(any) error { return cfg.OAuth2.ValidateWithContext(ctx) })),
	)
}
