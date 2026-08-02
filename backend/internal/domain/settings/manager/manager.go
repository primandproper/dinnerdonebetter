package manager

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	settingskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/keys"

	platformerrors "github.com/primandproper/platform-go/v9/errors"
	"github.com/primandproper/platform-go/v9/filtering"
	"github.com/primandproper/platform-go/v9/observability"
	"github.com/primandproper/platform-go/v9/observability/logging"
	"github.com/primandproper/platform-go/v9/observability/tracing"
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
	tracer tracing.Tracer
	logger logging.Logger
	repo   settingsRepo
}

// NewSettingsDataManager returns a new SettingsDataManager implementing settings.Repository.
//
// Data change events are enqueued into the outbox by the repository, inside the same
// transaction as the write they describe; see internal/repositories/postgres/events.
func NewSettingsDataManager(
	ctx context.Context,
	tracerProvider tracing.TracerProvider,
	logger logging.Logger,
	repo settingsRepo,
) (SettingsDataManager, error) {
	return &settingsManager{
		tracer: tracing.NewNamedTracer(tracerProvider, o11yName),
		logger: logging.NewNamedLogger(logger, o11yName),
		repo:   repo,
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
		return observability.PrepareAndLogError(err, logger, span, "archive service setting")
	}

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
		return observability.PrepareAndLogError(err, logger, span, "update service setting configuration")
	}

	return nil
}

func (m *settingsManager) ArchiveServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) error {
	ctx, span := m.tracer.StartSpan(ctx)
	defer span.End()

	logger := m.logger.WithSpan(span).WithValue(settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)
	tracing.AttachToSpan(span, settingskeys.ServiceSettingConfigurationIDKey, serviceSettingConfigurationID)

	if err := m.repo.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfigurationID); err != nil {
		return observability.PrepareAndLogError(err, logger, span, "archive service setting configuration")
	}

	return nil
}
