package uploadedmedia

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	"github.com/primandproper/platform-go/v14/mediaregistry"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "uploaded_media_db_client"
)

// ProvideUploadedMediaRepository provides platform's upload registry, with this
// application's recording hung off its writes.
func ProvideUploadedMediaRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (mediaregistry.Store, error) {
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	store, err := mediaregistry.NewSQLStore(
		client,
		&hooks{
			logger:   logging.NewNamedLogger(logger, o11yName),
			recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
		},
		mediaregistry.WithTablePrefix(branding.TablePrefix),
		mediaregistry.WithStoreLogger(logger),
		mediaregistry.WithStoreTracerProvider(tracerProvider),
		mediaregistry.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the upload registry store")
	}

	return store, nil
}
