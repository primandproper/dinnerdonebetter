package reportartifacts

import (
	"context"

	encryptioncfg "github.com/primandproper/platform-go/v9/cryptography/encryption/config"
	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/metrics"
	"github.com/primandproper/platform-go/v9/observability/tracing"
	"github.com/primandproper/platform-go/v9/uploads/objectstorage"
)

// ProvideStore builds a Store with an upload manager of its own.
//
// The upload manager is not taken from the injector on purpose. Every process that touches
// disclosure artifacts also registers exactly one uploads.UploadManager for its ordinary media,
// pointed at a different bucket — so a Store handed the ambient one would write reports into the
// avatar bucket, or read for them there and find nothing. The bucket that holds everything known
// about a person is worth naming explicitly.
func ProvideStore(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.TracerProvider,
	metricsProvider metrics.Provider,
	storageConfig *objectstorage.Config,
	encryptionConfig *encryptioncfg.Config,
	encryptionKey string,
) (Store, error) {
	if encryptionKey == "" {
		return nil, platformerrors.New("no disclosure artifact encryption key provided")
	}

	uploadManager, err := objectstorage.NewUploadManager(
		ctx,
		storageConfig,
		objectstorage.WithLogger(logger),
		objectstorage.WithTracerProvider(tracerProvider),
		objectstorage.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "initializing disclosure artifact upload manager")
	}

	encDec, err := encryptioncfg.NewEncryptorDecryptor(
		ctx,
		encryptionConfig,
		[]byte(encryptionKey),
		encryptioncfg.WithLogger(logger),
		encryptioncfg.WithTracerProvider(tracerProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "initializing disclosure artifact encryptor")
	}

	return NewStore(logger, tracerProvider, uploadManager, encDec), nil
}
