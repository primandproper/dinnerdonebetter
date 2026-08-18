package authcfg

import (
	"context"
	"time"

	tokenscfg "github.com/primandproper/platform-go/v11/authentication/tokens/config"
	webauthncfg "github.com/primandproper/platform-go/v11/authentication/webauthn/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// TokensConfig is the platform's token issuer configuration plus the token
	// lifetimes this application enforces.
	//
	// platform-go v9 dropped MaxAccessTokenLifetime and MaxRefreshTokenLifetime
	// because nothing in that module read them. The refresh flow in
	// internal/authentication does, so they live here. The embedding keeps the
	// JSON and env var shape identical to what v8 produced.
	TokensConfig struct {
		tokenscfg.Config

		MaxAccessTokenLifetime  time.Duration `env:"MAX_ACCESS_TOKEN_LIFETIME"  json:"maxAccessTokenLifetime,omitempty"`
		MaxRefreshTokenLifetime time.Duration `env:"MAX_REFRESH_TOKEN_LIFETIME" json:"maxRefreshTokenLifetime,omitempty"`
	}

	// Config is our configuration.
	Config struct {
		_ struct{} `json:"-"`

		Tokens TokensConfig `envPrefix:"TOKENS_" json:"tokens,omitzero"`

		// Passkey is the WebAuthn relying party and the ceremony store beneath it. Both
		// halves are the platform's, which is what collapses the three timeouts this
		// application used to keep separately — the row's TTL, the timeout asked of the
		// browser, and the deadline the library enforces — into RelyingParty.CeremonyTimeout.
		//
		// Its Provider defaults to the database rather than to memory. A ceremony spans two
		// requests and nothing pins them to a replica, so a per-process store fails a
		// fraction of passkey logins in a way that reads as a browser bug.
		Passkey webauthncfg.Config `envPrefix:"PASSKEY_" json:"passkey,omitzero"`

		Debug                 bool  `env:"DEBUG"                   json:"debug,omitempty"`
		EnableUserSignup      bool  `env:"ENABLE_USER_SIGNUP"      json:"enableUserSignup,omitempty"`
		MinimumUsernameLength uint8 `env:"MINIMUM_USERNAME_LENGTH" json:"minimumUsernameLength,omitempty"`
		MinimumPasswordLength uint8 `env:"MINIMUM_PASSWORD_LENGTH" json:"minimumPasswordLength,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
		validation.Field(&cfg.Tokens, validation.Required),
		validation.Field(&cfg.Passkey, validation.By(func(any) error {
			return cfg.Passkey.ValidateWithContext(ctx)
		})),
	)
}
