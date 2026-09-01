package uploadedmedia

import (
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbuploadedmedia "github.com/primandproper/dinnerdonebetter/backend/internal/domain/uploadedmedia"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	"github.com/primandproper/platform-go/v13/uploads/registry"
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
	registry.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// ProvideUploadedMediaRepository provides a new upload registry.
func ProvideUploadedMediaRepository(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (registry.Store, error) {
	store, err := registry.NewSQLStore(
		client,
		registry.WithTablePrefix(ddbuploadedmedia.TablePrefix),
		registry.WithStoreLogger(logger),
		registry.WithStoreTracerProvider(tracerProvider),
		registry.WithStoreMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the upload registry store")
	}

	return &repository{
		Store:             store,
		client:            client,
		tracer:            tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:            logging.NewNamedLogger(logger, o11yName),
		auditLogEntryRepo: auditLogEntryRepo,
		events:            eventEmitter,
	}, nil
}
