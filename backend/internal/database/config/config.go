// Package dbcfg is the platform's database configuration plus the fields this
// application keeps alongside it.
//
// platform-go v9 removed Encryption and OAuth2TokenEncryptionKey from its own
// database config: nothing in that module consumed either. This application's
// OAuth2 client repository does, so they are declared here. The embedding keeps
// the JSON and env var shape identical to what v8 produced, including the
// promoted ValidateWithContext.
package dbcfg

import (
	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	databasecfg "github.com/primandproper/platform-go/v9/database/config"
)

// Config is the database configuration.
type Config struct {
	// Encryption selects the cipher used for at-rest encryption of OAuth2 tokens.
	Encryption encryptioncfg.Config `envPrefix:"ENCRYPTION_" json:"encryption,omitzero"`

	// OAuth2TokenEncryptionKey is the key OAuth2 client tokens are encrypted with.
	OAuth2TokenEncryptionKey string `env:"OAUTH2_TOKEN_ENCRYPTION_KEY" json:"oauth2TokenEncryptionKey,omitempty"`
	databasecfg.Config
}
