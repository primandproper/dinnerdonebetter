package config

import (
	capitalismcfg "github.com/primandproper/platform-go/v13/capitalism/config"
)

// Config holds payments service configuration.
//
// Capitalism is platform-go's own config, and carries both providers' credentials: Stripe's
// for the web checkout endpoint and RevenueCat's for the mobile store one. Neither provider is
// modeled here any more — capitalism gained RevenueCat in v13 — so the only thing this struct
// adds to it is the second selector.
type Config struct {
	Capitalism capitalismcfg.Config `envPrefix:"CAPITALISM_" json:"capitalism,omitzero"`

	// MobileProvider selects the processor mounted at the mobile store webhook endpoint, from
	// the same vocabulary Capitalism.Provider uses for the web one: `revenuecat` builds the
	// real thing, `noop` builds the stub, and anything else — an unset value included — is an
	// error.
	//
	// It is a second selector rather than a second capitalismcfg.Config because capitalism's
	// config names one provider and this service takes webhooks from two. Running one adapter
	// per endpoint is the shape capitalism/revenuecat's package doc prescribes, over a merged
	// manager that would have to pretend one provider could answer for the other; a second
	// whole config would have brought a second set of every credential with it, three quarters
	// of them inert.
	MobileProvider string `env:"MOBILE_PROVIDER" json:"mobileProvider,omitempty"`
}
