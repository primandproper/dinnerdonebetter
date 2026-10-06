/*
Package settings is platform-go's settings store, recording through platform's hooks.

The catalog and the values are platform's, and so is the recording: settings.RecordingHooks
writes an audit entry and emits an event for every write, on the write's transaction, through the
recording.Recorder this application registers. A value's entries are filed under the person whose
setting it is, because the Recorder files by subject; a definition's are filed where the write
ran, because a definition belongs to nobody.

What this package still decides is the table prefix, which has to match the prefix the migration
was rendered with — see renderSettingsDDL.
*/
package settings

import (
	"context"

	"github.com/primandproper/dinnerdonebetter/backend/internal/branding"

	platformrecording "github.com/primandproper/platform-go/v15/recording"
	platformsettings "github.com/primandproper/platform-go/v15/settings"
	settingscfg "github.com/primandproper/platform-go/v15/settings/config"
	"github.com/primandproper/primitives-go/v2/database"
	platformerrors "github.com/primandproper/primitives-go/v2/errors"
	"github.com/primandproper/primitives-go/v2/observability/logging"
	"github.com/primandproper/primitives-go/v2/observability/metrics"
	"github.com/primandproper/primitives-go/v2/observability/tracing"
)

// ProvideSettingsRepository provides platform's settings store, with platform's recording hooks
// on its writes.
//
// The store is assembled through platform's own settings/config rather than by naming
// settings.NewSQLStore's options here, so the knobs are stated once upstream.
func ProvideSettingsRepository(
	ctx context.Context,
	logger logging.Logger,
	tracerProvider tracing.Provider,
	metricsProvider metrics.Provider,
	client database.Client,
	recorder *platformrecording.Recorder,
) (platformsettings.Store, error) {
	hooks, err := platformsettings.NewRecordingHooks(recorder)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the settings recording hooks")
	}

	store, err := settingscfg.NewStore(
		ctx,
		&settingscfg.Config{TablePrefix: branding.TablePrefix},
		client,
		settingscfg.WithLogger(logger),
		settingscfg.WithTracerProvider(tracerProvider),
		settingscfg.WithMetricsProvider(metricsProvider),
		settingscfg.WithStoreOptions(platformsettings.WithHooks(hooks)),
	)
	if err != nil {
		return nil, platformerrors.Wrap(err, "building the settings store")
	}

	return store, nil
}
