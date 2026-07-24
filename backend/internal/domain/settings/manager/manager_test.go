package manager

import (
	"context"
	"testing"

	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/converters"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/fakes"
	settingskeys "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/keys"
	settingsmock "github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/domain/settings/mock"
	"github.com/verygoodsoftwarenotvirus/dinnerdonebetter/backend/internal/testutils"

	"github.com/primandproper/platform-go/v6/messagequeue"
	msgconfig "github.com/primandproper/platform-go/v6/messagequeue/config"
	mockpublishers "github.com/primandproper/platform-go/v6/messagequeue/mock"
	loggingnoop "github.com/primandproper/platform-go/v6/observability/logging/noop"
	tracingnoop "github.com/primandproper/platform-go/v6/observability/tracing/noop"
	"github.com/primandproper/platform-go/v6/reflection"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func buildSettingsManagerForTest(t *testing.T) *settingsManager {
	t.Helper()

	ctx := t.Context()
	queueCfg := &msgconfig.QueuesConfig{
		DataChangesTopicName: t.Name(),
	}

	mpp := &mockpublishers.PublisherProviderMock{
		NewPublisherFunc: func(_ context.Context, _ string) (messagequeue.Publisher, error) {
			return &mockpublishers.PublisherMock{
				PublishAsyncFunc: func(_ context.Context, _ any) {},
			}, nil
		},
	}

	m, err := NewSettingsDataManager(
		ctx,
		tracingnoop.NewTracerProvider(),
		loggingnoop.NewLogger(),
		&settingsmock.Repository{},
		queueCfg,
		mpp,
	)
	require.NoError(t, err)

	return m.(*settingsManager)
}

func setupExpectationsForSettingsManager(
	manager *settingsManager,
	repoSetupFunc func(repo *settingsmock.Repository),
	eventTypeMaps ...map[string][]string,
) []any {
	repo := &settingsmock.Repository{}
	if repoSetupFunc != nil {
		repoSetupFunc(repo)
	}
	manager.repo = repo

	mp := &mockpublishers.PublisherMock{
		PublishAsyncFunc: func(_ context.Context, _ any) {},
	}
	manager.dataChangesPublisher = mp

	return []any{repo}
}

func TestSettingsManager_CreateServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		expected := fakes.BuildFakeServiceSetting()
		input := converters.ConvertServiceSettingToServiceSettingDatabaseCreationInput(expected)

		expectations := setupExpectationsForSettingsManager(
			sm,
			func(repo *settingsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateServiceSetting), testutils.ContextMatcher, testutils.MatchType[*settings.ServiceSettingDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				settings.ServiceSettingCreatedServiceEventType: {settingskeys.ServiceSettingIDKey},
			},
		)

		actual, err := sm.CreateServiceSetting(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestSettingsManager_ArchiveServiceSetting(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		serviceSettingID := fakes.BuildFakeID()

		expectations := setupExpectationsForSettingsManager(
			sm,
			func(repo *settingsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveServiceSetting), testutils.ContextMatcher, serviceSettingID).Return(nil)
			},
			map[string][]string{
				settings.ServiceSettingArchivedServiceEventType: {settingskeys.ServiceSettingIDKey},
			},
		)

		err := sm.ArchiveServiceSetting(ctx, serviceSettingID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
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

		expectations := setupExpectationsForSettingsManager(
			sm,
			func(repo *settingsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.CreateServiceSettingConfiguration), testutils.ContextMatcher, testutils.MatchType[*settings.ServiceSettingConfigurationDatabaseCreationInput]()).Return(expected, nil)
			},
			map[string][]string{
				settings.ServiceSettingConfigurationCreatedServiceEventType: {settingskeys.ServiceSettingConfigurationIDKey},
			},
		)

		actual, err := sm.CreateServiceSettingConfiguration(ctx, input)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestSettingsManager_UpdateServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		updated := fakes.BuildFakeServiceSettingConfiguration()

		expectations := setupExpectationsForSettingsManager(
			sm,
			func(repo *settingsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.UpdateServiceSettingConfiguration), testutils.ContextMatcher, updated).Return(nil)
			},
			map[string][]string{
				settings.ServiceSettingConfigurationUpdatedServiceEventType: {settingskeys.ServiceSettingConfigurationIDKey},
			},
		)

		err := sm.UpdateServiceSettingConfiguration(ctx, updated)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}

func TestSettingsManager_ArchiveServiceSettingConfiguration(t *testing.T) {
	t.Parallel()

	t.Run("standard", func(t *testing.T) {
		t.Parallel()

		ctx := t.Context()
		sm := buildSettingsManagerForTest(t)

		serviceSettingConfigurationID := fakes.BuildFakeID()

		expectations := setupExpectationsForSettingsManager(
			sm,
			func(repo *settingsmock.Repository) {
				repo.On(reflection.GetMethodName(repo.ArchiveServiceSettingConfiguration), testutils.ContextMatcher, serviceSettingConfigurationID).Return(nil)
			},
			map[string][]string{
				settings.ServiceSettingConfigurationArchivedServiceEventType: {settingskeys.ServiceSettingConfigurationIDKey},
			},
		)

		err := sm.ArchiveServiceSettingConfiguration(ctx, serviceSettingConfigurationID)
		assert.NoError(t, err)

		mock.AssertExpectationsForObjects(t, expectations...)
	})
}
