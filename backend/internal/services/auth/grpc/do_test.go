package grpc

import (
	"testing"
	"time"

	authcfg "github.com/primandproper/dinnerdonebetter/backend/internal/authentication/config"
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/config"

	platformwebauthn "github.com/primandproper/platform-go/v12/authentication/webauthn"
	webauthncfg "github.com/primandproper/platform-go/v12/authentication/webauthn/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProvidePasskeyConfig(T *testing.T) {
	T.Parallel()

	T.Run("standard", func(t *testing.T) {
		t.Parallel()

		configured := webauthncfg.Config{
			Provider: webauthncfg.ProviderDatabase,
			RelyingParty: platformwebauthn.Config{
				RPID:            "dinnerdonebetter.com",
				RPDisplayName:   "Somewhere Else",
				RPOrigins:       []string{"https://dinnerdonebetter.com"},
				CeremonyTimeout: 2 * time.Minute,
			},
		}

		actual := ProvidePasskeyConfig(&config.APIServiceConfig{Auth: authcfg.Config{Passkey: configured}})

		require.NotNil(t, actual)
		assert.Equal(t, configured, *actual)
	})

	// Local dev renders no relying party, and a ceremony cannot be begun without one.
	T.Run("with no relying party configured", func(t *testing.T) {
		t.Parallel()

		actual := ProvidePasskeyConfig(&config.APIServiceConfig{})

		require.NotNil(t, actual)
		assert.Equal(t, "localhost", actual.RelyingParty.RPID)
		assert.Equal(t, branding.CompanyName, actual.RelyingParty.RPDisplayName)
		assert.NotEmpty(t, actual.RelyingParty.RPOrigins)
	})

	// The defaulting fills in the relying party and nothing else. An omitted provider stays
	// omitted, so it resolves to the platform's default — the table — rather than to a
	// per-process store that works on one replica.
	T.Run("does not default the ceremony store to a process-local one", func(t *testing.T) {
		t.Parallel()

		actual := ProvidePasskeyConfig(&config.APIServiceConfig{})
		require.NotNil(t, actual)

		actual.EnsureDefaults()
		assert.Equal(t, webauthncfg.ProviderDatabase, actual.Provider)
		assert.Positive(t, actual.SweepInterval)
	})

	// The fallback must not smuggle in a config the relying party then refuses to be built
	// from, which would turn a missing config block into a failure at the first ceremony.
	T.Run("produces a valid config in either case", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()

		actual := ProvidePasskeyConfig(&config.APIServiceConfig{})
		require.NotNil(t, actual)

		actual.EnsureDefaults()
		assert.NoError(t, actual.ValidateWithContext(ctx))
	})
}
