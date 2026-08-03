package authcfg

import (
	"context"
	"time"

	webauthncfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/webauthn/config"

	tokenscfg "github.com/primandproper/platform-go/v9/authentication/tokens/config"

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

	// PasskeyConfig holds WebAuthn Relying Party configuration for passkey registration and authentication.
	PasskeyConfig struct {
		_             struct{} `json:"-"`
		RPID          string   `env:"RP_ID"           json:"rpID,omitempty"`
		RPDisplayName string   `env:"RP_DISPLAY_NAME" json:"rpDisplayName,omitempty"`
		RPOrigins     []string `env:"RP_ORIGINS"      json:"rpOrigins,omitempty"`
	}

	// Config is our configuration.
	Config struct {
		_                     struct{}           `json:"-"`
		SessionStore          webauthncfg.Config `envPrefix:"SESSION_STORE_"    json:"sessionStore,omitzero"`
		Passkey               PasskeyConfig      `envPrefix:"PASSKEY_"          json:"passkey,omitzero"`
		Tokens                TokensConfig       `envPrefix:"TOKENS_"           json:"tokens,omitzero"`
		Debug                 bool               `env:"DEBUG"                   json:"debug,omitempty"`
		EnableUserSignup      bool               `env:"ENABLE_USER_SIGNUP"      json:"enableUserSignup,omitempty"`
		MinimumUsernameLength uint8              `env:"MINIMUM_USERNAME_LENGTH" json:"minimumUsernameLength,omitempty"`
		MinimumPasswordLength uint8              `env:"MINIMUM_PASSWORD_LENGTH" json:"minimumPasswordLength,omitempty"`
	}
)

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(ctx, cfg,
		validation.Field(&cfg.MinimumUsernameLength, validation.Required),
		validation.Field(&cfg.MinimumPasswordLength, validation.Required),
		validation.Field(&cfg.Tokens, validation.Required),
		validation.Field(&cfg.SessionStore, validation.By(func(value any) error {
			if c, ok := value.(webauthncfg.Config); ok {
				return (&c).ValidateWithContext(ctx)
			}
			return nil
		})),
	)
}
