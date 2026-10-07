package authentication

import (
	"testing"

	oauth2servercfg "github.com/primandproper/platform-go/v15/authentication/oauth2serverstore/config"
	ratelimitingcfg "github.com/primandproper/primitives-go/v2/ratelimiting/config"

	"github.com/stretchr/testify/assert"
)

func TestConfig_Validate(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		cfg := &Config{
			Debug:                 false,
			EnableUserSignup:      false,
			MinimumUsernameLength: 123,
			MinimumPasswordLength: 123,
			OAuth2: oauth2servercfg.Config{
				Provider: oauth2servercfg.ProviderMemory,
				Issuer:   "https://example.com",
			},
			RateLimiting: ratelimitingcfg.Config{Provider: ratelimitingcfg.ProviderMemory},
		}

		assert.NoError(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with no rate limiter", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		// The anonymous doors are throttled in every deployment, so a config that names no
		// limiter is one that forgot, and it fails here rather than at the first sign-in.
		cfg := &Config{
			MinimumUsernameLength: 123,
			MinimumPasswordLength: 123,
			OAuth2: oauth2servercfg.Config{
				Provider: oauth2servercfg.ProviderMemory,
				Issuer:   "https://example.com",
			},
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})

	T.Run("with an unknown oauth2 provider", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		// Whether the issuer is a legal issuer is oauth2server.NewServer's answer rather than
		// the config's — see TestProvideOAuth2Server. What the config decides is the provider,
		// and an unrecognized one has to fail here: the alternative is a server that comes up
		// on neither store.
		cfg := &Config{
			Debug:                 false,
			EnableUserSignup:      false,
			MinimumUsernameLength: 123,
			MinimumPasswordLength: 123,
			OAuth2: oauth2servercfg.Config{
				Provider: "postgres",
				Issuer:   "https://example.com",
			},
			RateLimiting: ratelimitingcfg.Config{Provider: ratelimitingcfg.ProviderMemory},
		}

		assert.Error(t, cfg.ValidateWithContext(ctx))
	})
}
