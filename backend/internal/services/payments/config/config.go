package config

import (
	capitalismcfg "github.com/primandproper/platform-go/v11/capitalism/config"
)

// Config holds payments service configuration.
//
// Capitalism carries the web billing provider (Stripe today) and is the platform's own
// config, so the Stripe client we talk to is the one platform-go maintains and bumps.
// RevenueCat stays ours: mobile in-app purchases aren't something capitalism models, and
// its webhooks are a different shape entirely.
type Config struct {
	RevenueCat *RevenueCatConfig    `env:"init"              envPrefix:"REVENUECAT_"    json:"revenueCat,omitempty"`
	Capitalism capitalismcfg.Config `envPrefix:"CAPITALISM_" json:"capitalism,omitzero"`
}

// RevenueCatConfig holds RevenueCat-specific configuration.
type RevenueCatConfig struct {
	APIKey            string `env:"API_KEY"             json:"apiKey,omitempty"`
	WebhookAuthHeader string `env:"WEBHOOK_AUTH_HEADER" json:"webhookAuthHeader,omitempty"`
}
