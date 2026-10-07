package authentication

import (
	"context"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"

	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	ratelimitingcfg "github.com/primandproper/primitives-go/v2/ratelimiting/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// Config is our configuration.
	Config struct {
		_ struct{} `json:"-"`

		Tokens authcfg.TokensConfig `envPrefix:"TOKENS_" json:"tokens,omitzero"`
		// RateLimiting throttles the doors a caller reaches without a credential: signing in,
		// signing up, and asking for a mail. Each door has a budget of its own per address.
		RateLimiting ratelimitingcfg.Config `envPrefix:"RATE_LIMITING_" json:"rateLimiting,omitzero"`
		OAuth2       oauth2servercfg.Config `envPrefix:"OAUTH2_"        json:"oauth2,omitzero"`
		Debug        bool                   `env:"DEBUG"                json:"debug,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.Tokens, validation.Required),
		// Called explicitly rather than left to ozzo's nested-struct handling: the platform's
		// Config implements ValidatableWithContext on its pointer receiver, and ozzo hands
		// the dereferenced value to that check — so the nested rules would be silently
		// skipped, and an unusable authorization server config would validate clean.
		validation.Field(&cfg.OAuth2, validation.By(func(any) error { return cfg.OAuth2.ValidateWithContext(ctx) })),
		// The same, for the same reason.
		validation.Field(&cfg.RateLimiting, validation.By(func(any) error { return cfg.RateLimiting.ValidateWithContext(ctx) })),
	)
}
