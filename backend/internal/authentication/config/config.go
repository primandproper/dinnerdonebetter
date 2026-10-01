package authcfg

import (
	"context"

	webauthncfg "github.com/primandproper/platform-go/v14/authentication/webauthnsessions/config"
	tokenscfg "github.com/primandproper/primitives-go/v2/authentication/tokens/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type (
	// TokensConfig is the platform's token issuer configuration.
	//
	// The issuer signs every token this application hands out — SignInService's sign-ins and
	// the tokens the OAuth2 authorization server's login hands to it. How long one lives is
	// the sign-in service's to say, not the issuer's; see internal/authentication.
	TokensConfig struct {
		tokenscfg.Config
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
