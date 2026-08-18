package manager

import (
	"context"
	"testing"

	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/converters"
	"github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingsmock "github.com/primandproper/dinnerdonebetter/backend/internal/domain/settings/mock"

	loggingnoop "github.com/primandproper/platform-go/v11/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v11/observability/tracing/noop"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildSettingsManagerForTest(t *testing.T) *settingsManager {
	t.Helper()

	ctx := t.Context()

	m, err := NewSettingsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&settingsmock.RepositoryMock{},
	)
	require.NoError(t, err)

	return m.(*settingsManager)
}

// attachRepositoryToSettingsManager wires a configured repository mock into the manager under test.
func attachRepositoryToSettingsManager(manager *settingsManager, repo *settingsmock.RepositoryMock) {
	manager.repo = repo
}

func TestSettingsManager_CreateServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		expected := fakes.BuildFakeServiceSetting()
		input := converters.ConvertServiceSettingToServiceSettingDatabaseCreationInput(expected)

		repo := &settingsmock.RepositoryMock{
			CreateServiceSettingFunc: func(_ context.Context, _ *settings.ServiceSettingDatabaseCreationInput) (*settings.ServiceSetting, error) {
				return expected, nil
			},
		}
		attachRepositoryToSettingsManager(sm, repo)

		actual, err := sm.CreateServiceSetting(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateServiceSettingCalls(), 1)
	})
}

func TestSettingsManager_ArchiveServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		serviceSettingID := fakes.BuildFakeID()

		repo := &settingsmock.RepositoryMock{
			ArchiveServiceSettingFunc: func(_ context.Context, id string) error {
				assert.Equal(t, serviceSettingID, id)

				return nil
			},
		}
		attachRepositoryToSettingsManager(sm, repo)

		err := sm.ArchiveServiceSetting(ctx, serviceSettingID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveServiceSettingCalls(), 1)
	})
}

func TestSettingsManager_CreateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		expected := fakes.BuildFakeServiceSettingConfiguration()
		input := converters.ConvertServiceSettingConfigurationToServiceSettingConfigurationDatabaseCreationInput(expected)

		repo := &settingsmock.RepositoryMock{
			CreateServiceSettingConfigurationFunc: func(_ context.Context, _ *settings.ServiceSettingConfigurationDatabaseCreationInput) (*settings.ServiceSettingConfiguration, error) {
				return expected, nil
			},
		}
		attachRepositoryToSettingsManager(sm, repo)

		actual, err := sm.CreateServiceSettingConfiguration(ctx, input)
		require.NoError(t, err)
		assert.Equal(t, expected, actual)

		assert.Len(t, repo.CreateServiceSettingConfigurationCalls(), 1)
	})
}

func TestSettingsManager_UpdateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		updated := fakes.BuildFakeServiceSettingConfiguration()

		repo := &settingsmock.RepositoryMock{
			UpdateServiceSettingConfigurationFunc: func(_ context.Context, input *settings.ServiceSettingConfiguration) error {
				assert.Equal(t, updated, input)

				return nil
			},
		}
		attachRepositoryToSettingsManager(sm, repo)

		err := sm.UpdateServiceSettingConfiguration(ctx, updated)
		require.NoError(t, err)

		assert.Len(t, repo.UpdateServiceSettingConfigurationCalls(), 1)
	})
}

func TestSettingsManager_ArchiveServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		serviceSettingConfigurationID := fakes.BuildFakeID()

		repo := &settingsmock.RepositoryMock{
			ArchiveServiceSettingConfigurationFunc: func(_ context.Context, id string) error {
				assert.Equal(t, serviceSettingConfigurationID, id)

				return nil
			},
		}
		attachRepositoryToSettingsManager(sm, repo)

		err := sm.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfigurationID)
		require.NoError(t, err)

		assert.Len(t, repo.ArchiveServiceSettingConfigurationCalls(), 1)
	})
}
