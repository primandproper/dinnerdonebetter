package uploadedmedia

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
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

// repository is platform's upload registry with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the five
// reads this package adds nothing to are the platform's own rather than five
// forwarding stubs that could drift from it.
type repository struct {
	mediaregistry.Store
	client   database.Client
	tracer   tracing.Tracer
	logger   logging.Logger
	recorder *recording.Recorder
}

// ProvideUploadedMediaRepository provides a new upload mediaregistry.
func ProvideUploadedMediaRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (mediaregistry.Store, error) {
	store, err := mediaregistry.NewSQLStore(
		client,
		mediaregistry.WithTablePrefix(ddbuploadedmedia.TablePrefix),
		mediaregistry.WithStoreLogger(logger),
		mediaregistry.WithStoreTracerProvider(tracerProvider),
		mediaregistry.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the upload registry store")
	}

	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return &repository{
		Store:    store,
		client:   client,
		tracer:   tracer,
		logger:   logging.NewNamedLogger(logger, o11yName),
		recorder: recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter),
	}, nil
}
