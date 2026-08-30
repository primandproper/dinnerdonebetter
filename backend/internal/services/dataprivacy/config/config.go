/*
Package config configures this application's half of platform-go's data privacy
machinery: where export artifacts are stored, what they are encrypted with, and
the timings of the request state machine that produces them.

The same struct configures two processes, because two of them touch artifacts:
the API server reads them back for the subject, and the scheduler writes and
expires them. They must agree on the bucket, the cipher, and the table prefix, or
the artifact written by one is unreadable to the next and the sweep meant to
destroy it deletes nothing and reports success.
*/
package config

import (
	"context"

	"github.com/primandproper/platform-go/v13/compression"
	encryptioncfg "github.com/primandproper/platform-go/v13/cryptography/encryption/config"
	platformdataprivacycfg "github.com/primandproper/platform-go/v13/dataprivacy/config"
	uploadscfg "github.com/primandproper/platform-go/v13/uploads/config"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// CompressionAlgorithm is what export artifacts are compressed with before they
// are encrypted.
//
// Zstandard rather than S2: an artifact is written once and read at most once,
// so the ratio is worth more than the decompression speed, and what a bucket
// bills for is bytes. It is a constant rather than a knob because the Worker
// that writes and the Service that reads must agree, and the only way to
// discover that they do not is a subject opening a file that will not
// decompress.
const CompressionAlgorithm = compression.AlgorithmZstd

// Config configures artifact storage, artifact encryption, and the request
// lifecycle.
type Config struct {
	_ struct{} `json:"-"`

	// Encryption selects the cipher used for at-rest encryption of disclosure artifacts. A
	// disclosure artifact is everything the system knows about one person in a single object,
	// so it is never written in the clear.
	Encryption encryptioncfg.Config `envPrefix:"ENCRYPTION_" json:"encryption,omitzero"`

	// ArtifactEncryptionKey is the key disclosure artifacts are encrypted with. Rotating it
	// makes every artifact written under the old key unreadable — which for objects that expire
	// in a week is a survivable way to revoke them, but is not a decision to make by accident.
	//
	// It is not validated as required here, the same way the OAuth2 token key is not: a rendered
	// config for a real environment carries a blank secret and takes the value from the
	// environment. An empty key is caught where it matters instead, when the machinery is
	// constructed at startup.
	ArtifactEncryptionKey string `env:"ARTIFACT_ENCRYPTION_KEY" json:"artifactEncryptionKey,omitempty"`

	Uploads uploadscfg.Config `envPrefix:"UPLOADS_" json:"uploads,omitzero"`

	// Requests carries the platform's own knobs: the response windows a deadline is
	// stamped from, the confirmation window, the artifact TTL, and the fulfillment
	// loop's timings.
	//
	// Dialect and TablePrefix are filled in rather than read from the environment —
	// see prepare in do.go. Neither has a second legal value, and a deployment that
	// set one differently would not be configuring anything, it would be pointing the
	// Store at a table that does not exist.
	Requests platformdataprivacycfg.Config `envPrefix:"REQUESTS_" json:"requests,omitzero"`
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
