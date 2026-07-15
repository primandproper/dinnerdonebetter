package manager

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/audit"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/keys"

	platformerrors "github.com/primandproper/platform-go/v4/errors"
	"github.com/primandproper/platform-go/v4/filtering"
	"github.com/primandproper/platform-go/v4/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v4/messagequeue/config"
	"github.com/primandproper/platform-go/v4/observability"
	"github.com/primandproper/platform-go/v4/observability/logging"
	"github.com/primandproper/platform-go/v4/observability/tracing"
)

const (
	o11yName = "settings_data_manager"
)

// settingsRepo avoids wire cycles: manager takes this interface and produces settings.Repository.
type settingsRepo interface {
	settings.Repository
}

var (
	_ settings.Repository = (*settingsManager)(nil)
	_ SettingsDataManager = (*settingsManager)(nil)
)

type settingsManager struct {
	tracer               tracing.Tracer
	logger               logging.Logger
	repo                 settingsRepo
	dataChangesPublisher messagequeue.Publisher
}

// NewSettingsDataManager returns a new SettingsDataManager implementing settings.Repository.
func NewSettingsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo settingsRepo,
	cfg *msgconfig.QueuesConfig,
	publisherProvider messagequeue.PublisherProvider,
) (SettingsDataManager, error) {
	dataChangesPublisher, err := publisherProvider.NewPublisher(ctx, cfg.DataChangesTopicName)
	if err != nil {
		return nil, fmt.Errorf("failed to provide publisher for data changes topic: %w", err)
	}

	return &settingsManager{
		tracer:               tracing.NewNamedTracer(tracerProvider, o11yName),
		logger:               logging.NewNamedLogger(logger, o11yName),
		repo:                 repo,
		dataChangesPublisher: dataChangesPublisher,
	}, nil
}

// CreateServiceSetting creates a service setting.
func (m *settingsManager) CreateServiceSetting(ctx context.Context, input *settings.ServiceSettingDatabaseCreationInput) (*settings.ServiceSetting, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingIDKey, input.ID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, input.ID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating service setting creation input")
	}

	created, err := m.repo.CreateServiceSetting(ctx, input)
	if err != nil {
		return nil, err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, settings.ServiceSettingCreatedServiceEventType, map[string]any{
		settingskeys.ServiceSettingIDKey: created.ID,
	}))

	return created, nil
}

func (m *settingsManager) ServiceSettingExists(ctx context.Context, serviceSettingID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.ServiceSettingExists(ctx, serviceSettingID)
}

func (m *settingsManager) GetServiceSetting(ctx context.Context, serviceSettingID string) (*settings.ServiceSetting, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSetting(ctx, serviceSettingID)
}

func (m *settingsManager) GetServiceSettings(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettings(ctx, filter)
}

func (m *settingsManager) SearchForServiceSettings(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.SearchForServiceSettings(ctx, query, filter)
}

func (m *settingsManager) ArchiveServiceSetting(ctx context.Context, serviceSettingID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingIDKey, serviceSettingID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingIDKey, serviceSettingID)

	if err := m.repo.ArchiveServiceSetting(ctx, serviceSettingID); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, settings.ServiceSettingArchivedServiceEventType, map[string]any{
		settingskeys.ServiceSettingIDKey: serviceSettingID,
	}))

	return nil
}

// ServiceSettingConfigurationExists checks the existence of a service setting configuration.
func (m *settingsManager) ServiceSettingConfigurationExists(ctx context.Context, serviceSettingConfigurationID string) (bool, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.ServiceSettingConfigurationExists(ctx, serviceSettingConfigurationID)
}

func (m *settingsManager) GetServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) (*settings.ServiceSettingConfiguration, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettingConfiguration(ctx, serviceSettingConfigurationID)
}

func (m *settingsManager) GetServiceSettingConfigurationForUserByName(ctx context.Context, userID, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettingConfigurationForUserByName(ctx, userID, serviceSettingConfigurationName)
}

func (m *settingsManager) GetServiceSettingConfigurationForAccountByName(ctx context.Context, accountID, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettingConfigurationForAccountByName(ctx, accountID, serviceSettingConfigurationName)
}

func (m *settingsManager) GetServiceSettingConfigurationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettingConfigurationsForUser(ctx, userID, filter)
}

func (m *settingsManager) GetServiceSettingConfigurationsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()
	return m.repo.GetServiceSettingConfigurationsForAccount(ctx, accountID, filter)
}

func (m *settingsManager) CreateServiceSettingConfiguration(ctx context.Context, input *settings.ServiceSettingConfigurationDatabaseCreationInput) (*settings.ServiceSettingConfiguration, error) {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if input == nil {
		return nil, platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, input.ID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, input.ID)

	if err := input.ValidateWithContext(ctx); err != nil {
		return nil, observability.PrepareAndLogError(err, logger, span, "validating service setting configuration creation input")
	}

	created, err := m.repo.CreateServiceSettingConfiguration(ctx, input)
	if err != nil {
		return nil, err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, settings.ServiceSettingConfigurationCreatedServiceEventType, map[string]any{
		settingskeys.ServiceSettingConfigurationIDKey: created.ID,
	}))

	return created, nil
}

func (m *settingsManager) UpdateServiceSettingConfiguration(ctx context.Context, updated *settings.ServiceSettingConfiguration) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	if updated == nil {
		return platformerrors.ErrNilInputParameter
	}
	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, updated.ID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, updated.ID)

	if err := m.repo.UpdateServiceSettingConfiguration(ctx, updated); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, settings.ServiceSettingConfigurationUpdatedServiceEventType, map[string]any{
		settingskeys.ServiceSettingConfigurationIDKey: updated.ID,
	}))

	return nil
}

func (m *settingsManager) ArchiveServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)

	if err := m.repo.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfigurationID); err != nil {
		return err
	}

	m.dataChangesPublisher.PublishAsync(ctx, audit.BuildDataChangeMessageFromContext(ctx, logger, settings.ServiceSettingConfigurationArchivedServiceEventType, map[string]any{
		settingskeys.ServiceSettingConfigurationIDKey: serviceSettingConfigurationID,
	}))

	return nil
}
