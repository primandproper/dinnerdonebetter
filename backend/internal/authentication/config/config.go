package authcfg

import (
	"context"
	"time"

	tokenscfg "github.com/primandproper/platform-go/v13/authentication/tokens/config"
	webauthncfg "github.com/primandproper/platform-go/v13/authentication/webauthn/config"

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

	// SessionsConfig is the expiry policy the user session store enforces.
	//
	// It is not platform-go's sessionscfg.Config, and the two things missing from it are
	// the reason. That type chooses where sessions live, and this application cannot take
	// either position: "sign out my other devices" needs an index on the holder, which
	// only sessions/database has — sessions/cache answers ErrNoPrincipalIndex from all
	// three of List, Revoke, and RevokeAllExcept — so a provider knob here would be a knob
	// whose other setting turns three RPCs into errors.
	//
	// And it cannot express the sweeper this deployment wants, which is none: the session
	// table is swept by `ddb job db-cleaner`, one pass for the fleet, for the same reason
	// the authorization server's and the password reset store's tables are. Its
	// SweepInterval field documents a non-positive value as "no sweeper", but
	// EnsureDefaults rewrites zero to five minutes before the value reaches WithSweeper,
	// so through that config every replica sweeps — see platform-go#456. Building the
	// store directly is how the interval stays absent rather than defaulted.
	//
	// What is left after removing both is the expiry policy, which is these three.
	SessionsConfig struct {
		_ struct{} `json:"-"`

		// AbsoluteTimeout bounds a session's total lifetime from the moment it was
		// established. Nothing extends it — not activity, not a token refresh — which
		// is what makes it the only bound on a stolen refresh token.
		AbsoluteTimeout time.Duration `env:"ABSOLUTE_TIMEOUT" json:"absoluteTimeout,omitempty"`

		// IdleTimeout bounds how long a session may go unauthenticated-against. A
		// request bearing the session's access token is what counts as a read.
		IdleTimeout time.Duration `env:"IDLE_TIMEOUT" json:"idleTimeout,omitempty"`

		// TouchInterval is how much of the idle window has to elapse before a read
		// bothers refreshing the idle deadline. Zero writes on every authenticated
		// request; see sessions.Policy for why that is not the default.
		TouchInterval time.Duration `env:"TOUCH_INTERVAL" json:"touchInterval,omitempty"`
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

		// Sessions is the server-side session store's expiry policy. Zero values take
		// platform-go's defaults, which are the ones sessions.Policy documents.
		Sessions SessionsConfig `envPrefix:"SESSIONS_" json:"sessions,omitzero"`

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
