package config

import (
	"context"

	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	platformdataprivacy "github.com/primandproper/platform-go/v14/dataprivacy"
	platformdataprivacycfg "github.com/primandproper/platform-go/v14/dataprivacy/config"
	"github.com/primandproper/platform-go/v14/operations"
	"github.com/primandproper/primitives-go/v2/compression"
	"github.com/primandproper/primitives-go/v2/cryptography/encryption"
	encryptioncfg "github.com/primandproper/primitives-go/v2/cryptography/encryption/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
	"github.com/primandproper/primitives-go/v2/uploads"
	"github.com/primandproper/primitives-go/v2/uploads/objectstorage"

	"github.com/samber/do/v2"
)

type (
	// ArtifactUploadManager is the bucket export artifacts live in, wrapped so the
	// injector can tell it apart from the one holding user avatars. Every process
	// that touches artifacts also registers an upload manager for ordinary media, and
	// two registrations of one interface type is how the wrong bucket gets used.
	ArtifactUploadManager struct{ uploads.UploadManager }

	// ArtifactEncryptorDecryptor is the cipher artifacts are written and read with,
	// wrapped for the same reason.
	ArtifactEncryptorDecryptor struct{ encryption.EncryptorDecryptor }

	// ArtifactCompressor compresses an artifact before it is encrypted, wrapped for
	// the same reason.
	ArtifactCompressor struct{ compression.Compressor }
)

// RegisterArtifactStorage registers what every process that touches an export
// artifact needs: the bucket, the cipher, the compressor, and the request store.
//
// Every such process calls this, so the four are chosen in one place rather than
// once per process. Prerequisite: *Config and database.Client.
func RegisterArtifactStorage(i do.Injector) {
	do.Provide(i, func(i do.Injector) (ArtifactUploadManager, error) {
		cfg := do.MustInvoke[*Config](i)

		manager, err := objectstorage.NewUploadManager(
			do.MustInvoke[context.Context](i),
			&cfg.Uploads.Storage,
			objectstorage.WithLogger(do.MustInvoke[logging.Logger](i)),
			objectstorage.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			objectstorage.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
		if err != nil {
			return ArtifactUploadManager{}, platformerrors.Wrap(err, "initializing disclosure artifact upload manager")
		}

		return ArtifactUploadManager{UploadManager: manager}, nil
	})

	do.Provide(i, func(i do.Injector) (ArtifactEncryptorDecryptor, error) {
		cfg := do.MustInvoke[*Config](i)

		// Caught here rather than during validation, because a rendered config for a
		// real environment carries a blank secret and takes the value from the
		// environment. Startup is the last moment at which "no key" is a crash rather
		// than an artifact nobody can open.
		if cfg.ArtifactEncryptionKey == "" {
			return ArtifactEncryptorDecryptor{}, platformerrors.New("no disclosure artifact encryption key provided")
		}

		// One key, named by the configured current key ID. Rotating means adding the new key
		// to this set and pointing CurrentKeyID at it; artifacts already written keep opening
		// under the key their ciphertext names.
		encDec, err := encryptioncfg.NewKeyring(
			do.MustInvoke[context.Context](i),
			&cfg.Encryption,
			encryption.Keyset{
				encryption.KeyID(cfg.Encryption.CurrentKeyID): encryption.MasterKey(cfg.ArtifactEncryptionKey),
			},
			encryptioncfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			encryptioncfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
		)
		if err != nil {
			return ArtifactEncryptorDecryptor{}, platformerrors.Wrap(err, "initializing disclosure artifact encryptor")
		}

		return ArtifactEncryptorDecryptor{EncryptorDecryptor: encDec}, nil
	})

	do.Provide(i, func(do.Injector) (ArtifactCompressor, error) {
		compressor, err := compression.NewCompressor(CompressionAlgorithm)
		if err != nil {
			return ArtifactCompressor{}, platformerrors.Wrap(err, "initializing disclosure artifact compressor")
		}

		return ArtifactCompressor{Compressor: compressor}, nil
	})

	do.Provide(i, func(i do.Injector) (platformdataprivacy.Store, error) {
		client := do.MustInvoke[database.Client](i)

		return platformdataprivacycfg.NewStore(
			do.MustInvoke[context.Context](i),
			PlatformConfig(do.MustInvoke[*Config](i), client),
			client,
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
		)
	})
}

// RegisterRequestService registers the Service subjects submit requests through and
// read their artifacts back from.
//
// Prerequisite: RegisterArtifactStorage.
func RegisterRequestService(i do.Injector) {
	do.Provide(i, func(i do.Injector) (platformdataprivacy.Service, error) {
		client := do.MustInvoke[database.Client](i)

		// WithCompressor and WithEncryptor are what keep the reader's codecs the same
		// as the writer's: NewFulfiller writes with them and NewService reads with
		// them, from this one option set. Getting them apart is not a startup failure
		// — it is an artifact that decodes to noise, discovered by a subject rather
		// than by us. v14 replaced the EnsurePackaging helper that used to return the
		// paired option slices, which is a strictly better shape: there is no longer a
		// second slice a caller could forget to pass on.
		serviceOpts := []platformdataprivacy.ServiceOption{}

		// The upload manager is the read path, not a delivery path. Artifacts are
		// encrypted, so Download is refused outright by platform-go and Open — which
		// reads the object, decrypts, and decompresses — is the only way a subject
		// gets their export. See FetchUserDataReport.
		serviceOpts = append(serviceOpts,
			platformdataprivacy.WithServiceUploadManager(do.MustInvoke[ArtifactUploadManager](i).UploadManager),
		)

		return platformdataprivacycfg.NewService(
			do.MustInvoke[context.Context](i),
			PlatformConfig(do.MustInvoke[*Config](i), client),
			client,
			do.MustInvoke[platformdataprivacy.Store](i),
			// v10 fulfills a privacy request as an operation, so submitting one is starting
			// one. The kinds it starts have to be registered in this process's registry or
			// Start refuses them — see dataprivacybuild.RegisterOperationsRegistry.
			do.MustInvoke[operations.Service](i),
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
			platformdataprivacycfg.WithCompressor(do.MustInvoke[ArtifactCompressor](i).Compressor),
			platformdataprivacycfg.WithEncryptor(do.MustInvoke[ArtifactEncryptorDecryptor](i).EncryptorDecryptor),
			platformdataprivacycfg.WithServiceOptions(serviceOpts...),
		)
	})
}

// PlatformConfig returns a copy of the platform config with the three fields that
// are ours to decide rather than an operator's.
//
// Pinned, not validated. The prefixes have to equal the ones the migrations
// rendered the tables under, and the dialect has to be the client's. A deployment
// that set any of them differently would not be configuring anything, it would be
// pointing the Store at a table that does not exist — and a Store reading a table
// that isn't there finds no pending requests forever, which looks exactly like
// nobody having asked.
//
// Copied rather than mutated in place: the Config is shared with whatever else
// reads it, and several providers writing the same fields is a race that only
// happens to be benign.
func PlatformConfig(cfg *Config, client database.Client) *platformdataprivacycfg.Config {
	requests := cfg.Requests
	requests.TablePrefix = ddbdataprivacy.TablePrefix
	requests.Dialect = client.Dialect()
	requests.AuditErasure.TablePrefix = ddbaudit.TablePrefix

	return &requests
}
