package config

import (
	"context"

	ddbaudit "github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbdataprivacy "github.com/primandproper/dinnerdonebetter/backend/internal/domain/dataprivacy"

	"github.com/primandproper/platform-go/v10/compression"
	"github.com/primandproper/platform-go/v10/cryptography/encryption"
	encryptioncfg "github.com/primandproper/platform-go/v10/cryptography/encryption/config"
	"github.com/primandproper/platform-go/v10/database"
	platformdataprivacy "github.com/primandproper/platform-go/v10/dataprivacy"
	platformdataprivacycfg "github.com/primandproper/platform-go/v10/dataprivacy/config"
	platformerrors "github.com/primandproper/platform-go/v10/errors"
	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/metrics"
	"github.com/primandproper/platform-go/v10/observability/tracing"
	"github.com/primandproper/platform-go/v10/operations"
	"github.com/primandproper/platform-go/v10/uploads"
	"github.com/primandproper/platform-go/v10/uploads/objectstorage"

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

		// EnsurePackaging is what keeps the reader's compressor and cipher the same as
		// the writer's. Getting them apart is not a startup failure — it is an artifact
		// that decodes to noise, discovered by a subject rather than by us.
		_, serviceOpts := platformdataprivacycfg.EnsurePackaging(
			do.MustInvoke[ArtifactCompressor](i).Compressor,
			do.MustInvoke[ArtifactEncryptorDecryptor](i).EncryptorDecryptor,
		)

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
			do.MustInvoke[platformdataprivacy.Store](i),
			// v10 fulfills a privacy request as an operation, so submitting one is starting
			// one. The kinds it starts have to be registered in this process's registry or
			// Start refuses them — see dataprivacybuild.RegisterOperationsRegistry.
			do.MustInvoke[operations.Service](i),
			platformdataprivacycfg.WithLogger(do.MustInvoke[logging.Logger](i)),
			platformdataprivacycfg.WithTracerProvider(do.MustInvoke[tracing.Provider](i)),
			platformdataprivacycfg.WithMetricsProvider(do.MustInvoke[metrics.Provider](i)),
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
