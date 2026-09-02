package grpc

import (
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"
	_ "github.com/primandproper/dinnerdonebetter/backend/internal/services/settings/errors"

	"github.com/primandproper/platform-go/v13/observability/logging"
	"github.com/primandproper/platform-go/v13/observability/tracing"
	platformsettings "github.com/primandproper/platform-go/v13/settings"
)

const (
	o11yName = "settings_service"
)

var _ settingssvc.SettingsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		settingssvc.UnimplementedSettingsServiceServer
		tracer   tracing.Tracer
		logger   logging.Logger
		settings platformsettings.Store
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	settingsStore platformsettings.Store,
) settingssvc.SettingsServiceServer {
	return &serviceImpl{
		logger:   logging.NewNamedLogger(logger, o11yName),
		tracer:   tracing.NewNamedTracer(tracerProvider, o11yName),
		settings: settingsStore,
	}
}
