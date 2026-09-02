package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	ddbsettings "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"

	"github.com/primandproper/platform-go/v13/database"
	platformerrors "github.com/primandproper/platform-go/v13/errors"
	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/metrics"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformsettings "github.com/primandproper/platform-go/v13/settings"
	settingscfg "github.com/primandproper/platform-go/v13/settings/config"
)

const (
	o11yName = "settings_db_client"
)

// repository is platform's settings store with this application's recording
// around it.
//
// The store is embedded rather than held in a named field so that the reads —
// the catalog, a subject's answers, and every resolution — are the platform's own
// rather than forwarding stubs that could drift from it.
type repository struct {
	platformsettings.Store
	client            database.Client
	tracer            tracing.Tracer
	logger            logging.Logger
	auditLogEntryRepo audit.Repository
	events            *events.Emitter
}

// ProvideSettingsRepository provides a new settings store.
//
// The store is assembled through platform's own settings/config rather than by
// naming settings.NewSQLStore's options here, so the knobs are stated once
// upstream. The table prefix is the one thing this application decides, and it
// has to match the prefix the migration was rendered with — see
// internal/repositories/postgres/migrations.
func ProvideSettingsRepository(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	auditLogEntryRepo audit.Repository,
	client database.Client,
	eventEmitter *events.Emitter,
) (platformsettings.Store, error) {
	store, err := settingscfg.NewStore(
		ctx,
		&settingscfg.Config{TablePrefix: ddbsettings.TablePrefix},
		client,
		settingscfg.WithLogger(logger),
		settingscfg.WithTracerProvider(tracerProvider),
		settingscfg.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the settings store")
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
