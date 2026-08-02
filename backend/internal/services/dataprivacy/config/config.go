package config

import (
	"context"

	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	uploadscfg "github.com/primandproper/platform-go/v9/uploads/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// Config configures the service.
//
// The same struct configures three processes, because three of them touch disclosure artifacts:
// the API server reads them, the async message handler writes them, and the scheduler destroys
// them. They must agree on the bucket and the cipher or the artifact written by one is
// unreadable to the next.
type Config struct {
	_ struct{} `json:"-"`

	// Encryption selects the cipher used for at-rest encryption of disclosure artifacts. A
	// disclosure artifact is everything the system knows about one person in a single object,
	// so it is never written in the clear.
	Encryption encryptioncfg.Config `envPrefix:"ENCRYPTION_" json:"encryption"`

	// ArtifactEncryptionKey is the key disclosure artifacts are encrypted with. Rotating it
	// makes every artifact written under the old key unreadable — which for objects that expire
	// in a week is a survivable way to revoke them, but is not a decision to make by accident.
	//
	// It is not validated as required here, the same way the OAuth2 token key is not: a rendered
	// config for a real environment carries a blank secret and takes the value from the
	// environment. An empty key is caught where it matters instead, when the artifact store is
	// constructed at startup.
	ArtifactEncryptionKey string `env:"ARTIFACT_ENCRYPTION_KEY" json:"artifactEncryptionKey"`

	Uploads uploadscfg.Config `envPrefix:"UPLOADS_" json:"uploads"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config struct.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return validation.ValidateStructWithContext(
		ctx,
		cfg,
		validation.Field(&cfg.Uploads, validation.Required),
		validation.Field(&cfg.Encryption, validation.Required),
	)
}
