package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/events"
	"github.com/primandproper/dinnerdonebetter/backend/internal/repositories/postgres/recording"

	platformsettings "github.com/primandproper/platform-go/v14/settings"
	settingscfg "github.com/primandproper/platform-go/v14/settings/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

const (
	o11yName = "settings_db_client"
)

// ProvideSettingsRepository provides platform's settings store, with this
// application's recording hung off its writes.
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
	tracer := tracing.NewNamedTracer(tracerProvider, o11yName)

	return newStore(ctx, logger, tracerProvider, metricsProvider, client, recording.NewRecorder(tracer, auditLogEntryRepo, eventEmitter))
}

// newStore builds the store with recorder behind its hooks. It is
// ProvideSettingsRepository with the recorder already assembled, which is the
// seam a test that wants the recording to fail reaches for.
func newStore(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	recorder *recording.Recorder,
) (platformsettings.Store, error) {
	store, err := settingscfg.NewStore(
		ctx,
		&settingscfg.Config{TablePrefix: branding.TablePrefix},
		client,
		&hooks{
			logger:   logging.NewNamedLogger(logger, o11yName),
			recorder: recorder,
		},
		settingscfg.WithLogger(logger),
		settingscfg.WithTracerProvider(tracerProvider),
		settingscfg.WithMetricsProvider(metricsProvider),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the settings store")
	}

	return store, nil
}
