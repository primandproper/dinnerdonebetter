// Package dbcfg is the platform's database configuration plus the fields this
// application keeps alongside it.
//
// There are none left. It held Encryption and OAuth2TokenEncryptionKey, which
// configured the at-rest encryption of the oauth2_client_tokens columns — a
// table that no longer exists: the authorization server's records moved to the
// platform's own, which store a SHA-256 digest of each credential rather than a
// reversible copy, so there is nothing to key.
//
// The wrapper stays because the whole application names this type, and because
// the next application-shaped database setting has somewhere obvious to go.
package dbcfg

import (
	databasecfg "github.com/primandproper/platform-go/v13/database/config"
)

// Config is the database configuration.
type Config struct {
	databasecfg.Config
}
