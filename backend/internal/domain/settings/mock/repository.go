package mock

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"

	"github.com/primandproper/platform-go/v5/filtering"

	"github.com/stretchr/testify/mock"
)

var _ settings.Repository = (*Repository)(nil)

type Repository struct {
	mock.Mock
}

// CreateServiceSetting is a mock function.
func (m *Repository) CreateServiceSetting(ctx context.Context, input *settings.ServiceSettingDatabaseCreationInput) (*settings.ServiceSetting, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*settings.ServiceSetting), args.Error(1)
}

// ServiceSettingExists is a mock function.
func (m *Repository) ServiceSettingExists(ctx context.Context, serviceSettingID string) (bool, error) {
	args := m.Called(ctx, serviceSettingID)
	return args.Bool(0), args.Error(1)
}

// GetServiceSetting is a mock function.
func (m *Repository) GetServiceSetting(ctx context.Context, serviceSettingID string) (*settings.ServiceSetting, error) {
	args := m.Called(ctx, serviceSettingID)
	return args.Get(0).(*settings.ServiceSetting), args.Error(1)
}

// GetServiceSettings is a mock function.
func (m *Repository) GetServiceSettings(ctx context.Context, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
	args := m.Called(ctx, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[settings.ServiceSetting]), args.Error(1)
}

// SearchForServiceSettings is a mock function.
func (m *Repository) SearchForServiceSettings(ctx context.Context, query string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSetting], error) {
	args := m.Called(ctx, query, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[settings.ServiceSetting]), args.Error(1)
}

// ArchiveServiceSetting is a mock function.
func (m *Repository) ArchiveServiceSetting(ctx context.Context, serviceSettingID string) error {
	args := m.Called(ctx, serviceSettingID)
	return args.Error(0)
}

// CreateServiceSettingConfiguration is a mock function.
func (m *Repository) CreateServiceSettingConfiguration(ctx context.Context, input *settings.ServiceSettingConfigurationDatabaseCreationInput) (*settings.ServiceSettingConfiguration, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*settings.ServiceSettingConfiguration), args.Error(1)
}

// ServiceSettingConfigurationExists is a mock function.
func (m *Repository) ServiceSettingConfigurationExists(ctx context.Context, serviceSettingConfigurationID string) (bool, error) {
	args := m.Called(ctx, serviceSettingConfigurationID)
	return args.Bool(0), args.Error(1)
}

// GetServiceSettingConfiguration is a mock function.
func (m *Repository) GetServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) (*settings.ServiceSettingConfiguration, error) {
	args := m.Called(ctx, serviceSettingConfigurationID)
	return args.Get(0).(*settings.ServiceSettingConfiguration), args.Error(1)
}

// GetServiceSettingConfigurationForAccountByName is a mock function.
func (m *Repository) GetServiceSettingConfigurationForAccountByName(ctx context.Context, accountID, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
	args := m.Called(ctx, accountID, serviceSettingConfigurationName)
	return args.Get(0).(*settings.ServiceSettingConfiguration), args.Error(1)
}

// GetServiceSettingConfigurationForUserByName is a mock function.
func (m *Repository) GetServiceSettingConfigurationForUserByName(ctx context.Context, userID, serviceSettingConfigurationName string) (*settings.ServiceSettingConfiguration, error) {
	args := m.Called(ctx, userID, serviceSettingConfigurationName)
	return args.Get(0).(*settings.ServiceSettingConfiguration), args.Error(1)
}

// GetServiceSettingConfigurationsForAccount is a mock function.
func (m *Repository) GetServiceSettingConfigurationsForAccount(ctx context.Context, accountID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
	args := m.Called(ctx, accountID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration]), args.Error(1)
}

// GetServiceSettingConfigurationsForUser is a mock function.
func (m *Repository) GetServiceSettingConfigurationsForUser(ctx context.Context, userID string, filter *filtering.QueryFilter) (*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration], error) {
	args := m.Called(ctx, userID, filter)
	return args.Get(0).(*filtering.QueryFilteredResult[settings.ServiceSettingConfiguration]), args.Error(1)
}

// UpdateServiceSettingConfiguration is a mock function.
func (m *Repository) UpdateServiceSettingConfiguration(ctx context.Context, input *settings.ServiceSettingConfiguration) error {
	args := m.Called(ctx, input)
	return args.Error(0)
}

// ArchiveServiceSettingConfiguration is a mock function.
func (m *Repository) ArchiveServiceSettingConfiguration(ctx context.Context, serviceSettingConfigurationID string) error {
	args := m.Called(ctx, serviceSettingConfigurationID)
	return args.Error(0)
}
