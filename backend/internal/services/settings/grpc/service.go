package grpc

import (
	settingsmanager "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/manager"
	settingssvc "github.com/primandproper/dinnerdonebetter/backend/internal/grpc/generated/services/settings"

	"github.com/primandproper/platform-go/v10/observability/logging"
	"github.com/primandproper/platform-go/v10/observability/tracing"
)

const (
	o11yName = "configuration_service"
)

var _ settingssvc.SettingsServiceServer = (*serviceImpl)(nil)

type (
	serviceImpl struct {
		settingssvc.UnimplementedSettingsServiceServer
		tracer          tracing.Tracer
		logger          logging.Logger
		settingsManager settingsmanager.SettingsDataManager
	}
)

func NewService(
	logger logging.Logger,
	tracerProvider tracing.Provider,
	settingsManager settingsmanager.SettingsDataManager,
) settingssvc.SettingsServiceServer {
	return &serviceImpl{
		logger:          logging.NewNamedLogger(logger, o11yName),
		tracer:          tracing.NewNamedTracer(tracerProvider, o11yName),
		settingsManager: settingsManager,
	}
}
