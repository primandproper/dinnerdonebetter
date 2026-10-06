/*
Package config configures this application's half of platform-go's data privacy
machinery: the one key export artifacts are sealed under, and the pins that keep
platform's block pointed at this application's tables.

Everything else is platform's own dataprivacy/config block — where artifacts are
stored, what they are encrypted with, and the timings of the request state machine
that produces them — and platform builds from it: the artifact storage through
RegisterArtifactStorage, the store, the fulfiller, the service and the sweep.

The same values configure two processes, because two of them touch artifacts: the
API server reads them back for the subject, and the scheduler writes and expires
them. They must agree on the bucket, the cipher, and the table prefix, or the
artifact written by one is unreadable to the next and the sweep meant to destroy it
deletes nothing and reports success.
*/
package config

import (
	"context"

	platformdataprivacycfg "github.com/primandproper/platform-go/v15/dataprivacy/config"
	"github.com/primandproper/primitives-go/v2/compression"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

// CompressionAlgorithm is what export artifacts are compressed with before they are sealed.
const CompressionAlgorithm = compression.AlgorithmZstd

// Config is the API server's data privacy configuration. The scheduler carries the same two
// values as its service.Config's DataPrivacy block and a key of its own field.
type Config struct {
	_ struct{} `json:"-"`

	// ArtifactEncryptionKey is the key artifacts are sealed under, filed under
	// Platform.Artifacts.Encryption's CurrentKeyID. platform builds the keyring over an
	// encryption.Keyset the container supplies rather than from configuration, so this is
	// where the one key this deployment has comes from — see RegisterKeyset.
	ArtifactEncryptionKey string `env:"ARTIFACT_ENCRYPTION_KEY" json:"artifactEncryptionKey,omitempty"`

	// Platform is platform's own block. Its Artifacts half names the bucket artifacts are
	// written to and the keyring they are sealed with; its table prefix and dialect are
	// pinned in code by Pin rather than read from here.
	Platform platformdataprivacycfg.Config `envPrefix:"PLATFORM_" json:"platform,omitzero"`
}

var _ validation.ValidatableWithContext = (*Config)(nil)

// ValidateWithContext validates a Config.
//
// Artifacts is required whole. A deployment that left out the storage would write exports into
// the shared media bucket, and one that left out the encryption would write them there in the
// clear; platform reads both absences as choices, and this application has made neither.
func (cfg *Config) ValidateWithContext(ctx context.Context) error {
	return ValidatePlatform(ctx, &cfg.Platform)
}

// ValidatePlatform validates a platform data privacy block the way this application runs it:
// artifacts stored in a bucket of their own and sealed under a keyring of their own.
//
// The key itself is not checked. It is a secret, so a rendered config never carries one — it
// arrives from the environment at startup — and RegisterKeyset refuses to build without it.
func ValidatePlatform(ctx context.Context, cfg *platformdataprivacycfg.Config) error {
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return err
	}

	return validation.Errors{
		"Artifacts": validation.Validate(cfg.Artifacts, validation.Required),
		"Artifacts.Storage": validation.Validate(cfg.Artifacts, validation.When(cfg.Artifacts != nil, validation.By(func(any) error {
			return validation.Validate(cfg.Artifacts.Storage, validation.Required)
		}))),
		"Artifacts.Encryption": validation.Validate(cfg.Artifacts, validation.When(cfg.Artifacts != nil, validation.By(func(any) error {
			return validation.Validate(cfg.Artifacts.Encryption, validation.Required)
		}))),
	}.Filter()
}
