package config

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformdataprivacycfg "github.com/primandproper/platform-go/v15/dataprivacy/config"
	"github.com/primandproper/primitives-go/v2/compression"
	"github.com/primandproper/primitives-go/v2/cryptography/encryption"
	"github.com/primandproper/primitives-go/v2/database/dialect"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"

	"github.com/samber/do/v2"
)

// Pin overwrites the two settings in platform's block that are not this deployment's to choose:
// the table prefix, which has to be the one the migrations rendered the tables under, and the
// dialect, which is Postgres because the migrations are. See docs/configuration.md for why a
// prefix is never read from configuration here.
//
// It writes through cfg, so it is called on the block a process is about to register, before
// anything has read it.
func Pin(cfg *platformdataprivacycfg.Config) {
	if cfg == nil {
		return
	}

	cfg.TablePrefix = branding.TablePrefix
	cfg.Dialect = dialect.Postgres
	cfg.AuditErasure.TablePrefix = branding.TablePrefix
}

// RegisterKeyset registers the encryption.Keyset platform's RegisterArtifactStorage builds the
// artifact keyring over: this deployment's one key, under the CurrentKeyID the block names.
//
// One key is all there has ever been. A rotation would add the next one here beside it rather
// than replace it, because an artifact sealed before the rotation is otherwise unreadable — and
// found to be so by the subject who asked for it.
//
// Prerequisites: *platformdataprivacycfg.Config.
func RegisterKeyset(i do.Injector, key string) {
	do.Provide(i, func(i do.Injector) (encryption.Keyset, error) {
		if key == "" {
			return nil, platformerrors.New("no data privacy artifact encryption key provided")
		}

		cfg, err := do.Invoke[*platformdataprivacycfg.Config](i)
		if err != nil {
			return nil, err
		}

		if cfg.Artifacts == nil || cfg.Artifacts.Encryption == nil {
			return nil, platformerrors.New("data privacy artifacts name no keyring to file the key under")
		}

		return encryption.Keyset{
			encryption.KeyID(cfg.Artifacts.Encryption.CurrentKeyID): encryption.MasterKey(key),
		}, nil
	})
}

// RegisterCompressor registers the compression.Compressor platform's fulfiller writes artifacts
// with and its service reads them back with. Both resolve the one registration, so the codec an
// artifact was written with is by construction the one it is read with.
func RegisterCompressor(i do.Injector) {
	do.Provide(i, func(do.Injector) (compression.Compressor, error) {
		compressor, err := compression.NewCompressor(CompressionAlgorithm)
		if err != nil {
			return nil, platformerrors.Wrap(err, "initializing data privacy artifact compressor")
		}

		return compressor, nil
	})
}

// RegisterRequestService registers what the API server needs of platform's data privacy
// machinery: the request store, the artifact storage, and the service a subject's request is
// submitted to and their export read back through.
//
// The API server builds these by hand rather than from a service.Config's DataPrivacy block,
// because that block requires an Operations block beside it, and an Operations block registers
// the operations worker and its reapers too. The worker runs in the scheduler, and an API server
// that also claimed operations would fulfill privacy requests on the request path's replicas.
//
// They are platform's own registrations all the same, so the API server and the scheduler build
// the artifact storage the same way from the same block. The fulfiller is registered too, and not
// to run anything: building it is what registers the privacy kinds into the operations registry,
// and the service depends on it for exactly that ordering — a kind is resolved at submission, so
// an API server without the kinds would refuse every request.
//
// Prerequisites: everything platform's RegisterStore, RegisterArtifactStorage, RegisterFulfiller
// and RegisterService name, with the *platformdataprivacycfg.Config this registers from cfg.
func RegisterRequestService(i do.Injector, cfg *Config) {
	platform := cfg.Platform
	Pin(&platform)

	do.ProvideValue(i, &platform)
	RegisterKeyset(i, cfg.ArtifactEncryptionKey)
	RegisterCompressor(i)

	platformdataprivacycfg.RegisterStore(i)
	platformdataprivacycfg.RegisterArtifactStorage(i)
	platformdataprivacycfg.RegisterFulfiller(i)
	platformdataprivacycfg.RegisterService(i)
}
